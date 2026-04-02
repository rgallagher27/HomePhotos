# Implementation Status

## What Has Been Built

### Phase 1: Auth & Users — Complete

User authentication and role-based access control are fully implemented.

### Phase 2: Photo Scanning & Indexing — Complete

Photo scanning, EXIF extraction, and browsing APIs are fully implemented.

### Phase 3: Caching & Thumbnails — Complete

Thumbnail/preview generation (background + on-demand) and image serving endpoint are fully implemented.

### Backend (Go)

| Component | Path | Status |
|-----------|------|--------|
| Entry point | `backend/cmd/server/main.go` | Done — signal handling, graceful shutdown, envconfig |
| Config | `backend/config/config.go` | Done — `HOMEPHOTOS_` prefix, all env vars defined |
| SQLite connection | `backend/database/sqlite/sqlite.go` | Done — WAL mode, foreign keys, busy timeout, single writer |
| Migrations | `backend/database/sqlite/migrations/` | Done — `000001_init` (schema_info), `000002_add_users_table` (users), `000003_add_photos_table` (photos with EXIF columns) |
| sqlc | `backend/database/sqlite/sqlc.yaml` | Done — configured for SQLite, generates to `database/sqlite/` |
| sqlc queries | `backend/database/sqlite/queries/` | Done — `health.sql` (Ping), `users.sql` (7 queries), `photos.sql` (CreatePhoto, UpdatePhoto, GetPhotoByID, GetPhotoByFilePath, ListAllFingerprints, ListPendingPhotos, UpdatePhotoCacheStatus) |
| User repository | `backend/database/sqlite/user_repository.go` | Done — implements `user.Repository`, maps sqlc models to domain entities |
| Photo repository | `backend/database/sqlite/photo_repository.go` | Done — implements `photo.Repository`, cursor pagination with dynamic SQL, orphan cleanup, list pending, update cache status |
| User domain | `backend/domain/user/` | Done — `User` entity, `Repository` interface, `Role` type, sentinel errors |
| Photo domain | `backend/domain/photo/` | Done — `Photo` entity with EXIF fields, `Repository` interface with `ListParams`/`ListResult`, sentinel errors |
| Auth service | `backend/services/auth/service.go` | Done — Register (bcrypt, first-user-admin), Login (verify + JWT) |
| Token service | `backend/services/auth/token.go` | Done — HMAC-SHA256 JWT, configurable expiry |
| Auth errors | `backend/services/auth/errors.go` | Done — `ErrInvalidCredentials`, `ErrRegistrationClosed` |
| Scanner service | `backend/services/scanner/service.go` | Done — walks source directory, diffs fingerprints, extracts EXIF, upserts photos, deletes orphans |
| Scanner scheduler | `backend/services/scanner/scheduler.go` | Done — background goroutine with configurable interval and on-startup scan |
| EXIF extraction | `backend/services/scanner/exif.go` | Done — extracts camera, lens, GPS, exposure data via `rwcarlsen/goexif` |
| Image processing | `backend/services/imaging/imaging.go` | Done — decode (JPEG/PNG/TIFF/RAW), resize via Lanczos, JPEG encode |
| EXIF orientation | `backend/services/imaging/orientation.go` | Done — applies EXIF orientation transforms (values 1-8) |
| RAW extraction | `backend/services/imaging/raw.go` | Done — pure Go TIFF/IFD parser, extracts embedded JPEG from ARW/DNG |
| Cache service | `backend/services/cache/service.go` | Done — content-addressable cache, generate thumb+preview, on-demand generation |
| Cache worker pool | `backend/services/cache/worker.go` | Done — background batch processing of pending photos |
| OpenAPI server | `backend/ports/rest/server.gen.go` | Generated — `ServerInterface` with health, auth, user, photo, scanner, and image endpoints |
| Server struct | `backend/ports/rest/server.go` | Done — holds `*sql.DB`, auth service, token service, user repository, photo repository, scanner service, cache service |
| Server init | `backend/ports/rest/rest.go` | Done — builds services, wires JWT auth, sets up middleware stack, constructs scanner and cache |
| Health handler | `backend/ports/rest/health_handlers.go` | Done — pings DB, returns ok/degraded |
| Auth handlers | `backend/ports/rest/auth_handlers.go` | Done — register, login, me endpoints |
| User handlers | `backend/ports/rest/user_handlers.go` | Done — list users, update role (admin only) |
| Photo handlers | `backend/ports/rest/photo_handlers.go` | Done — list photos (cursor pagination, filters), get photo detail with full EXIF |
| Scanner handlers | `backend/ports/rest/scanner_handlers.go` | Done — trigger scan (admin), get scan status (admin) |
| Image handlers | `backend/ports/rest/image_handlers.go` | Done — serve photo images (thumb/preview/full), on-demand generation, cache headers |
| Auth middleware | `backend/ports/rest/auth_middleware.go` | Done — JWT validation (OAPI), context injection (HTTP middleware), `RequireAdmin` |
| Middleware | `backend/ports/rest/middleware.go` | Done — CORS, JSON content-type |
| Error helpers | `backend/ports/rest/error.go` | Done — nested `{"error": {"code": "...", "message": "..."}}` format with SCREAMING_SNAKE codes |
| Hot reload | `backend/.air.toml` | Done — loads `.env` + `.env.local`, watches `.go` and `.yaml` |
| Backend Makefile | `backend/Makefile` | Done — setup, generate, test, lint, migration/create |

