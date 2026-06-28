# Embalses — Migration Guide: From Windows to Ubuntu WSL

> **Date:** 2026-06-28  
> **Branch to checkout:** `mvp/05-frontend-mvp`  
> **Status:** Phases 0–5 complete. Ready for UI fine-tuning in WSL.

---

## What is Embalses?

**Embalses** is a Spanish reservoir data platform — an open-source alternative to the official MITECO (Ministerio para la Transición Ecológica) water portal. It provides:

- **Real-time reservoir levels** across all Spanish hydrographic basins
- **Historical time-series analysis** with 6-month rolling data
- **Interactive maps** (MapLibre GL) showing reservoir locations
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
│  └── 18 seeded reservoirs with 468 weekly readings           │
└──────────────────────────────────────────────────────────────┘
```

---

## Repository Layout

```
embalses/
├── api/
│   └── openapi.yaml               # OpenAPI 3 spec (Issues #29, #31)
├── cmd/
│   ├── api/
│   │   └── main.go                # API server entry point
│   ├── migrate/
│   │   └── main.go                # Custom migration runner
│   ├── parser/
│   │   └── main.go                # Batch data ingestion (Phase 2)
│   └── seed/
│       └── main.go                # 6-month seed data (20 reservoirs)
├── configs/
│   ├── parser_config.json         # SNCZI + IGN parser configs
│   └── snczi_basins.csv           # Basin mapping overrides
├── docs/
│   ├── issues.md                  # All 52 project issues
│   └── query-intent.md            # Safe Query Planner grammar
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
│   │   └── config.go              # DATABASE_URL from env
│   ├── db/
│   │   └── db.go                  # pgxpool connection wrapper
│   ├── geo/
│   │   ├── parser.go              # SNCZI / IGN XML parsers
│   │   ├── parser_test.go         # 10 parser tests
│   │   └── transform.go           # ETRS89 → 4326 reprojection
│   ├── health/
│   │   └── health.go              # /healthz handler
│   ├── parser/
│   │   └── main.go                # (old) ingestion entry
│   └── planner/
│       ├── intent.go              # QueryIntent schema + allow-lists
│       ├── validator.go           # Injection guard + allow-list validation
│       ├── compiler.go            # Intent → parameterized SQL
│       └── planner_test.go        # 26 adversarial tests
├── migrations/
│   ├── 000001_init.up.sql         # Core tables (reservoirs, basins, etc.)
│   ├── 000002_geo_schema.up.sql   # PostGIS + spatial data
│   └── 000003_api_v1.up.sql     # readings, api_keys, metering tables
├── scripts/                       # PowerShell (deprecated — use WSL now)
│   ├── setup.ps1
│   ├── start.ps1
│   └── stop.ps1
├── web/                           # Frontend (React + Vite + TS)
│   ├── index.html
│   ├── vite.config.ts             # Vite + proxy to localhost:8080
│   ├── public/
│   │   ├── favicon.ico
│   │   └── manifest.json          # PWA manifest
│   ├── src/
│   │   ├── App.tsx                # Router + QueryClientProvider
│   │   ├── App.test.tsx           # 4 frontend tests
│   │   ├── main.tsx               # Entry point
│   │   ├── api/
│   │   │   └── client.ts          # Fetch API client (X-API-Key header)
│   │   ├── components/
│   │   │   └── Layout.tsx         # Header + nav + footer (govt-style)
│   │   ├── hooks/
│   │   │   └── useQueries.ts      # TanStack Query hooks
│   │   ├── i18n/
│   │   │   ├── index.ts           # i18next config (ES/EN)
│   │   │   └── locales/
│   │   │       ├── es.json        # Spanish translations
│   │   │       └── en.json        # English translations
│   │   ├── pages/
│   │   │   ├── Basins.tsx         # Basin ranking table
│   │   │   ├── Comparator.tsx     # Multi-reservoir comparison
│   │   │   ├── DataQuality.tsx   # Data quality KPI grid
│   │   │   ├── Home.tsx           # National KPIs + rankings
│   │   │   ├── MapPage.tsx        # MapLibre map with markers
│   │   │   ├── NotFound.tsx       # 404 page
│   │   │   ├── ReservoirDetail.tsx # Detail + Recharts line chart
│   │   │   ├── Reservoirs.tsx     # Paginated table
│   │   │   ├── Settings.tsx       # Language switch (ES/EN)
│   │   │   └── Sources.tsx        # Attribution / data sources
│   │   ├── styles/
│   │   │   └── global.css         # Government design system (MITECO colors)
│   │   └── types/
│   │       └── index.ts           # TypeScript types (APIResponse, Reservoir, etc.)
│   ├── package.json               # Dependencies
│   ├── tailwind.config.js         # Tailwind CSS config
│   ├── tsconfig.app.json          # TS compiler options
│   └── tsconfig.json
├── .env.example
├── docker-compose.yml             # PostgreSQL + PostGIS
├── Dockerfile
├── Makefile
├── go.mod                         # Go 1.23, chi, pgx, gobl
├── go.sum
├── LICENSE                        # MIT
├── README.md
└── .github/workflows/ci.yml       # GitHub Actions CI
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
| #10 | Migration runner | `cmd/migrate` + `migrations/000001_init.up.sql` |
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
| #21 | ETRS89→4326 | `transform.go` reprojection using `gobl` package |
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
  - 🟢 60-80% — Good (light green `#65a30d`)
  - 🟢 > 80% — Excellent (dark green `#059669`)
