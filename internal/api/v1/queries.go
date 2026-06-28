package v1

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ReservoirSummary is a lightweight reservoir view for lists.
type ReservoirSummary struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	ExternalID    string  `json:"external_id"`
	BasinName     string  `json:"basin_name,omitempty"`
	ProvinceName  string  `json:"province_name,omitempty"`
	CapacityHM3   float64 `json:"capacity_hm3,omitempty"`
	LatestFillPct float64 `json:"latest_fill_pct,omitempty"`
}

// ReservoirDetail is the full reservoir view.
type ReservoirDetail struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	ExternalID    string  `json:"external_id"`
	BasinName     string  `json:"basin_name,omitempty"`
	ProvinceName  string  `json:"province_name,omitempty"`
	CapacityHM3   float64 `json:"capacity_hm3,omitempty"`
	LatestVolume  float64 `json:"latest_volume_hm3,omitempty"`
	LatestFillPct float64 `json:"latest_fill_pct,omitempty"`
	DamName       string  `json:"dam_name,omitempty"`
}

// Reading is a time-series data point.
type Reading struct {
	ObservedAt         string  `json:"observed_at"`
	VolumeHM3          float64 `json:"volume_hm3"`
	CapacityHM3        float64 `json:"capacity_hm3"`
	FillPct            float64 `json:"fill_pct"`
	WeeklyVariationHM3 float64 `json:"weekly_variation_hm3,omitempty"`
	IsProvisional      bool    `json:"is_provisional"`
}

// SourceResponse is the public source representation.
type SourceResponse struct {
	Name        string `json:"name"`
	Organism    string `json:"organism"`
	Licence     string `json:"licence"`
	Attribution string `json:"attribution"`
	URL         string `json:"url,omitempty"`
}

// BasinResponse is the public basin representation.
type BasinResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Code string `json:"code,omitempty"`
}

// RankingItem is a reservoir ranking entry.
type RankingItem struct {
	Rank        int     `json:"rank"`
	ReservoirID int     `json:"reservoir_id"`
	Name        string  `json:"name"`
	Value       float64 `json:"value"`
	Metric      string  `json:"metric"`
}

// DataQualityReport is a data quality summary.
type DataQualityReport struct {
	TotalReservoirs        int    `json:"total_reservoirs"`
	ReservoirsWithReadings int    `json:"reservoirs_with_readings"`
	LatestReadingDate      string `json:"latest_reading_date,omitempty"`
	OldestReadingDate      string `json:"oldest_reading_date,omitempty"`
	ProvisionalCount       int    `json:"provisional_count"`
	OfficialCount          int    `json:"official_count"`
}

// QuerySources returns all registered sources with attribution.
func QuerySources(ctx context.Context, pool *pgxpool.Pool) ([]SourceResponse, error) {
	rows, err := pool.Query(ctx, `
		SELECT name, organism, licence, attribution, url
		FROM sources
		ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("query sources: %w", err)
	}
	defer rows.Close()

	var results []SourceResponse
	for rows.Next() {
		var s SourceResponse
		if err := rows.Scan(&s.Name, &s.Organism, &s.Licence, &s.Attribution, &s.URL); err != nil {
			return nil, err
		}
		results = append(results, s)
	}
	return results, rows.Err()
}

// QueryReservoirs returns a paginated list of reservoirs.
func QueryReservoirs(ctx context.Context, pool *pgxpool.Pool, offset, limit int) ([]ReservoirSummary, int, error) {
	var total int
	err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM reservoirs`).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count reservoirs: %w", err)
	}

	rows, err := pool.Query(ctx, `
		SELECT
			r.id, r.name, r.external_id,
			b.name AS basin_name,
			p.name AS province_name,
			COALESCE(r.capacity_hm3, d.capacity_hm3) AS capacity_hm3,
			(
				SELECT fill_pct FROM readings
				WHERE reservoir_id = r.id
				ORDER BY observed_at DESC
				LIMIT 1
			) AS latest_fill_pct
		FROM reservoirs r
		LEFT JOIN basins b ON r.basin_id = b.id
		LEFT JOIN provinces p ON r.province_id = p.id
		LEFT JOIN dams d ON r.dam_id = d.id
		ORDER BY r.name
		OFFSET $1 LIMIT $2
	`, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("query reservoirs: %w", err)
	}
	defer rows.Close()

	var results []ReservoirSummary
	for rows.Next() {
		var rs ReservoirSummary
		var fillPct *float64
		if err := rows.Scan(&rs.ID, &rs.Name, &rs.ExternalID, &rs.BasinName, &rs.ProvinceName, &rs.CapacityHM3, &fillPct); err != nil {
			return nil, 0, err
		}
		if fillPct != nil {
			rs.LatestFillPct = *fillPct
		}
		results = append(results, rs)
	}
	return results, total, rows.Err()
}

