package main

import (
	"archive/zip"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const (
	mitecoZIPURL = "https://www.miteco.gob.es/content/dam/miteco/es/agua/temas/evaluacion-de-los-recursos-hidricos/boletin-hidrologico/Historico-de-embalses/BD-Embalses.zip"
	mitecoTable  = "T_Datos Embalses 1988-2026"
)

// fetchMITECOHistorical downloads the MITECO historical BD-Embalses ZIP,
// exports the Access table to CSV, and upserts readings into SQLite.
func fetchMITECOHistorical(db *sql.DB) error {
	rawDir := "data/raw/miteco"
	if err := ensureDir(rawDir); err != nil {
		return fmt.Errorf("create raw dir: %w", err)
	}

	zipPath := filepath.Join(rawDir, "BD-Embalses.zip")
	mdbPath := filepath.Join(rawDir, "BD-Embalses.mdb")
	csvPath := filepath.Join(rawDir, "BD-Embalses.csv")

	// 1. Download ZIP if missing or older than 7 days.
	needDownload := true
	if info, err := os.Stat(zipPath); err == nil {
		if time.Since(info.ModTime()) < 7*24*time.Hour {
			needDownload = false
		}
	}
	if needDownload {
		log.Printf("Downloading MITECO historical data from %s", mitecoZIPURL)
		if err := downloadFile(zipPath, mitecoZIPURL); err != nil {
			return fmt.Errorf("download MITECO ZIP: %w", err)
		}
	} else {
		log.Printf("Using existing MITECO ZIP: %s", zipPath)
	}

	// 2. Unzip if MDB is missing or ZIP is newer.
	if needDownload || fileMissing(mdbPath) {
		log.Printf("Unzipping %s", zipPath)
		if err := unzipFile(zipPath, mdbPath); err != nil {
			return fmt.Errorf("unzip MITCO file: %w", err)
		}
	}

	// 3. Export MDB table to CSV if CSV is missing or MDB is newer.
	if needDownload || fileMissing(csvPath) || fileOlder(csvPath, mdbPath) {
		log.Printf("Exporting MDB table %q to CSV", mitecoTable)
		if err := exportMDBTableToCSV(mdbPath, csvPath, mitecoTable); err != nil {
			return fmt.Errorf("export MDB to CSV: %w", err)
		}
	}

	// 4. Parse CSV and upsert readings.
	log.Println("Importing MITECO historical readings into SQLite (this may take a few minutes)")
	records, err := parseMITECOCSV(csvPath)
	if err != nil {
		return fmt.Errorf("parse MITECO CSV: %w", err)
	}

	if err := upsertMITECOReadings(db, records); err != nil {
		return fmt.Errorf("upsert MITECO readings: %w", err)
	}

	return nil
}

// fetchMITECOWeekly is still a stub for the weekly PDF bulletin.
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

	// TODO: implement weekly PDF ingestion.
	_, _ = db.Exec(`
		INSERT INTO updater_state (source_name, last_fetch_date, last_fetch_status, records_count)
		VALUES ('MITECO_WEEKLY', ?, 'skipped', 0)
		ON CONFLICT(source_name) DO UPDATE SET
			last_fetch_date = excluded.last_fetch_date,
			last_fetch_status = excluded.last_fetch_status,
			records_count = excluded.records_count,
			updated_at = datetime('now')
	`, time.Now().Format("2006-01-02"))

	return fmt.Errorf("MITECO weekly fetch not yet implemented")
}

// mitecoRecord represents a single row from the MITECO historical CSV.
type mitecoRecord struct {
	BasinName     string
	ReservoirName string
	ObservedAt    string // YYYY-MM-DD
	CapacityHM3   float64
	VolumeHM3     float64
	FillPct       float64
	ElectricoFlag bool
}

func downloadFile(dst, url string) error {
	out, err := os.Create(dst + ".tmp")
	if err != nil {
		return err
	}
	defer os.Remove(dst + ".tmp")

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return os.Rename(dst+".tmp", dst)
}

