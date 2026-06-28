package main

import (
	"database/sql"
	"fmt"
	"log"
)

// fetchSAIHAll checks all SAIH basins for new data.
// Incremental: only fetches data newer than the last stored reading per basin.
func fetchSAIHAll(db *sql.DB) error {
	basins := []struct {
		name string
		url  string
	}{
		{"Ebro", "https://www.chebro.es/"},
		{"Jucar", "https://aps.chj.es/down/html/descargas.html"},
		{"Guadalquivir", "https://www.chguadalquivir.es/saih/"},
		{"Guadiana", "https://www.saihguadiana.com"},
		{"Segura", "https://www.chsegura.es/es/cuenca/redes-de-control/saih/"},
		{"Duero", ""},
		{"Tajo", ""},
		{"Mino-Sil", ""},
		{"Cantabrico", ""},
	}

	for _, b := range basins {
		if b.url == "" {
			log.Printf("SAIH %s: URL unverified, skipping", b.name)
			continue
		}
		log.Printf("SAIH %s: checking for updates...", b.name)
		if err := fetchSAIHBasin(db, b.name, b.url); err != nil {
			log.Printf("  SAIH %s error: %v", b.name, err)
		}
	}
	return nil
}

// fetchSAIHBasin fetches real-time data for a single basin.
func fetchSAIHBasin(db *sql.DB, basinName, url string) error {
	var lastDate string
	err := db.QueryRow(`
		SELECT last_fetch_date FROM updater_state WHERE source_name = ?
	`, "SAIH_"+basinName).Scan(&lastDate)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("check updater state: %w", err)
	}
	if lastDate == "" {
		lastDate = "2024-01-01"
	}

	// TODO: Implement SAIH basin-specific fetch
	// Steps:
	// 1. HTTP GET or POST to the basin portal
	// 2. Parse HTML/JSON response (often requires session cookies)
	// 3. Extract reservoir readings (nivel, volumen, % fill)
	// 4. Insert into readings table
	// 5. Update updater_state

	_, _ = db.Exec(`
		INSERT INTO updater_state (source_name, last_fetch_date, last_fetch_status, records_count)
		VALUES (?, ?, 'skipped', 0)
		ON CONFLICT(source_name) DO UPDATE SET
			last_fetch_date = excluded.last_fetch_date,
			last_fetch_status = excluded.last_fetch_status,
			records_count = excluded.records_count,
			updated_at = datetime('now')
	`, "SAIH_"+basinName, "2024-01-01")

	return fmt.Errorf("SAIH %s fetch not yet implemented", basinName)
}
