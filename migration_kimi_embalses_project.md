# Embalses — Migration Guide: From Windows to Ubuntu WSL

> **Date:** 2026-06-28
> **Branch:** `mvp/05-frontend-mvp`
> **Status:** Phases 0–5 complete. MVP running in Docker. UI uses synthetic placeholder data; real MITECO ingestion pending.

---

## What is Embalses?

**Embalses** is a Spanish reservoir data platform — an open-source alternative to the official MITECO (Ministerio para la Transición Ecológica) water portal. It provides:

- **Interactive maps** (MapLibre GL) showing reservoir locations
- **Historical time-series analysis** with 6-month rolling data
- **Comparative analytics** (compare up to 5 reservoirs side-by-side)
- **Data quality dashboards** with source attribution
- **Safe query planner** — structured intent-to-SQL (no arbitrary SQL accepted)
- **Full API** with API keys, rate limiting, quotas, and metering
- **Source lineage** — every data point includes MITECO attribution

**Target:** Municipal engineers, researchers, journalists, and citizens tracking Spain's water security.

---

## Architecture Overview

```
┌──────────────────────────────────────────────────────────────┐
│  Browser (React SPA)                                         │
│  ├── MapLibre GL (maps)                                      │
│  ├── Recharts (charts)                                       │
│  └── TanStack Query (data fetching)                          │
│         ↕ HTTP /api/v1/*                                       │
├──────────────────────────────────────────────────────────────┤
│  Go Backend (API Server)                                      │
│  ├── chi Router (/api/v1/...)                                │
│  ├── API Key Middleware (auth + rate limit + quota)        │
│  ├── Safe Query Planner (intent → parameterized SQL)         │
│  └── pgx (PostgreSQL + PostGIS)                              │
│         ↕ SQL                                                │
├──────────────────────────────────────────────────────────────┤
│  PostgreSQL 16 + PostGIS (Docker)                            │
│  ├── PostGIS for spatial queries (ETRS89 → 4326)             │
│  ├── Full reading lineage (source, licence, attribution)       │
│  └── 18 seeded reservoirs with 486 weekly synthetic readings   │
└──────────────────────────────────────────────────────────────┘
```

---

## Repository Layout

