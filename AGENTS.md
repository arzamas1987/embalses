# AGENTS.md — Embalses

This file is for AI coding agents. It summarises the project architecture, build/test commands, conventions, and safety rules. Read this before making non-trivial changes.

## Project overview

Embalses is a local-first Spanish reservoir data platform. It aggregates, exposes, and visualises official Spanish hydrological data through a REST API, a React single-page application, and (in scaffolding) an MCP server.

Current branch: `phase/06-real-data`. The SQLite-backed runtime is the actively developed path for local development; the PostgreSQL/PostGIS path is the documented reference architecture and the one exercised by CI.

## Technology stack

### Backend

- **Go 1.23** (`go.mod`)
- **Router:** chi v5 (`github.com/go-chi/chi/v5`)
- **Databases:**
  - PostgreSQL 16 + PostGIS (primary/reference stack, `pgx/v5`)
  - SQLite (`modernc.org/sqlite`) for local-first file-based backend
- **Migrations:** `golang-migrate` for Postgres; in-process schema setup for SQLite
- **Testing:** standard `testing`, `testify` (indirect dependency)

### Frontend

- **React 19 + TypeScript 6**
- **Build tool:** Vite 8
- **Styling:** Tailwind CSS 4
- **Routing:** React Router DOM 7
- **Data fetching / caching:** TanStack Query 5
- **Charts:** Recharts
- **Maps:** MapLibre GL
- **i18n:** i18next + react-i18next (Spanish and English)
- **Linting:** Oxlint
- **Testing:** Vitest + jsdom + React Testing Library

### Infrastructure / tooling

- Docker and Docker Compose
- GitHub Actions CI
- `Makefile` for common tasks

## Repository layout

```
embalses/
├── api/                    # OpenAPI 3.0 spec (openapi.yaml)
├── bin/                    # Compiled Go binaries (git-ignored artefacts)
├── cmd/                    # Application entrypoints
│   ├── api/               # REST API server (PostgreSQL)
│   ├── api-sqlite/        # REST API server (SQLite) — current local focus
│   ├── ingest/            # GeoJSON ingestion CLI for SNCZI / IGN (partial)
│   ├── mcp/               # MCP server entrypoint (stub)
│   ├── seed/              # Synthetic 6-month reservoir dataset seeder (Postgres)
│   └── updater/           # SQLite updater / orchestrator (geo import, synthetic readings)
├── data/                   # SQLite database and source GeoJSON fixtures
├── internal/               # Private Go packages
│   ├── api/v1/            # PostgreSQL REST handlers, queries, middleware
│   ├── api/v1sqlite/      # SQLite REST handlers (reuses v1 response types)
│   ├── config/            # Env-var configuration loader
│   ├── db/                # PostgreSQL pgxpool wrapper
│   ├── geo/               # PostGIS spatial joins
│   ├── geo/ign/           # IGN GeoJSON parser
│   ├── geo/snczi/         # SNCZI dam/reservoir GeoJSON parser
│   ├── health/            # Health check handler
│   ├── planner/           # Safe query planner (intent → validated plan → SQL)
│   ├── storage/sqlite/    # SQLite open / migrate / schema / queries
│   └── version/           # Build-time version info
├── migrations/             # golang-migrate SQL up/down files
├── scripts/                # Windows helper PowerShell scripts
├── test/fixtures/          # GeoJSON test fixtures
├── web/                    # React + Vite + TypeScript SPA
└── docs/                   # Architecture, roadmap, data sources, licensing, etc.
```

## Build and run commands

All commands assume the repository root as working directory.

### Setup

```bash
cp .env.example .env
make setup          # downloads Go deps and runs npm install in web/
```

### PostgreSQL stack (reference architecture)

```bash
make dev            # docker compose up --build -d db api mcp web
curl http://localhost:8080/healthz   # API
curl http://localhost:8082/healthz   # MCP (external port)
```

Run migrations separately when needed:

```bash
make db-migrate-up
make db-migrate-down
```

Seed synthetic data for testing:

```bash
go run ./cmd/seed
```

### SQLite stack (current local development focus)

Populate the SQLite database (requires `data/embalses.db` to exist; the updater creates the schema):

```bash
make updater
./bin/updater -db data/embalses.db -geo-only -seed-readings
```

Import real MITECO historical reservoir data (1988–present). This downloads `BD-Embalses.zip` from MITECO and exports the included Microsoft Access database to CSV using `mdb-export` (system `mdbtools` or Docker fallback):

```bash
./bin/updater -db data/embalses.db -miteco
```

Start the SQLite API and the frontend together:

```bash
./scripts/start.sh [api-base-port] [frontend-base-port]
# defaults: API 8082, frontend 5173; picks the next free port if the default is taken
```

