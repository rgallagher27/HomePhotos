# Implementation Status

## What Has Been Built

### Phase 1: Auth & Users — Complete

User authentication and role-based access control are fully implemented.

### Phase 2: Photo Scanning & Indexing — Complete

Photo scanning, EXIF extraction, and browsing APIs are fully implemented.

### Phase 3: Caching & Thumbnails — Complete

Thumbnail/preview generation (background + on-demand) and image serving endpoint are fully implemented.

### Phase 4: Tagging — Complete

Tag-based photo organization: tags, tag groups, photo-tag associations, CRUD APIs, bulk tagging, and tag-based filtering on the photo list endpoint.

### Backend (Go)

| Component | Path | Status |
|-----------|------|--------|
| Entry point | `backend/cmd/server/main.go` | Done — signal handling, graceful shutdown, envconfig |
| Config | `backend/config/config.go` | Done — `HOMEPHOTOS_` prefix, all env vars defined |
| SQLite connection | `backend/database/sqlite/sqlite.go` | Done — WAL mode, foreign keys, busy timeout, single writer |
| Auto-migrations | `backend/database/sqlite/migrate.go` | Done — embeds SQL files, applies on startup, version tracking |
| Migrations | `backend/database/sqlite/migrations/` | Done — `000001_init` (schema_info), `000002_add_users_table` (users), `000003_add_photos_table` (photos with EXIF columns), `000004_add_tagging_tables` (tag_groups, tags, photo_tags) |
| sqlc | `backend/database/sqlite/sqlc.yaml` | Done — configured for SQLite, generates to `database/sqlite/` |
| sqlc queries | `backend/database/sqlite/queries/` | Done — `health.sql` (Ping), `users.sql` (7 queries), `photos.sql` (7 queries), `tags.sql` (tag groups CRUD, tags CRUD with joins, photo-tag associations) |
| User repository | `backend/database/sqlite/user_repository.go` | Done — implements `user.Repository`, maps sqlc models to domain entities |
| Photo repository | `backend/database/sqlite/photo_repository.go` | Done — implements `photo.Repository`, cursor pagination with dynamic SQL, orphan cleanup, list pending, update cache status, tag filtering (OR/AND modes) |
| Tag repository | `backend/database/sqlite/tag_repository.go` | Done — implements `tag.Repository`, tags/groups/photo-tags CRUD, batch operations, ListTagsForPhotos |
| User domain | `backend/domain/user/` | Done — `User` entity, `Repository` interface, `Role` type, sentinel errors |
| Photo domain | `backend/domain/photo/` | Done — `Photo` entity with EXIF fields, `Repository` interface with `ListParams`/`ListResult` (includes TagIDs/TagMode), sentinel errors |
| Tag domain | `backend/domain/tag/` | Done — `Tag` and `TagGroup` entities, `Repository` interface, sentinel errors |
| Auth service | `backend/services/auth/service.go` | Done — Register (bcrypt, first-user-admin), Login (verify + JWT) |
| Token service | `backend/services/auth/token.go` | Done — HMAC-SHA256 JWT, configurable expiry |
| Auth errors | `backend/services/auth/errors.go` | Done — `ErrInvalidCredentials`, `ErrRegistrationClosed` |
| Scanner service | `backend/services/scanner/service.go` | Done — walks source directory, skips zero-byte files, diffs fingerprints, extracts EXIF, upserts photos, deletes orphans, tracks scan results (added/updated/unchanged/deleted/skipped) |
| Scanner scheduler | `backend/services/scanner/scheduler.go` | Done — background goroutine with configurable interval and on-startup scan |
| EXIF extraction | `backend/services/scanner/exif.go` | Done — extracts camera, lens, GPS, exposure data via `rwcarlsen/goexif` |
| Image processing | `backend/services/imaging/imaging.go` | Done — decode (JPEG/PNG/TIFF/RAW), resize via Lanczos, JPEG encode |
| EXIF orientation | `backend/services/imaging/orientation.go` | Done — applies EXIF orientation transforms (values 1-8) |
| RAW extraction | `backend/services/imaging/raw.go` | Done — pure Go TIFF/IFD parser, extracts embedded JPEG from ARW/DNG |
| Cache service | `backend/services/cache/service.go` | Done — content-addressable cache, generate thumb+preview, on-demand generation |
| Cache worker pool | `backend/services/cache/worker.go` | Done — background batch processing of pending photos |
| OpenAPI server | `backend/ports/rest/server.gen.go` | Generated — `ServerInterface` with health, auth, user, photo, scanner, image, tag, and photo-tag endpoints |
| Server struct | `backend/ports/rest/server.go` | Done — holds `*sql.DB`, auth service, token service, user repository, photo repository, tag repository, scanner service, cache service |
| Server init | `backend/ports/rest/rest.go` | Done — builds services, wires JWT auth, sets up middleware stack, constructs scanner and cache |
| Health handler | `backend/ports/rest/health_handlers.go` | Done — pings DB, returns ok/degraded |
| Auth handlers | `backend/ports/rest/auth_handlers.go` | Done — register, login, me endpoints |
| User handlers | `backend/ports/rest/user_handlers.go` | Done — list users, update role (admin only) |
| Photo handlers | `backend/ports/rest/photo_handlers.go` | Done — list photos (cursor pagination, filters, tag filtering), get photo detail with full EXIF and tags |
| Tag handlers | `backend/ports/rest/tag_handlers.go` | Done — CRUD for tags and tag groups (admin only for writes) |
| Photo-tag handlers | `backend/ports/rest/photo_tag_handlers.go` | Done — assign tags, remove tag, bulk tag (auth required) |
| Scanner handlers | `backend/ports/rest/scanner_handlers.go` | Done — trigger scan (admin), get scan status with results (admin) |
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

