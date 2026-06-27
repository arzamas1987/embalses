# Roadmap — Implementation Phases

> Planning document (no code). Anchor date **2026-06-27**. Phases are
> incremental and each ends in a demoable, tested state. Branch names are
> suggestions (`feat/…`, `chore/…`). Acceptance criteria are binary gates — a
> phase is "done" only when all its criteria pass in CI. Issue numbers refer to
> `docs/issues.md`.

## Phase 0 — Foundations & scaffolding
**Goal:** a runnable, tested empty skeleton with CI and licence gating.
**Branches / tasks:** `chore/repo-scaffold`, `chore/ci`, `chore/compose`, `chore/licence-gate` (Issues #1–#8).
**Acceptance criteria:**
- `docker compose up` starts Postgres+PostGIS, API, web, MCP stubs; all healthchecks green.
- `go test ./...` and frontend test command run green in CI (GitHub Actions).
- Licence-scan job enforces the allow/deny lists from `licensing.md` and fails on a planted GPL dep.
- `.env.example` committed; no secrets in repo; `.gitignore` excludes `.env`.
- `golang-migrate` runs an initial empty migration up/down cleanly.

## Phase 1 — Data model & MITECO ingestion (P0 data)
**Goal:** load the canonical national reservoir time series + schema + lineage.
**Branches / tasks:** `feat/schema`, `feat/source-registry`, `feat/ingest-miteco`, `feat/lineage` (Issues #9–#16).
**Acceptance criteria:**
- Schema migrations create `sources`, `basins`, `reservoirs`, `dams`, `readings`, `data_quality`.
- `sources` registry seeded with MITECO/IGN/SNCZI entries incl. licence + attribution strings.
- MITECO historical XLSX ingested idempotently; re-running does not duplicate rows.
- Every `readings` row carries `source_id`, `published_at`, `fetched_at`, `is_official`.
- Integration test ingests a fixture XLSX and asserts row counts + lineage fields.

## Phase 2 — Geo layer: SNCZI + IGN (P0 data)
**Goal:** dam inventory + reservoir/basin geometries, reprojected and joined.
**Branches / tasks:** `feat/ingest-snczi`, `feat/ingest-ign`, `feat/geo-joins` (Issues #17–#22).
**Acceptance criteria:**
- SNCZI dam Shapefile ingested → `dams` with ETRS89→4326 reprojection.
- IGN hydrography/boundaries ingested → `basins` + reservoir polygons (4326).
- PostGIS spatial join links reservoirs↔basins↔provinces; test asserts a known reservoir maps to the correct basin/province.
- IGN CC-BY attribution string stored in `sources` and exposed.

## Phase 3 — REST API v1
**Goal:** read-only public API over loaded data, with lineage in responses.
**Branches / tasks:** `feat/api-core`, `feat/api-rankings`, `feat/api-openapi`, `feat/api-keys-metering` (Issues #23–#30).
**Acceptance criteria:**
- Endpoints: reservoirs (list/detail), basins, readings (time series), rankings, comparator, data-quality — all `/api/v1`.
- Every response includes `source`, `licence`, and lineage fields.
- OpenAPI 3 spec committed and matches handlers (contract test).
- API-key middleware + per-key quota/rate-limit enforced; metering counter increments (no billing yet).
- Handler + integration tests pass; pagination + error shapes covered.

## Phase 4 — Safe query planner
**Goal:** structured Query Intent → validated plan → parameterised SQL. No arbitrary SQL.
**Branches / tasks:** `feat/query-intent-schema`, `feat/planner-validate`, `feat/planner-compile` (Issues #31–#36).
**Acceptance criteria:**
- Query Intent JSON schema defined (metric/entity/filters/time_range/aggregation/sort/limit) with allow-lists.
- Validator rejects unknown fields/operators and out-of-bounds limits; tests assert rejection.
- Compiler emits only **parameterised** queries (sqlc-backed); a test asserts no string-interpolated SQL path exists.
- `/api/v1/query` accepts an intent, returns results **plus the executed plan**.
- Adversarial tests (injection-style intents) are all rejected.

## Phase 5 — Frontend MVP
**Goal:** usable SPA — map, reservoir pages, charts, comparator, sources page.
**Branches / tasks:** `feat/web-shell`, `feat/web-map`, `feat/web-reservoir`, `feat/web-charts`, `feat/web-i18n`, `feat/web-sources` (Issues #37–#46).
**Acceptance criteria:**
- React+Vite SPA with MapLibre map colour-coded by fill %, click-through to reservoir detail.
- Reservoir detail: current + historical chart (multi-year overlay), comparison vs last year/average.
- Comparator view (multiple reservoirs). National KPIs on home.
- i18n ES/EN; **"Fuentes de datos" page** + map-footer CC-BY attribution for IGN.
- PWA installable; Vitest component tests + build pass in CI.

## Phase 6 — MCP server
**Goal:** read-only MCP tools incl. the safe `query` tool.
**Branches / tasks:** `feat/mcp-core`, `feat/mcp-tools`, `feat/mcp-query`, `feat/mcp-metering` (Issues #47–#52).
**Acceptance criteria:**
- MCP server (Go SDK) runs over stdio + HTTP; lists tools.
- Tools: list/get reservoirs, get readings, compare, basin summary, rankings, data-quality, and `query` (intent in, results+plan out).
- `query` uses the same planner; no raw SQL path; adversarial test passes.
- Access keys + metering mirror the REST API.
- Tool-call integration tests pass against a seeded DB.

## Phase 7 — Gemini NL assistant (optional layer)
**Goal:** NL question → Query Intent → answer + transparent plan; graceful degradation.
**Branches / tasks:** `feat/assistant-gemini`, `feat/assistant-fallback`, `feat/assistant-ui` (Issues #53–#58).
**Acceptance criteria:**
- Assistant maps NL (ES/EN) to a valid Query Intent; Gemini never emits SQL or sees DB creds.
- With no `GEMINI_API_KEY`/API down, assistant disables cleanly and the rest of the platform works (graceful-degradation test).
- Deterministic fallback parser answers a defined subset of common questions.
- UI shows the answer **and** the executed query plan (transparency).
- Gemini key read only from env/secret manager; no secrets in repo (CI check).

## Phase 8 — SAIH real-time (P1 data) + data-quality reports
**Goal:** real-time enrichment (Ebro + Júcar) and first-class quality reporting.
**Branches / tasks:** `feat/ingest-saih-ebro`, `feat/ingest-saih-jucar`, `feat/data-quality` (Issues #59–#64).
**Acceptance criteria:**
- SAIH Ebro + Júcar adapters ingest real-time readings idempotently with lineage; per-source failure is isolated.
- Data-quality service computes freshness, completeness, anomaly flags, provisional-vs-validated; exposed via API + MCP + UI page.
- CI tests adapters against recorded fixtures (no live scraping in CI).

## Phase 9 — Hardening & monetisation seams
**Goal:** production-readiness without enabling billing.
**Branches / tasks:** `feat/exports`, `feat/quota-tiers`, `chore/observability`, `chore/security-review` (Issues #65–#72).
**Acceptance criteria:**
- CSV/JSON export endpoints (quota-aware) with embedded attribution/lineage.
- Quota tiers configurable per API/MCP key; usage metering reportable (billing-ready, billing-off).
- Structured logging + basic metrics; rate-limit + abuse protections tested.
- Security review checklist passed (secrets, input validation, planner allow-lists, dependency licence scan, attribution compliance).

## Phase 10 — Optional / later
- AEMET precipitation enrichment (P2); remaining SAIH basins (P2); alerts/thresholds; embeddable charts; deployment packaging.

## Dependency order (summary)
`P0 → P1 → P2 → P3 → P4 → P5 → P6` (P5 & P6 can partly parallelise after P3) `→ P7 (needs P4) → P8 → P9 → P10`.