Or manually:

```bash
make api-sqlite
DATABASE_URL=data/embalses.db ./bin/api-sqlite
cd web && npm run dev
```

### Individual Go binaries

```bash
make api            # cmd/api → bin/api
make api-sqlite     # cmd/api-sqlite → bin/api-sqlite
make mcp            # cmd/mcp → bin/mcp
make ingest         # cmd/ingest → bin/ingest
make admin          # cmd/admin → bin/admin
make updater        # cmd/updater → bin/updater
```

### Frontend

```bash
cd web
npm run dev         # Vite dev server on 5173
npm run build       # type-check and build to web/dist
npm run preview     # preview production build
npm run test        # Vitest
npm run lint        # Oxlint
```

## Environment variables

Copy `.env.example` to `.env` and fill values. Never commit `.env`.

| Variable | Purpose | Required |
|---|---|---|
| `DATABASE_URL` | Postgres connection string OR path to SQLite file | Yes |
| `API_ADDR` | API bind address | No (default `:8080`) |
| `MCP_ADDR` | MCP bind address | No (default `:8081`) |
| `WEB_PORT` | Frontend dev server port | No (default `5173`) |
| `GEMINI_API_KEY` | Optional Google Gemini API key | No |
| `GEMINI_MODEL` | Optional Gemini model name | No |
| `AEMET_API_KEY` | Optional AEMET OpenData key | No |
| `APP_ENV` | `development` or `production` | No (default `development`) |

The Makefile detects `sqlite` in `DATABASE_URL` and treats the backend as SQLite.

## API surface

Base path: `/api/v1`. OpenAPI spec is in `api/openapi.yaml`.

| Endpoint | Description |
|---|---|
| `GET /healthz` | Health check |
| `GET /readyz` | Database readiness |
| `GET /sources` | Data sources with attribution |
| `GET /reservoirs` | Paginated reservoir list |
| `GET /reservoirs/{slug}` | Reservoir detail |
| `GET /reservoirs/{slug}/readings` | Time-series readings (`since`, `until`, `page`, `per_page`) |
| `GET /basins` | List basins |
| `GET /basins/summary` | Aggregated fill statistics per basin |
| `GET /basins/{slug}` | Basin detail with reservoirs and aggregate fill |
| `GET /rankings/reservoirs?metric=fullest\|emptiest&limit=N` | Rankings |
| `GET /compare?reservoir=A&reservoir=B` | Compare up to 5 reservoirs |
| `GET /data-quality` | Data quality report |
| `POST /query` | Safe structured query via the planner |
| `POST /admin/readings/import` | Import CSV readings for reservoirs missing MITECO data |

Authentication: API key via `X-API-Key` header or `api_key` query parameter. The test key `test-key-123` is seeded. Middleware enforces per-key rate limits and daily quotas.

Every response uses the envelope:

```json
{
  "data": ..., 
  "meta": {...}, 
  "error": {...}, 
  "lineage": {...}
}
```

## Code style guidelines

### Go

- Format with `gofmt`. CI fails on unformatted files.
- Standard project layout: `cmd/` for entrypoints, `internal/` for private packages.
- Keep handlers in `internal/api/v1` and `internal/api/v1sqlite` thin; business logic belongs in dedicated packages (`planner`, `storage/sqlite`, etc.).
- Use parameterised queries only. Never concatenate user input into SQL.
- Handlers reuse the `v1` response envelope types even in the SQLite backend (`internal/api/v1sqlite`).

### Frontend

- TypeScript with strict-ish settings (`noUnusedLocals`, `noUnusedParameters`, `verbatimModuleSyntax`, `erasableSyntaxOnly`).
- Components in `web/src/components/` and `web/src/pages/` are default-exported function components.
- API calls go through `web/src/api/client.ts`; custom hooks live in `web/src/hooks/useQueries.ts`.
- Tailwind utility classes are mixed with custom design-system classes prefixed `gov-*` (e.g. `gov-card`, `gov-btn`) in `web/src/styles/global.css`.
- i18n keys are namespaced (`home.*`, `map.*`, `reservoir.*`, etc.) and mirrored in `es.json` / `en.json`.
- Inline SVG components are preferred over icon libraries.
- UI routes are Spanish (`/mapa`, `/embalses`, `/comparar`, `/cuencas`, `/fuentes`, `/calidad-datos`, `/ajustes`).

## Testing instructions

### Go

```bash
# All tests
make test

# Unit only (skips integration tests when -short is honoured)
make test-unit

# Integration (requires DATABASE_URL pointing to a reachable Postgres/PostGIS)
make test-integration
```

