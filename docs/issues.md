# GitHub Issue Plan

> Planning document (no code; nothing created on GitHub). Anchor date
> **2026-06-27**. Each issue is **small, independently buildable, and testable**.
> Issues are grouped by roadmap phase (`docs/roadmap.md`). Suggested labels:
> `phase-N`, `backend`, `frontend`, `data`, `mcp`, `infra`, `security`,
> `licensing`. Every issue lists its **Acceptance criteria** (the binary gate).

## Phase 0 — Foundations

**#1 — Repo scaffold & module layout** `infra`
Create Go module + directory layout (`/cmd`, `/internal`, `/migrations`) and a Vite+React+TS app under `/web`.
*Accept:* `go build ./...` and `web` build succeed locally; README documents layout.

**#2 — `.env.example` + secrets hygiene** `security`
Add `.env.example` (placeholders only: `DATABASE_URL`, `GEMINI_API_KEY`, `AEMET_API_KEY`), ensure `.gitignore` excludes `.env`/keys.
*Accept:* no secret values in repo; CI check greps for accidental secrets and passes.

**#3 — Docker Compose: Postgres+PostGIS** `infra`
Add `db` service (stock `postgis/postgis:16-*`) with healthcheck + named volume.
*Accept:* `docker compose up db` becomes healthy; `psql` shows PostGIS extension available.

**#4 — Compose: API/web/mcp/ingest stubs** `infra`
Add stub services that start and expose a healthcheck endpoint.
*Accept:* `docker compose up` → all services healthy; `/healthz` returns 200 on api/mcp.

**#5 — golang-migrate wiring + empty initial migration** `backend` `data`
Integrate golang-migrate; add `0001_init` up/down.
*Accept:* migrate up then down runs clean in CI against the Compose DB.

**#6 — GitHub Actions: build + test** `infra`
CI runs `go test ./...` and frontend tests with a Postgres/PostGIS service container.
*Accept:* CI green on an empty test; status check required on PRs.

**#7 — CI licence-scan gate** `licensing` `infra`
Add `go-licenses` (Go) + npm SBOM/licence checker with allow/deny lists from `docs/licensing.md`.
*Accept:* job fails when a planted GPL/AGPL dep is added; passes on the clean tree.

**#8 — Lint & formatting gate** `infra`
Add `gofmt`/`go vet`/golangci-lint + ESLint/Prettier in CI.
*Accept:* CI fails on a formatting violation; passes clean.

## Phase 1 — Data model & MITECO

**#9 — Core schema migration** `data` `backend`
Migration for `basins`, `reservoirs`, `dams`, `readings`, `data_quality` (geometry cols 4326).
*Accept:* migrate up/down clean; tables + PostGIS geometry columns exist (test query).

**#10 — `sources` registry table + seed** `data`
Table for source name/organism/licence/attribution/url/last_fetched_at; seed MITECO/IGN/SNCZI.
*Accept:* seed migration inserts rows with correct CC-BY/PSI licence + attribution strings; test asserts presence.

**#11 — Lineage helper + columns** `backend`
Add `source_id`, `published_at`, `fetched_at`, `is_official`, `is_provisional` to `readings`; helper to stamp them.
*Accept:* unit test stamps a record and reads back all lineage fields.

**#12 — sqlc setup + first typed queries** `backend`
Configure sqlc; generate typed accessors for reservoirs/readings.
*Accept:* `sqlc generate` is reproducible in CI; generated code compiles and a query test passes.

**#13 — MITECO XLSX fetcher** `data`
Resolve + download the historical reservoir XLSX/ZIP; store raw snapshot.
*Accept:* fetcher saves a file; unit test uses a fixture and verifies parse-ready output (handles URL-token rotation gracefully).

**#14 — MITECO parser → normalised readings** `data`
Parse XLSX rows into reservoir + reading records.
*Accept:* fixture parse yields expected reservoir count + sample values; malformed rows are skipped with logs.

**#15 — Idempotent MITECO upsert** `data` `backend`
Upsert keyed on `(source, external_id, observed_at)`.
*Accept:* integration test ingests fixture twice → no duplicate rows; lineage fields populated.

**#16 — Ingestion CLI (`cmd/ingest`)** `backend` `infra`
`ingest run --source=miteco` entrypoint, usable from Compose.
*Accept:* CLI runs against the Compose DB and reports inserted/updated counts.

## Phase 2 — Geo (SNCZI + IGN)

**#17 — SNCZI Shapefile fetcher** `data`
Download dam inventory Shapefile (handle token rotation).
*Accept:* fetch saves zip; test on a small fixture extracts features.