- **Status badges:** Rounded pill badges with color-coded background
- **Typography:** Inter via Google Fonts (system fallback), proper hierarchy
- **Charts:** Recharts line charts with dual-axis (fill % + volume)

**Pages:**

| Route | Page | Key Features |
|-------|------|--------------|
| `/` | Home | 4 KPI cards, fullest/emptiest rankings with progress bars |
| `/mapa` | Map | MapLibre map with OSM tiles, reservoir markers, popup info |
| `/embalses` | Reservoirs | Table with sortable columns, progress bars, status badges |
| `/embalses/:slug` | Detail | 5 KPI cards, Recharts historical chart (6 months) |
| `/comparar` | Comparator | Multi-select dropdown, metric toggle, color-coded comparison chart |
| `/cuencas` | Basins | Basin ranking table |
| `/fuentes` | Sources | Attribution cards with MITECO licence |
| `/calidad-datos` | Data Quality | 6 KPI cards (total, with readings, dates, provisional/official) |
| `/ajustes` | Settings | Language switch (ES/EN), persisted to localStorage |

**Seeded data:**
- 18 real Spanish reservoirs (Mequinenza, Sau, La Serena, Alcántara, etc.)
- 468 weekly readings spanning 6 months (Dec 2025 → Jun 2026)
- Seasonal patterns: winter high → spring peak → summer low → autumn recovery
- 15 basins, 46 provinces, 20 dams

---

## WSL / Ubuntu Setup Guide

### 1. Clone the repo in WSL

```bash
# In your WSL terminal
cd ~
git clone https://github.com/arzamas1987/embalses.git
cd embalses

# Checkout the correct branch
git checkout mvp/05-frontend-mvp
```

### 2. Install dependencies

```bash
# Go 1.23+
# Install via apt or download from https://go.dev/dl/
sudo apt update
sudo apt install -y golang-go

# Or use the official installer:
# wget https://go.dev/dl/go1.23.6.linux-amd64.tar.gz
# sudo rm -rf /usr/local/go
# sudo tar -C /usr/local -xzf go1.23.6.linux-amd64.tar.gz
# echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
# source ~/.bashrc

# Node.js 22+
curl -fsSL https://deb.nodesource.com/setup_22.x | sudo -E bash -
sudo apt install -y nodejs

# Verify
node -v  # should show v22.x
npm -v   # should show 10.x
go version  # should show go1.23.x

# Docker (optional but recommended for DB)
# Install Docker Desktop with WSL2 integration:
# https://docs.docker.com/desktop/wsl/
# Then enable WSL integration in Docker Desktop settings
```

