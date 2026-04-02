# HomePhotos

A self-hosted photo management application for browsing and organizing photos stored on a TrueNAS server. Designed as a lightweight, family-friendly alternative to Google Photos and Immich.

## Why HomePhotos?

- **Large RAW files** (Sony ARW from an A7RV, ~50MB each) don't sync well to phones, tablets, or laptops
- **No cloud dependency** — runs entirely on your home network, accessible remotely via Tailscale
- **Read-only** — never modifies your original files on disk
- **Optimistic caching** — generates and serves optimized thumbnails and previews so browsing is fast
- **Multi-user** — account-scoped access for family members with role-based permissions
- **Tagging** — organize photos with tags and tag groups without touching the source files
- **Lightroom integration** (stretch goal) — import keywords and develop settings from your Lightroom Classic catalog

## Documentation

See the [`docs/`](docs/) directory for detailed feature specifications and architecture:

| Doc | Description |
|-----|-------------|
| [00 - Project Overview](docs/00-project-overview.md) | Vision, goals, personas, and scope |
| [01 - Architecture](docs/01-architecture.md) | System design, data flow, and key decisions |
| [02 - Tech Stack](docs/02-tech-stack.md) | Technology choices and rationale |
| [03 - User Management](docs/03-feature-user-management.md) | Auth, roles, and user model |
| [04 - Photo Browsing](docs/04-feature-photo-browsing.md) | Scanning, browse views, and image serving |
| [05 - Caching & Thumbnails](docs/05-feature-caching-thumbnails.md) | Thumbnail pipeline, cache layout, and performance |
| [06 - Tagging](docs/06-feature-tagging.md) | Tags, tag groups, and filtering |
| [07 - Lightroom Integration](docs/07-feature-lightroom-integration.md) | Catalog sync, keyword import, develop settings |
| [08 - Non-Functional Requirements](docs/08-non-functional-requirements.md) | Security, performance, reliability |
| [09 - API Design](docs/09-api-design.md) | REST API specification |
| [10 - Deployment](docs/10-deployment.md) | Docker Compose setup and configuration |

## Quick Start

### Prerequisites

- Go 1.24+
- Node.js 20+
- [air](https://github.com/air-verse/air) (Go hot reload)
- [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) (OpenAPI code gen)
- [sqlc](https://sqlc.dev/) (SQL code gen)
- [migrate](https://github.com/golang-migrate/migrate) (database migrations)

### Setup

```bash
make setup        # Install all tools and dependencies
make generate     # Generate server stubs and API client
make dev          # Start backend + frontend (uses tmux)
```

Backend runs at `http://localhost:8080`, frontend at `http://localhost:5173`.

### Project Structure

```
HomePhotos/
├── backend/              # Go API server
│   ├── cmd/server/       # Entry point
│   ├── config/           # Config via envconfig
│   ├── domain/user/      # User entity, repository interface, errors
│   ├── domain/photo/     # Photo entity, repository interface, errors
│   ├── domain/tag/       # Tag/TagGroup entities, repository interface, errors
│   ├── services/auth/    # Auth service (register, login, JWT)
│   ├── services/scanner/ # Photo scanner, EXIF extraction, scheduler
│   ├── services/imaging/ # Image decode, resize, orientation, RAW extraction
│   ├── services/cache/   # Thumbnail/preview cache, background worker pool
│   ├── ports/rest/       # HTTP handlers, middleware, generated server
│   ├── database/sqlite/  # SQLite, migrations, sqlc queries, repositories
│   └── Makefile
├── frontend/             # SvelteKit web app
│   └── src/
│       ├── routes/       # SvelteKit pages
│       └── lib/api/      # API client (manual + generated)
├── openapi/              # Split OpenAPI spec (source of truth)
├── infra/                # Dockerfile
├── docker-compose.yml    # Production deployment
├── Makefile              # Root orchestrator
└── docs/                 # Feature specs and architecture
```

## Status

Phases 1-4 are complete. The backend has user auth, role-based access control, filesystem scanning with EXIF extraction, photo browsing APIs with cursor pagination, thumbnail/preview generation with image serving, and tag-based photo organization with filtering. See [`docs/implementation.md`](docs/implementation.md) for full details.
