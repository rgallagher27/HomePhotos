# HomePhotos

Self-hosted photo management app. Go backend, SvelteKit frontend, SQLite database.

## Package Manager

Use `npm` for frontend. Run `npm install` from within `frontend/`, not from the repo root.

## Making API Changes

1. Edit files in `openapi/paths/` or `openapi/components/schemas/`
2. Run `make generate` (validates, bundles, generates all code)
3. Implement handlers/components

## Generated Code (Do Not Edit)

- `openapi.yaml` (bundled from `openapi/`)
- `backend/ports/rest/server.gen.go`
- `backend/database/sqlite/db.go`
- `backend/database/sqlite/models.go`
- `backend/database/sqlite/queries.sql.go`
- `frontend/src/lib/api/gen/*`

## Backend Patterns

- **Config**: `HOMEPHOTOS_` env var prefix via `kelseyhightower/envconfig`
- **Architecture**: Clean Architecture — `ports/rest/` (HTTP) → `services/` (business logic) → `domain/` (entities + repos)
- **Database**: SQLite via `modernc.org/sqlite`, type-safe queries via sqlc
- **HTTP**: `net/http` with oapi-codegen generated server interface
- **Migrations**: golang-migrate with sequential numbered files in `backend/database/sqlite/migrations/`

## Feature Documentation

Feature docs live in `docs/` — see README.md for the full index.
