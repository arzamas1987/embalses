package sqlite

import (
	"database/sql"
	"fmt"
)

// Migrate creates the SQLite schema. It is idempotent (uses IF NOT EXISTS).
func Migrate(db *sql.DB) error {
	for i, stmt := range schemaStatements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("schema stmt %d: %w", i, err)
		}
	}
	return nil
}

var schemaStatements = []string{
	// Sources (attribution / licence)
	`CREATE TABLE IF NOT EXISTS sources (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		organism TEXT NOT NULL,
		licence TEXT NOT NULL,
		attribution TEXT NOT NULL,
		url TEXT,
		last_fetched_at TEXT
	)`,

	// Basins (cuencas hidrográficas)
	`CREATE TABLE IF NOT EXISTS basins (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		code TEXT
	)`,

	// Provinces (with comunidad autónoma)
	`CREATE TABLE IF NOT EXISTS provinces (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		code TEXT,
		comunidad_autonoma TEXT
	)`,

	// Dams (SNCZI inventory + GPS)
	`CREATE TABLE IF NOT EXISTS dams (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		external_id TEXT,
		risk_category TEXT,
		river TEXT,
		municipality TEXT,
		province TEXT,
		basin TEXT,
		basin_id INTEGER REFERENCES basins(id),
		province_id INTEGER REFERENCES provinces(id),
		basin_area_km2 REAL,
		capacity_hm3 REAL,
		nmn_elevation REAL,
		dam_type TEXT,
		dam_height_m REAL,
		latitude REAL,
		longitude REAL,
		source_id INTEGER REFERENCES sources(id)
	)`,

	// Reservoirs (linked to dams for geometry, enhanced with GPS + slug)
	`CREATE TABLE IF NOT EXISTS reservoirs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE,
		external_id TEXT,
		slug TEXT NOT NULL UNIQUE,
		basin_id INTEGER REFERENCES basins(id),
		province_id INTEGER REFERENCES provinces(id),
		dam_id INTEGER REFERENCES dams(id),
		capacity_hm3 REAL,
		latitude REAL,
		longitude REAL,
		source_id INTEGER REFERENCES sources(id)
	)`,

	// Readings (time-series from MITECO / SAIH)
	`CREATE TABLE IF NOT EXISTS readings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		reservoir_id INTEGER NOT NULL REFERENCES reservoirs(id),
		source_id INTEGER REFERENCES sources(id),
		observed_at TEXT NOT NULL,
		volume_hm3 REAL,
		capacity_hm3 REAL,
		fill_pct REAL,
		weekly_variation_hm3 REAL,
		is_provisional INTEGER DEFAULT 0,
		is_official INTEGER DEFAULT 1,
		published_at TEXT,
		fetched_at TEXT DEFAULT CURRENT_TIMESTAMP,
		UNIQUE(reservoir_id, observed_at, source_id)
	)`,

	// Indexes for readings
	`CREATE INDEX IF NOT EXISTS idx_readings_reservoir ON readings(reservoir_id)`,
	`CREATE INDEX IF NOT EXISTS idx_readings_observed ON readings(observed_at)`,
	`CREATE INDEX IF NOT EXISTS idx_readings_reservoir_date ON readings(reservoir_id, observed_at)`,

	// API Keys
	`CREATE TABLE IF NOT EXISTS api_keys (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		key_hash TEXT NOT NULL UNIQUE,
		name TEXT NOT NULL,
		tier TEXT NOT NULL DEFAULT 'free',
		daily_quota INTEGER NOT NULL DEFAULT 100,
		rate_limit_per_minute INTEGER NOT NULL DEFAULT 60,
		is_active INTEGER DEFAULT 1,
		created_at TEXT DEFAULT CURRENT_TIMESTAMP,
		expires_at TEXT
	)`,

	// Metering
	`CREATE TABLE IF NOT EXISTS metering (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		api_key_id INTEGER REFERENCES api_keys(id),
		endpoint TEXT NOT NULL,
		method TEXT NOT NULL,
		status_code INTEGER,
		response_time_ms INTEGER,
		created_at TEXT DEFAULT CURRENT_TIMESTAMP
	)`,

	// Updater state (incremental fetch tracking)
	`CREATE TABLE IF NOT EXISTS updater_state (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		source_name TEXT NOT NULL UNIQUE,
		last_fetch_date TEXT,
		last_fetch_status TEXT,
		records_count INTEGER DEFAULT 0,
		updated_at TEXT DEFAULT CURRENT_TIMESTAMP
	)`,
}
