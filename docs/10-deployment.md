# Deployment

## Overview

HomePhotos is deployed as a Docker Compose stack on the home network. The application reads photos from a TrueNAS SMB share mounted on the Docker host, stores metadata in a SQLite database, and uses Clerk for authentication. Remote access is provided through Tailscale.

---

## Prerequisites

Before deploying HomePhotos, ensure the following are in place:

- **TrueNAS with SMB share** configured and accessible from the Docker host. The share should contain your photo library.
- **Docker and Docker Compose** installed on the host machine (Docker Engine 20.10+ and Compose V2 recommended).
- **Tailscale** installed and configured on the host for secure remote access without exposing ports to the public internet.
- **Clerk account** for authentication. The free tier is sufficient for family use.

---

## Docker Compose

Create a `docker-compose.yml` file in your project directory:

```yaml
version: "3.8"

services:
  homephotos:
    build: .
    # Or use a published image:
    # image: ghcr.io/your-org/homephotos:latest
    ports:
      - "8080:8080"
    volumes:
      - /mnt/photos:/source:ro        # SMB mount — READ-ONLY
      - homephotos-cache:/cache        # Generated thumbnails and previews
      - homephotos-db:/data            # SQLite database
    environment:
      - HOMEPHOTOS_SOURCE_PATH=/source
      - HOMEPHOTOS_CACHE_PATH=/cache
      - HOMEPHOTOS_DB_PATH=/data/homephotos.db
      - HOMEPHOTOS_LISTEN_ADDR=:8080
      - CLERK_SECRET_KEY=${CLERK_SECRET_KEY}
      - CLERK_PUBLISHABLE_KEY=${CLERK_PUBLISHABLE_KEY}
      # Optional: Lightroom integration
      # - HOMEPHOTOS_LRCAT_PATH=/source/.lightroom/catalog.lrcat
      # - HOMEPHOTOS_LR_PATH_MAP={"D:\\Photos": "/source"}
    restart: unless-stopped

volumes:
  homephotos-cache:
  homephotos-db:
```

### Key points

