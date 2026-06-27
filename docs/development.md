# Development Guide

## Repository Layout

We use a **single `web/` directory** for the frontend (not `apps/web`).
This keeps the root flat and matches the simple Docker Compose setup.
If we later add mobile apps or a separate docs site, we can migrate to `apps/`.

```
embalses/
├── cmd/
│   ├── api/           # REST API server entrypoint
│   ├── mcp/           # MCP server entrypoint
│   ├── ingest/        # Data ingestion CLI
│   └── admin/         # Administrative CLI
├── internal/
│   ├── config/        # Environment-based configuration
│   ├── health/        # Health check handlers
│   ├── version/       # Build-time version info
│   └── db/            # PostgreSQL connection pool
├── api/               # OpenAPI specs and DTOs (placeholder)
├── migrations/        # golang-migrate SQL files
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
| DB driver | pgx v5 | MIT | Fast, feature-complete Postgres driver |
| Migrations | golang-migrate | MIT | Industry standard, SQL up/down |
| Frontend | React + Vite + TS | MIT | Large ecosystem, fast HMR, easy PWA |
| Maps | MapLibre GL JS | BSD-3 | Community fork of Mapbox, free |
| Charts | Recharts | MIT | React-native, simple API |

## Environment Variables

All configuration is env-based. See `.env.example` for placeholders.

| Variable | Purpose | Required |
|----------|---------|----------|
| `DATABASE_URL` | Postgres connection string | Yes |
| `API_ADDR` | API server bind address | No (default `:8080`) |
| `MCP_ADDR` | MCP server bind address | No (default `:8081`) |
| `WEB_PORT` | Frontend dev server port | No (default `5173`) |
| `GEMINI_API_KEY` | Google Gemini API (optional) | No |
| `GEMINI_MODEL` | Gemini model name | No |
| `AEMET_API_KEY` | AEMET OpenData key (optional) | No |
| `APP_ENV` | `development` or `production` | No (default `development`) |

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
