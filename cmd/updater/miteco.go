package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// fetchMITECOHistorical imports historical MITECO data from local Excel files.
// Expected: data/raw/miteco/BD-Embalses_*.xlsx or *.csv files.
func fetchMITECOHistorical(db *sql.DB) error {
	rawDir := "data/raw/miteco"
	if _, err := os.Stat(rawDir); os.IsNotExist(err) {
		return fmt.Errorf("raw data directory not found: %s", rawDir)
	}

	files, err := filepath.Glob(filepath.Join(rawDir, "*.xlsx"))
	if err != nil {
		return fmt.Errorf("glob xlsx: %w", err)
	}
	csvFiles, _ := filepath.Glob(filepath.Join(rawDir, "*.csv"))
	files = append(files, csvFiles...)

	if len(files) == 0 {
		return fmt.Errorf("no Excel/CSV files found in %s", rawDir)
	}

	log.Printf("Found %d MITECO historical file(s) to import", len(files))
	for _, f := range files {
		log.Printf("  - %s", f)
		// TODO: parse Excel/CSV and insert into readings table
		// This requires an Excel parser (github.com/xuri/excelize/v2) or CSV parser
	}
	return fmt.Errorf("MITECO historical import not yet implemented — see TODO in cmd/updater/miteco.go")
}

// fetchMITECOWeekly fetches the latest MITECO weekly bulletin (PDF or HTML).
// Incremental: only fetches if the latest bulletin date is newer than the last stored reading.
func fetchMITECOWeekly(db *sql.DB) error {
	var lastDate string
	err := db.QueryRow(`SELECT last_fetch_date FROM updater_state WHERE source_name = 'MITECO_WEEKLY'`).Scan(&lastDate)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("check updater state: %w", err)
	}
	if lastDate == "" {
		lastDate = "2024-01-01"
	}
	log.Printf("MITECO weekly: last fetch was %s", lastDate)

	// TODO: Fetch MITECO bulletin page, discover latest PDF URL, parse it
	// This requires:
	// 1. HTTP fetch of https://www.miteco.gob.es/es/agua/temas/evaluacion-de-los-recursos-hidricos/boletin-hidrologico.html
	// 2. Extract latest PDF link
	// 3. Download PDF
	// 4. Parse PDF (github.com/ledongthuc/pdf or similar)
	// 5. Insert readings into database

	// Update state
	_, _ = db.Exec(`
		INSERT INTO updater_state (source_name, last_fetch_date, last_fetch_status, records_count)
		VALUES ('MITECO_WEEKLY', ?, 'skipped', 0)
		ON CONFLICT(source_name) DO UPDATE SET
			last_fetch_date = excluded.last_fetch_date,
			last_fetch_status = excluded.last_fetch_status,
			records_count = excluded.records_count,
			updated_at = datetime('now')
	`, time.Now().Format("2006-01-02"))

	return fmt.Errorf("MITECO weekly fetch not yet implemented — see TODO in cmd/updater/miteco.go")
}