func unzipFile(zipPath, dstPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if strings.HasSuffix(strings.ToLower(f.Name), ".mdb") {
			rc, err := f.Open()
			if err != nil {
				return err
			}
			defer rc.Close()

			out, err := os.Create(dstPath + ".tmp")
			if err != nil {
				return err
			}
			if _, err := io.Copy(out, rc); err != nil {
				out.Close()
				os.Remove(dstPath + ".tmp")
				return err
			}
			if err := out.Close(); err != nil {
				return err
			}
			return os.Rename(dstPath+".tmp", dstPath)
		}
	}
	return fmt.Errorf("no .mdb file found in %s", zipPath)
}

func exportMDBTableToCSV(mdbPath, csvPath, table string) error {
	// Prefer system mdb-export.
	if _, err := exec.LookPath("mdb-export"); err == nil {
		out, err := os.Create(csvPath + ".tmp")
		if err != nil {
			return err
		}
		defer os.Remove(csvPath + ".tmp")

		cmd := exec.Command("mdb-export", mdbPath, table)
		cmd.Stdout = out
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			out.Close()
			return fmt.Errorf("mdb-export: %w", err)
		}
		if err := out.Close(); err != nil {
			return err
		}
		return os.Rename(csvPath+".tmp", csvPath)
	}

	// Fallback to Docker with mdbtools.
	if _, err := exec.LookPath("docker"); err != nil {
		return fmt.Errorf("mdb-export not found and docker not available")
	}

	absMDB, err := filepath.Abs(mdbPath)
	if err != nil {
		return err
	}
	absCSV, err := filepath.Abs(csvPath)
	if err != nil {
		return err
	}
	dataDir := filepath.Dir(absMDB)
	csvName := filepath.Base(absCSV)

	cmd := exec.Command("docker", "run", "--rm",
		"-v", dataDir+":/data",
		"ubuntu:24.04",
		"bash", "-c",
		fmt.Sprintf("apt-get update -qq && apt-get install -y -qq mdbtools >/dev/null 2>&1 && mdb-export /data/%s %q > /data/%s.tmp && mv /data/%s.tmp /data/%s",
			filepath.Base(absMDB), table, csvName, csvName, csvName),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker mdb-export: %w", err)
	}
	return nil
}

func parseMITECOCSV(path string) ([]mitecoRecord, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// mdb-export uses the runtime locale for decimal separator; the MITECO file uses comma decimals.
	// Set LC_ALL=C so numbers are not reformatted.
	reader := csv.NewReader(f)
	reader.Comma = ','
	reader.LazyQuotes = true
	reader.ReuseRecord = true

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	colIndex := make(map[string]int, len(header))
	for i, h := range header {
		colIndex[strings.TrimSpace(h)] = i
	}
	required := []string{"AMBITO_NOMBRE", "EMBALSE_NOMBRE", "FECHA", "AGUA_TOTAL", "AGUA_ACTUAL", "ELECTRICO_FLAG"}
	for _, col := range required {
		if _, ok := colIndex[col]; !ok {
			return nil, fmt.Errorf("missing required column: %s", col)
		}
	}

	const expectedCols = 6
	var records []mitecoRecord
	line := 1
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("row %d: %w", line, err)
		}
		line++
		if len(row) < expectedCols {
			continue
		}

		observedAt, err := parseMITECODate(strings.TrimSpace(row[colIndex["FECHA"]]))
		if err != nil {
			continue // skip rows with bad dates
		}

		capacity, err := parseSpanishFloat(strings.TrimSpace(row[colIndex["AGUA_TOTAL"]]))
		if err != nil {
			continue
		}
		volume, err := parseSpanishFloat(strings.TrimSpace(row[colIndex["AGUA_ACTUAL"]]))
		if err != nil {
			continue
		}

		// Skip zero-capacity rows to avoid division by zero.
		if capacity <= 0 {
			continue
		}

		fillPct := (volume / capacity) * 100
		if fillPct < 0 {
			fillPct = 0
		}
		if fillPct > 100 {
			fillPct = 100
		}

		records = append(records, mitecoRecord{
			BasinName:     strings.TrimSpace(row[colIndex["AMBITO_NOMBRE"]]),
			ReservoirName: strings.TrimSpace(row[colIndex["EMBALSE_NOMBRE"]]),
			ObservedAt:    observedAt,
			CapacityHM3:   capacity,
			VolumeHM3:     volume,
			FillPct:       fillPct,
			ElectricoFlag: parseFlag(strings.TrimSpace(row[colIndex["ELECTRICO_FLAG"]])),
		})
	}

	return records, nil
}

