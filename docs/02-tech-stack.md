# Tech Stack

## Technology Choices

| Layer | Choice | Rationale |
|-------|--------|-----------|
| **Backend** | Go | Excellent concurrency model (goroutines) for parallel thumbnail generation across thousands of files. Compiles to a single static binary -- ideal for home server deployment. Low memory footprint under load. CGo support enables direct LibRaw bindings for RAW image processing. |
| **Frontend** | SvelteKit | Lightweight framework with a small bundle size, fast runtime performance, and excellent developer experience. Less boilerplate than React/Vue for a relatively simple UI. Built-in routing and SSR/SPA flexibility. |
| **UI Components** | shadcn-svelte | Copy-paste Svelte 5 component library built on bits-ui headless primitives. Components live in `src/lib/components/ui/` (not node_modules), giving full control while providing consistent, accessible primitives (Button, Input, Card, Tabs, Sheet, Table, AlertDialog, Tooltip, etc.). Styled with Tailwind CSS variables for theming. |
| **Database** | SQLite (via `modernc.org/sqlite` or `mattn/go-sqlite3`) | Single-file database requiring no separate service or configuration. Perfect for 1-5 concurrent users. WAL mode provides good read concurrency. Symmetry with Lightroom Classic's `.lrcat` format (also SQLite) simplifies future catalog integration. |
| **Image Processing** | LibRaw (via CGo or shelling out to `dcraw_emu`) | Industry-standard RAW processing library with broad camera support. Sony A7RV ARW files embed a full-resolution JPEG; LibRaw's `unpack_thumb()` extracts it in milliseconds without decoding Bayer data. Go's standard `image` library + `disintegration/imaging` handles JPEG/PNG resizing. |
| **Cache Storage** | Local filesystem | Cached thumbnails and previews stored as `cache/{hash-prefix}/{hash}/{size}.jpg`. Simple, fast, no additional service. Content-addressable structure avoids conflicts and supports deduplication. |
| **Auth** | Built-in (bcrypt + JWT) | Simple username/password authentication with bcrypt password hashing and JWT tokens. No external service dependency. Appropriate for a small self-hosted app with 1-5 users. |
| **Deployment** | Docker Compose | Single `docker-compose.yml` defining the HomePhotos service with volume mounts: SMB share (read-only) and cache directory (read-write). Simple to run on TrueNAS or any Linux host. |
| **Reverse Proxy** | Caddy (optional) | Caddy can provide automatic HTTPS and clean URLs if desired. However, Go's built-in `net/http` server is production-quality, and Tailscale handles the network boundary -- so a reverse proxy is optional, not required. |

## Alternatives Considered

### Backend Language

| Alternative | Why not |
|-------------|---------|
| **Rust** | Excellent performance and memory safety, but slower development iteration. Borrow checker overhead is less justified for a personal project where development speed matters more than zero-cost abstractions. |
| **Python / FastAPI** | Rich ecosystem for image processing (Pillow, rawpy). However, the GIL limits true parallel execution for CPU-bound thumbnail generation. Would require multiprocessing or async workarounds that add complexity. |
| **Node.js** | Good for I/O-bound web servers but poor at CPU-bound image processing. Sharp (libvips bindings) helps, but RAW support is limited. No good LibRaw bindings. |

### Frontend Framework

| Alternative | Why not |
|-------------|---------|
| **React / Next.js** | Heavier runtime and larger bundle than needed for this use case. Next.js adds server-side complexity that is unnecessary when the Go backend already serves the API. |
| **Vue / Nuxt** | Viable alternative. SvelteKit was preferred for its smaller bundle size and less boilerplate. |
| **Plain HTML/JS** | Feasible for v1 but would slow down development of interactive features (lightbox, infinite scroll, filtering). |

### Database

| Alternative | Why not |
|-------------|---------|
| **PostgreSQL** | Excellent database, but overkill for 1-5 users. Requires a separate service, configuration, and backup strategy. Adds operational complexity to a home server deployment. |
| **MySQL / MariaDB** | Same concerns as PostgreSQL. No advantage over SQLite at this scale. |
| **Embedded key-value store (BoltDB, BadgerDB)** | Good for simple key-value access but lacks the relational query capabilities needed for filtering by tags, dates, and metadata combinations. |