**#18 — SNCZI parser + reprojection ETRS89→4326** `data`
Parse dam features, reproject, map fields to `dams`.
*Accept:* fixture yields dams with valid 4326 geometry + key attributes; test asserts a known dam's coords within tolerance.

**#19 — IGN hydrography/boundaries fetcher (WFS/download)** `data`
Fetch basins + reservoir polygons + admin boundaries.
*Accept:* fetch returns features; test parses a fixture into basins/polygons.

**#20 — IGN parser + CC-BY attribution capture** `data` `licensing`
Load geometries; record the exact CC-BY 4.0 attribution string in `sources`.
*Accept:* basins/reservoir polygons loaded; attribution string stored and retrievable.

**#21 — Spatial joins reservoir↔basin↔province** `backend` `data`
PostGIS join populating relations.
*Accept:* test asserts a known reservoir resolves to the correct basin + province.

**#22 — Geo ingestion CLI sources** `backend`
Extend `cmd/ingest` with `--source=snczi|ign`.
*Accept:* both run idempotently and report counts.

## Phase 3 — REST API v1

**#23 — chi router + `/api/v1` skeleton + healthz** `backend`
*Accept:* server boots; `/healthz` 200; 404 shape defined; handler test passes.

**#24 — Reservoirs endpoints (list + detail)** `backend`
*Accept:* list paginates; detail returns reservoir + latest reading + lineage; tests cover both + not-found.

**#25 — Basins + provinces endpoints** `backend`
*Accept:* returns basins/provinces with aggregates; test asserts known aggregate.

**#26 — Readings (time-series) endpoint** `backend`
*Accept:* returns series with time-range filter; pagination + bounds tested.

**#27 — Rankings endpoint** `backend`
*Accept:* fullest/emptiest + biggest weekly rise/fall; test on seeded data.

**#28 — Comparator endpoint** `backend`
*Accept:* accepts N reservoir ids → aligned series; cap enforced; tested.

**#29 — OpenAPI 3 spec + contract test** `backend`
*Accept:* committed spec; contract test fails if a handler diverges from the spec.

**#30 — API keys + quota/rate-limit + metering** `backend` `security`
*Accept:* unauthenticated→limited tier; per-key quota enforced (429 on exceed); metering counter increments; tested.

## Phase 4 — Safe query planner

**#31 — Query Intent JSON schema** `backend`
*Accept:* schema defines metric/entity/filters/time_range/aggregation/sort/limit with allow-lists; schema validation test passes.

**#32 — Intent validator (allow-list)** `backend` `security`
*Accept:* rejects unknown fields/operators + out-of-bounds limits; table-driven tests cover accept/reject.

**#33 — Plan compiler → parameterised SQL** `backend` `security`
*Accept:* compiles valid intents to parameterised queries only; test asserts no string-interpolated SQL path; results correct on seeded data.

**#34 — `/api/v1/query` endpoint (intent in, results+plan out)** `backend`
*Accept:* returns results **and** the executed plan; tested.

**#35 — Adversarial planner tests** `security`
*Accept:* injection-style/oversized/unknown intents all rejected; suite green.

**#36 — Planner docs (intent grammar)** `backend`
*Accept:* `docs/query-intent.md` documents allowed fields/operators; matches the schema (lint test).

## Phase 5 — Frontend MVP

**#37 — App shell + routing + API client** `frontend`
*Accept:* SPA boots; TanStack Query client; routes for home/reservoir/compare/sources; build + smoke test pass.

**#38 — National KPIs home** `frontend`
*Accept:* shows total volume, avg fill %, # reservoirs, capacity from API; component test with mocked API.

**#39 — MapLibre map colour-coded by fill** `frontend`
*Accept:* renders reservoirs coloured by fill %; click→reservoir route; uses MapLibre (not Mapbox); test asserts marker→route.

**#40 — Reservoir detail + historical chart** `frontend`
*Accept:* detail page with multi-year overlay chart (Recharts/ECharts); comparison vs last year/average; component test.

**#41 — Comparator view** `frontend`
*Accept:* select multiple reservoirs → overlaid chart + table; cap enforced; test.

**#42 — i18n ES/EN** `frontend`
*Accept:* language switcher toggles ES/EN; default ES; test asserts string swap.

**#43 — "Fuentes de datos" page + attribution** `frontend` `licensing`
*Accept:* lists every source + licence + required attribution; IGN CC-BY at map footer; test asserts attribution present.

**#44 — Data-quality UI page (placeholder→wired)** `frontend`
*Accept:* renders freshness/completeness from API; test with mocked data.

