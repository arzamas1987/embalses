# Development Guide

## Repository Layout

We use a **single `web/` directory** for the frontend (not `apps/web`).
This keeps the root flat and matches the simple Docker Compose setup.
If we later add mobile apps or a separate docs site, we can migrate to `apps/`.

```
embalses/
├── cmd/
│   ├── api/           # REST API server entrypoint (PostgreSQL target)
│   ├── api-sqlite/    # REST API server entrypoint (MVP SQLite)
│   ├── updater/       # Data ingestion CLI (MITECO, IGN, SNCZI, optional regional)
│   ├── mcp/           # MCP server entrypoint
│   ├── ingest/        # Legacy ingestion CLI
│   ├── admin/         # Administrative CLI
│   └── seed/          # Seed fixtures / synthetic data
├── internal/
│   ├── config/        # Environment-based configuration
│   ├── api/v1sqlite/  # SQLite-backed API handlers
│   ├── storage/sqlite/# SQLite store + queries
│   ├── health/        # Health check handlers
│   ├── version/       # Build-time version info
│   └── db/            # PostgreSQL connection pool
├── api/               # OpenAPI specs and DTOs (placeholder)
├── migrations/        # golang-migrate SQL files (Postgres target)
├── scripts/           # Utility scripts
├── test/fixtures/     # Test fixtures
├── web/               # React + Vite + TypeScript frontend
├── docs/              # Research and planning docs
└── .github/workflows/ # CI/CD
```

## Technology Choices

| Layer | Choice | License | Rationale |
|-------|--------|---------|-----------|
| Router | chi v5 | MIT | net/http-native, minimal deps |
| DB driver (MVP) | modernc.org/sqlite | Public domain / BSD | SQLite driver in pure Go, no CGO |
| DB driver (target) | pgx v5 | MIT | Fast, feature-complete Postgres driver |
| Migrations (target) | golang-migrate | MIT | Industry standard, SQL up/down |
| Frontend | React + Vite + TS | MIT | Large ecosystem, fast HMR, easy PWA |
| Maps | MapLibre GL JS | BSD-3 | Community fork of Mapbox, free |
| Charts | Recharts | MIT | React-native, simple API |

## Environment Variables

All configuration is env-based. See `.env.example` for placeholders.

| Variable | Purpose | Required |
|----------|---------|----------|
| `DATABASE_URL` | SQLite file path (MVP) or Postgres connection string | Yes (defaults to `data/embalses.db`) |
| `API_ADDR` | API server bind address | No (default `:8080`) |
| `MCP_ADDR` | MCP server bind address | No (default `:8081`) |
| `WEB_PORT` | Frontend dev server port | No (default `5173`) |
| `GEMINI_API_KEY` | Google Gemini API (optional) | No |
| `GEMINI_MODEL` | Gemini model name | No |
| `AEMET_API_KEY` | AEMET OpenData key (optional) | No |
| `APP_ENV` | `development` or `production` | No (default `development`) |

## Updating data in the MVP

The MVP stores data in `data/embalses.db` (SQLite). All writes go through the
`cmd/updater` binary. The public UI/API only surfaces reservoirs whose latest
**MITECO** reading is less than 6 months old.

```mermaid
%%{init: {'theme': 'base', 'themeVariables': { 'fontFamily': 'Inter, sans-serif' }}%%
flowchart LR
    subgraph SRC ["Official MITECO sources"]
        HIST["BD-Embalses historical ZIP"]
        WEEKLY["Weekly bulletin PDF / XLSX"]
    end

    subgraph UPD ["Updater"]
        CLI["cmd/updater"]
    end

    subgraph STORE ["MVP storage"]
        DB[("SQLite<br/>data/embalses.db")]
    end

    HIST -->|./updater -miteco| CLI
    WEEKLY -->|./updater -weekly<br/>(not implemented yet)| CLI
    CLI -->|idempotent upsert| DB
    API["cmd/api-sqlite"] -->|reads| DB
    WEB["web browser"] -->|/api/v1/*| API

    classDef source fill:#4ECDC4,stroke:#2C3E50,stroke-width:2px,color:#fff,rx:10,ry:10;
    classDef ingest fill:#FF9F43,stroke:#2C3E50,stroke-width:2px,color:#fff,rx:10,ry:10;
    classDef backend fill:#45B7D1,stroke:#2C3E50,stroke-width:2px,color:#fff,rx:10,ry:10;
    classDef frontend fill:#96CEB4,stroke:#2C3E50,stroke-width:2px,color:#2C3E50,rx:10,ry:10;
    classDef db fill:#FF6B6B,stroke:#2C3E50,stroke-width:2px,color:#fff,rx:10,ry:10;
    classDef default rx:10,ry:10;

    class HIST,WEEKLY source;
    class CLI ingest;
    class API backend;
    class WEB frontend;
    class DB db;
```

### Common updater commands

```bash
# Build the updater
go build -o bin/updater ./cmd/updater

# One-time MITECO historical import (BD-Embalses ZIP)
./bin/updater -db data/embalses.db -miteco

# Full import (geo fixtures + MITECO historical)
./bin/updater -db data/embalses.db -full

# Seed synthetic 6-month readings (for UI testing only)
./bin/updater -db data/embalses.db -seed-readings

# Regional sources (ACA, CHD, CHJ) — disabled for the MVP public view,
# but can still be ingested for later releases
./bin/updater -db data/embalses.db -regional
```

> **Note:** The weekly MITECO bulletin fetcher (`fetchMITECOWeekly`) is still a
> stub. For now the dataset is kept current by re-running `-miteco` when MITECO
> publishes a new historical/weekly file, or by manually importing a CSV through
> `/api/v1/admin/readings/import`.

## Adding a Migration

```bash
# Create new migration files
migrate create -ext sql -dir migrations -seq add_reservoirs_table

# Run up
make db-migrate-up

# Run down
make db-migrate-down
```

## Testing

```bash
# Go tests
go test ./...

# Frontend tests
cd web && npm run test

# Integration tests (require DB)
go test -run Integration ./...
```

## CI Gates

1. `go build ./...` must pass
2. `go test ./...` must pass
3. `go vet ./...` must pass
4. `gofmt` check must pass
5. Frontend `npm run test` must pass
6. Frontend `npm run build` must pass
7. Migration up/down must work against PostGIS
8. License gate must pass (no GPL/AGPL/SSPL/BSL)
9. Secrets check must pass (no committed API keys)

## Security Rules

- Never execute arbitrary SQL. Use parameterized queries only.
- All query plans must be validated against an allow-list.
- Never commit secrets. Use `.env` (git-ignored).
- Never commit downloaded raw data.