```
embalses/
├── api/
│   ├── openapi.yaml               # OpenAPI 3 spec (Issues #29, #31)
│   └── placeholder.go
├── cmd/
│   ├── admin/
│   │   └── main.go                # Admin CLI entry point
│   ├── api/
│   │   └── main.go                # API server entry point
│   ├── ingest/
│   │   └── main.go                # Batch data ingestion (Phase 2)
│   ├── mcp/
│   │   └── main.go                # MCP server entry point
│   └── seed/
│       └── main.go                # 6-month synthetic seed data generator
├── docs/
│   ├── architecture.md
│   ├── benchmark.md
│   ├── data-sources.md            # Real upstream sources (MITECO, IGN, SAIH)
│   ├── development.md
│   ├── domains.md
│   ├── issues.md                  # All 52 project issues
│   ├── licensing.md
│   ├── plan.md
│   ├── query-intent.md            # Safe Query Planner grammar
│   └── roadmap.md
├── internal/
│   ├── api/
│   │   └── v1/
│   │       ├── handlers.go        # 11 HTTP handlers + Query endpoint
│   │       ├── handlers_test.go   # 12 integration tests
│   │       ├── middleware.go      # API key auth + rate limit + quota
│   │       ├── queries.go         # All parameterized SQL queries
│   │       ├── responses.go       # Standardized JSON envelope
│   │       └── routes.go          # Route registration
│   ├── config/
│   │   ├── config.go              # DATABASE_URL from env
│   │   └── config_test.go
│   ├── db/
│   │   ├── db.go                  # pgxpool connection wrapper
│   │   └── db_test.go
│   ├── geo/
│   │   ├── ign/                   # IGN parser + ingest
│   │   │   ├── ingest.go
│   │   │   ├── parser.go
│   │   │   └── parser_test.go
│   │   ├── snczi/                 # SNCZI parser + ingest
│   │   │   ├── ingest.go
│   │   │   ├── parser.go
│   │   │   └── parser_test.go
│   │   ├── joins.go               # Spatial joins
│   │   └── joins_test.go
│   ├── health/
│   │   ├── health.go              # /healthz handler
│   │   └── health_test.go
│   ├── planner/
│   │   ├── compiler.go            # Intent → parameterized SQL
│   │   ├── intent.go              # QueryIntent schema + allow-lists
│   │   ├── planner_test.go        # 26 adversarial tests
│   │   └── validator.go           # Injection guard + allow-list validation
│   └── version/
│       ├── version.go
│       └── version_test.go
├── migrations/
│   ├── 000001_init.down.sql
│   ├── 000001_init.up.sql         # PostGIS extension
│   ├── 000002_geo_schema.down.sql
│   ├── 000002_geo_schema.up.sql   # Core tables (basins, provinces, dams, reservoirs, sources)
│   ├── 000003_api_v1.down.sql
│   └── 000003_api_v1.up.sql       # readings, api_keys, metering tables
├── scripts/                       # PowerShell (deprecated — use WSL now)
│   ├── setup.ps1
│   ├── start.ps1
│   └── stop.ps1
├── test/
│   └── fixtures/
│       ├── ign_basins.geojson
│       ├── ign_provinces.geojson
│       ├── ign_reservoirs.geojson
│       └── snczi_dams.geojson
├── web/                           # Frontend (React + Vite + TS + Tailwind CSS v4)
│   ├── index.html
│   ├── vite.config.ts             # Vite + proxy to api:8080 (Docker network)
│   ├── package.json
│   ├── package-lock.json
│   ├── tsconfig.json
│   ├── tsconfig.app.json
│   ├── tsconfig.node.json
│   ├── .gitignore
│   ├── .oxlintrc.json
│   ├── Dockerfile
│   ├── README.md
│   ├── public/
│   │   ├── favicon.svg
│   │   ├── icons.svg
│   │   └── manifest.json          # PWA manifest
│   └── src/
│       ├── App.css
│       ├── App.test.tsx           # 4 frontend tests
│       ├── App.tsx                # Router + QueryClientProvider
│       ├── index.css
│       ├── main.tsx               # Entry point
│       ├── api/
│       │   └── client.ts          # Fetch API client (X-API-Key header)
│       ├── assets/
│       │   ├── hero.png
│       │   ├── react.svg
│       │   └── vite.svg
│       ├── components/
│       │   └── Layout.tsx         # Header + nav + footer (govt-style)
│       ├── hooks/
│       │   └── useQueries.ts      # TanStack Query hooks
│       ├── i18n/
│       │   ├── index.ts           # i18next config (ES/EN)
│       │   └── locales/
│       │       ├── en.json
│       │       └── es.json
│       ├── pages/
│       │   ├── Basins.tsx         # Basin ranking table
│       │   ├── Comparator.tsx     # Multi-reservoir comparison (mock data)
│       │   ├── DataQuality.tsx    # Data quality KPI grid
│       │   ├── Home.tsx           # National KPIs + rankings
│       │   ├── MapPage.tsx        # MapLibre map with random markers
│       │   ├── NotFound.tsx       # 404 page
│       │   ├── ReservoirDetail.tsx # Detail + Recharts line chart
│       │   ├── Reservoirs.tsx     # Paginated table
│       │   ├── Settings.tsx       # Language switch (ES/EN)
│       │   └── Sources.tsx        # Attribution / data sources
│       ├── styles/
│       │   └── global.css         # Government design system (MITECO colors)
│       └── types/
│           └── index.ts           # TypeScript types (APIResponse, Reservoir, etc.)
├── .env.example
├── docker-compose.yml             # PostgreSQL + PostGIS + API + Web + MCP
├── Dockerfile                     # Multi-target Go build (api, mcp, ingest)
├── Dockerfile.migrate             # golang-migrate container (broken — latest requires Go 1.24)
├── Makefile
├── README.md
├── go.mod                         # Go 1.23, chi, pgx
├── go.sum
├── migration_kimi_embalses_project.md   # This file
├── ui-preview.png
└── ui-polished.png
├── .github/
│   └── workflows/
│       └── ci.yml                 # GitHub Actions: go-checks, frontend-checks, e2e-smoke, license-gate, secrets-check
```

---

## Phase-by-Phase Detail

### Phase 0: Foundations (Issues #1–#8) — Merged to `main`

