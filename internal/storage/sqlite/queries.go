package sqlite

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
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
	Latitude      float64 `json:"latitude,omitempty"`
	Longitude     float64 `json:"longitude,omitempty"`
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

// BasinSummary provides aggregated fill data for a basin.
type BasinSummary struct {
	ID               int     `json:"id"`
	Name             string  `json:"name"`
	Code             string  `json:"code,omitempty"`
	ReservoirCount   int     `json:"reservoir_count"`
	TotalCapacity    float64 `json:"total_capacity_hm3"`
	TotalVolume      float64 `json:"total_volume_hm3"`
	AvgFillPct       float64 `json:"avg_fill_pct"`
	LatestObservedAt string  `json:"latest_observed_at,omitempty"`
}

// BasinDetail extends BasinSummary with the reservoirs in the basin.
type BasinDetail struct {
	BasinSummary
	Reservoirs []ReservoirSummary `json:"reservoirs"`
}

// RankingItem mirrors the v1 ranking type.
type RankingItem struct {
	Rank        int     `json:"rank"`
	ReservoirID int     `json:"reservoir_id"`
	Name        string  `json:"name"`
	Slug        string  `json:"slug"`
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

// recentMitecoReservoirs filters to reservoirs whose latest MITECO reading is
// within the last 6 months and has a non-zero fill percentage. The MVP public
// view relies only on current MITECO data; zero-fill or outdated reservoirs
// are tracked separately in docs/reservoirs-excluded-from-mvp.md.
const recentMitecoReservoirs = `
JOIN (
	SELECT rd.reservoir_id AS id
	FROM (
		SELECT rd.reservoir_id, rd.observed_at, rd.fill_pct,
		       ROW_NUMBER() OVER (PARTITION BY rd.reservoir_id ORDER BY rd.observed_at DESC) AS rn
		FROM readings rd
		JOIN sources s ON s.id = rd.source_id
		WHERE s.name = 'MITECO'
	) rd
	WHERE rd.rn = 1
	  AND rd.observed_at >= date('now', '-6 months')
	  AND rd.fill_pct > 0
) miteco ON miteco.id = r.id
`

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
	if err := db.QueryRow(`
		SELECT COUNT(*)
		FROM reservoirs r
		` + recentMitecoReservoirs + `
	`).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count reservoirs: %w", err)
	}

	rows, err := db.Query(`
		SELECT
			r.id, r.name, r.slug, COALESCE(r.external_id, '') AS external_id,
			b.name AS basin_name,
			COALESCE(p.name, '') AS province_name,
			COALESCE(r.capacity_hm3, d.capacity_hm3) AS capacity_hm3,
			(
				SELECT fill_pct FROM readings rd2
				JOIN sources s2 ON s2.id = rd2.source_id
				WHERE rd2.reservoir_id = r.id AND s2.name = 'MITECO'
				ORDER BY rd2.observed_at DESC
				LIMIT 1
			) AS latest_fill_pct,
			COALESCE(r.latitude, d.latitude) AS latitude,
			COALESCE(r.longitude, d.longitude) AS longitude
		FROM reservoirs r
		`+recentMitecoReservoirs+`
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
		var fillPct, lat, lon sql.NullFloat64
		if err := rows.Scan(&rs.ID, &rs.Name, &rs.Slug, &rs.ExternalID, &rs.BasinName, &rs.ProvinceName, &rs.CapacityHM3, &fillPct, &lat, &lon); err != nil {
			return nil, 0, err
		}
		if fillPct.Valid {
			rs.LatestFillPct = fillPct.Float64
		}
		if lat.Valid {
			rs.Latitude = lat.Float64
		}
		if lon.Valid {
			rs.Longitude = lon.Float64
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
			r.id, r.name, COALESCE(r.external_id, '') AS external_id,
			b.name AS basin_name,
			COALESCE(p.name, '') AS province_name,
			COALESCE(r.capacity_hm3, d.capacity_hm3) AS capacity_hm3,
			(
				SELECT volume_hm3 FROM readings rd2
				JOIN sources s2 ON s2.id = rd2.source_id
				WHERE rd2.reservoir_id = r.id AND s2.name = 'MITECO'
				ORDER BY rd2.observed_at DESC
				LIMIT 1
			) AS latest_volume,
			(
				SELECT fill_pct FROM readings rd2
				JOIN sources s2 ON s2.id = rd2.source_id
				WHERE rd2.reservoir_id = r.id AND s2.name = 'MITECO'
				ORDER BY rd2.observed_at DESC
				LIMIT 1
			) AS latest_fill_pct,
			d.name AS dam_name,
			COALESCE(r.latitude, d.latitude) AS latitude,
			COALESCE(r.longitude, d.longitude) AS longitude
		FROM reservoirs r
		` + recentMitecoReservoirs + `
		LEFT JOIN basins b ON r.basin_id = b.id
		LEFT JOIN provinces p ON r.province_id = p.id
		LEFT JOIN dams d ON r.dam_id = d.id
		WHERE r.slug = ? OR r.name = ?
		LIMIT 1
	`
	var latestVol sql.NullFloat64
	var lat, lon sql.NullFloat64
	err := db.QueryRow(query, slug, slug).Scan(
		&rd.ID, &rd.Name, &rd.ExternalID, &rd.BasinName, &rd.ProvinceName,
		&rd.CapacityHM3, &latestVol, &rd.LatestFillPct, &rd.DamName,
		&lat, &lon,
	)
	if latestVol.Valid {
		rd.LatestVolume = latestVol.Float64
	}
	if lat.Valid {
		rd.Latitude = lat.Float64
	}
	if lon.Valid {
		rd.Longitude = lon.Float64
	}
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
		SELECT COUNT(*) FROM readings rd
		JOIN sources s ON s.id = rd.source_id
		WHERE rd.reservoir_id = ? AND s.name = 'MITECO' AND rd.observed_at BETWEEN ? AND ?
	`, reservoirID, since.Format("2006-01-02"), until.Format("2006-01-02")).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("count readings: %w", err)
	}

	rows, err := db.Query(`
		SELECT
			rd.observed_at, rd.volume_hm3, rd.capacity_hm3, rd.fill_pct,
			rd.weekly_variation_hm3, rd.is_provisional
		FROM readings rd
		JOIN sources s ON s.id = rd.source_id
		WHERE rd.reservoir_id = ? AND s.name = 'MITECO' AND rd.observed_at BETWEEN ? AND ?
		ORDER BY rd.observed_at DESC
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

// QueryBasinSummaries returns aggregated fill statistics for every basin.
func QueryBasinSummaries(db *sql.DB) ([]BasinSummary, error) {
	rows, err := db.Query(`
		SELECT
			b.id,
			b.name,
			b.code,
			COUNT(DISTINCT r.id) AS reservoir_count,
			COALESCE(SUM(latest.capacity_hm3), 0) AS total_capacity,
			COALESCE(SUM(latest.volume_hm3), 0) AS total_volume,
			CASE WHEN COALESCE(SUM(latest.capacity_hm3), 0) > 0
				THEN (COALESCE(SUM(latest.volume_hm3), 0) / COALESCE(SUM(latest.capacity_hm3), 0)) * 100
				ELSE 0
			END AS avg_fill_pct,
			MAX(latest.observed_at) AS latest_observed_at
		FROM basins b
		LEFT JOIN reservoirs r ON r.basin_id = b.id
		` + recentMitecoReservoirs + `
		LEFT JOIN (
			SELECT
				rd.reservoir_id,
				rd.observed_at,
				rd.volume_hm3,
				rd.capacity_hm3
			FROM readings rd
			JOIN sources s ON s.id = rd.source_id
			WHERE s.name = 'MITECO'
			  AND (rd.reservoir_id, rd.observed_at) IN (
				SELECT reservoir_id, MAX(observed_at)
				FROM readings rd2
				JOIN sources s2 ON s2.id = rd2.source_id
				WHERE s2.name = 'MITECO'
				GROUP BY rd2.reservoir_id
			)
		) latest ON latest.reservoir_id = r.id
		GROUP BY b.id, b.name, b.code
		ORDER BY b.name
	`)
	if err != nil {
		return nil, fmt.Errorf("query basin summaries: %w", err)
	}
	defer rows.Close()

	var results []BasinSummary
	for rows.Next() {
		var bs BasinSummary
		var latestObs sql.NullString
		if err := rows.Scan(&bs.ID, &bs.Name, &bs.Code, &bs.ReservoirCount, &bs.TotalCapacity, &bs.TotalVolume, &bs.AvgFillPct, &latestObs); err != nil {
			return nil, err
		}
		if latestObs.Valid {
			bs.LatestObservedAt = latestObs.String
		}
		results = append(results, bs)
	}
	return results, rows.Err()
}

// QueryBasinDetail returns a basin summary plus its reservoirs with latest readings.
func QueryBasinDetail(db *sql.DB, slug string) (BasinDetail, error) {
	var bd BasinDetail

	// Basin summary.
	row := db.QueryRow(`
		SELECT
			b.id,
			b.name,
			b.code,
			COUNT(DISTINCT r.id) AS reservoir_count,
			COALESCE(SUM(latest.capacity_hm3), 0) AS total_capacity,
			COALESCE(SUM(latest.volume_hm3), 0) AS total_volume,
			CASE WHEN COALESCE(SUM(latest.capacity_hm3), 0) > 0
				THEN (COALESCE(SUM(latest.volume_hm3), 0) / COALESCE(SUM(latest.capacity_hm3), 0)) * 100
				ELSE 0
			END AS avg_fill_pct,
			MAX(latest.observed_at) AS latest_observed_at
		FROM basins b
		LEFT JOIN reservoirs r ON r.basin_id = b.id
		`+recentMitecoReservoirs+`
		LEFT JOIN (
			SELECT
				rd.reservoir_id,
				rd.observed_at,
				rd.volume_hm3,
				rd.capacity_hm3
			FROM readings rd
			JOIN sources s ON s.id = rd.source_id
			WHERE s.name = 'MITECO'
			  AND (rd.reservoir_id, rd.observed_at) IN (
				SELECT reservoir_id, MAX(observed_at)
				FROM readings rd2
				JOIN sources s2 ON s2.id = rd2.source_id
				WHERE s2.name = 'MITECO'
				GROUP BY rd2.reservoir_id
			)
		) latest ON latest.reservoir_id = r.id
		WHERE b.name = ?
		GROUP BY b.id, b.name, b.code
	`, slug)
	var latestObs sql.NullString
	err := row.Scan(&bd.ID, &bd.Name, &bd.Code, &bd.ReservoirCount, &bd.TotalCapacity, &bd.TotalVolume, &bd.AvgFillPct, &latestObs)
	if err != nil {
		if err == sql.ErrNoRows {
			return bd, fmt.Errorf("basin not found: %s", slug)
		}
		return bd, fmt.Errorf("query basin detail: %w", err)
	}
	if latestObs.Valid {
		bd.LatestObservedAt = latestObs.String
	}

	// Reservoirs in basin.
	resRows, err := db.Query(`
		SELECT
			r.id, r.name, r.slug, COALESCE(r.external_id, '') AS external_id,
			b.name AS basin_name,
			COALESCE(p.name, '') AS province_name,
			COALESCE(r.capacity_hm3, d.capacity_hm3) AS capacity_hm3,
			latest.fill_pct AS latest_fill_pct,
			COALESCE(r.latitude, d.latitude) AS latitude,
			COALESCE(r.longitude, d.longitude) AS longitude
		FROM reservoirs r
		`+recentMitecoReservoirs+`
		LEFT JOIN basins b ON r.basin_id = b.id
		LEFT JOIN provinces p ON r.province_id = p.id
		LEFT JOIN dams d ON r.dam_id = d.id
		LEFT JOIN (
			SELECT rd.reservoir_id, rd.fill_pct
			FROM readings rd
			JOIN sources s ON s.id = rd.source_id
			WHERE s.name = 'MITECO'
			  AND (rd.reservoir_id, rd.observed_at) IN (
				SELECT reservoir_id, MAX(observed_at)
				FROM readings rd2
				JOIN sources s2 ON s2.id = rd2.source_id
				WHERE s2.name = 'MITECO'
				GROUP BY rd2.reservoir_id
			)
		) latest ON latest.reservoir_id = r.id
		WHERE b.name = ?
		ORDER BY r.name
	`, slug)
	if err != nil {
		return bd, fmt.Errorf("query basin reservoirs: %w", err)
	}
	defer resRows.Close()

	for resRows.Next() {
		var rs ReservoirSummary
		var fillPct, lat, lon sql.NullFloat64
		if err := resRows.Scan(&rs.ID, &rs.Name, &rs.Slug, &rs.ExternalID, &rs.BasinName, &rs.ProvinceName, &rs.CapacityHM3, &fillPct, &lat, &lon); err != nil {
			return bd, err
		}
		if fillPct.Valid {
			rs.LatestFillPct = fillPct.Float64
		}
		if lat.Valid {
			rs.Latitude = lat.Float64
		}
		if lon.Valid {
			rs.Longitude = lon.Float64
		}
		bd.Reservoirs = append(bd.Reservoirs, rs)
	}

	return bd, resRows.Err()
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
		SELECT r.id, r.name, r.slug, rd.fill_pct
		FROM reservoirs r
		%s
		JOIN (
			SELECT rd2.reservoir_id, rd2.fill_pct
			FROM readings rd2
			JOIN sources s ON s.id = rd2.source_id
			WHERE s.name = 'MITECO'
			  AND (rd2.reservoir_id, rd2.observed_at) IN (
				SELECT reservoir_id, MAX(observed_at)
				FROM readings rd3
				JOIN sources s3 ON s3.id = rd3.source_id
				WHERE s3.name = 'MITECO'
				GROUP BY rd3.reservoir_id
			)
		) rd ON rd.reservoir_id = r.id
		WHERE rd.fill_pct IS NOT NULL
		ORDER BY rd.fill_pct %s
		LIMIT ?
	`, recentMitecoReservoirs, order)

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
		if err := rows.Scan(&ri.ReservoirID, &ri.Name, &ri.Slug, &val); err != nil {
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

// ImportReadingsCSV parses a CSV reader and upserts readings into the database.
// It expects the column index map to contain at least reservoir_slug and observed_at.
func ImportReadingsCSV(db *sql.DB, reader *csv.Reader, colIdx map[string]int) (int, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// Ensure MANUAL source exists.
	var sourceID int
	err = tx.QueryRow(`SELECT id FROM sources WHERE name = 'MANUAL'`).Scan(&sourceID)
	if err != nil {
		if err == sql.ErrNoRows {
			res, err := tx.Exec(`INSERT INTO sources (name, organism, licence, attribution, url) VALUES (?, ?, ?, ?, ?)`,
				"MANUAL",
				"User-supplied data",
				"N/A",
				"Datos aportados manualmente",
				"",
			)
			if err != nil {
				return 0, fmt.Errorf("insert MANUAL source: %w", err)
			}
			id, _ := res.LastInsertId()
			sourceID = int(id)
		} else {
			return 0, fmt.Errorf("lookup MANUAL source: %w", err)
		}
	}

	// Build slug -> id map.
	reservoirMap := make(map[string]int)
	rows, err := tx.Query(`SELECT id, slug FROM reservoirs WHERE slug IS NOT NULL AND slug != ''`)
	if err != nil {
		return 0, fmt.Errorf("query reservoirs: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var slug string
		if err := rows.Scan(&id, &slug); err != nil {
			return 0, err
		}
		reservoirMap[slug] = id
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}

	slugCol := colIdx["reservoir_slug"]
	dateCol := colIdx["observed_at"]
	volCol := colIdx["volume_hm3"]
	capCol := colIdx["capacity_hm3"]
	fillCol := colIdx["fill_pct"]

	upsertStmt, err := tx.Prepare(`
		INSERT INTO readings (reservoir_id, source_id, observed_at, volume_hm3, capacity_hm3, fill_pct, weekly_variation_hm3, is_provisional, is_official, fetched_at)
		VALUES (?, ?, ?, ?, ?, ?, NULL, 1, 1, datetime('now'))
		ON CONFLICT(reservoir_id, observed_at, source_id) DO UPDATE SET
			volume_hm3 = excluded.volume_hm3,
			capacity_hm3 = excluded.capacity_hm3,
			fill_pct = excluded.fill_pct,
			fetched_at = excluded.fetched_at
	`)
	if err != nil {
		return 0, fmt.Errorf("prepare upsert: %w", err)
	}
	defer upsertStmt.Close()

	var count int
	line := 1
	for {
		row, err := reader.Read()
		if err == csv.ErrFieldCount || err == csv.ErrBareQuote {
			line++
			continue
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return 0, fmt.Errorf("row %d: %w", line, err)
		}
		line++

		if len(row) <= slugCol || len(row) <= dateCol {
			continue
		}
		slug := strings.TrimSpace(row[slugCol])
		observedAt := strings.TrimSpace(row[dateCol])
		if slug == "" || observedAt == "" {
			continue
		}

		reservoirID, ok := reservoirMap[slug]
		if !ok {
			// Try fallback variants.
			if id, ok := reservoirMap["embalse-de-"+slug]; ok {
				reservoirID = id
			} else if id, ok := reservoirMap[strings.TrimPrefix(slug, "embalse-de-")]; ok {
				reservoirID = id
			} else {
				continue
			}
		}

		var volume, capacity, fillPct float64
		hasValue := false
		if volCol >= 0 && capCol >= 0 && len(row) > volCol && len(row) > capCol {
			vStr := strings.TrimSpace(row[volCol])
			cStr := strings.TrimSpace(row[capCol])
			if vStr != "" && cStr != "" {
				volume, _ = strconv.ParseFloat(strings.ReplaceAll(vStr, ",", "."), 64)
				capacity, _ = strconv.ParseFloat(strings.ReplaceAll(cStr, ",", "."), 64)
				if capacity > 0 {
					fillPct = (volume / capacity) * 100
					hasValue = true
				}
			}
		}
		if !hasValue && fillCol >= 0 && len(row) > fillCol {
			fStr := strings.TrimSpace(row[fillCol])
			if fStr != "" {
				fillPct, _ = strconv.ParseFloat(strings.ReplaceAll(fStr, ",", "."), 64)
				hasValue = true
			}
		}
		if !hasValue {
			continue
		}

		if _, err := upsertStmt.Exec(reservoirID, sourceID, observedAt, volume, capacity, fillPct); err != nil {
			return 0, fmt.Errorf("upsert row %d: %w", line, err)
		}
		count++
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return count, nil
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