### Phase 5: Frontend — Complete

SvelteKit frontend with auth, photo browsing, tag filtering, and admin panel.

### Phase 6: Polish & Hardening — Complete

Auto-migrations on startup, error handling and loading states across all components, custom error page, keyboard/accessibility fixes, arrow key photo navigation.

### Frontend (SvelteKit)

| Component | Path | Status |
|-----------|------|--------|
| SvelteKit project | `frontend/` | Done — TypeScript, Svelte 5 runes |
| Tailwind CSS v4 | `frontend/vite.config.ts` | Done — via `@tailwindcss/vite` plugin |
| shadcn-svelte | `frontend/src/lib/components/ui/` | Done — Button, Input, Label, Card, Tabs, Breadcrumb, Sheet, Badge, Table, Progress, Skeleton, Spinner, AlertDialog, Tooltip |
| Root layout | `frontend/src/routes/+layout.svelte` | Done — auth context, SDK init, route guard, header nav, Spinner loading state |
| Auth state | `frontend/src/lib/auth.svelte.ts` | Done — reactive auth class with localStorage + cookie sync |
| SDK setup | `frontend/src/lib/api/setup.ts` | Done — client config with Bearer auth |
| Login page | `frontend/src/routes/login/+page.svelte` | Done — username/password form |
| Register page | `frontend/src/routes/register/+page.svelte` | Done — username/password/email form |
| Image proxy | `frontend/src/routes/img/[id]/[size]/+server.ts` | Done — server-side auth proxy for images |
| Image helpers | `frontend/src/lib/image.ts` | Done — thumbUrl, previewUrl, fullUrl |
| Photo grid page | `frontend/src/routes/(app)/+page.svelte` | Done — infinite scroll, tag filtering, Sheet sidebar (mobile), date/folder grouping toggle |
| PhotoGrid | `frontend/src/lib/components/PhotoGrid.svelte` | Done — responsive grid, date/folder grouping, IntersectionObserver, Skeleton loading |
| PhotoCard | `frontend/src/lib/components/PhotoCard.svelte` | Done — thumbnail with hover overlay |
| Photo detail | `frontend/src/routes/(app)/photos/[id]/+page.svelte` | Done — preview/full image, EXIF panel, tag editing, arrow key navigation, Tooltip on nav arrows, Skeleton loading |
| Photo navigation | `frontend/src/lib/photoNav.svelte.ts` | Done — stores photo ID list for prev/next cycling |
| Error page | `frontend/src/routes/+error.svelte` | Done — custom error page with status code and home link |
| ExifPanel | `frontend/src/lib/components/ExifPanel.svelte` | Done — camera, lens, exposure, GPS metadata |
| PhotoTags | `frontend/src/lib/components/PhotoTags.svelte` | Done — add/remove tags, admin can create inline |
| TagSidebar | `frontend/src/lib/components/TagSidebar.svelte` | Done — grouped tags, AND/OR toggle, Skeleton loading |
| FolderBreadcrumb | `frontend/src/lib/components/FolderBreadcrumb.svelte` | Done — shadcn Breadcrumb, navigable path segments |
| FolderGrid | `frontend/src/lib/components/FolderGrid.svelte` | Done — grid of folder tiles for drill-down browsing |
| VirtualGroup | `frontend/src/lib/components/VirtualGroup.svelte` | Done — IntersectionObserver wrapper for group-level virtual scrolling |
| TagChip | `frontend/src/lib/components/TagChip.svelte` | Done — reusable colored tag chip |
| Admin layout | `frontend/src/routes/(app)/admin/+layout.svelte` | Done — admin-only guard |
| Admin page | `frontend/src/routes/(app)/admin/+page.svelte` | Done — shadcn Tabs: Scanner, Users, Tags |
| ScannerPanel | `frontend/src/lib/components/ScannerPanel.svelte` | Done — run scan, Badge status, Progress bar, scan results display |
| UserManagement | `frontend/src/lib/components/UserManagement.svelte` | Done — shadcn Table, AlertDialog for role changes, Skeleton loading |
| TagManagement | `frontend/src/lib/components/TagManagement.svelte` | Done — tag/group CRUD, AlertDialog delete confirmations |
| Generated API client | `frontend/src/lib/api/gen/` | Done — generated by `@hey-api/openapi-ts` |
| OpenAPI bundling | `package.json` scripts | Done — `validate-openapi`, `bundle-openapi`, `api:generate` via `@redocly/cli` + `@hey-api/openapi-ts` |

