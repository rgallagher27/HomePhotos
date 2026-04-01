# Non-Functional Requirements

## Read-Only Filesystem Access (CRITICAL)

This is the **single most important constraint in the entire system**.

The source photo library on the NAS is a collection of irreplaceable originals. Any accidental modification, corruption, or deletion would be catastrophic and potentially unrecoverable. HomePhotos enforces read-only access at multiple layers to make accidental writes structurally impossible.

### Enforcement Layers

1. **OS-level mount flags**: The SMB share MUST be mounted read-only at the operating system level.
   - Native mount: `mount -o ro //nas/photos /mnt/source`
   - Docker Compose: `read_only: true` on the volume definition

   ```yaml
   volumes:
     - type: bind
       source: /mnt/nas/photos
       target: /source
       read_only: true
   ```

2. **Application-level path validation**: The application code MUST validate all file operations and reject any write attempt targeting the source root. This acts as a defense-in-depth measure even if the mount flags are misconfigured.

3. **Code review policy**: Any code change that introduces filesystem write operations must be reviewed with explicit attention to the target path. Write operations to the source path are never acceptable.

### Cache Separation

The cache directory (for thumbnails, previews, and database) is a **completely separate, writable path** -- never located on the source mount.

| Path       | Mount     | Access     | Purpose                              |
| ---------- | --------- | ---------- | ------------------------------------ |
| `/source`  | SMB share | Read-only  | Original photos (irreplaceable)      |
| `/cache`   | Local vol | Read-write | Thumbnails, previews, SQLite DB      |

---

## Performance

### Target Response Times

| Operation                          | Target             |
| ---------------------------------- | ------------------ |
| Photo grid: first 50 thumbnails    | Under 2 seconds    |
| Photo detail: preview image        | Under 1 second     |
| Search and filter results          | Under 500ms        |
| Cold start to serving cached photos| Under 5 seconds    |

### Background Processing

The background filesystem scanner and thumbnail generator must not degrade UI responsiveness:

- Use a **bounded pool of worker goroutines** for thumbnail generation (e.g., `GOMAXPROCS` or a configurable concurrency limit).
- API request handling takes priority over background work. Workers should yield (via channel backpressure or semaphore) when the server is under API load.
- Thumbnail generation is I/O-bound (reading source files, writing cache files). Limit concurrent disk reads to avoid saturating the SMB connection.

---

## Security

### Network Boundary

HomePhotos is designed for **no public internet exposure**. Tailscale provides the network boundary:

- The server runs on a Tailscale node accessible only to devices on the user's tailnet.
- No port forwarding, no public DNS, no reverse proxy to the internet.
- Tailscale handles encrypted transport (WireGuard) between all nodes.

### Authentication

All API endpoints require authentication via **Clerk JWT verification**. There are no anonymous or public endpoints (except the health check).

Clerk handles the security-critical concerns that should not be implemented from scratch:

- Password hashing and storage
- Brute force protection and rate limiting
- Session management and token rotation
- Multi-factor authentication (if configured)

### Path Traversal Prevention

All file path parameters received from API requests must be validated against the configured source root and cache root directories:

- Reject any path containing `..` components.
- Resolve the final absolute path and confirm it falls within an allowed directory.
- Return a `400 Bad Request` for any path that fails validation. Never return filesystem error details to the client.

```go
// Example validation logic
func isPathSafe(requested string, allowedRoot string) bool {
    abs := filepath.Clean(filepath.Join(allowedRoot, requested))
    return strings.HasPrefix(abs, filepath.Clean(allowedRoot))
}
```

### HTTP Security Headers

- **CORS**: Restrict allowed origins to the configured frontend origin. Do not use wildcard (`*`) origins.
- **Content-Security-Policy**: Set on all frontend responses to prevent XSS, inline script injection, and unauthorized resource loading.
- **X-Content-Type-Options**: `nosniff`
- **X-Frame-Options**: `DENY`

---

## Reliability

### Unavailable SMB Mount

