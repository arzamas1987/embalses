# Query Intent Grammar

> **Security-critical document.** The Query Intent system is the **only** way
> user-defined queries reach the database. Arbitrary SQL is never accepted,
> generated, compiled, or executed.

## Overview

Query Intent is a structured JSON format that describes what data the caller
wants, not how to fetch it. A validated intent is compiled into a
**parameterized SQL query** using only allow-listed identifiers. The caller
receives both the results **and** the executed plan for transparency.

## Architecture

```
User JSON → Parse → Validate (allow-lists) → Compile (parameterized SQL)
→ Execute (pgx with $N params) → Return {results, plan}
```

**No LLM is involved.** The planner is deterministic, fully allow-listed, and
operates on structured JSON only.

## Intent Schema

```json
{
  "entity": "reservoir",
  "metrics": ["fill_percent", "stored_hm3"],
  "filters": {
    "slugs": ["Embalse de Mequinenza"],
    "basin": "Ebro",
    "province": "Zaragoza",
    "since": "2024-01-01",
    "until": "2024-12-31"
  },
  "aggregation": "latest",
  "sort": {
    "field": "fill_percent",
    "order": "desc"
  },
  "limit": 20,
  "chart_hint": "table"
}
```

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `entity` | string | Yes | Queryable entity (see allow-list) |
| `metrics` | string[] | Yes | Metrics to return (see allow-list) |
| `filters` | object | No | Filter criteria (validated for injection) |
| `aggregation` | string | Yes | Aggregation mode (see allow-list) |
| `sort` | object | No | `{field, order}` (see allow-list) |
| `limit` | integer | No | Max results (hard cap: **500**) |
| `chart_hint` | string | No | Suggested visualization type |

### Allow-Lists

#### Entities

- `reservoir` — individual reservoirs
- `basin` — hydrographic basins
- `province` — administrative provinces
- `community` — autonomous communities (future)
- `national` — national aggregates

#### Metrics per Entity

| Metric | Reservoir | Basin | Province | National |
|--------|-----------|-------|----------|----------|
| `fill_percent` | ✅ | ✅ | ✅ | ✅ |
| `stored_hm3` | ✅ | ✅ | ✅ | ✅ |
| `capacity_hm3` | ✅ | ✅ | ✅ | ✅ |
| `change_hm3` | ✅ | ❌ | ❌ | ❌ |

#### Aggregations

- `latest` — most recent reading per entity
- `timeseries` — full time series
- `ranking` — ordered list (fullest/emptiest)
- `compare` — side-by-side comparison
- `summary` — aggregate statistics

#### Sort Fields

`name`, `fill_percent`, `stored_hm3`, `capacity_hm3`, `change_hm3`,
`observed_at`, `basin_name`, `province_name`

#### Sort Orders

`asc`, `desc` (default: `asc`)

#### Chart Hints

`table`, `line`, `bar`, `map`, `none` (default: `none`)

### Filter Validation

All filter values are checked for SQL injection patterns before reaching the
compiler. Disallowed patterns include:

- Semicolons (`;`)
- SQL comments (`--`, `/*`, `*/`)
- DDL/DML keywords (`DROP`, `DELETE`, `INSERT`, `UPDATE`, `SELECT`, `UNION`)
- Stored procedure prefixes (`xp_`, `sp_`)
- Quotes (`'`, `"`)
- Null bytes (`\x00`, `\x1a`)

### Limits

- **Default**: 20
- **Hard maximum**: 500
- Negative limits are rejected

## Response Format

```json
{
  "data": {
    "results": [...],
    "plan": {
      "intent": {...},
      "entity_table": "reservoirs r LEFT JOIN ...",
      "selected_columns": [...],
      "where_clause": "WHERE ...",
      "order_by": "ORDER BY ...",
      "limit": 20,
      "parameters": ["Ebro", "2024-01-01"],
      "query_sql": "SELECT ... FROM ... WHERE ... LIMIT 20"
    },
    "count": 42
  },
  "lineage": {
    "source": "MITECO",
    "licence": "Ley 37/2007 + RD 1495/2011",
    "attribution": "Fuente: MITECO"
  }
}
```

## Security Guarantees

1. **No arbitrary SQL accepted**: Only allow-listed JSON fields are recognized.
2. **No string interpolation**: All user values become `$N` parameters.
3. **No LLM involved**: The planner is deterministic code, not a language model.
4. **Validation before compilation**: Invalid intents are rejected before any SQL is generated.
5. **Audit trail**: Every query plan includes the compiled SQL for inspection.
6. **Hard limits**: Result count, query complexity, and execution time are bounded.

## Example Queries

### Latest fill % for all reservoirs in Ebro basin

```json
{
  "entity": "reservoir",
  "metrics": ["fill_percent"],
  "filters": { "basin": "Ebro" },
  "aggregation": "latest",
  "limit": 50
}
```

### Top 10 fullest reservoirs

```json
{
  "entity": "reservoir",
  "metrics": ["fill_percent", "stored_hm3"],
  "aggregation": "ranking",
  "sort": { "field": "fill_percent", "order": "desc" },
  "limit": 10
}
```

### Time series for a specific reservoir

```json
{
  "entity": "reservoir",
  "metrics": ["fill_percent", "stored_hm3"],
  "filters": {
    "slugs": ["Embalse de Mequinenza"],
    "since": "2024-01-01",
    "until": "2024-06-30"
  },
  "aggregation": "timeseries",
  "limit": 500
}
```

## Endpoint

```
POST /api/v1/query
Content-Type: application/json
X-API-Key: <your-key>
```

## Related

- `api/openapi.yaml` — OpenAPI specification
- `internal/planner/` — Go implementation
- `internal/planner/planner_test.go` — Adversarial test suite
