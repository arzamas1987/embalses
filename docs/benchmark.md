# Feature Benchmark — vs `estadoembalses.es`

> Research anchor date: **2026-06-27**. The reference site was visited **only**
> to observe public-facing features. **No data, text, or design assets were
> copied.** This is a competitive feature inventory, not a data source — we
> ingest the same *official upstreams* independently (see `data-sources.md`).
> Items not confirmed on inspected pages are marked **UNVERIFIED**.

## 1. What the benchmark is
A third-party aggregator/visualisation site run by an individual (a telecom
engineer / ex dam inspector). It monitors ~**449 reservoirs across 16 basins**,
holds **600,000+ records since 2006**, refreshing roughly every 15 minutes.

**Crucially, every data source it uses is official public-sector open data we
can independently obtain** (MITECO historical volume; SAIH real-time sensors;
SNCZI dam fiches; plus geographic enrichment). The site itself states its
% fill, basin aggregations, variations and rankings are **its own derived
calculations, "no son cifras oficiales."** It publishes a transparency
"chain of custody" (SAIH sensor → Confederación → site → user).

**Implication:** we compete on **engineering and UX**, not on data exclusivity.
Nothing in its feature set depends on data we cannot get ourselves.

## 2. Feature inventory

**Legend:** **[COMMON]** = any site over the same public data could offer it;
**[DIFF]** = a potential differentiator (harder to do well / not universal).

| # | Observed feature | Class | Notes |
|---|---|---|---|
| 1 | National KPIs: total stored volume (hm³), national avg fill %, # reservoirs, total capacity | **[COMMON]** | Standard headline for this data class. |
| 2 | Per-reservoir detail pages (volume, level, % fill, inflow/outflow, rainfall) | **[COMMON]** | Baseline expectation. |
| 3 | Per-confederation / per-cuenca breakdown (16 basins) | **[COMMON]** | Inherent in the data. |
| 4 | Per-province / per-comunidad-autónoma aggregation | **[COMMON]** → mild **[DIFF]** | Geographic re-aggregation beyond native basins is a modest value-add. |
| 5 | Interactive map colour-coded by fill (green/amber/red), click-through | **[COMMON]** baseline, **[DIFF]** in execution | Maps are common; a fast, mobile, real-time-coloured map is where quality shows. |
| 6 | Historical charts with **interannual comparison (up to 8 years)** | **[DIFF]** | Deep history (since 2006) + multi-year overlay. |
| 7 | **Comparator** — up to 10 reservoirs side-by-side (overlay chart + table) | **[DIFF]** | Genuine differentiator; not universal. |
| 8 | Rankings — fullest / emptiest, biggest weekly rise / fall | **[COMMON]**, mild **[DIFF]** | Easy to compute. |
| 9 | Weekly variation vs ~7 days ago | **[COMMON]** | Standard derived metric. |
| 10 | Comparison vs same week last year / multi-year average | **[COMMON]** | Expected. |
| 11 | **Real-time (SAIH) cadence ~5–15 min** vs weekly MITECO | **[DIFF]** | Real-time freshness vs weekly-bulletin-only sites. |
| 12 | **Embeddable charts via iframe** ("Datos embebibles") | **[DIFF]** | Distribution/SEO lever; not common. |
| 13 | Installable as an app (**PWA**) | **[DIFF]** (modest) | Mobile-first UX. |
| 14 | **Dam technical context** (type, year, owner, height, use) from SNCZI | **[DIFF]** | Enriching hydrology with infrastructure metadata. |
| 15 | Search / find a reservoir | **[COMMON]** | Expected; **UNVERIFIED** as a dedicated UX in inspected pages. |
| 16 | Methodology + data-source transparency ("chain of custody") page | **[DIFF]** (trust) | Strong credibility/SEO play. |
| 17 | FAQ / explanatory content (SAIH vs MITECO) | **[COMMON]** | Content marketing. |
| 18 | Spanish-language UI only | **[COMMON]** | **No language switcher observed → multi-language is open for us.** |
| 19 | Alerts / notifications / thresholds | — | **UNVERIFIED — not observed.** Likely a differentiation gap we can own. |
| 20 | Ads / monetisation | — | **UNVERIFIED — none observed;** appears a free passion project. |
| 21 | Public API / CSV export / open data | — | **UNVERIFIED — not advertised.** A documented API + MCP + NL assistant is our strongest wedge. |

## 3. Common public-data functionality (we must match — table stakes)
- National KPIs; per-reservoir and per-basin pages.
- Fill-% colour map with click-through.
- Historical charts; comparison vs last year and vs multi-year average.
- Weekly variation; rankings; search.

## 4. Our differentiators (explicit)

These are the wedges this project is built around. The first four appear
**absent** from the benchmark (UNVERIFIED) and none require data we cannot get.

1. **REST API** — a documented, versioned, rate-limited public API over the same
   official data. *(Benchmark exposes iframe embeds, not an open API.)*
2. **MCP server** — exposes reservoir data + safe query tools to AI agents/clients
   over the Model Context Protocol. Novel in this domain.
3. **Natural-language query assistant (Gemini)** — ask in plain Spanish/English
   ("¿qué embalses del Júcar están por debajo del 20 %?") and get answers.
4. **Query-plan transparency** — the assistant compiles NL → a **structured,
   validated query plan** (never arbitrary SQL) and **shows the plan** it ran:
   which fields, filters, sources, and aggregations. Auditable and safe.
5. **Source lineage** — every figure carries its origin (MITECO/SAIH/SNCZI/IGN),
   publication date, fetch timestamp, and an **official-vs-derived** flag.
6. **Data-quality reports** — freshness, completeness, gaps, anomaly flags, and
   provisional-vs-validated status surfaced as a first-class report.

**Secondary differentiators to match/beat the benchmark:** real-time SAIH
cadence, multi-year overlays, multi-reservoir comparator, embeddable charts,
PWA, dam-infrastructure metadata, transparency/methodology pages, **plus
multi-language UI and alerts/thresholds** (open gaps).

## 5. Strategic takeaway
- **Match** the common functionality cleanly and quickly (it's expected).
- **Win** on the API/MCP/assistant/transparency/lineage/quality layer — the
  things that turn a public-data viewer into a **platform** with a monetisation
  path (API quotas, MCP access, exports, assistant usage).
- **Never** copy the benchmark; replicate value by ingesting its official
  upstreams and out-engineering the presentation and the developer surface.

### Verification caveats
**UNVERIFIED (not observed on inspected pages):** benchmark alerts/notifications,
a dedicated search UX, any public API/CSV export, and ad/monetisation presence —
treated here as differentiation opportunities, to be re-confirmed before any
public competitive claim.