**Key dependencies:**
- `kelseyhightower/envconfig` — config from env vars
- `oapi-codegen/runtime` + `oapi-codegen/nethttp-middleware` — OpenAPI server + validation
- `getkin/kin-openapi` — OpenAPI spec loading
- `samber/slog-http` — structured logging + recovery middleware
- `modernc.org/sqlite` — pure Go SQLite driver
- `golang-migrate/migrate/v4` — database migrations (CLI only)
- `golang-jwt/jwt/v5` — JWT signing/verification
- `golang.org/x/crypto/bcrypt` — password hashing
- `rwcarlsen/goexif` — EXIF metadata extraction
- `disintegration/imaging` — image resizing, orientation transforms

### Frontend (SvelteKit)

| Component | Path | Status |
|-----------|------|--------|
| SvelteKit project | `frontend/` | Done — TypeScript, minimal template |
| Tailwind CSS v4 | `frontend/vite.config.ts` | Done — via `@tailwindcss/vite` plugin |
| Layout | `frontend/src/routes/+layout.svelte` | Scaffold — header with "HomePhotos" title |
| Home page | `frontend/src/routes/+page.svelte` | Scaffold — welcome placeholder |
| API client | `frontend/src/lib/api/client.ts` | Scaffold — manual `fetchHealth()` helper |
| Generated API client | `frontend/src/lib/api/gen/` | Done — generated by `@hey-api/openapi-ts`, includes auth and user endpoints |
| OpenAPI bundling | `package.json` scripts | Done — `validate-openapi`, `bundle-openapi`, `api:generate` via `@redocly/cli` + `@hey-api/openapi-ts` |

### OpenAPI

| Component | Path | Status |
|-----------|------|--------|
| Split spec (source of truth) | `openapi/openapi.yaml` | Done — health, auth, user, photo, scanner, and image endpoints |
| Path definitions | `openapi/paths/` | Done — health, auth (register/login/me), users, photos (list/detail), scanner (run/status), image (serve) |
| Schema definitions | `openapi/components/schemas/` | Done — auth, user, photo (list item, detail, list response), scanner (run/status responses), error |
| Error responses | `openapi/components/responses/` | Done — 400, 401, 403, 404, 409, 500 |
| Bundled spec | `openapi.yaml` (root) | Generated — single-file version used by code generators |

### Infrastructure