## LibRaw Specifics

### Sony A7RV ARW Support

The Sony A7RV produces ARW files (Sony's variant of the TIFF-based RAW format). Each ARW file is typically 40-60 MB and contains:

- **Bayer sensor data** -- the actual RAW pixel values from the image sensor.
- **An embedded full-resolution JPEG** -- a camera-processed JPEG at the full sensor resolution (approximately 9504 x 6336 pixels). This is what the camera shows on its LCD for review.

### Extraction Path: `unpack_thumb` vs Full Decode

```
┌─────────────────────────────────────────────────────────┐
│  ARW File (~50 MB)                                      │
│                                                         │
│  ┌─────────────────────┐  ┌──────────────────────────┐  │
│  │  Bayer Sensor Data  │  │  Embedded JPEG (~8-12MB) │  │
│  │  (RAW pixel values) │  │  (full resolution)       │  │
│  └─────────────────────┘  └──────────────────────────┘  │
└─────────────────────────────────────────────────────────┘

Option A: unpack_thumb()          Option B: Full RAW decode
─────────────────────────         ─────────────────────────
- Extracts embedded JPEG          - Decodes Bayer data
- ~5-20 ms per file               - ~2-10 seconds per file
- Minimal memory usage            - ~500 MB+ memory per file
- Output: camera-processed JPEG   - Output: linear RGB image
- Sufficient for viewing          - Needed for editing
                                    (not our use case)
```

**HomePhotos uses Option A.** Since the goal is viewing -- not editing -- the embedded JPEG is the right source. It is already color-processed, white-balanced, and tone-mapped by the camera's image processor.

### Known Issues: Black Borders on Sony Thumbnails

Some Sony camera models (and some firmware versions) produce embedded thumbnails with thin black borders or slight dimensional mismatches. This issue has been documented in other projects, notably Immich. Mitigation strategies:

- Detect and crop black borders after extraction (a few pixels on each edge).
- Use the largest available embedded image (some ARWs contain multiple sizes).
- Validate extracted dimensions against expected sensor output dimensions.

This should be tested with actual A7RV files during development and handled as needed.

## Key Go Libraries

| Library | Purpose |
|---------|---------|
| `disintegration/imaging` | Image resizing, cropping, and JPEG encoding. Used to generate thumbnails (300px) and previews (1600px) from extracted JPEGs. Supports Lanczos resampling for high-quality downscaling. |
| `rwcarlsen/goexif` | EXIF metadata parsing. Extracts camera model, lens, focal length, aperture, shutter speed, ISO, date/time, GPS coordinates, and other fields from image files. |
| `modernc.org/sqlite` | Pure Go SQLite driver (no CGo required). Simplifies cross-compilation and deployment. Slightly slower than the CGo alternative but eliminates the C toolchain dependency for the database layer. |
| `mattn/go-sqlite3` | CGo-based SQLite driver. Faster than the pure Go alternative. Since LibRaw already requires CGo, the C toolchain is available anyway -- making this a viable choice if performance matters. |
| `golang-jwt/jwt/v5` | JWT signing and verification. Used in API middleware to issue tokens on login and verify them on every request. |
| `golang.org/x/crypto/bcrypt` | Password hashing. Used to hash passwords on registration and verify them on login. |

### SQLite Driver Choice

Since LibRaw integration already requires CGo (either via direct bindings or by shelling out to `dcraw_emu`), the CGo toolchain will be present in the build environment. This makes `mattn/go-sqlite3` a practical choice -- there is no additional build complexity, and it offers better performance than the pure Go alternative. However, if LibRaw is accessed by shelling out to an external binary (avoiding CGo entirely), then `modernc.org/sqlite` keeps the Go build pure and simplifies cross-compilation.

The choice can be deferred and swapped later since both drivers implement the standard `database/sql` interface.