**#45 — PWA manifest + offline shell** `frontend`
*Accept:* installable; offline shell loads cached data; Lighthouse/PWA check in CI or documented manual check.

**#46 — Frontend test+build CI integration** `frontend` `infra`
*Accept:* Vitest + build run in CI as a required check.

## Phase 6 — MCP server

**#47 — MCP server core (Go SDK, stdio+HTTP)** `mcp`
*Accept:* server starts on both transports; lists tools; smoke test connects and lists.

**#48 — Read tools: reservoirs/basins/readings** `mcp`
*Accept:* tools return data matching the API; integration test against seeded DB.

**#49 — Tools: compare/rankings/data-quality** `mcp`
*Accept:* each returns expected shape; tested.

**#50 — MCP `query` tool (planner-backed)** `mcp` `security`
*Accept:* accepts Query Intent, runs validated plan, returns results+plan; no raw SQL; adversarial test passes.

**#51 — MCP access keys + metering** `mcp` `security`
*Accept:* key required; quota enforced; metering increments; tested.

**#52 — MCP usage docs** `mcp`
*Accept:* `docs/mcp.md` documents tools + connection; example client call verified.

## Phase 7 — Gemini assistant (optional)

**#53 — Gemini client (env key only)** `backend` `security`
*Accept:* client reads key from env/secret manager only; CI asserts no secret in repo; unit test with mocked transport.

**#54 — NL → Query Intent mapping** `backend`
*Accept:* sample ES/EN questions map to valid intents (mocked Gemini); invalid output rejected by validator.

**#55 — Deterministic fallback parser** `backend`
*Accept:* a defined subset of common questions answered without Gemini; tested.

**#56 — Graceful degradation** `backend`
*Accept:* with no key/API down, assistant disabled cleanly; API/MCP/UI still work; test asserts platform health.

**#57 — Assistant UI (answer + plan)** `frontend`
*Accept:* chat-style input shows answer **and** executed query plan; component test with mocked endpoint.

**#58 — Assistant usage metering** `backend`
*Accept:* per-key assistant call metering; reportable; tested.

## Phase 8 — SAIH real-time + data quality

**#59 — SAIH Ebro adapter** `data`
*Accept:* ingests Ebro real-time readings idempotently with lineage; fixture test; failure isolated.

**#60 — SAIH Júcar adapter** `data`
*Accept:* same as #59 for Júcar (WMS/data endpoint); fixture test.

**#61 — Per-source failure isolation** `backend` `data`
*Accept:* one failing source does not abort others; test simulates a failing adapter.

**#62 — Data-quality service (freshness/completeness/anomaly)** `backend`
*Accept:* computes metrics + provisional-vs-validated; unit tests on seeded data.

**#63 — Data-quality API + MCP exposure** `backend` `mcp`
*Accept:* endpoint + MCP tool return quality report; tested.

**#64 — Ingestion scheduler in Compose** `infra`
*Accept:* scheduled ingestion runs (cron in `ingest`); documented cadence; smoke test.

## Phase 9 — Hardening & monetisation seams

**#65 — CSV/JSON export endpoints (quota-aware)** `backend`
*Accept:* exports include embedded attribution/lineage; quota-counted; tested.

**#66 — Quota tiers config** `backend`
*Accept:* tiers configurable per key; enforcement tested; billing disabled.

**#67 — Structured logging + metrics** `infra`
*Accept:* structured logs + basic metrics endpoint; documented; smoke test.

**#68 — Rate-limit/abuse protections** `security`
*Accept:* limits + basic abuse guards tested under load fixture.

**#69 — Security review checklist** `security`
*Accept:* checklist (secrets, input validation, planner allow-lists, licence scan, attribution) completed + linked in repo.

**#70 — Attribution compliance audit** `licensing`
*Accept:* automated check that API/export/UI surfaces required attribution per source; test passes.

**#71 — Backup/restore for DB volume** `infra`
*Accept:* documented + scripted dump/restore; round-trip test.

**#72 — Release packaging (compose profiles / prod build)** `infra`
*Accept:* prod build profile produces static frontend + binaries; documented run; smoke test.

## Phase 10 — Optional / later (parking lot)
- **#73** AEMET precipitation adapter (P2) · **#74** remaining SAIH basins (P2) ·
  **#75** alerts/thresholds · **#76** embeddable charts (iframe) ·
  **#77** deployment/IaC.

---

### Notes
- Issues are numbered for reference only; create them in any order respecting
  the phase dependencies in `docs/roadmap.md`.
- Keep each PR scoped to one issue; require CI (tests + licence scan + lint) green to merge.
- No issue requires scraping `estadoembalses.es`; all data work targets official sources only.