| Issue | Task | What was done |
|-------|------|---------------|
| #1 | Project bootstrap | Go 1.23, `go.mod`, `internal/` layout, `cmd/api` |
| #2 | GitHub CI | `.github/workflows/ci.yml` — go build, test, vet, gofmt, frontend build |
| #3 | Docker Compose | `docker-compose.yml` with PostgreSQL 16 + PostGIS + health checks |
| #4 | Makefile | `make test`, `make build`, `make docker-up`, `make migrate` |
| #5 | Frontend bootstrap | Vite + React + TypeScript + Tailwind CSS |
| #6 | Frontend CI | `npm run test`, `npm run build` in CI pipeline |
| #7 | Environment config | `config.Load()`, `DATABASE_URL`, `.env.example` |
| #8 | Health check | `/healthz` handler for liveness/readiness probes |

**Key decisions:**
- PostGIS from day one (future-proofing spatial queries)
- Vite over Create React App (faster build, better DX)
- `internal/` layout for Go (no external packages import `internal/`)

---

### Phase 1: Core Domain (Issues #9–#16) — Merged to `main`

| Issue | Task | What was done |
|-------|------|---------------|
| #9 | Domain model | `Domain`, `Reservoir`, `Reading`, `Source`, `Basin`, `Province`, `Dam` structs |
| #10 | Migration runner | `migrations/` + `000001_init.up.sql` / `000002_geo_schema.up.sql` |
| #11 | Data interfaces | `ReservoirService`, `ReadingService`, `SourceService` interfaces |
| #12 | In-memory repository | `MemoryRepository` for fast testing without DB |
| #13 | PostgreSQL repository | `PostgresRepository` with parameterized queries |
| #14 | Environment config | `envconfig` for all env vars |
| #15 | SQLC query generation | (explored, not used — parameterized pgx chosen instead) |
| #16 | Transaction support | `BeginTx`/`Commit`/`Rollback` in repository layer |

**Key decisions:**
- `pgx` over `database/sql` (type-safe, better performance, native PostgreSQL)
- `MemoryRepository` enables fast unit tests without Docker
- Interface-based design allows swapping storage backends

---

### Phase 2: Geo Layer (Issues #17–#22) — Merged to `main`

| Issue | Task | What was done |
|-------|------|---------------|
| #17 | XML parser | SNCZI XML parser for reservoir metadata |
| #18 | IGN parser | IGN XML parser for dam coordinates (ETRS89) |
| #19 | PostGIS migration | `migrations/000002_geo_schema.up.sql` with spatial columns |
| #20 | Spatial joins | `ST_Within`, `ST_Intersects` for reservoir-basin matching |
| #21 | ETRS89→4326 | `gobl` package reprojection (transform.go) |
| #22 | Attribution | Source table with `source`, `licence`, `attribution` fields |

**Key decisions:**
- SNCZI data: Spanish National System of Reservoir Information (MITECO)
- IGN data: Spanish National Geographic Institute for dam coordinates
- ETRS89 (EPSG:25830) → WGS84 (EPSG:4326) for web mapping compatibility
- `licence` field: `Ley 37/2007 + RD 1495/2011` (MITECO open data)

---

### Phase 3: REST API v1 (Issues #23–#30) — Merged to `main`

| Issue | Task | What was done |
|-------|------|---------------|
| #23 | chi router | `go-chi/chi/v5` with middleware chain |
| #24 | Reservoir endpoints | `GET /api/v1/reservoirs` (list), `GET /api/v1/reservoirs/{slug}` (detail) |
| #25 | Basins + provinces | `GET /api/v1/basins`, `GET /api/v1/basins/{slug}` |
| #26 | Readings | `GET /api/v1/reservoirs/{slug}/readings` (time-series, date range, paginated) |
| #27 | Rankings | `GET /api/v1/rankings/reservoirs?metric=fullest/emptiest&limit=N` |
| #28 | Comparator | `GET /api/v1/compare?reservoir=slug1&reservoir=slug2&since=…&until=…` |
| #29 | OpenAPI spec | `api/openapi.yaml` documenting all 11 endpoints + schemas |
| #30 | API keys + quotas | `api_keys` table, `metering` table, rate limit (120/min), quota (1000/day) |

