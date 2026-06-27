# Data Sources — Official Spanish Reservoir & Hydrological Data

> Research anchor date: **2026-06-27** (Europe/Madrid). Facts below were verified
> live via web search/fetch on that date. Items not confirmable live are marked
> **UNVERIFIED — confirm before relying**. No URLs or licence names were invented.
>
> **Hard rule:** We ingest only from official upstream sources. We do **NOT**
> scrape or copy `estadoembalses.es` (it is a feature benchmark only — see
> `benchmark.md`). All sources below are public-sector open data reusable under
> the Spanish PSI-reuse regime (Ley 37/2007 + RD 1495/2011) unless noted.

## How to read this document

- **Owner** — the public body that publishes/maintains the data.
- **License/terms** — the legal basis for reuse + any per-source notice.
- **Ingestion difficulty** — Easy / Medium / Hard with the reason.
- **MVP priority** — `P0` (build first), `P1` (next), `P2` (later/optional).

---

## P0 — Core MVP sources

### 1. MITECO — Boletín Hidrológico Semanal / Estado de los embalses
- **Owner:** Ministerio para la Transición Ecológica y el Reto Demográfico (MITECO), Dirección General del Agua. Aggregates data from the Confederaciones Hidrográficas, intracommunity water authorities, AEMET and Red Eléctrica.
- **URLs:**
  - Bulletin home: `https://www.miteco.gob.es/es/agua/temas/evaluacion-de-los-recursos-hidricos/boletin-hidrologico.html`
  - ArcGIS dashboard: `https://miteco.maps.arcgis.com/apps/dashboards/912dfee767264e3884f7aea8eb1e0673`
  - Historical Excel/ZIP: the page states it provides an Excel with capacity & water reserve of every peninsular reservoir > 5 hm³ **since 1988**. The historically distributed file was `BD-Embalses_1988-2022.zip`. **Exact current static file URL: UNVERIFIED — resolve off the live page during ingestion build.**
- **Data type:** Weekly reservoir storage (hm³), total capacity, % fill, hydro energy, main-river flows, weekly precipitation. Coverage: peninsular reservoirs > 5 hm³ (~374 reported 2024–2026).
- **Format:** Weekly **PDF** bulletin (published Tuesdays) + **XLSX/ZIP** historical series since 1988 + ArcGIS web dashboard (FeatureServer REST/JSON may exist but is undocumented — **UNVERIFIED**).
- **Update frequency:** Weekly (Tuesdays). Bulletin warns data are *provisional, subject to revision and validation*.
- **License/terms:** PSI reuse — Ley 37/2007 + RD 1495/2011, general "Aviso legal" at `www.datos.gob.es/avisolegal`. Commercial reuse allowed. No explicit per-page CC label observed (**UNVERIFIED**).
- **Attribution:** `Fuente: Ministerio para la Transición Ecológica y el Reto Demográfico (MITECO)`. Exact mandated string not found verbatim.
- **Fields:** embalse name, demarcación/cuenca, capacity (hm³), volume (hm³), % fill, weekly variation, energy, comparisons vs prior week / prior years / 5- & 10-yr means, date.
- **Ingestion difficulty:** **Easy–Medium.** Historical XLSX trivial once the URL is pinned; weekly PDF needs parsing; ArcGIS REST path is fragile.
- **MVP priority:** **P0** — the canonical national time series; backbone of the dataset.

### 2. IGN / CNIG — Geographic reference (hydrography, boundaries, basemaps)
- **Owner:** Instituto Geográfico Nacional (IGN) + Centro Nacional de Información Geográfica (CNIG).
- **URLs:**
  - Centro de Descargas: `https://centrodedescargas.cnig.es/CentroDescargas/home`
  - Hydrography (IGR-HY) metadata: `https://www.idee.es/csw-codsi-idee/srv/api/records/spaignHIDROGRAFIA_IGR`
  - PNOA WMS: `https://www.ign.es/wms-inspire/pnoa-ma` · WMTS: `https://www.ign.es/wmts/pnoa-ma`
- **Data type:** Hydrography (rivers, reservoir polygons, dam points, demarcations), administrative boundaries (+ INE codes), ortho-imagery (PNOA), toponymy (Nomenclátor).
- **Format:** Vector (Shapefile/GeoPackage), **WMS / WMTS / WFS (INSPIRE)**, COG GeoTIFF, CSV for high-value datasets.
- **Update frequency:** Hydrography (IGR-HY) **annual**; ortho ~biannual/triannual.
- **License/terms:** **CC-BY 4.0** equivalent, per **Orden FOM/2807/2015** (+ high-value-data regime, RD-ley 24/2021). Free incl. **commercial** reuse and redistribution. Registration only for bulk download > 20 files/session. Some reference datasets carry no licence at all.
- **Attribution (verified exact form):** `CC-BY 4.0 scne.es` / `© Instituto Geográfico Nacional`; derived works prefix `Obra derivada de …`. Product/attribution table at `scne.es/productos`.
- **Fields:** river-network topology, reservoir polygons, dam points, admin geometries + INE codes, toponyms.
- **Ingestion difficulty:** **Easy.** Cleanest licence, standard OGC formats, stable endpoints.
- **MVP priority:** **P0** — provides geometry (reservoir polygons, basins, admin boundaries) and the basemap.