Integration tests create unique temporary schemas and run migration files directly; they skip gracefully if the database is unavailable.

### Frontend

```bash
cd web && npm run test
```

Uses Vitest globals, jsdom, and mocked `globalThis.fetch`.

### CI

`.github/workflows/ci.yml` runs on pushes to `main`/`mvp/*` and on PRs to `main`:

1. `go-checks`: `go mod tidy` check, `go build ./...`, `go vet ./...`, `gofmt` check, `go test` against PostGIS, migration up/down/status.
2. `frontend-checks`: `npm ci`, `npm run lint`, `npm run test`, `npm run build`.
3. `license-gate`: placeholder licence scan (see `docs/licensing.md` for allow/deny lists).
4. `secrets-check`: greps for `sk-*` and `AIza*` patterns.
5. `e2e-smoke`: full Ubuntu end-to-end with Postgres, migrations, seed, API, and frontend preview, plus curl smoke tests.

## Security considerations

- **Never execute arbitrary SQL.** All user queries must go through the safe query planner in `internal/planner`. The planner validates Query Intents against strict allow-lists and compiles only parameterised SQL.
- **Never commit secrets.** Use `.env` (git-ignored). CI rejects `sk-*` and `AIza*` patterns.
- **API keys and metering:** `api_keys` and `metering` tables exist. Rate limits are in-memory per key/minute; daily quotas are DB-backed. Free usage is currently ungated in local development.
- **Map library:** MapLibre GL JS (BSD-3) is used. Do not introduce Mapbox GL JS v2+ (proprietary licence).
- **PostGIS:** stock, unmodified PostGIS is acceptable as a service over SQL. Do not fork or redistribute PostGIS code.

## Data sources and licensing

The project uses official Spanish public-sector sources only. `estadoembalses.es` is treated as a benchmark, not a source to scrape.

Core sources:

- **MITECO Boletín Hidrológico Semanal** — weekly reservoir storage / capacity / fill data.
- **IGN / CNIG** — hydrography, basins, provinces, reservoir polygons (CC-BY 4.0).
- **SNCZI Inventario de Presas y Embalses** — dam/reservoir inventory and geolocation.

Planned: SAIH real-time data (Ebro, Júcar first), AEMET OpenData.

Current real data in `data/embalses.db`: 368 SNCZI reservoirs, 16 basins, 47 provinces.

Dependency licence policy is permissive-first. Allowed: MIT, Apache-2.0, BSD-2/3-Clause, ISC, PostgreSQL, MPL-2.0 (with file-level tracking). Denied: GPL-*, AGPL-*, SSPL, BUSL, Commons Clause, unlicensed, proprietary. See `docs/licensing.md`.

Attribution is mandatory in API responses, MCP tools, exports, and the UI.

## Deployment and operations

### Docker Compose (PostgreSQL stack)

Services in `docker-compose.yml`:

- `db`: `postgis/postgis:16-3.4` on port `5432`
- `migrate`: one-shot migration container (`--profile migrate`)
- `api`: Go REST API on port `8080`
- `mcp`: Go MCP server on external port `8082` (internal `8081`)
- `web`: Vite dev server on port `5173`

Run: `make dev` or `docker compose up --build -d db api mcp web`.

### Local SQLite quick start

For the current branch the fastest local path is:

```bash
make updater
./bin/updater -db data/embalses.db -geo-only -seed-readings
./scripts/start.sh 8082 5173
```

`scripts/start.sh` kills stale processes, verifies the database, builds `bin/api-sqlite` if missing, picks free ports, starts the API, installs frontend deps if needed, and starts Vite. A Windows equivalent is `scripts/start.ps1`.

### Migrations

Create a new Postgres migration:

```bash
migrate create -ext sql -dir migrations -seq <name>
```

Run up/down:

```bash
make db-migrate-up
make db-migrate-down
```

## Current state and known limitations

- **Implemented and working:** PostgreSQL REST API, SQLite REST API, web frontend, safe query planner, real SNCZI reservoir import into SQLite.
- **Stubs / partial:** MCP server (`cmd/mcp`), admin CLI (`cmd/admin`), MITECO ingestion, SAIH real-time ingestion.
- **README vs. branch:** `README.md` still describes the Postgres-first quick start, while `scripts/start.sh` + SQLite is the current local demo path.
- **CI licence scan:** placeholder; not yet enforced with real tooling.
- **Migration up/down in CI:** marked non-blocking with a reference to issue #5.

When adding features, prefer to extend the SQLite path (`cmd/api-sqlite`, `internal/api/v1sqlite`, `internal/storage/sqlite`) if the change is data-model related, and keep the PostgreSQL path (`cmd/api`, `internal/api/v1`) in sync where practical.
