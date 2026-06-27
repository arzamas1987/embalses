# Architecture — Spanish Reservoir Data Platform

> Planning document (no code). Anchor date **2026-06-27**. Dependency choices
> follow `licensing.md` (permissive-first). Data choices follow
> `data-sources.md`. The platform is **local-first**: the full stack runs on a
> developer machine via Docker Compose with no cloud dependency except the
> optional Gemini API.

## 1. Goals & constraints

- **Local-first**: `docker compose up` brings up DB + backend + frontend + MCP.
  Ingestion runs against public official sources; the app is fully usable
  offline once data is loaded. The NL assistant (Gemini) is **optional** and
  degrades gracefully.
- **Safety**: the assistant **never executes arbitrary SQL**. NL → a validated,
  whitelisted **query plan** → parameterised queries only.
- **Permissive licensing** end-to-end (see `licensing.md`).
- **Lineage & quality** are first-class, not afterthoughts.
- **Monetisation-ready**: API quotas, MCP access tiers, exports, assistant usage
  metering are designed-in seams, not retrofits.

## 2. High-level diagram

```
                         ┌──────────────────────────────────────────────┐
                         │                Frontend (SPA)                 │
                         │   React + Vite + TS, MapLibre GL, Recharts     │
                         │   PWA · multi-language · "Fuentes" page        │
                         └───────────────┬───────────────┬──────────────┘
                                         │ HTTPS/JSON     │
                                         ▼                ▼
┌──────────────┐   MCP (stdio/HTTP)  ┌───────────────────────────────────┐
│ AI agents /  │ ──────────────────▶ │            Go backend             │
│ MCP clients  │                     │  chi router · REST API (versioned)│
└──────────────┘                     │  Query Planner (safe, whitelisted)│
                                      │  Assistant svc → Gemini (optional)│
        ┌──────────────┐   call       │  Ingestion workers / scheduler    │
        │  Gemini API  │ ◀────────────│  Lineage + Data-quality services  │
        │  (optional)  │   plan only   └───────────────┬───────────────────┘
        └──────────────┘                               │ pgx / sqlc
                                                        ▼
                                      ┌───────────────────────────────────┐
                                      │   PostgreSQL 16 + PostGIS (stock)  │
                                      │  reservoirs · readings · dams ·    │
                                      │  geometries · sources · quality    │
                                      └───────────────────────────────────┘
                                                        ▲
   Official sources ── ingestion ──────────────────────┘
   MITECO XLSX/PDF · SNCZI Shapefile · IGN WFS/WMS · SAIH (Ebro,Júcar) · AEMET
```

## 3. Backend (Go)

- **Language/runtime:** Go (single static binary; great for local-first + ops).
- **Router:** **chi** (MIT) — net/http-native, minimal. (gin/echo also MIT if preferred.)
- **DB access:** **pgx** (MIT) + **sqlc** (MIT) for type-safe, compile-checked
  queries from `.sql` files. No ORM — keeps queries explicit and auditable
  (supports the "never arbitrary SQL" stance).
- **Migrations:** **golang-migrate** (MIT), versioned SQL up/down.
- **Config:** env vars only; `.env.example` committed, real `.env` git-ignored.
- **Layout (planned):**
  ```
  /cmd/api          # REST API server entrypoint
  /cmd/mcp          # MCP server entrypoint
  /cmd/ingest       # ingestion CLI / scheduler
  /internal/domain  # reservoir, dam, reading, source, basin models
  /internal/store   # pgx + sqlc generated queries
  /internal/api     # chi handlers, DTOs, OpenAPI
  /internal/planner # safe query planner (intent → plan → SQL builder)
  /internal/assistant # Gemini client + NL→intent mapping
  /internal/ingest  # per-source adapters (miteco, snczi, ign, saih, aemet)
  /internal/lineage # source registry + record stamping
  /internal/quality # freshness/completeness/anomaly reports
  /migrations        # golang-migrate SQL
  ```