### OpenAPI

| Component | Path | Status |
|-----------|------|--------|
| Split spec (source of truth) | `openapi/openapi.yaml` | Done — health, auth, user, photo, scanner, image, tag, and photo-tag endpoints |
| Path definitions | `openapi/paths/` | Done — health, auth (register/login/me), users, photos (list/detail), scanner (run/status), image (serve), tags (CRUD), tag-groups (CRUD), photo-tags (assign/remove/bulk) |
| Schema definitions | `openapi/components/schemas/` | Done — auth, user, photo (list item with tag summaries, detail with full tags, list response), scanner, tag (response, list, create/update), tag group, photo-tag (assign, bulk), error |
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
| `database/sqlite` | 38 tests | User repository (8), photo repository (16): CRUD, cursor pagination, filters, tag filtering (OR/AND/combined), orphan cleanup, list pending, update cache status. Tag repository (11): group CRUD, tag CRUD, duplicates, cascades, photo-tag operations, batch queries. Migrations (3): fresh DB, idempotent, partial |
| `services/auth` | 7 tests | Token: generate/validate, expired, invalid sig, malformed. Service: register/login |
| `services/scanner` | 14 tests | EXIF extraction (4), scanner service (10): new/incremental/changed/orphaned files, concurrency, scheduling |
| `services/imaging` | 11 tests | Resize (3), orientation (1), encode roundtrip (1), decode (1), RAW extraction (3), edge cases (2) |
| `services/cache` | 9 tests | Generate JPEG (1), corrupt file (1), has/path (1), cache dir layout (1), generate if needed (1), source reader (1), worker pool (2), context cancellation (1) |
| `ports/rest` | 54 tests | Auth (8), users (7), photos (9), scanner (5), images (5), tags (13), photo-tags (9): full endpoint coverage with auth checks |
| **Total** | **133 tests** | |

### Documentation

| Doc | Status |
|-----|--------|
| `CLAUDE.md` | Done — project instructions for AI assistants |
| `README.md` | Done — quick start, project structure |
| `docs/00-10` | Done — full feature specs, all updated to reflect built-in auth (no Clerk) |

---

## What Needs To Be Built Next

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