| Component | Path | Status |
|-----------|------|--------|
| Dockerfile | `infra/Dockerfile` | Done — multi-stage Alpine, `CGO_ENABLED=0` |
| Docker Compose | `docker-compose.yml` | Done — source (ro), cache, db volumes |
| Root Makefile | `Makefile` | Done — setup, dev, generate, test, lint, db targets |
| Dev script | `scripts/dev.sh` | Done — tmux-based, handles nested sessions |
| Env template | `.env.example` | Done |

### Tests

| Suite | Count | Coverage |
|-------|-------|----------|
| `database/sqlite` | 21 tests | User repository (8), photo repository (13): CRUD, cursor pagination, filters, orphan cleanup, list pending, update cache status |
| `services/auth` | 7 tests | Token: generate/validate, expired, invalid sig, malformed. Service: register/login |
| `services/scanner` | 11 tests | EXIF extraction (4), scanner service (7): new/incremental/changed/orphaned files, concurrency, scheduling |
| `services/imaging` | 11 tests | Resize (3), orientation (1), encode roundtrip (1), decode (1), RAW extraction (3), edge cases (2) |
| `services/cache` | 9 tests | Generate JPEG (1), corrupt file (1), has/path (1), cache dir layout (1), generate if needed (1), source reader (1), worker pool (2), context cancellation (1) |
| `ports/rest` | 30 tests | Auth (8), users (7), photos (6), scanner (5), images (5): full endpoint coverage with auth checks |
| **Total** | **89 tests** | |

### Documentation

| Doc | Status |
|-----|--------|
| `CLAUDE.md` | Done — project instructions for AI assistants |
| `README.md` | Done — quick start, project structure |
| `docs/00-10` | Done — full feature specs, all updated to reflect built-in auth (no Clerk) |

---

## What Needs To Be Built Next

Ordered by dependency — earlier items unblock later ones.

### Phase 4: Tagging

Organization layer.

1. **Tags migration** — `tags`, `tag_groups`, `photo_tags` tables (see `docs/06-feature-tagging.md`)
2. **Tag domain + repository + service**
3. **Tag API** — CRUD for tags and tag groups, photo tagging, bulk tagging
4. **Photo filtering by tags** — extend `GET /photos` with `tags` and `tag_mode` query params

### Phase 5: Frontend

Build out the SvelteKit UI.

1. **Auth pages** — login and registration forms
2. **Photo grid** — infinite scroll with thumbnails, date grouping
3. **Photo detail / lightbox** — preview image, EXIF metadata panel
4. **Tag sidebar** — filter by tags, tag management (admin)
5. **Admin panel** — scanner controls, user management, system stats

### Phase 6: Lightroom Integration (Stretch)

Optional, only after core features are solid.

1. **Lightroom catalog reader** — open `.lrcat` SQLite file, extract keywords/ratings/labels
2. **Path mapping** — translate Lightroom catalog paths to source mount paths
3. **Sync service** — match photos by path, import metadata
4. **Lightroom API** — sync trigger + status endpoints

---

## Development Workflow

```bash
# Day-to-day development
make dev                    # Start backend (air) + frontend (vite) in tmux

# After changing OpenAPI spec
make generate               # Bundle spec → generate backend stubs + frontend client

# After adding a new migration
make db/migrate/create name=AddUsersTable
# Edit the generated .up.sql and .down.sql files
make db/migrate             # Apply pending migrations

# Before committing
make lint                   # go vet + svelte-check
make test                   # go test + svelte-check
```

## Architecture Pattern

Follow the established clean architecture from the scaffold:

```
OpenAPI spec (openapi/)
    ↓ generate
ports/rest/         ← HTTP handlers, request/response mapping
    ↓ calls
services/           ← Business logic, orchestration
    ↓ calls
domain/             ← Entities, repository interfaces, errors
    ↓ implements
database/sqlite/    ← sqlc-generated queries, repository implementations
```

Each new feature follows this pattern:
1. Add endpoints to OpenAPI spec
2. Run `make generate`
3. Create domain entities and repository interface
4. Add sqlc queries and migration
5. Run `make generate` again (for sqlc)
6. Implement the service layer
7. Implement the handlers (the generated `ServerInterface` tells you what to implement)