func upsertMITECOReadings(db *sql.DB, records []mitecoRecord) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Ensure source exists.
	var sourceID int
	err = tx.QueryRow(`SELECT id FROM sources WHERE name = 'MITECO'`).Scan(&sourceID)
	if err != nil {
		if err == sql.ErrNoRows {
			res, err := tx.Exec(`INSERT INTO sources (name, organism, licence, attribution, url) VALUES (?, ?, ?, ?, ?)`,
				"MITECO",
				"Ministerio para la Transición Ecológica y el Reto Demográfico",
				"Ley 37/2007 + RD 1495/2011",
				"Fuente: Ministerio para la Transición Ecológica y el Reto Demográfico (MITECO)",
				"https://www.miteco.gob.es/es/agua/temas/evaluacion-de-los-recursos-hidricos/boletin-hidrologico.html",
			)
			if err != nil {
				return fmt.Errorf("insert MITECO source: %w", err)
			}
			id, _ := res.LastInsertId()
			sourceID = int(id)
		} else {
			return fmt.Errorf("lookup MITECO source: %w", err)
		}
	}

	// Build lookup maps.
	basinMap, err := loadBasinMap(tx)
	if err != nil {
		return fmt.Errorf("load basin map: %w", err)
	}
	reservoirMap, err := loadReservoirMap(tx)
	if err != nil {
		return fmt.Errorf("load reservoir map: %w", err)
	}

	// Prepared statements.
	insertBasin, err := tx.Prepare(`INSERT INTO basins (name, code) VALUES (?, ?) ON CONFLICT(name) DO UPDATE SET code=excluded.code RETURNING id`)
	if err != nil {
		return err
	}
	defer insertBasin.Close()

	insertReservoir, err := tx.Prepare(`
		INSERT INTO reservoirs (name, slug, basin_id, capacity_hm3, source_id)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(name) DO UPDATE SET
			slug = excluded.slug,
			basin_id = excluded.basin_id,
			capacity_hm3 = excluded.capacity_hm3,
			source_id = excluded.source_id
		RETURNING id
	`)
	if err != nil {
		return err
	}
	defer insertReservoir.Close()

	upsertReading, err := tx.Prepare(`
		INSERT INTO readings (reservoir_id, source_id, observed_at, volume_hm3, capacity_hm3, fill_pct, weekly_variation_hm3, is_provisional, is_official, fetched_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, 0, 1, datetime('now'))
		ON CONFLICT(reservoir_id, observed_at, source_id) DO UPDATE SET
			volume_hm3 = excluded.volume_hm3,
			capacity_hm3 = excluded.capacity_hm3,
			fill_pct = excluded.fill_pct,
			weekly_variation_hm3 = excluded.weekly_variation_hm3,
			fetched_at = excluded.fetched_at
	`)
	if err != nil {
		return err
	}
	defer upsertReading.Close()

	var latestDate string
	for i, rec := range records {
		// Basin.
		bNorm := normalizeBasinName(rec.BasinName)
		basinID, ok := basinMap[bNorm]
		if !ok {
			var id int
			if err := insertBasin.QueryRow(rec.BasinName, rec.BasinName).Scan(&id); err != nil {
				return fmt.Errorf("insert basin %q: %w", rec.BasinName, err)
			}
			basinID = id
			basinMap[bNorm] = id
		}

		// Reservoir.
		rNorm := normalizeReservoirName(rec.ReservoirName)
		reservoirID, ok := reservoirMap[rNorm]
		if !ok {
			slug := slugify(rec.ReservoirName)
			var id int
			if err := insertReservoir.QueryRow(rec.ReservoirName, slug, basinID, rec.CapacityHM3, sourceID).Scan(&id); err != nil {
				return fmt.Errorf("insert reservoir %q: %w", rec.ReservoirName, err)
			}
			reservoirID = id
			reservoirMap[rNorm] = id
		}

		// Weekly variation: difference vs previous week for same reservoir.
		// We don't have ordered per-reservoir data here; leave null.
		if _, err := upsertReading.Exec(reservoirID, sourceID, rec.ObservedAt, rec.VolumeHM3, rec.CapacityHM3, rec.FillPct, nil); err != nil {
			return fmt.Errorf("upsert reading for %s on %s: %w", rec.ReservoirName, rec.ObservedAt, err)
		}

		if rec.ObservedAt > latestDate {
			latestDate = rec.ObservedAt
		}

		if (i+1)%50000 == 0 {
			log.Printf("  ... imported %d / %d readings", i+1, len(records))
		}
	}

	// Update updater state.
	_, err = tx.Exec(`
		INSERT INTO updater_state (source_name, last_fetch_date, last_fetch_status, records_count, updated_at)
		VALUES ('MITECO_HISTORICAL', ?, 'ok', ?, datetime('now'))
		ON CONFLICT(source_name) DO UPDATE SET
			last_fetch_date = excluded.last_fetch_date,
			last_fetch_status = excluded.last_fetch_status,
			records_count = excluded.records_count,
			updated_at = excluded.updated_at
	`, latestDate, len(records))
	if err != nil {
		return fmt.Errorf("update updater_state: %w", err)
	}

	return tx.Commit()
}

