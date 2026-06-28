# Embalses MVP

Local-first Spanish reservoir data platform. Phase 0 — Foundations & scaffolding.

## Quick Start

```bash
# 1. Copy environment
cp .env.example .env

# 2. Install dependencies and start the stack
make setup
make dev
```

## Local Development Commands

```bash
# Build all Go binaries
make api
make mcp
make ingest
make admin

# Run tests
make test
make test-unit
make test-integration

# Lint
make lint

# Database migrations (requires running DB)
make db-migrate-up
make db-migrate-down

# Frontend (from web/ directory)
cd web && npm run dev
```

## Docker Compose

```bash
# Start the full stack (DB + API + MCP + web)
docker compose up --build -d

# Check health
curl http://localhost:8080/healthz   # API
curl http://localhost:8082/healthz  # MCP (port 8082 due to local conflict)
```

## Architecture

- **Backend**: Go 1.23 + chi router + pgx/Postgres + golang-migrate
- **Frontend**: React 19 + Vite + TypeScript
- **Database**: PostgreSQL 16 + PostGIS
- **MCP**: Model Context Protocol server (stub)

## Documentation

- [Architecture](docs/architecture.md)
- [Roadmap](docs/roadmap.md)
- [Issues](docs/issues.md)
- [Data Sources](docs/data-sources.md)
- [Licensing](docs/licensing.md)
- [Benchmark](docs/benchmark.md)
- [Development](docs/development.md)

## License

See [docs/licensing.md](docs/licensing.md) for dependency and data attribution policies.
