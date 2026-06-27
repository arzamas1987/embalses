# Licensing Policy — Dependencies & Data Attribution

> Research anchor date: **2026-06-27**. Licences below were verified live where
> possible. Items not confirmable live are marked **UNVERIFIED — confirm before
> adopting**. This is engineering guidance, not legal advice; have counsel review
> before commercial launch.

## 1. Goals

The product is a hosted **SaaS / REST API / MCP server / NL assistant** with a
future commercial model (API quotas, MCP access, exports, assistant usage).
That deployment model makes **network-copyleft licences uniquely dangerous**, so
the dependency policy is deliberately conservative and **permissive-first**.

## 2. Dependency licence policy

### 2.1 SAFE — allowed by default
Permissive licences; we may use, link, bundle and ship freely, keeping notices.

| Licence | SPDX | Notes |
|---|---|---|
| MIT | `MIT` | Keep copyright/licence notice. |
| Apache-2.0 | `Apache-2.0` | Permissive **+ explicit patent grant**. Keep `NOTICE`. |
| BSD-2-Clause / BSD-3-Clause | `BSD-2-Clause` / `BSD-3-Clause` | BSD-3 adds no-endorsement clause. |
| ISC | `ISC` | Equivalent to MIT/BSD-2. |
| PostgreSQL Licence | `PostgreSQL` | MIT/BSD-style. |

### 2.2 ACCEPTABLE WITH CARE — middle case
| Licence | SPDX | Condition of use |
|---|---|---|
| MPL-2.0 | `MPL-2.0` | **File-level (weak) copyleft.** Safe to *use/link* inside a proprietary or differently-licensed product. **Catch:** if you modify an MPL-2.0 *file*, you must publish that file's source. **Track which files are MPL; never bury our own logic inside them.** |

### 2.3 RISKY / AVOID — do not depend on without explicit legal clearance
| Licence / type | SPDX / kind | Why it's a problem for us |
|---|---|---|
| GPL-2.0 / GPL-3.0 | `GPL-2.0-only` / `GPL-3.0-only` | Strong copyleft: linking/including in a **distributed** binary forces the whole combined work under GPL. Do not bundle GPL libraries into our Go binary or frontend bundle. |
| **AGPL-3.0** | `AGPL-3.0-only` | **Network copyleft (§13).** The single most dangerous licence for us — see §2.4. |
| **SSPL** | `SSPL-1.0` (non-OSI) | "Service-side" copyleft aimed at SaaS providers; would force open-sourcing the whole service stack. |
| Commons Clause (rider) | source-available | Adds "may not sell" on top of another licence — kills commercial use. |
| BSL / BUSL-1.1 / "source-available" | e.g. `BUSL-1.1` | Time-delayed / usage-restricted; not OSS at adoption; often forbids competing hosted use. |
| No LICENSE / unclear | "all rights reserved" | No licence = no rights granted. Treat as do-not-use until clarified. |
| Commercial-only / proprietary | — | Only with a paid licence we explicitly hold (e.g. Mapbox GL JS v2+). |

### 2.4 Why AGPL / SSPL specifically matter (network-use copyleft)
Plain GPL's trigger is **distribution of a binary**. Running modified GPL
software on a server *for users over a network* is not "distribution", so the
source-sharing obligation isn't triggered (the "SaaS loophole").

- **AGPL-3.0 §13 (Remote Network Interaction)** closes that loophole: if users
  interact with a modified AGPL program **over a network**, you must offer them
  the **complete corresponding source** of your modified version. Our entire
  product — REST API, MCP server, NL assistant — *is* that deployment model.
  An AGPL library inside the server could force publishing our whole server's
  source to every API/MCP user. **→ AGPL is the #1 avoid.**
- **SSPL** goes further: to offer the software "as a service" you must
  open-source effectively the **entire service-management stack**. Non-OSI,
  non-free. **→ Avoid.**

**Rule of thumb:** No AGPL, SSPL, BSL, Commons Clause, GPL, or unlicensed code
as a dependency we link, bundle, or run inside our own service code. Permissive
preferred; MPL-2.0 acceptable with file-level tracking.

## 3. Verified licences of the planned stack