func loadBasinMap(tx *sql.Tx) (map[string]int, error) {
	rows, err := tx.Query(`SELECT id, name FROM basins`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := make(map[string]int)
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		m[normalizeBasinName(name)] = id
	}
	return m, rows.Err()
}

func loadReservoirMap(tx *sql.Tx) (map[string]int, error) {
	rows, err := tx.Query(`SELECT id, name FROM reservoirs`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	m := make(map[string]int)
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, err
		}
		m[normalizeReservoirName(name)] = id
	}
	return m, rows.Err()
}

// normalizeBasinName returns a comparable key for basin names.
func normalizeBasinName(s string) string {
	s = removeAccents(strings.ToUpper(s))
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "  ", " ")
	return strings.TrimSpace(s)
}

// normalizeReservoirName returns a comparable key for reservoir names.
func normalizeReservoirName(s string) string {
	s = removeAccents(strings.ToLower(s))
	s = strings.TrimPrefix(s, "embalse de ")
	s = strings.TrimPrefix(s, "presa de ")
	s = strings.TrimPrefix(s, "pantano de ")
	s = strings.Trim(s, " ,")
	return s
}

func removeAccents(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	out, _, _ := transform.String(t, s)
	return out
}

func parseMITECODate(s string) (string, error) {
	// MITECO BD-Embalses uses MM/DD/YY (e.g. 06/23/26 = 23 Jun 2026).
	t, err := time.Parse("01/02/06 15:04:05", s)
	if err != nil {
		return "", err
	}
	return t.Format("2006-01-02"), nil
}

func parseSpanishFloat(s string) (float64, error) {
	// MITECO uses comma as decimal separator and no thousands separator.
	s = strings.ReplaceAll(s, ".", "") // safety
	s = strings.ReplaceAll(s, ",", ".")
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty number")
	}
	return strconv.ParseFloat(s, 64)
}

func parseFlag(s string) bool {
	return s == "1" || strings.EqualFold(s, "true") || strings.EqualFold(s, "yes")
}

func fileMissing(path string) bool {
	_, err := os.Stat(path)
	return os.IsNotExist(err)
}

func fileOlder(a, b string) bool {
	ai, err := os.Stat(a)
	if err != nil {
		return true
	}
	bi, err := os.Stat(b)
	if err != nil {
		return false
	}
	return ai.ModTime().Before(bi.ModTime())
}
