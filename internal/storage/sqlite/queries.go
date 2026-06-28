package sqlite

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// ReservoirSummary mirrors the v1 type for list views.
type ReservoirSummary struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	Slug          string  `json:"slug"`
	ExternalID    string  `json:"external_id"`
	BasinName     string  `json:"basin_name,omitempty"`
	ProvinceName  string  `json:"province_name,omitempty"`
	CapacityHM3   float64 `json:"capacity_hm3,omitempty"`
	LatestFillPct float64 `json:"latest_fill_pct,omitempty"`
}

// ReservoirDetail mirrors the v1 detail type.
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
	Latitude      float64 `json:"latitude,omitempty"`
	Longitude     float64 `json:"longitude,omitempty"`
}

// Reading mirrors the v1 reading type.
type Reading struct {
	ObservedAt         string  `json:"observed_at"`
	VolumeHM3          float64 `json:"volume_hm3"`
	CapacityHM3        float64 `json:"capacity_hm3"`
	FillPct            float64 `json:"fill_pct"`
	WeeklyVariationHM3 float64 `json:"weekly_variation_hm3,omitempty"`
	IsProvisional      bool    `json:"is_provisional"`
}

// SourceResponse mirrors the v1 source type.
type SourceResponse struct {
	Name        string `json:"name"`
	Organism    string `json:"organism"`
	Licence     string `json:"licence"`
	Attribution string `json:"attribution"`
	URL         string `json:"url,omitempty"`
}

// BasinResponse mirrors the v1 basin type.
type BasinResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Code string `json:"code,omitempty"`
}

// RankingItem mirrors the v1 ranking type.
type RankingItem struct {
	Rank        int     `json:"rank"`
	ReservoirID int     `json:"reservoir_id"`
	Name        string  `json:"name"`
	Value       float64 `json:"value"`
	Metric      string  `json:"metric"`
}

// DataQualityReport mirrors the v1 data quality type.
type DataQualityReport struct {
	TotalReservoirs        int    `json:"total_reservoirs"`
	ReservoirsWithReadings int    `json:"reservoirs_with_readings"`
	LatestReadingDate      string `json:"latest_reading_date,omitempty"`
	OldestReadingDate      string `json:"oldest_reading_date,omitempty"`
	ProvisionalCount       int    `json:"provisional_count"`
	OfficialCount          int    `json:"official_count"`
}

// Lineage mirrors the v1 lineage type.
type Lineage struct {
	Source      string    `json:"source"`
	Licence     string    `json:"licence"`
	Attribution string    `json:"attribution"`
	FetchedAt   time.Time `json:"fetched_at,omitempty"`
}