### 3.1 Go backend — all permissive (SAFE)
| Dependency | Licence | SPDX | Verdict |
|---|---|---|---|
| Go stdlib / toolchain | BSD-3-Clause | `BSD-3-Clause` | SAFE |
| chi (`go-chi/chi`) | MIT | `MIT` | SAFE — recommended (net/http-native, zero deps) |
| gin (`gin-gonic/gin`) | MIT | `MIT` | SAFE |
| echo (`labstack/echo`) | MIT | `MIT` | SAFE |
| pgx (`jackc/pgx`) | MIT | `MIT` | SAFE — Postgres driver |
| sqlc (`sqlc-dev/sqlc`) | MIT | `MIT` | SAFE — build-time codegen |
| golang-migrate | MIT | `MIT` | SAFE — confirm imported DB-driver sub-packages |
| testify (`stretchr/testify`) | MIT | `MIT` | SAFE — test-only |

### 3.2 Database layer
| Component | Licence | Implication |
|---|---|---|
| PostgreSQL | `PostgreSQL` (permissive) | SAFE — run, connect, ship freely. |
| **PostGIS** | **GPL-2.0** | ⚠️ GPL, **but low practical risk** — see below. |

**PostGIS GPL-2.0 — practical implication (per official PostGIS GPL FAQ):**
- PostGIS runs as an extension inside the **separate PostgreSQL service**; we
  talk to it **over SQL / a network connection** and do **not** link its C
  library into our binary.
- The FAQ is explicit: software *"can use a PostgreSQL/PostGIS database as much
  as it wants and be under any license you like."* Loading data and running
  queries does **not** make our application code GPL.
- **Only trigger:** modifying the PostGIS source **and distributing that
  modified PostGIS** — then you share *that* modified source, never your app.
- **Action:** use **stock, unmodified PostGIS** as a DB service; never
  fork-and-redistribute it. No impact on our code's licence. ✅

### 3.3 Frontend — charts & maps
| Dependency | Licence | SPDX | Verdict |
|---|---|---|---|
| React | MIT | `MIT` | SAFE |
| Vite | MIT | `MIT` | SAFE |
| Recharts | MIT | `MIT` | SAFE |
| Chart.js | MIT | `MIT` | SAFE |
| Apache ECharts | Apache-2.0 | `Apache-2.0` | SAFE (keep NOTICE) |
| **MapLibre GL JS** | BSD-3-Clause (some sources label MIT — both permissive) | `BSD-3-Clause` | ✅ **SAFE — recommended map lib** |
| **Leaflet** | BSD-2-Clause | `BSD-2-Clause` | ✅ SAFE (lightweight raster/marker) |
| **Mapbox GL JS v2+** | **Proprietary / commercial** | proprietary | 🚫 **AVOID** unless we buy a Mapbox plan |

**⚠️ Map licence trap (verified):** In **December 2020 Mapbox relicensed Mapbox
GL JS from BSD to proprietary at v2.0** (account + token required, billed per
map load). The community forked the last BSD release into **MapLibre GL JS**
(BSD-3, Linux Foundation governance). **Leaflet stays BSD-2.**
**Decision: use MapLibre GL JS or Leaflet — never Mapbox GL JS v2+.**
Also: the **map library** licence is separate from **basemap/tile** terms —
self-host or use openly-licensed tiles (OSM, or a provider with clear terms) and
attribute them separately.

### 3.4 MCP SDKs
| SDK | Licence | Verdict |
|---|---|---|
| MCP Go SDK (`modelcontextprotocol/go-sdk`) | MIT | SAFE |
| MCP TypeScript SDK (`@modelcontextprotocol/sdk`) | MIT (v1.x); v2 branch dual MIT/Apache-2.0 | SAFE — v1.x recommended for production |

### 3.5 Gemini API — a paid service, not an OSS dependency
- The **Google Gemini API is a paid Google Cloud / Google AI service** — no OSS
  licence to vet. It is governed by **Google's commercial API Terms of Service /
  Generative AI usage policies**. Treat as a **third-party service dependency**.
- **Review current Gemini API commercial/data-use terms before production —
  UNVERIFIED here.** In particular how query text / any data are processed.
- **API-key handling:** keys are **secrets** → §5.
- Our **safe query planner** (Gemini emits only structured, validated query
  intents; never arbitrary SQL) is a security control as well as a UX choice.
- Gemini is a runtime/operational cost; design **graceful degradation** — the
  core platform must fully work without the assistant.

## 4. Data-source attribution policy (Spanish PSI reuse)

**Legal framework (verified):** **Ley 37/2007** (PSI reuse; amended by Ley
18/2015) + **RD 1495/2011** establish a default authorisation to reuse
public-sector documents **including commercially**, under the general legal
notice at `www.datos.gob.es/avisolegal`. Reuse includes copy, distribute,
modify, adapt, extract, reorder, combine — provided you **cite source + date,
don't present it as official activity, and don't distort the data**.