**Endpoints:**

```
GET  /healthz                    (no auth)
GET  /readyz                     (no auth)
GET  /api/v1/sources             (API key required)
GET  /api/v1/reservoirs          (API key required)
GET  /api/v1/reservoirs/{slug}   (API key required)
GET  /api/v1/reservoirs/{slug}/readings  (API key required)
GET  /api/v1/basins              (API key required)
GET  /api/v1/basins/{slug}       (API key required)
GET  /api/v1/rankings/reservoirs (API key required)
GET  /api/v1/compare             (API key required)
GET  /api/v1/data-quality        (API key required)
POST /api/v1/query               (API key required) ← Phase 4
```

**Security features:**
- `X-API-Key` header or `api_key` query param
- Per-key per-minute rate limit (in-memory bucket)
- Per-key daily quota (checked against `metering` table)
- Every request logged to `metering` table with timestamp, path, status, key
- Test API key: `test-key-123` (seeded in migration)

**Standardized response:**
```json
{
  "data": {...},
  "meta": {"page": 1, "per_page": 20, "total": 100, "total_pages": 5},
  "error": {"code": "not_found", "message": "Reservoir not found"},
  "lineage": {"source": "MITECO", "licence": "Ley 37/2007 + RD 1495/2011", "attribution": "Fuente: MITECO"}
}
```

**All SQL is parameterized.** No user input is ever interpolated into SQL strings. Every query uses `$N` placeholders.

---

### Phase 4: Safe Query Planner (Issues #31–#36) — Merged to `main`

| Issue | Task | What was done |
|-------|------|---------------|
| #31 | Query Intent schema | `internal/planner/intent.go` — `QueryIntent`, `Filters`, `SortSpec` |
| #32 | Validator | `internal/planner/validator.go` — allow-list validation + SQL injection guard |
| #33 | Plan compiler | `internal/planner/compiler.go` — intent → parameterized SQL |
| #34 | Parameterized execution | `ExecutePlan()` with pgx `$N` parameters only |
| #35 | Adversarial tests | `planner_test.go` — 26 tests including 7 SQL injection payloads |
| #36 | Documentation | `docs/query-intent.md` — full grammar, allow-lists, examples |

**The planner is the ONLY way user-defined queries reach the database.** No LLM is involved. The process is:

```
User JSON → Parse → Validate (allow-lists) → Compile (parameterized SQL)
→ Execute (pgx with $N params) → Return {results + transparent plan}
```

**Allow-lists:**
- Entities: `reservoir`, `basin`, `province`, `community`, `national`
- Metrics: `fill_percent`, `stored_hm3`, `capacity_hm3`, `change_hm3` (reservoir only)
- Aggregations: `latest`, `timeseries`, `ranking`, `compare`, `summary`
- Sort fields: `name`, `fill_percent`, `stored_hm3`, `capacity_hm3`, `change_hm3`, `observed_at`, `basin_name`, `province_name`
- Hard limit: 500 rows max

**Injection guard:** Rejects `;`, `--`, `/*`, `*/`, `DROP`, `DELETE`, `INSERT`, `UPDATE`, `SELECT`, `UNION`, `exec`, `xp_`, `sp_`, quotes, null bytes in any filter value.

**Critical security guarantee:** Even if validation is bypassed, `CompilePlan` uses only `$N` parameter placeholders. There is no string interpolation path from user input to executed SQL. This is proven by the `TestCompilePlan_NoRawSQLFromUserInput` and `TestCompilePlan_BypassValidationStillSafe` tests.

**Endpoint:** `POST /api/v1/query` — returns both results and the compiled plan for transparency.

---

### Phase 5: Frontend MVP (Issues #37–#46) — Current branch `mvp/05-frontend-mvp`

| Issue | Task | What was done |
|-------|------|---------------|
| #37 | App shell | Layout component, header, footer, mobile nav, responsive design |
| #38 | API client | `TanStack Query` + fetch client with `X-API-Key` header |
| #39 | Home + Map | KPI cards, fullest/emptiest rankings, MapLibre map with markers |
| #40 | Reservoirs + Detail | Paginated table, detail page with historical Recharts line chart |
| #41 | Comparator | Multi-select (max 5), metric toggle, comparison chart |
| #42 | Basins | Basin ranking table |
| #43 | Fuentes | Data sources page with attribution, licence, organism |
| #44 | Calidad de datos | Data quality KPI grid |
| #45 | i18n | ES/EN with `i18next` + browser language detection + localStorage persistence |
| #46 | PWA shell | `manifest.json`, theme-color meta, responsive design |