### 3.1 Ingestion
- **Per-source adapters** (one package each) implementing a common interface:
  `Fetch() → Parse() → Normalise() → Upsert(with lineage)`.
- **Idempotent upserts** keyed by `(source, external_id, observed_at)`.
- **Scheduler**: cron-like in-process (or a `cmd/ingest run --source=...` cron in
  Compose). MITECO weekly; IGN/SNCZI on-demand/periodic; SAIH (Ebro/Júcar) more
  frequent in P1.
- **Lineage stamping** on every row: `source_id`, `published_at`, `fetched_at`,
  `is_official` vs `is_derived`.

### 3.2 Safe query planner (core differentiator)
A two-stage, **no-arbitrary-SQL** pipeline:

1. **Intent extraction**: NL question → a constrained **Query Intent** object
   (JSON) — Gemini (or a deterministic parser fallback) maps text onto a fixed
   schema: `metric ∈ {volume, fill_pct, inflow, ...}`, `entity ∈ {reservoir,
   basin, province, ...}`, `filters` (whitelisted fields + operators),
   `time_range`, `aggregation`, `sort`, `limit`.
2. **Plan validation + compilation**: the intent is validated against an
   **allow-list** (known fields, operators, bounded limits). Only then is it
   compiled to a **parameterised** query via sqlc-backed builders. Anything
   outside the schema is rejected. The **plan is returned to the caller** for
   transparency (query-plan transparency feature).

> The LLM **never** emits SQL and never touches the DB directly. Worst case, a
> malformed intent is rejected — it cannot run unsafe queries.

## 4. Database (PostgreSQL + PostGIS)

- **PostgreSQL 16** (permissive licence) + **stock PostGIS** (GPL-2.0, used as a
  service over SQL only — see `licensing.md`; **never modified/redistributed**).
- **Core tables (planned):**
  - `sources` — registry (name, organism, licence, attribution string, url, last_fetched_at).
  - `basins` (demarcaciones) — with `geometry(MultiPolygon, 4326)` from IGN.
  - `reservoirs` — identity, basin FK, capacity, NMN, `geometry(Point/Polygon, 4326)`.
  - `dams` — SNCZI inventory fields, `geometry(Point, 4326)`, reservoir FK.
  - `readings` — time series: reservoir FK, metric, value, `observed_at`,
    `source_id`, `is_official`, `is_provisional`. Indexed on `(reservoir_id, observed_at)`.
  - `data_quality` — per source/dataset freshness, completeness, anomaly flags.
- **Geo**: store in EPSG:4326; reproject IGN/SNCZI ETRS89 on ingest. PostGIS for
  spatial joins (reservoir ↔ basin ↔ province) and map queries.

## 5. Frontend (React/Vite — justified)

**Choice: React + Vite + TypeScript.** Rationale: largest ecosystem, permissive
(MIT), excellent map/chart library support, easy PWA, fast HMR via Vite. (A
justified alternative — SvelteKit (MIT) — is lighter but has a smaller hiring
pool; React/Vite is the lower-risk default for an MVP. Avoid Next.js SSR
complexity for a local-first SPA.)

- **Maps:** **MapLibre GL JS** (BSD-3) with self-hosted or openly-licensed tiles
  (OSM / MapTiler with proper terms) + IGN WMS overlays. **Never Mapbox GL v2+.**
- **Charts:** **Recharts** (MIT) or **ECharts** (Apache-2.0) for multi-year overlays.
- **State/data:** TanStack Query (MIT) for API caching.
- **PWA:** installable, offline shell for already-loaded data.
- **i18n:** Spanish + English from day one (a benchmark gap).
- Mandatory **"Fuentes de datos" (lineage/attribution) page** + map-footer CC-BY
  attribution for IGN data.

## 6. REST API