// QuerySources returns all registered sources.
func QuerySources(db *sql.DB) ([]SourceResponse, error) {
	rows, err := db.Query(`SELECT name, organism, licence, attribution, url FROM sources ORDER BY name`)
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

// QueryReservoirs returns a paginated list.
func QueryReservoirs(db *sql.DB, offset, limit int) ([]ReservoirSummary, int, error) {
	var total int
	if err := db.QueryRow(`SELECT COUNT(*) FROM reservoirs`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count reservoirs: %w", err)
	}

	rows, err := db.Query(`
		SELECT
			r.id, r.name, r.slug, r.external_id,
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
		LIMIT ? OFFSET ?
	`, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query reservoirs: %w", err)
	}
	defer rows.Close()

	var results []ReservoirSummary
	for rows.Next() {
		var rs ReservoirSummary
		var fillPct sql.NullFloat64
		if err := rows.Scan(&rs.ID, &rs.Name, &rs.Slug, &rs.ExternalID, &rs.BasinName, &rs.ProvinceName, &rs.CapacityHM3, &fillPct); err != nil {
			return nil, 0, err
		}
		if fillPct.Valid {
			rs.LatestFillPct = fillPct.Float64
		}
		results = append(results, rs)
	}
	return results, total, rows.Err()
}

// QueryReservoirBySlug returns a single reservoir by slug.
func QueryReservoirBySlug(db *sql.DB, slug string) (ReservoirDetail, error) {
	var rd ReservoirDetail
	// Support both slug and full name lookups for backward compatibility
	query := `
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
			d.name AS dam_name,
			COALESCE(r.latitude, d.latitude) AS latitude,
			COALESCE(r.longitude, d.longitude) AS longitude
		FROM reservoirs r
		LEFT JOIN basins b ON r.basin_id = b.id
		LEFT JOIN provinces p ON r.province_id = p.id
		LEFT JOIN dams d ON r.dam_id = d.id
		WHERE r.slug = ? OR r.name = ?
		LIMIT 1
	`
	err := db.QueryRow(query, slug, slug).Scan(
		&rd.ID, &rd.Name, &rd.ExternalID, &rd.BasinName, &rd.ProvinceName,
		&rd.CapacityHM3, &rd.LatestVolume, &rd.LatestFillPct, &rd.DamName,
		&rd.Latitude, &rd.Longitude,
	)
	if err != nil {
		return rd, fmt.Errorf("query reservoir: %w", err)
	}
	return rd, nil
}

// QueryReadings returns time-series readings for a reservoir.
func QueryReadings(db *sql.DB, reservoirName string, since, until time.Time, offset, limit int) ([]Reading, int, error) {
	var reservoirID int
	err := db.QueryRow(`SELECT id FROM reservoirs WHERE name = ? OR slug = ?`, reservoirName, reservoirName).Scan(&reservoirID)
	if err != nil {
		return nil, 0, fmt.Errorf("reservoir not found: %w", err)
	}

	var total int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM readings
		WHERE reservoir_id = ? AND observed_at BETWEEN ? AND ?
	`, reservoirID, since.Format("2006-01-02"), until.Format("2006-01-02")).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count readings: %w", err)
	}

	rows, err := db.Query(`
		SELECT
			observed_at, volume_hm3, capacity_hm3, fill_pct,
			weekly_variation_hm3, is_provisional
		FROM readings
		WHERE reservoir_id = ? AND observed_at BETWEEN ? AND ?
		ORDER BY observed_at DESC
		LIMIT ? OFFSET ?
	`, reservoirID, since.Format("2006-01-02"), until.Format("2006-01-02"), limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("query readings: %w", err)
	}
	defer rows.Close()

	var results []Reading
	for rows.Next() {
		var rd Reading
		var vol, cap, fill, weekly sql.NullFloat64
		var prov sql.NullInt64
		var obs string
		if err := rows.Scan(&obs, &vol, &cap, &fill, &weekly, &prov); err != nil {
			return nil, 0, err
		}
		rd.ObservedAt = obs
		if vol.Valid {
			rd.VolumeHM3 = vol.Float64
		}
		if cap.Valid {
			rd.CapacityHM3 = cap.Float64
		}
		if fill.Valid {
			rd.FillPct = fill.Float64
		}
		if weekly.Valid {
			rd.WeeklyVariationHM3 = weekly.Float64
		}
		if prov.Valid && prov.Int64 != 0 {
			rd.IsProvisional = true
		}
		results = append(results, rd)
	}
	return results, total, rows.Err()
}

// QueryBasins returns all basins.
func QueryBasins(db *sql.DB) ([]BasinResponse, error) {
	rows, err := db.Query(`SELECT id, name, code FROM basins ORDER BY name`)
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
func QueryBasinBySlug(db *sql.DB, slug string) (BasinResponse, error) {
	var b BasinResponse
	err := db.QueryRow(`SELECT id, name, code FROM basins WHERE name = ?`, slug).Scan(&b.ID, &b.Name, &b.Code)
	if err != nil {
		return b, fmt.Errorf("query basin: %w", err)
	}
	return b, nil
}

// QueryRankings returns reservoir rankings.
func QueryRankings(db *sql.DB, metric string, limit int) ([]RankingItem, error) {
	var order string
	switch metric {
	case "fullest":
		order = "DESC"
	case "emptiest":
		order = "ASC"
	default:
		return nil, fmt.Errorf("unknown metric: %s", metric)
	}

	sqlStr := fmt.Sprintf(`
		SELECT r.id, r.name, rd.fill_pct
		FROM reservoirs r
		JOIN (
			SELECT reservoir_id, fill_pct FROM readings
			WHERE (reservoir_id, observed_at) IN (
				SELECT reservoir_id, MAX(observed_at) FROM readings GROUP BY reservoir_id
			)
		) rd ON rd.reservoir_id = r.id
		WHERE rd.fill_pct IS NOT NULL
		ORDER BY rd.fill_pct %s
		LIMIT ?
	`, order)

	rows, err := db.Query(sqlStr, limit)
	if err != nil {
		return nil, fmt.Errorf("query rankings: %w", err)
	}
	defer rows.Close()

	var results []RankingItem
	rank := 1
	for rows.Next() {
		var ri RankingItem
		var val sql.NullFloat64
		if err := rows.Scan(&ri.ReservoirID, &ri.Name, &val); err != nil {
			return nil, err
		}
		ri.Rank = rank
		ri.Metric = metric
		if val.Valid {
			ri.Value = val.Float64
		}
		results = append(results, ri)
		rank++
	}
	return results, rows.Err()
}

// QueryComparator returns aligned readings for multiple reservoirs.
func QueryComparator(db *sql.DB, names []string, since, until time.Time) (map[string][]Reading, error) {
	if len(names) > 5 {
		return nil, fmt.Errorf("maximum 5 reservoirs allowed")
	}
	result := make(map[string][]Reading)
	for _, name := range names {
		readings, _, err := QueryReadings(db, name, since, until, 0, 1000)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				continue // skip unknown reservoirs
			}
			return nil, fmt.Errorf("readings for %s: %w", name, err)
		}
		result[name] = readings
	}
	return result, nil
}

// QueryDataQuality returns a data quality report.
func QueryDataQuality(db *sql.DB) (DataQualityReport, error) {
	var r DataQualityReport

	_ = db.QueryRow(`SELECT COUNT(*) FROM reservoirs`).Scan(&r.TotalReservoirs)
	_ = db.QueryRow(`SELECT COUNT(DISTINCT reservoir_id) FROM readings`).Scan(&r.ReservoirsWithReadings)
	_ = db.QueryRow(`SELECT MAX(observed_at), MIN(observed_at) FROM readings`).Scan(&r.LatestReadingDate, &r.OldestReadingDate)
	_ = db.QueryRow(`SELECT COUNT(*) FROM readings WHERE is_provisional = 1`).Scan(&r.ProvisionalCount)
	_ = db.QueryRow(`SELECT COUNT(*) FROM readings WHERE is_official = 1`).Scan(&r.OfficialCount)

	return r, nil
}

// QueryLineage returns lineage for a source.
func QueryLineage(db *sql.DB, sourceName string) (Lineage, error) {
	var l Lineage
	var fetchedAt sql.NullString
	err := db.QueryRow(`
		SELECT name, licence, attribution, last_fetched_at
		FROM sources
		WHERE name = ?
	`, sourceName).Scan(&l.Source, &l.Licence, &l.Attribution, &fetchedAt)
	if err != nil {
		return l, fmt.Errorf("query lineage: %w", err)
	}
	if fetchedAt.Valid && fetchedAt.String != "" {
		if t, err := time.Parse(time.RFC3339, fetchedAt.String); err == nil {
			l.FetchedAt = t
		}
	}
	return l, nil
}