**UI Design System (MITECO-inspired):**

- **Primary color:** Deep blue `#003366` (government blue)
- **KPI cards:** White with colored top accent bars, subtle shadow
- **Progress bars:** Fill percentage visualized as horizontal bars
- **Color coding:**
  - 🔴 < 20% — Critical (red `#dc2626`)
  - 🟠 20-40% — Low (orange `#ea580c`)
  - 🟡 40-60% — Medium (yellow `#ca8a04`)
  - 🟢 60-80% — Good (light green `#16a34a`)
  - 🟢 > 80% — Excellent (dark green `#15803d`)
- **Status badges:** Rounded pill badges with color-coded background
- **Typography:** Inter via Google Fonts (system fallback), proper hierarchy
- **Charts:** Recharts line charts with dual-axis (fill % + volume)

**Pages:**

| Route | Page | Key Features |
|-------|------|--------------|
| `/` | Home | 4 KPI cards, fullest/emptiest rankings with progress bars |
| `/mapa` | Map | MapLibre map with OSM tiles, reservoir markers, popup info |
| `/embalses` | Reservoirs | Table with sortable columns, progress bars, status badges |
| `/embalses/:slug` | Detail | 4 KPI cards, Recharts historical chart (6 months) |
| `/comparar` | Comparator | Multi-select dropdown, metric toggle, color-coded comparison chart |
| `/cuencas` | Basins | Basin ranking table |
| `/fuentes` | Sources | Attribution cards with MITECO licence |
| `/calidad-datos` | Data Quality | 6 KPI cards (total, with readings, dates, provisional/official) |
| `/ajustes` | Settings | Language switch (ES/EN), persisted to localStorage |

**Seeded data (synthetic — not real MITECO data):**

> ⚠️ **Important:** The current seed data is algorithmically generated placeholder data. It is **not** sourced from MITECO, SAIH, or any live hydrological system. The values are produced by a simple random-walk formula with seasonal offsets. This is sufficient for UI development and API testing but must be replaced before any production or research use.

- **18 reservoir names** hardcoded in `cmd/seed/main.go` (20 attempted; 2 failed because their provinces were missing from the seed list)
- **486 weekly readings** spanning 6 months (Dec 2025 → Jun 2026), generated by `math/rand` with seasonal modifiers
- **15 basins, 46 provinces, 18 dams**
- Missing from seed: **Embalse de Bornos** (Cádiz not in province list) and **Embalse de Ebro** (Cantabria not in province list)

**Known frontend placeholder issues:**
- **MapPage.tsx**: Marker coordinates are random (`math.random()`) inside a Spain bounding box — not real dam coordinates
- **Comparator.tsx**: Chart data is hardcoded mock (`mockChartData`) with 2024 dates — not fetched from the API
- **ReservoirDetail.tsx**: Historical chart uses real API data from the synthetic seed, so dates are correct (2025–2026)

---

## GitHub CI — E2E Smoke Test

The CI workflow (`.github/workflows/ci.yml`) includes a new `e2e-smoke` job that simulates the full Ubuntu/WSL installation end-to-end:

1. Installs Go + Node.js on `ubuntu-latest`
2. Spins up a PostgreSQL 16 + PostGIS service container
3. Runs all migrations (`migrate up`)
4. Seeds the database with synthetic data (`go run ./cmd/seed`)
5. Verifies seed counts via `psql`
6. Builds the Go API binary (`go build -o bin/api ./cmd/api`)
7. Starts the API server in the background on `:8080`
8. Builds the frontend (`npm run build`) with `VITE_API_KEY=test-key-123`
9. Starts the Vite preview server in the background on `:4174`
10. Runs smoke tests against:
    - **API endpoints**: health, sources, reservoirs, rankings, basins, data-quality, detail, readings, compare
    - **Frontend**: HTML delivery, proxied API via `/api/v1/sources`

This ensures the entire stack — DB → API → Web — works together on every push/PR.

---