// QueryReservoirBySlug returns a single reservoir by name (slug).
func QueryReservoirBySlug(ctx context.Context, pool *pgxpool.Pool, slug string) (ReservoirDetail, error) {
	var rd ReservoirDetail
	err := pool.QueryRow(ctx, `
		SELECT
			r.id, r.name, r.external_id,
			b.name AS basin_name,
			p.name AS province_name,
			COALESCE(r.capacity_hm3, d.capacity_hm3) AS capacity_hm3,
			(
				SELECT volume_hm3 FROM readings
				WHERE reservoir_id = r.id
				ORDER BY observed_at DESC
				LIMIT 1
			) AS latest_volume,
			(
				SELECT fill_pct FROM readings
				WHERE reservoir_id = r.id
				ORDER BY observed_at DESC
				LIMIT 1
			) AS latest_fill_pct,
			d.name AS dam_name
		FROM reservoirs r
		LEFT JOIN basins b ON r.basin_id = b.id
		LEFT JOIN provinces p ON r.province_id = p.id
		LEFT JOIN dams d ON r.dam_id = d.id
		WHERE r.name = $1
	`, slug).Scan(&rd.ID, &rd.Name, &rd.ExternalID, &rd.BasinName, &rd.ProvinceName, &rd.CapacityHM3, &rd.LatestVolume, &rd.LatestFillPct, &rd.DamName)
	if err != nil {
		return rd, fmt.Errorf("query reservoir: %w", err)
	}
	return rd, nil
}

// QueryReadings returns time-series readings for a reservoir.
func QueryReadings(ctx context.Context, pool *pgxpool.Pool, reservoirName string, since, until time.Time, offset, limit int) ([]Reading, int, error) {
	// First, resolve reservoir_id
	var reservoirID int
	err := pool.QueryRow(ctx, `SELECT id FROM reservoirs WHERE name = $1`, reservoirName).Scan(&reservoirID)
	if err != nil {
		return nil, 0, fmt.Errorf("reservoir not found: %w", err)
	}

	var total int
	err = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM readings
		WHERE reservoir_id = $1 AND observed_at BETWEEN $2 AND $3
	`, reservoirID, since, until).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count readings: %w", err)
	}

	rows, err := pool.Query(ctx, `
		SELECT
			observed_at, volume_hm3, capacity_hm3, fill_pct,
			weekly_variation_hm3, is_provisional
		FROM readings
		WHERE reservoir_id = $1 AND observed_at BETWEEN $2 AND $3
		ORDER BY observed_at DESC
		OFFSET $4 LIMIT $5
	`, reservoirID, since, until, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("query readings: %w", err)
	}
	defer rows.Close()

	var results []Reading
	for rows.Next() {
		var rd Reading
		var vol, cap, fill, weekly *float64
		var obs time.Time
		if err := rows.Scan(&obs, &vol, &cap, &fill, &weekly, &rd.IsProvisional); err != nil {
			return nil, 0, err
		}
		rd.ObservedAt = obs.Format("2006-01-02")
		if vol != nil {
			rd.VolumeHM3 = *vol
		}
		if cap != nil {
			rd.CapacityHM3 = *cap
		}
		if fill != nil {
			rd.FillPct = *fill
		}
		if weekly != nil {
			rd.WeeklyVariationHM3 = *weekly
		}
		results = append(results, rd)
	}
	return results, total, rows.Err()
}

// QueryBasins returns all basins.
func QueryBasins(ctx context.Context, pool *pgxpool.Pool) ([]BasinResponse, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, name, code FROM basins ORDER BY name
	`)
	if err != nil {
		return nil, fmt.Errorf("query basins: %w", err)
	}
	defer rows.Close()

	var results []BasinResponse
	for rows.Next() {
		var b BasinResponse
		if err := rows.Scan(&b.ID, &b.Name, &b.Code); err != nil {
			return nil, err
		}
		results = append(results, b)
	}
	return results, rows.Err()
}

