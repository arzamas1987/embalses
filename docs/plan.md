# Implementation Planning — Spanish Reservoir Data Platform (research blueprint)

> Status: planning/research only. No code is written. The existing repository
> files (`README.md`, `.gitignore`, `.git`) are NOT modified. All deliverables
> are created under `docs/`. Nothing is committed or pushed.
>
> Anchor date for all freshness/verification claims: **2026-06-27** (Europe/Madrid).

## Stages

### Stage 1 — Research (parallel)
- **Worker R1 — Data sources**: official Spanish hydrological/reservoir data
  (MITECO, SAIH confederaciones, SNCZI/Inventario de Presas y Embalses, IGN,
  AEMET, datos.gob.es). Capture URL, owner, data type, format, update
  frequency, license/terms, attribution, fields, ingestion difficulty.
- **Worker R2 — Benchmark + licensing**: feature inventory of estadoembalses.es
  (benchmark only, NO scraping/copying), plus open-source dependency licence
  policy and data-source attribution policy.
- **Verification V1 — Domains** (done by orchestrator with live tools): brand
  candidates and availability across `.es / .com / .dev / .app`, SEO and
  trademark/confusion notes.

### Stage 2 — Writing (orchestrator authored from validated research)
Produce: data-sources.md, licensing.md, benchmark.md, architecture.md,
domains.md, roadmap.md, and the GitHub issue plan.

## Hard rules carried into every stage
- Do NOT scrape or copy Estado Embalses; benchmark only.
- Prioritise official sources; research licences/attribution carefully.
- Prefer permissive deps (MIT/Apache-2.0/BSD/ISC); flag GPL/AGPL/SSPL/Commons
  Clause/unclear/commercial-only as risky.
- No secrets. Do not claim domain availability unless actually checked.
- Safe query planner — never arbitrary SQL.