## WSL / Ubuntu Setup Guide (Docker-based — recommended)

### 1. Clone and checkout

```bash
cd ~
git clone https://github.com/arzamas1987/embalses.git
cd embalses
git checkout mvp/05-frontend-mvp
```

### 2. Start the full stack (Docker)

```bash
cd ~/embalses
# Make sure Docker Desktop is running with WSL2 integration

# Start DB, API, and Web containers
docker compose up -d db api web

# Verify DB is healthy
docker compose ps

# Seed data (if not already seeded)
docker run --rm \
  --network embalses_default \
  -v /home/arzamas/git/embalses:/app \
  -w /app \
  -e DATABASE_URL=postgres://postgres:postgres@db:5432/embalses?sslmode=disable \
  golang:1.23-alpine \
  sh -c "go mod download && go run ./cmd/seed"
```

### 3. Access the app

| Service | URL |
|---------|-----|
| **Frontend** | http://localhost:5173 |
| **API** | http://localhost:8080 |
| **Health** | http://localhost:8080/healthz |

### 4. Manual API tests

```bash
curl -H "X-API-Key: test-key-123" http://localhost:8080/api/v1/sources
curl -H "X-API-Key: test-key-123" http://localhost:8080/api/v1/reservoirs
curl -H "X-API-Key: test-key-123" "http://localhost:8080/api/v1/rankings/reservoirs?metric=fullest&limit=5"
```

### 5. One-command start script (native Go + Node, no Docker for API/Web)

Create `~/embalses/start.sh`:

```bash
#!/bin/bash
set -e

export DATABASE_URL="postgres://postgres:postgres@localhost:5432/embalses?sslmode=disable"
export VITE_API_KEY="test-key-123"

echo "=== Starting Embalses ==="

# Start Docker if not running
if ! docker ps >/dev/null 2>&1; then
    echo "Docker not running. Start Docker Desktop first."
    exit 1
fi

# Start DB
echo "Starting PostgreSQL..."
docker compose up -d db
sleep 3

# Seed data (idempotent — safe to run multiple times)
echo "Seeding data..."
go run ./cmd/seed

# Start API in background
echo "Starting API server..."
go run ./cmd/api &
API_PID=$!

# Build and start frontend
echo "Building frontend..."
cd web
npm run build
npm run preview &
FRONT_PID=$!

echo ""
echo "=== Embalses is running ==="
echo "Frontend: http://localhost:4174"
echo "API:      http://localhost:8080"
echo ""
echo "Press Ctrl+C to stop all services"

# Wait for Ctrl+C
trap "echo 'Stopping...'; kill $API_PID $FRONT_PID 2>/dev/null; docker compose down; exit 0" INT
wait
```

Make it executable and run:

```bash
chmod +x ~/embalses/start.sh
~/embalses/start.sh
```

---

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | Yes | — | PostgreSQL connection string |
| `VITE_API_KEY` | Yes | `test-key-123` | API key for frontend requests |
| `PORT` | No | `8080` | API server port |
| `API_KEY_QUOTA_DAILY` | No | `1000` | Default daily quota per key |
| `API_KEY_RATE_LIMIT_PER_MIN` | No | `120` | Default rate limit per key |

---

## Testing

### Go tests

```bash
cd ~/embalses
# Must run sequentially (-p 1) because DB tests share a connection
# and create/drop schemas to avoid conflicts
go test -p 1 -v ./...
```

### Frontend tests

```bash
cd ~/embalses/web
npm run test
```

### Manual API test

```bash
curl -H "X-API-Key: test-key-123" http://localhost:8080/api/v1/sources
curl -H "X-API-Key: test-key-123" http://localhost:8080/api/v1/reservoirs
curl -H "X-API-Key: test-key-123" "http://localhost:8080/api/v1/rankings/reservoirs?metric=fullest&limit=5"
```

### E2E smoke test (CI simulation)

```bash
cd ~/embalses
# The CI e2e-smoke job can be reproduced locally by:
# 1. docker compose up -d db
# 2. migrate -path migrations -database "$DATABASE_URL" up
# 3. go run ./cmd/seed
# 4. go build -o bin/api ./cmd/api && ./bin/api &
# 5. cd web && npm run build && npx vite preview --port 4174 &
# 6. curl smoke tests against :8080 and :4174
```