// QueryBasinBySlug returns a basin by name.
func QueryBasinBySlug(ctx context.Context, pool *pgxpool.Pool, slug string) (BasinResponse, error) {
	var b BasinResponse
	err := pool.QueryRow(ctx, `
		SELECT id, name, code FROM basins WHERE name = $1
	`, slug).Scan(&b.ID, &b.Name, &b.Code)
	if err != nil {
		return b, fmt.Errorf("query basin: %w", err)
	}
	return b, nil
}

// QueryRankings returns reservoir rankings by fill percentage.
func QueryRankings(ctx context.Context, pool *pgxpool.Pool, metric string, limit int) ([]RankingItem, error) {
	var sql string
	switch metric {
	case "fullest":
		sql = `
			SELECT r.id, r.name, rd.fill_pct
			FROM reservoirs r
			JOIN LATERAL (
				SELECT fill_pct FROM readings
				WHERE reservoir_id = r.id
				ORDER BY observed_at DESC
				LIMIT 1
			) rd ON true
			WHERE rd.fill_pct IS NOT NULL
			ORDER BY rd.fill_pct DESC
			LIMIT $1
		`
	case "emptiest":
		sql = `
			SELECT r.id, r.name, rd.fill_pct
			FROM reservoirs r
			JOIN LATERAL (
				SELECT fill_pct FROM readings
				WHERE reservoir_id = r.id
				ORDER BY observed_at DESC
				LIMIT 1
			) rd ON true
			WHERE rd.fill_pct IS NOT NULL
			ORDER BY rd.fill_pct ASC
			LIMIT $1
		`
	default:
		return nil, fmt.Errorf("unknown metric: %s", metric)
	}

	rows, err := pool.Query(ctx, sql, limit)
	if err != nil {
		return nil, fmt.Errorf("query rankings: %w", err)
	}
	defer rows.Close()

	var results []RankingItem
	rank := 1
	for rows.Next() {
		var ri RankingItem
		var val *float64
		if err := rows.Scan(&ri.ReservoirID, &ri.Name, &val); err != nil {
			return nil, err
		}
		ri.Rank = rank
		ri.Metric = metric
		if val != nil {
			ri.Value = *val
		}
		results = append(results, ri)
		rank++
	}
	return results, rows.Err()
}

// QueryComparator returns aligned readings for multiple reservoirs.
func QueryComparator(ctx context.Context, pool *pgxpool.Pool, names []string, since, until time.Time) (map[string][]Reading, error) {
	if len(names) > 5 {
		return nil, fmt.Errorf("maximum 5 reservoirs allowed")
	}
	result := make(map[string][]Reading)
	for _, name := range names {
		readings, _, err := QueryReadings(ctx, pool, name, since, until, 0, 1000)
		if err != nil {
			return nil, fmt.Errorf("readings for %s: %w", name, err)
		}
		result[name] = readings
	}
	return result, nil
}

// QueryDataQuality returns a data quality summary.
func QueryDataQuality(ctx context.Context, pool *pgxpool.Pool) (DataQualityReport, error) {
	var r DataQualityReport

	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM reservoirs`).Scan(&r.TotalReservoirs)
	_ = pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT reservoir_id) FROM readings
	`).Scan(&r.ReservoirsWithReadings)
	_ = pool.QueryRow(ctx, `
		SELECT MAX(observed_at), MIN(observed_at) FROM readings
	`).Scan(&r.LatestReadingDate, &r.OldestReadingDate)
	_ = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM readings WHERE is_provisional = TRUE
	`).Scan(&r.ProvisionalCount)
	_ = pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM readings WHERE is_official = TRUE
	`).Scan(&r.OfficialCount)

	return r, nil
}

// QueryLineage returns the lineage for a given source name.
func QueryLineage(ctx context.Context, pool *pgxpool.Pool, sourceName string) (Lineage, error) {
	var l Lineage
	err := pool.QueryRow(ctx, `
		SELECT name, licence, attribution, last_fetched_at
		FROM sources
		WHERE name = $1
	`, sourceName).Scan(&l.Source, &l.Licence, &l.Attribution, &l.FetchedAt)
	if err != nil {
		return l, fmt.Errorf("query lineage: %w", err)
	}
	return l, nil
}