### 3. SNCZI — Inventario de Presas y Embalses de España (dam inventory)
- **Owner:** MITECO — Subdirección General de Dominio Público Hidráulico e Infraestructuras (Seguridad de Presas). SNCZI implements EU Floods Directive 2007/60.
- **URLs:**
  - Download page: `https://www.miteco.gob.es/es/cartografia-y-sig/ide/descargas/agua/inventario-presas-embalses.html`
  - SNCZI viewer: `https://sig.miteco.gob.es/snczi/`
  - Dam point Shapefile (ETRS89): `https://www.miteco.gob.es/es/cartografia-y-sig/ide/descargas/egis_presa_geoetrs89_tcm30-175857.zip` (**confirm `tcm30-…` token still live — UNVERIFIED, may rotate**)
  - datos.gob.es entry: `https://datos.gob.es/es/catalogo/e0dat0002-inventario-de-presas`
- **Data type:** Dam inventory (reference, not time series): dam fiche + reservoir characteristics + safety-document status.
- **Format:** **Shapefile/ZIP** + **WMS** (`https://wms.mapama.gob.es/sig/agua/...`) + viewer.
- **Update frequency:** Periodic/irregular (effectively static reference).
- **License/terms:** PSI reuse (Ley 37/2007 + RD 1495/2011). No explicit CC string on the page (**UNVERIFIED**).
- **Attribution:** `SNCZI – Inventario de Presas y Embalses, MITECO`.
- **Fields:** dam name, owner, risk category (A/B/C), norms/emergency-plan status, river, municipality, basin, province, UTM30 coords, basin area, mean annual contribution & precipitation, design flood, reservoir surface & capacity at NMN, NMN elevation, dam type/height/crest, spillway & outlet capacities, photos, plans.
- **Ingestion difficulty:** **Easy–Medium.** One-shot Shapefile load → reproject → join to reservoir IDs.
- **MVP priority:** **P0** — dam metadata, geolocation, canonical capacity/NMN per dam.

---

## P1 — Next sources

### 4. SAIH — Sistemas Automáticos de Información Hidrológica (Confederaciones)
- **Owner:** Each Confederación Hidrográfica (CH) under MITECO. Heterogeneous portals.
- **URLs (verified subset):**
  - **SAIH Ebro (CHE):** `https://www.chebro.es/` (renewed 2024) + legacy `http://www.saihebro.com/`. Real-time 15-min; registered users get "Datos a la carta" (daily/monthly/annual/15-min, CSV export). **Best-documented open download.**
  - **SAIH Júcar (CHJ):** `https://aps.chj.es/down/html/descargas.html` — exposes Embalses/Presas data + cartography + **WMS**.
  - **SAIH Guadalquivir (CHG):** `https://www.chguadalquivir.es/saih/`
  - **SAIH Guadiana:** `https://www.saihguadiana.com` · visor `https://www.chguadiana.es/visorCHG/`
  - **SAIH Segura (CHS):** `https://www.chsegura.es/es/cuenca/redes-de-control/saih/`
  - **Tajo / Duero / Miño-Sil / Cantábrico / Andalucía Mediterránea (Junta de Andalucía):** portals exist; exact SAIH download URLs **UNVERIFIED — confirm per CH**.
- **Data type:** Real-time reservoir level/volume/% fill, inflow/outflow caudales, gauging stations (aforos), pluviometry, snow, drought indicators.
- **Format:** Web portals + **PDF "partes"**; CSV/Excel via interactive "datos a la carta" (often free registration); CHJ exposes WMS. **No uniform public REST/JSON across CHs.**
- **Update frequency:** Real-time (15-min/hourly public; sub-hourly for registered).
- **License/terms:** PSI reuse in principle; **per-CH "Aviso legal" varies — UNVERIFIED per CH.**
- **Attribution:** Cite the specific CH + `SAIH <cuenca>`.
- **Fields:** station code/name, variable (nivel, caudal, volumen, precipitación), 15-min timestamp, value, basin/subsystem.
- **Ingestion difficulty:** **Medium–Hard.** 9–10 heterogeneous portals; some need registration; many are web apps not clean APIs. **Ebro + Júcar most tractable.**
- **MVP priority:** **P1** — start with Ebro + Júcar for "live" enrichment; expand later.

