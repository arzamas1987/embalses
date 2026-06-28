package main

import (
	"database/sql"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// seedSyntheticReadings generates the same 6-month synthetic dataset as cmd/seed
// but writes it into the SQLite database. Used for UI testing before real MITECO data arrives.
func seedSyntheticReadings(db *sql.DB) error {
	// Get all reservoir IDs and names
	rows, err := db.Query(`SELECT id, name, capacity_hm3 FROM reservoirs`)
	if err != nil {
		return fmt.Errorf("query reservoirs: %w", err)
	}
	defer rows.Close()

	type res struct {
		id       int
		name     string
		capacity float64
	}
	var reservoirs []res
	for rows.Next() {
		var r res
		var cap sql.NullFloat64
		if err := rows.Scan(&r.id, &r.name, &cap); err != nil {
			return err
		}
		if cap.Valid {
			r.capacity = cap.Float64
		}
		reservoirs = append(reservoirs, r)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	if len(reservoirs) == 0 {
		return fmt.Errorf("no reservoirs found — run updater with -geo-only first")
	}

	endDate := time.Now().Truncate(24 * time.Hour)
	startDate := endDate.AddDate(0, -6, 0)

	totalCount := 0
	for _, r := range reservoirs {
		baseFill := 35 + float64(len(r.name)%45)
		capacity := r.capacity
		if capacity == 0 {
			capacity = 100 + rand.Float64()*500
		}

		for d := startDate; d.Before(endDate) || d.Equal(endDate); d = d.AddDate(0, 0, 7) {
			month := d.Month()
			var seasonal float64
			switch {
			case month == 12 || month == 1 || month == 2:
				seasonal = 15 + rand.Float64()*10
			case month == 3 || month == 4 || month == 5:
				seasonal = 30 + rand.Float64()*15
			case month == 6 || month == 7 || month == 8:
				seasonal = -30 - rand.Float64()*15
			case month == 9 || month == 10 || month == 11:
				seasonal = -5 + rand.Float64()*10
			}

			weekly := (rand.Float64() - 0.5) * 8
			fillPct := math.Max(8, math.Min(100, baseFill+seasonal+weekly))
			volume := capacity * fillPct / 100
			variation := weekly

			_, err := db.Exec(`
				INSERT INTO readings (reservoir_id, source_id, observed_at, volume_hm3, capacity_hm3, fill_pct, weekly_variation_hm3, is_provisional, is_official)
				VALUES (?, 1, ?, ?, ?, ?, ?, 0, 1)
				ON CONFLICT(reservoir_id, observed_at, source_id) DO UPDATE SET
					volume_hm3 = excluded.volume_hm3,
					capacity_hm3 = excluded.capacity_hm3,
					fill_pct = excluded.fill_pct,
					weekly_variation_hm3 = excluded.weekly_variation_hm3
			`, r.id, d.Format("2006-01-02"), volume, capacity, fillPct, variation)
			if err != nil {
				return fmt.Errorf("insert reading for %s on %s: %w", r.name, d.Format("2006-01-02"), err)
			}
			totalCount++
		}
	}

	return nil
}