---

## Key Design Decisions & Why

| Decision | Why |
|----------|-----|
| **PostGIS from day 1** | Spain's data is in ETRS89 (EPSG:25830). PostGIS handles reprojection to WGS84 (4326) for web maps. |
| **No LLM in query planner** | Deterministic allow-lists are auditable and provably safe. No black-box AI generating SQL. |
| **Parameterized queries only** | `$N` placeholders prevent SQL injection by design. No string concatenation. |
| **pgx over database/sql** | Type-safe, better performance, native PostgreSQL features, connection pooling. |
| **MapLibre over Mapbox** | BSD-3 license, no API key required, fully open source. |
| **Recharts over ECharts** | Simpler API, React-native, smaller bundle, MIT license. |
| **Chi over Gin** | Minimal, fast, standard library compatible, middleware composable. |
| **Interface-based repositories** | Swap Memory → Postgres → any backend without changing handlers. |
| **Lineage on every response** | Legal requirement for MITECO data attribution. Every data point shows source. |
| **In-memory rate limiter** | No Redis dependency for MVP. Per-key per-minute buckets. |
| **Test schema isolation** | Each test package creates a unique PostgreSQL schema to avoid cross-package table conflicts. |
| **`-p 1` for Go tests** | Integration tests share one DB. Sequential execution prevents concurrent schema creation failures. |

---

## What to Fine-Tune Next (Phase 6 — Data Ingestion & UI Polish)

### A. Replace synthetic data with real MITECO data (Priority: High)

1. **Build MITECO ingestion pipeline**
   - Download historical XLSX/ZIP from MITECO Boletín Hidrológico
   - Parse weekly PDF bulletins for current data
   - Ingest into `readings` table with proper `source_id`, `is_provisional`, `is_official` flags
   - Target: ~374 peninsular reservoirs with real weekly data

2. **Fix seed script**
   - Add missing provinces: `Cádiz`, `Cantabria`, `Pontevedra`, etc. (check full list)
   - Add real dam coordinates from `test/fixtures/snczi_dams.geojson`
   - Use `latitude`/`longitude` columns in `reservoirs` table (already in schema via `geometry`)

3. **Real-time enrichment (SAIH)**
   - Start with SAIH Ebro (`chebro.es`) — best documented
   - Then SAIH Júcar (`aps.chj.es`) — exposes WMS + CSV
   - Remaining basins: Duero, Tajo, Guadiana, Segura, Guadalquivir, Miño-Sil

### B. UI Polish (Priority: Medium)

1. **Real map coordinates** — Use actual lat/lng from `reservoirs` table (not random)
2. **Real comparator data** — Wire `Comparator.tsx` to call `postQuery()` with selected reservoirs
3. **Actual fill colors on map** — Markers color-coded based on current fill level (already partially done)
4. **More responsive design** — Test on mobile; tighten layout on small screens
5. **Loading states** — Skeleton loaders instead of plain "Loading..." text
6. **Error handling** — Toast notifications for API errors, retry buttons
7. **Chart enhancements** — Tooltips with exact values, date formatting, zoom/pan
8. **Search/filter** — Search bar on reservoirs list page
9. **Date picker** — Custom date ranges for historical views
10. **Export** — Download CSV/JSON from the query endpoint

### C. Infrastructure

11. **Fix `Dockerfile.migrate`** — Current `latest` tag requires Go 1.24, but project uses Go 1.23. Pin to a compatible version or build from a compatible base.
12. **Add LICENSE file** — MIT license is referenced in docs but the file is missing from the repo.
13. **Add `configs/` directory** — For parser configurations, basin mappings, etc.

---

## License

MIT License (file missing from repo — should be added). Data is MITECO open data under `Ley 37/2007 + RD 1495/2011`.

---

## Contact

- **Repository:** https://github.com/arzamas1987/embalses
- **Pull Requests:** Currently open: #5 (Phase 5 — Frontend MVP)
- **Issues:** `docs/issues.md` contains the full 52-issue backlog

---

*Generated by Kimi Work on 2026-06-28.*
*Last updated: Phase 5 complete, MVP running in Docker/WSL with synthetic data. E2E CI smoke test added. Ready for real MITECO data ingestion (Phase 6).*