**Per-source attribution:**
- **IGN / CNIG** (Orden FOM/2807/2015 → **CC-BY 4.0**). Required string formats
  (verified from the IGN licence PDF):
  - General: `«<producto> <fecha> CC-BY 4.0 <atribución>»` → e.g. `BTN25 2014-2015 CC-BY 4.0 ign.es`
  - Abbreviated: `CC-BY 4.0 scne.es 2010`
  - Derived: prefix `Obra derivada de …` → `Obra derivada de PNOA 2010-2013 CC-BY scne.es`
  - Must be **visible with the data** (map footer); for modified datasets, embed
    in metadata fields **abstract, lineage, AccessConstraints**. Values from
    `scne.es/productos`.
- **AEMET:** AEMET **retains IP**; comply with the per-product legal notice;
  **keep the AEMET logo where present**; some products are cost-bearing. Cite
  `© AEMET`. **Confirm per-product terms — partly UNVERIFIED.**
- **MITECO / SAIH / SNCZI / datos.gob.es:** general `datos.gob.es/avisolegal`
  regime → cite source + date, don't misrepresent or distort.

**Recommended platform approach (lineage as a first-class feature):**
1. **Machine-readable source registry** (`sources` table/config): official name,
   organism, licence (SPDX or named: `CC-BY-4.0`, `datos.gob.es-avisolegal`,
   `AEMET-legal-notice`), licence URL, required attribution string, retrieval
   URL, `last_fetched_at`.
2. **Stamp lineage on every record/dataset:** source id, original publication
   date, fetch timestamp, and a **`derived`/`official` flag** (our % fill,
   aggregations and rankings are our calculations, not official MITECO figures).
3. **Surface attribution in the UI:** a "Fuentes de datos" credits page; IGN
   CC-BY attribution at the map footer; `Fuente: SAIH / MITECO (datos.gob.es)`
   on charts.
4. **Embed attribution in API/MCP/export outputs:** include source + licence
   fields in responses; for IGN-derived geodata include the CC-BY expression in
   metadata.
5. **Keep an explicit "official vs derived" disclaimer** (legal + trust
   requirement under Ley 37/2007).

## 5. Secrets policy
- **No secrets in the repository.** Gemini API key, DB credentials, AEMET API
  key → environment variables / a secret manager only.
- Ensure `.gitignore` excludes `.env` and key files; never commit a populated
  `.env`. Provide a committed `.env.example` with placeholders only.
- Rotate keys; scope them minimally.

## 6. Risks to avoid (checklist)
- 🚫 No **AGPL-3.0** dependency anywhere (server, MCP, API, frontend) — §13 network copyleft.
- 🚫 No **SSPL / BSL / Commons Clause / source-available / commercial-only** runtime dependency without legal clearance.
- 🚫 No **GPL-2.0/3.0** library bundled/linked into the shipped binary or frontend bundle.
- 🚫 Never **fork and redistribute modified PostGIS** (stock PostGIS over SQL is fine).
- 🚫 No **Mapbox GL JS v2+** without a paid plan — use **MapLibre (BSD-3)** or **Leaflet (BSD-2)**.
- 🚫 Don't confuse **map-library licence** with **tile/basemap terms** — vet tiles separately.
- 🚫 No dependency with **no LICENSE / unclear licence** — do-not-use until confirmed.
- 🚫 No **secrets** in the repo; use env/secret manager; rotate.
- 🚫 Never present **derived metrics as official** MITECO/CH figures.
- 🚫 Never ship **IGN-derived geodata without the CC-BY 4.0 attribution**.
- 🚫 Never strip **AEMET logos/legal notices**; confirm per-product AEMET terms.
- 🚫 Never **scrape/copy estadoembalses.es** data, text, or design — benchmark only.
- ⚠️ **MPL-2.0 acceptable but tracked** (file-level copyleft).

## 7. CI enforcement (operational governance)
- Add a **licence-scanning gate in CI**:
  - Go: `go-licenses` (or an SBOM tool) with an allow-list.
  - npm: SBOM + licence checker with an allow-list.
- **Allow-list:** MIT, Apache-2.0, BSD-2-Clause, BSD-3-Clause, ISC, MPL-2.0, PostgreSQL.
- **Deny-list:** GPL-*, AGPL-*, SSPL, BUSL, Commons-Clause, unlicensed.
- A risky transitive dependency must **fail the build** before it ships.

### Verification caveats
- **UNVERIFIED — confirm before relying:** current Gemini API commercial/data-use
  terms; per-product AEMET reuse conditions (some cost-bearing); exact licences of
  specific transitive sub-packages (pin versions + scan in CI).