### 5. datos.gob.es — National open-data catalogue
- **Owner:** Government of Spain (Iniciativa Aporta). Federates datasets; does not host most raw reservoir data.
- **URLs:** `https://datos.gob.es/` · reuse notice `https://www.datos.gob.es/avisolegal` · practical case `https://datos.gob.es/es/conocimiento/analisis-del-estado-y-evolucion-de-los-embalses-de-agua-nacionales`
- **Data type:** DCAT metadata + distributions that re-expose MITECO/CH/IGN sources; app registry.
- **Format:** DCAT metadata; distributions link out to CSV/XLSX/Shapefile/WMS at origin.
- **Update frequency:** Depends on source dataset.
- **License/terms (precise):** General "Aviso legal" under **RD 1495/2011 art. 8.1** developing **Ley 37/2007** — permits commercial + non-commercial reuse (copy, distribute, modify, adapt, extract, reorder, combine); non-exclusive worldwide cession of IP rights.
- **Attribution:** Cite source + date per the aviso legal.
- **Ingestion difficulty:** **Easy** as a discovery layer; **not a primary feed** — always ingest from the upstream origin.
- **MVP priority:** **P1** — provenance & licence governance reference, not an ingestion endpoint.

---

## P2 — Later / optional

### 6. AEMET — OpenData API (precipitation / meteorology)
- **Owner:** Agencia Estatal de Meteorología (AEMET), MITECO.
- **URLs:** API portal `https://opendata.aemet.es/` · RISP catalogue `https://www.aemet.es/es/datos_abiertos/catalogo`
- **Data type:** Precipitation, temperature, wind, station observations (last 24h), daily/monthly/annual climatology (~1,000 stations), forecasts, climate scenarios.
- **Format:** **REST/JSON** (two-step: request returns a `datos` URL); some products XML/CSV/Excel/PDF.
- **Update frequency:** Real-time/hourly observation; daily for many products.
- **License/terms:** Reuse under RD 1495/2011 + Ley 19/2013 + Plan RISP; AEMET **retains IP**; binding "Aviso Legal"; **requires a free API key**. Some products are cost-bearing; **logos/legal notices must be retained**. **Verify per-product terms — partly UNVERIFIED.**
- **Attribution:** `© AEMET` / `Fuente: Agencia Estatal de Meteorología (AEMET)`.
- **Fields:** `idema` (station), lat/lon/alt, `fint`, `prec`, `ta`, `vv`, humidity, pressure.
- **Ingestion difficulty:** **Medium.** API key + two-step fetch + rate limits + station→basin mapping.
- **MVP priority:** **P2** — precipitation context; defer past MVP.

### 7. Remaining SAIH basins (Duero, Tajo, Guadiana, Segura, Guadalquivir, Miño-Sil, Cantábrico, Andalucía)
- **MVP priority:** **P2** — staged rollout after Ebro + Júcar are proven. Each is a separate, partly undocumented portal.

### EU INSPIRE Directive (cross-cutting enabler)
- Directive **2007/2/EC**, transposed by **Ley 14/2010 (LISIGE)**. Hydrography is an INSPIRE high-value theme; IGN/CNIG IGR-HY is registered as such. Open Data Directive **2019/1024** + Reg. (EU) 2023/138 make hydrography a "high-value dataset" → must be free, machine-readable, API/bulk. **Practical impact:** prefer INSPIRE WFS (IGN/CNIG) for reservoir/river/basin geometry — most licence-clean and standards-based.

---

## Summary — MVP priority matrix

| Priority | Source | Data | Format | Freq. | Licence basis | Difficulty |
|---|---|---|---|---|---|---|
| **P0** | MITECO Boletín Hidrológico | Reservoir storage time series (1988→) | XLSX/ZIP + PDF | Weekly | Ley 37/2007 + RD 1495/2011 | Easy–Medium |
| **P0** | IGN / CNIG | Hydrography, boundaries, basemap | Shapefile/GPKG, WMS/WFS | Annual | **CC-BY 4.0** (FOM/2807/2015) | Easy |
| **P0** | SNCZI Inventario de Presas | Dam/reservoir reference + geolocation | Shapefile, WMS | Periodic | Ley 37/2007 + RD 1495/2011 | Easy–Medium |
| **P1** | SAIH (Ebro + Júcar) | Real-time level/inflow/outflow | Web/CSV, WMS | Real-time | PSI (per-CH, varies) | Medium–Hard |
| **P1** | datos.gob.es | Provenance/licence governance | DCAT | Varies | RD 1495/2011 avisolegal | Easy |
| **P2** | AEMET OpenData | Precipitation/meteorology | REST/JSON | Hourly/daily | AEMET Aviso Legal (key) | Medium |
| **P2** | Remaining SAIH basins | Full real-time coverage | Heterogeneous | Real-time | PSI (per-CH) | Hard |
| n/a | estadoembalses.es | **Benchmark only — DO NOT INGEST** | — | — | Third-party | — |

## Open items to confirm during ingestion build
1. Exact current static URL/filename of the MITECO historical reservoir XLSX/ZIP.
2. Whether the MITECO ArcGIS dashboard exposes a stable FeatureServer REST/JSON endpoint.
3. Current validity of the SNCZI dam Shapefile token (`egis_presa_geoetrs89_tcm30-175857.zip`).
4. Per-CH SAIH download URLs and explicit "Aviso legal" strings (esp. Duero, Tajo, Guadiana, Miño-Sil, Cantábrico, Andalucía).
5. Verbatim AEMET and MITECO attribution strings + per-product AEMET cost/logo conditions.