The NAS or SMB mount may become temporarily unavailable (network issues, NAS maintenance, etc.). HomePhotos must handle this gracefully:

- **Detect mount state** on startup and periodically during operation (e.g., check if the mount point is accessible and contains expected content).
- **Show a user-friendly error** in the UI ("Photo source is currently unavailable") rather than cryptic filesystem errors.
- **Continue serving cached content**: thumbnails, previews, and metadata already in the cache and database remain available.
- **Retry connection periodically** and resume normal operation when the mount comes back.

### Corrupt or Unreadable Files

The source library may contain corrupt files, unsupported RAW formats, or files that crash image processing libraries:

- **Log and skip** any file that cannot be read or processed. Include the file path and error in the log.
- **Never crash** the scanner or worker goroutine. Use `recover()` where appropriate to catch panics from image decoding libraries.
- **Track errors** in the database so the admin can review which files failed and why.

### Database Migrations

- Use **versioned migration files** (e.g., `001_initial_schema.sql`, `002_add_tags.sql`).
- Migrations run **automatically on startup**, applying any pending migrations in order.
- Migrations are **forward-only** -- no down migrations. If a migration needs to be reversed, write a new forward migration that undoes the change.
- Each migration runs in a transaction. If a migration fails, the transaction is rolled back and the application exits with a clear error message.

### Graceful Shutdown

On receiving a shutdown signal (SIGTERM, SIGINT):

- Stop accepting new API requests.
- **Finish in-progress thumbnail generation** to avoid leaving partial files in the cache.
- Close the database connection cleanly.
- Exit within a bounded timeout (e.g., 30 seconds). If in-progress work does not complete within the timeout, force exit.

---

## Scalability

HomePhotos is a **home application**, not a SaaS product. The design optimizes for simplicity and reliability over horizontal scalability.

| Dimension           | Design Target         |
| ------------------- | --------------------- |
| Photo count         | 1 -- 50,000           |
| Concurrent users    | 1 -- 5                |
| Deployment model    | Single server         |
| Clustering          | Not supported         |
| Load balancing      | Not needed            |
| Database            | Single SQLite file    |

If the photo count or user count grows beyond these targets, the appropriate response is to optimize the existing single-server architecture (better caching, query optimization, pagination), not to introduce distributed systems complexity.

---

## Observability

### Structured Logging

- All log output uses **JSON format** for easy parsing by log aggregation tools.
- Log level is configurable via the `HOMEPHOTOS_LOG_LEVEL` environment variable (e.g., `debug`, `info`, `warn`, `error`).
- Each log entry includes: timestamp, level, message, and relevant context fields (request ID, file path, duration, etc.).

### Health Check Endpoint

**`GET /api/v1/health`**

Returns the status of all critical subsystems:

```json
{
  "status": "healthy",
  "components": {
    "smb_mount": { "status": "ok", "path": "/source" },
    "database": { "status": "ok", "path": "/cache/homephotos.db" },
    "cache_directory": { "status": "ok", "path": "/cache", "free_space_mb": 10240 },
    "scan": { "status": "idle", "last_completed": "2025-01-15T03:00:00Z" }
  }
}
```

If any component is degraded, the top-level status reflects that (e.g., `"degraded"` or `"unhealthy"`).

### Metrics (Optional)

The following metrics are useful for understanding system behavior but are not required for v1:

| Metric                | Description                                          |
| --------------------- | ---------------------------------------------------- |
| Scan progress         | Total files discovered, processed, errors encountered|
| Cache hit rate        | Percentage of thumbnail/preview requests served from cache |
| Thumbnail queue depth | Number of thumbnails waiting to be generated         |
| Active connections    | Current number of active API connections              |

### Log Rotation

HomePhotos itself does **not** manage log rotation. It writes to stdout/stderr and defers log management to the container runtime or system logging infrastructure:

- **Docker**: Use Docker's built-in log drivers (`json-file` with `max-size` and `max-file`, or `local` driver).
- **Systemd**: Use `journald` log management.