### 3. Install Go dependencies

```bash
cd ~/embalses
go mod download
```

### 4. Install frontend dependencies

```bash
cd ~/embalses/web
npm install
```

### 5. Start PostgreSQL (Docker)

```bash
cd ~/embalses
# Make sure Docker Desktop is running with WSL integration
docker compose up -d db

# Verify
sudo apt install -y postgresql-client  # if psql not available
psql -h localhost -U postgres -d embalses -c "SELECT 1;"
# Password: postgres
```

### 6. Run migrations

```bash
cd ~/embalses
# Install migrate tool
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
export PATH=$PATH:$HOME/go/bin

# Run migrations
migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/embalses?sslmode=disable" up
```

### 7. Seed data

```bash
cd ~/embalses
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/embalses?sslmode=disable"
go run ./cmd/seed
# Expected output: 18 reservoirs, 468 readings, 15 basins, 46 provinces
```

### 8. Start the API server

```bash
cd ~/embalses
export DATABASE_URL="postgres://postgres:postgres@localhost:5432/embalses?sslmode=disable"
go run ./cmd/api
# Should print: API server listening on :8080
```

**Keep this terminal running.** The API server runs in the foreground.

### 9. Start the frontend (in a new terminal)

Open a **new WSL terminal** and run:

```bash
cd ~/embalses/web
export VITE_API_KEY="test-key-123"
npm run build
npm run preview
# Should print: http://localhost:4174
```

**Keep this terminal running too.**

### 10. Access the app

Open your browser:
- **Frontend:** http://localhost:4174
- **API docs:** http://localhost:8080/api/v1/ (try `/healthz`)
- **Test API:** http://localhost:8080/api/v1/sources with header `X-API-Key: test-key-123`

### 11. One-command start script (for convenience)

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
curl -H "X-API-Key: test-key-123" http://localhost:8080/api/v1/rankings/reservoirs?metric=fullest&limit=5
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

## What to Fine-Tune Next (UI)

The user is now working on `mvp/05-frontend-mvp` in WSL. The next steps are **UI polish**:

1. **Real map coordinates** — Currently using random lat/lng. Use actual coordinates from the `reservoirs` table (`latitude`, `longitude` columns). The `migrations/000002_geo_schema.up.sql` already has these columns.

2. **Real comparator data** — Currently using mock data. Wire `Comparator.tsx` to call `postQuery()` with the selected reservoirs and date range.

3. **Actual fill colors on map** — Markers should be color-coded based on current fill level (red/orange/yellow/green).

4. **More responsive design** — Test on mobile. The current layout works but could be tighter on small screens.

5. **Loading states** — Better skeleton loaders instead of plain "Loading..." text.

6. **Error handling** — Toast notifications for API errors, retry buttons.

7. **Chart enhancements** — Tooltips with exact values, date formatting, zoom/pan on charts.

8. **Search/filter** — Add a search bar on the reservoirs list page.

9. **Date picker** — Allow users to select custom date ranges for historical views.

10. **Export** — Download CSV/JSON from the query endpoint.

---

## License

MIT License. Data is MITECO open data under `Ley 37/2007 + RD 1495/2011`.

---

## Contact

- **Repository:** https://github.com/arzamas1987/embalses
- **Pull Requests:** Currently open: #5 (Phase 5 — Frontend MVP)
- **Issues:** `docs/issues.md` contains the full 52-issue backlog

---

*Generated by Kimi Work on 2026-06-28.*
*Last updated: Phase 5 complete, ready for WSL migration and UI fine-tuning.*