- Versioned (`/api/v1`), JSON, OpenAPI 3 spec committed.
- Read-only public endpoints (reservoirs, basins, readings, rankings, comparator,
  data-quality), plus the assistant/query-plan endpoint.
- **Monetisation seams:** API keys, per-key **quotas/rate limits**, usage
  metering middleware (counts toward future paid tiers). Free tier + paid tiers
  later — no billing in MVP, but the metering hooks exist.
- Every response includes **source + licence + lineage** fields.

## 7. MCP server

- Separate entrypoint (`/cmd/mcp`), **MCP Go SDK** (MIT).
- Transports: **stdio** (local clients) and **HTTP** (remote).
- **Tools exposed** (read-only, safe):
  - `list_reservoirs`, `get_reservoir`, `get_readings`, `compare_reservoirs`,
    `basin_summary`, `rankings`, `data_quality_report`.
  - `query` — accepts a **structured Query Intent** (same planner as §3.2), runs
    the validated plan, returns results **plus the executed plan**. No raw SQL.
- **Access control / metering** mirrors the REST API (keys, quotas) — future
  "MCP access" monetisation tier.

## 8. Gemini integration (optional)

- Used **only** to turn NL → Query Intent (and to phrase answers). It receives
  the **schema/allow-list and the user's question**, returns a structured intent;
  the backend validates and executes. Gemini never sees DB credentials and never
  emits SQL.
- Key handling: env/secret manager; **no secrets in repo**.
- **Graceful degradation:** if no key / API down, the assistant is disabled but
  the whole platform (API, MCP, UI) keeps working; a deterministic intent parser
  can cover a subset of common questions.

## 9. Local Docker Compose

```yaml
# illustrative only — not committed code
services:
  db:        # postgis/postgis:16-3.4 (stock image)
  migrate:   # runs golang-migrate, then exits
  api:       # Go REST API (depends_on db healthy)
  mcp:       # Go MCP server (http transport)
  ingest:    # one-shot/cron ingestion worker
  web:       # Vite dev server / static build via nginx
```
- Single `.env` (git-ignored) drives all services; `.env.example` documents keys
  (`DATABASE_URL`, `GEMINI_API_KEY` optional, `AEMET_API_KEY` optional).
- Healthchecks + `depends_on` ordering; named volume for Postgres data.

## 10. Testing & CI (GitHub Actions)
- **Backend:** `go test` (unit + integration against a Postgres/PostGIS service
  container), `testify` (MIT). Planner gets dedicated tests asserting unsafe
  intents are rejected and only parameterised SQL is produced.
- **Frontend:** Vitest + component tests; build check.
- **CI gates:** lint, test, build, **licence scan** (allow/deny lists from
  `licensing.md`), and migration up/down check.
- **Ingestion adapters:** tested against recorded fixtures (no live scraping in CI).

## 11. Cross-cutting: lineage & data quality
- **Lineage**: source registry + per-row stamping; surfaced in API/MCP/UI.
- **Data quality**: scheduled reports — freshness (last successful fetch per
  source), completeness (expected vs present reservoirs), anomaly detection
  (impossible jumps), and provisional-vs-validated flags from MITECO. Exposed as
  an endpoint, an MCP tool, and a UI page.

## 12. Key risks & mitigations
| Risk | Mitigation |
|---|---|
| SAIH portals heterogeneous/unstable | Adapter pattern; start with Ebro+Júcar; fixtures in CI; tolerate per-source failure. |
| MITECO file URL rotates | Resolve file link dynamically; alert on ingest failure; keep last-good snapshot. |
| LLM emits unsafe/garbage intent | Strict allow-list validation; reject+log; deterministic fallback parser. |
| Gemini cost/availability | Optional + graceful degradation; cache common intents; meter usage. |
| Licence creep via transitive dep | CI licence-scan gate (deny GPL/AGPL/SSPL/BSL/unlicensed). |
| Attribution non-compliance | Mandatory source registry + UI/API attribution; map-footer CC-BY for IGN. |