- The source volume is mounted **read-only** (`:ro`). HomePhotos never modifies your original photos.
- The cache and database volumes are named Docker volumes, persisting across container restarts and upgrades.
- Clerk keys are loaded from a `.env` file via Docker Compose variable substitution (see [Clerk Setup](#clerk-setup) below).
- Lightroom environment variables are commented out by default. Uncomment them if you have a Lightroom Classic catalog accessible on the SMB share.

---

## SMB Mount

The Docker host must mount the TrueNAS SMB share so the container can access the photo files.

### Linux (Docker host)

1. **Install CIFS utilities** (if not already installed):

   ```bash
   sudo apt install cifs-utils    # Debian/Ubuntu
   sudo dnf install cifs-utils    # Fedora/RHEL
   ```

2. **Create the mount point**:

   ```bash
   sudo mkdir -p /mnt/photos
   ```

3. **Create a credentials file** to avoid storing passwords in fstab:

   ```bash
   sudo nano /etc/samba/credentials
   ```

   Contents:

   ```
   username=your_smb_username
   password=your_smb_password
   ```

   Secure the file:

   ```bash
   sudo chmod 600 /etc/samba/credentials
   ```

4. **Add an fstab entry** for automatic mounting at boot:

   ```
   //truenas.local/photos /mnt/photos cifs credentials=/etc/samba/credentials,ro,uid=1000,gid=1000 0 0
   ```

   Replace `truenas.local` with your TrueNAS hostname or IP address. The `uid` and `gid` values should match the user running Docker.

5. **Mount the share**:

   ```bash
   sudo mount /mnt/photos
   ```

6. **Verify the mount**:

   ```bash
   # Should list your photo files:
   ls /mnt/photos

   # Should fail with "Read-only file system" — confirming the mount is read-only:
   touch /mnt/photos/test
   ```

---

## Clerk Setup

Clerk provides authentication so family members can sign in without you managing passwords.

1. **Create a Clerk application** at [clerk.com](https://clerk.com). Choose a sign-in method that works for your family (email, Google, etc.).

2. **Note the API keys** from the Clerk dashboard:
   - **Publishable Key** (starts with `pk_live_` or `pk_test_`)
   - **Secret Key** (starts with `sk_live_` or `sk_test_`)

3. **Configure allowed origins** in the Clerk dashboard under your application settings. Add your Tailscale hostname:

   ```
   http://homephotos.tailnet-name.ts.net:8080
   ```

   If you also access HomePhotos on the local network, add that origin too (e.g., `http://192.168.1.50:8080`).

4. **Create a `.env` file** in the same directory as `docker-compose.yml`:

   ```bash
   CLERK_SECRET_KEY=sk_live_xxxxxxxxxxxxxxxxxxxx
   CLERK_PUBLISHABLE_KEY=pk_live_xxxxxxxxxxxxxxxxxxxx
   ```

   **Never commit this file to version control.** Ensure `.env` is listed in your `.gitignore`.

---

## First-Run Walkthrough

Follow these steps to get HomePhotos running for the first time:

1. **Mount the SMB share** on the Docker host (see [SMB Mount](#smb-mount) above). Verify with `ls /mnt/photos`.

2. **Create the `.env` file** with your Clerk keys (see [Clerk Setup](#clerk-setup) above).

3. **Start the stack**:

   ```bash
   docker compose up -d
   ```

   On first run, the container will create the SQLite database and cache directory structure.

4. **Navigate to the application** in your browser:

   ```
   http://hostname:8080
   ```

   Replace `hostname` with your Docker host's IP, local hostname, or Tailscale hostname.

5. **Sign in via Clerk**. The first user to sign in is automatically granted the `admin` role.

6. **Start the initial scan**. Go to **Admin > Scanner > Start Scan** in the UI, or trigger it via the API:

   ```bash
   curl -X POST http://hostname:8080/api/v1/scanner/run \
     -H "Authorization: Bearer <your_jwt>"
   ```

7. **Browse photos** as thumbnails are generated. The scanner discovers files first, then background workers generate thumbnails and extract EXIF data. You can browse immediately while generation continues.

---

## Configuration Reference

All configuration is provided via environment variables.

| Variable | Description | Default | Required |
|---|---|---|---|
| `HOMEPHOTOS_SOURCE_PATH` | Path to the mounted photo source directory inside the container. | `/source` | Yes |
| `HOMEPHOTOS_CACHE_PATH` | Path to the cache directory for generated thumbnails and previews. | `/cache` | Yes |
| `HOMEPHOTOS_DB_PATH` | Path to the SQLite database file. | `/data/homephotos.db` | Yes |
| `HOMEPHOTOS_LISTEN_ADDR` | Address and port the server listens on. | `:8080` | No |
| `CLERK_SECRET_KEY` | Clerk backend secret key for JWT verification. | _(none)_ | Yes |
| `CLERK_PUBLISHABLE_KEY` | Clerk frontend publishable key, served to the SvelteKit frontend. | _(none)_ | Yes |
| `HOMEPHOTOS_LRCAT_PATH` | Path to the Lightroom Classic `.lrcat` catalog file. Enables the Lightroom integration. | _(none)_ | No |
| `HOMEPHOTOS_LR_PATH_MAP` | JSON object mapping Lightroom catalog paths to container paths (e.g., `{"D:\\Photos": "/source"}`). Required if `HOMEPHOTOS_LRCAT_PATH` is set. | _(none)_ | No |
| `HOMEPHOTOS_LOG_LEVEL` | Logging verbosity. Values: `debug`, `info`, `warn`, `error`. | `info` | No |
| `HOMEPHOTOS_CACHE_WORKERS` | Number of concurrent cache generation workers. | `4` | No |
| `HOMEPHOTOS_SCAN_EXTENSIONS` | Comma-separated list of file extensions to scan. | `arw,cr2,cr3,nef,dng,raf,orf,rw2,jpg,jpeg,tiff,tif,heic` | No |

---

## Backup

### SQLite database

The SQLite database (`/data/homephotos.db` inside the container) is the critical piece to back up. It contains all tags, tag assignments, user data, scan metadata, and Lightroom sync state.

To back up, copy the database file from the named volume:

```bash
docker compose exec homephotos cp /data/homephotos.db /data/homephotos.db.bak
# Or from the host, find the volume mount point:
docker compose cp homephotos:/data/homephotos.db ./homephotos-backup.db
```

Consider scheduling a daily backup via cron.

### Cache directory

The cache directory contains generated thumbnails and previews. It can be **fully regenerated** from the source files by running a scan, so backing it up is optional. However, regenerating a large cache (20,000+ photos) can take hours, so preserving it avoids that wait time after a restore.

### Source photos

Source photos live on TrueNAS and should be backed up via TrueNAS's own backup mechanisms (snapshots, replication, cloud sync tasks). HomePhotos never modifies source files.

---

## Updating

To update HomePhotos to the latest version:

```bash
docker compose pull
docker compose up -d
```

- **Database migrations** run automatically on startup. The application checks the current schema version and applies any pending migrations before accepting requests.
- **Cache is preserved** across updates. No need to rescan after upgrading.
- **Breaking changes** (if any) will be documented in the release notes. Check the changelog before upgrading major versions.

---

## Troubleshooting

### SMB mount not accessible

**Symptoms**: Container logs show errors reading from `/source`, scanner finds zero files, or the health endpoint reports `"smb_mount": "disconnected"`.

**Steps to resolve**:

1. Verify the mount is active on the host: `mount | grep /mnt/photos`
2. Test access from the host: `ls /mnt/photos`
3. If the mount dropped, remount: `sudo mount /mnt/photos`
4. Check credentials: ensure `/etc/samba/credentials` has the correct username and password.
5. Check network connectivity to TrueNAS: `ping truenas.local`
6. Verify the mount is read-only: `touch /mnt/photos/test` should fail. If it succeeds, update your fstab entry to include the `ro` option.

### Thumbnails not generating

**Symptoms**: Photos appear in the library with placeholder thumbnails. The UI shows a loading state that never resolves.

**Steps to resolve**:

1. Check scanner status via the API: `curl http://hostname:8080/api/v1/scanner/status`
2. Check container logs for LibRaw errors: `docker compose logs homephotos | grep -i "libraw\|error\|panic"`
3. Verify the cache directory is writable: the health endpoint should report `"cache_dir": "writable"`.
4. If a specific file fails, it may be a corrupted or unsupported RAW format. Check the scanner error count and logs for the file path.
5. Try increasing the number of cache workers via `HOMEPHOTOS_CACHE_WORKERS` if generation is slow but not erroring.

### Clerk authentication issues

**Symptoms**: Users cannot sign in, the UI shows authentication errors, or API calls return `401 Unauthorized` even with a valid session.

**Steps to resolve**:

1. Verify your keys in `.env` match the keys in the Clerk dashboard. Ensure there are no trailing spaces or newlines.
2. Check **allowed origins** in the Clerk dashboard. Your Tailscale hostname (including port) must be listed exactly as it appears in the browser address bar.
3. Check for CORS errors in the browser developer console. If you see CORS-related messages, the origin is not allowed in Clerk.
4. Ensure the container can reach the internet to verify JWTs against Clerk's JWKS endpoint. Test from inside the container: `docker compose exec homephotos wget -q -O- https://api.clerk.com`
5. If using `pk_test_` / `sk_test_` keys, make sure you are not mixing test and production keys.
6. Check container logs for JWT verification errors: `docker compose logs homephotos | grep -i "clerk\|jwt\|auth"`
