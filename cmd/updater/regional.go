package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// fetchRegional imports readings from regional open-data sources:
// - ACA (Catalonia) daily reservoir levels via Socrata
// - CHD (Duero) current reservoir status via CKAN JSON
// - CHJ (Júcar) current reservoir status via HTML scraping
func fetchRegional(db *sql.DB) error {
	if err := fetchACA(db); err != nil {
		log.Printf("ACA fetch (non-fatal): %v", err)
	}
	if err := fetchCHD(db); err != nil {
		log.Printf("CHD fetch (non-fatal): %v", err)
	}
	if err := fetchCHJ(db); err != nil {
		log.Printf("CHJ fetch (non-fatal): %v", err)
	}
	return nil
}

// --- ACA (Catalonia) ---

const acaDatasetURL = "https://analisi.transparenciacatalunya.cat/api/views/gn9e-3qhr/rows.json?accessType=DOWNLOAD"

func fetchACA(db *sql.DB) error {
	log.Println("--- Fetching ACA (Catalonia) reservoir data ---")
	resp, err := http.Get(acaDatasetURL)
	if err != nil {
		return fmt.Errorf("download ACA dataset: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ACA HTTP %d", resp.StatusCode)
	}

	var payload struct {
		Meta struct {
			View struct {
				Columns []struct {
					Name string `json:"name"`
				} `json:"columns"`
			} `json:"view"`
		} `json:"meta"`
		Data [][]interface{} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return fmt.Errorf("decode ACA JSON: %w", err)
	}

	// Column positions after metadata columns (8 metadata columns before actual data).
	// Verified structure: [:sid, :id, :position, :created_at, :created_meta, :updated_at, :updated_meta, :meta, Dia, Estació, Nivell absolut, Percentatge, Volum]
	const metaCols = 8
	dateIdx := metaCols + 0
	stationIdx := metaCols + 1
	pctIdx := metaCols + 3
	volIdx := metaCols + 4

	latest := make(map[string]regionalRecord)
	for _, row := range payload.Data {
		if len(row) <= volIdx {
			continue
		}
		dateStr, ok := row[dateIdx].(string)
		if !ok || dateStr == "" {
			continue
		}
		station, _ := row[stationIdx].(string)
		pctVal := parseJSONNumber(row[pctIdx])
		volVal := parseJSONNumber(row[volIdx])
		if pctVal == nil || volVal == nil {
			continue
		}

		observedAt, err := time.Parse("2006-01-02T15:04:05", dateStr)
		if err != nil {
			continue
		}

		name := normalizeRegionalName(station)
		rec := regionalRecord{
			SourceName:    "ACA",
			ReservoirName: station,
			ObservedAt:    observedAt.Format("2006-01-02"),
			VolumeHM3:     *volVal,
			FillPct:       *pctVal,
		}
		if existing, ok := latest[name]; !ok || rec.ObservedAt > existing.ObservedAt {
			latest[name] = rec
		}
	}

	records := make(map[string][]regionalRecord, len(latest))
	for k, rec := range latest {
		records[k] = []regionalRecord{rec}
	}
	return upsertRegionalReadings(db, records)
}

// --- CHD (Duero) ---

const chdEstadoURL = "https://datos.chduero.es/dataset/f2eebe21-10eb-4b04-bf5b-71578dc3562c/resource/9b642c38-59b8-4ab0-8abd-6580ed64d261/download/estado_embalses.json"

func fetchCHD(db *sql.DB) error {
	log.Println("--- Fetching CHD (Duero) reservoir data ---")
	resp, err := http.Get(chdEstadoURL)
	if err != nil {
		return fmt.Errorf("download CHD dataset: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("CHD HTTP %d", resp.StatusCode)
	}

	var items []struct {
		PuntoControl         string `json:"punto_control"`
		CapacidadHM3         string `json:"capacidad_hm3"`
		VolumenActualHM3     string `json:"volumen_actual_hm3"`
		VolumenActualPercent string `json:"volumen_actual_percent"`
		Actualizacion        string `json:"actualizacion"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return fmt.Errorf("decode CHD JSON: %w", err)
	}

	records := make(map[string][]regionalRecord)
	for _, it := range items {
		observedAt, err := time.Parse("2006-01-02 15:04:05", it.Actualizacion)
		if err != nil {
			continue
		}
		// CHD numbers may use comma as thousands separator and dot as decimal (e.g. "2,258.462").
		capacityPtr := parseDecimalFloat(it.CapacidadHM3)
		volumePtr := parseDecimalFloat(it.VolumenActualHM3)
		fillPctPtr := parseDecimalFloat(it.VolumenActualPercent)
		if volumePtr == nil || fillPctPtr == nil || *volumePtr == 0 || *fillPctPtr == 0 {
			continue
		}

		name := normalizeRegionalName(it.PuntoControl)
		records[name] = append(records[name], regionalRecord{
			SourceName:    "CHD",
			ReservoirName: strings.TrimPrefix(it.PuntoControl, "Embalse de "),
			ObservedAt:    observedAt.Format("2006-01-02"),
			VolumeHM3:     *volumePtr,
			CapacityHM3:   capacityPtr,
			FillPct:       *fillPctPtr,
		})
	}

	return upsertRegionalReadings(db, records)
}

// --- CHJ (Júcar) ---

const chjEmbalsesURL = "https://saih.chj.es/embalses"

func fetchCHJ(db *sql.DB) error {
	log.Println("--- Fetching CHJ (Júcar) reservoir data ---")
	resp, err := http.Get(chjEmbalsesURL)
	if err != nil {
		return fmt.Errorf("download CHJ page: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("CHJ HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read CHJ page: %w", err)
	}
	html := string(body)

	// The page contains rows like:
	// <td class="text-nowrap pe-2">EMBALSE DE FORATA</td>
	// ... <td class="text-center text-nowrap">30-06-2026 12:40</td>
	// ... <td class="text-center">37,34</td>  (volume?)
	// ... <td class="text-center"> ... 49,56%</td>
	// We extract name, date, and percentage from each row.
	records := make(map[string][]regionalRecord)

	// Find all rows containing EMBALSE DE ...
	rowRe := regexp.MustCompile(`<tr[^>]*>.*?EMBALSE DE ([^<]+).*?</tr>`)
	for _, m := range rowRe.FindAllString(html, -1) {
		name := extractText(regexp.MustCompile(`EMBALSE DE ([^<]+)`), m)
		dateStr := extractText(regexp.MustCompile(`>(\d{2}-\d{2}-\d{4} \d{2}:\d{2})<`), m)
		pctStr := extractText(regexp.MustCompile(`>(\d{1,3},\d+)%<`), m)
		if name == "" || dateStr == "" || pctStr == "" {
			continue
		}

		observedAt, err := time.Parse("02-01-2006 15:04", dateStr)
		if err != nil {
			continue
		}
		fillPct, err := parseSpanishFloat(pctStr)
		if err != nil || fillPct == 0 {
			continue
		}

		key := normalizeRegionalName(name)
		records[key] = append(records[key], regionalRecord{
			SourceName:    "CHJ",
			ReservoirName: strings.TrimSpace(name),
			ObservedAt:    observedAt.Format("2006-01-02"),
			FillPct:       fillPct,
		})
	}

	return upsertRegionalReadings(db, records)
}

// --- Shared helpers ---

type regionalRecord struct {
	SourceName    string
	ReservoirName string
	ObservedAt    string
	VolumeHM3     float64
	CapacityHM3   *float64
	FillPct       float64
}

func upsertRegionalReadings(db *sql.DB, records map[string][]regionalRecord) error {
	if len(records) == 0 {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Ensure sources exist.
	sourceIDs := make(map[string]int)
	for _, name := range []string{"ACA", "CHD", "CHJ"} {
		var id int
		err := tx.QueryRow(`SELECT id FROM sources WHERE name = ?`, name).Scan(&id)
		if err == sql.ErrNoRows {
			res, err := tx.Exec(`INSERT INTO sources (name, organism, licence, attribution, url) VALUES (?, ?, ?, ?, ?)`,
				name,
				regionalSourceOrganism(name),
				"Datos abiertos",
				fmt.Sprintf("Fuente: %s", name),
				regionalSourceURL(name),
			)
			if err != nil {
				return fmt.Errorf("insert source %s: %w", name, err)
			}
			id64, _ := res.LastInsertId()
			id = int(id64)
		} else if err != nil {
			return fmt.Errorf("lookup source %s: %w", name, err)
		}
		sourceIDs[name] = id
	}

	// Build reservoir name/slug lookup.
	reservoirMap := make(map[string]int)
	rows, err := tx.Query(`SELECT id, name, slug FROM reservoirs`)
	if err != nil {
		return fmt.Errorf("query reservoirs: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		var name, slug string
		if err := rows.Scan(&id, &name, &slug); err != nil {
			return err
		}
		reservoirMap[normalizeRegionalName(name)] = id
		if slug != "" {
			reservoirMap[normalizeRegionalName(slug)] = id
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	upsertStmt, err := tx.Prepare(`
		INSERT INTO readings (reservoir_id, source_id, observed_at, volume_hm3, capacity_hm3, fill_pct, weekly_variation_hm3, is_provisional, is_official, fetched_at)
		VALUES (?, ?, ?, ?, ?, ?, NULL, 0, 1, datetime('now'))
		ON CONFLICT(reservoir_id, observed_at, source_id) DO UPDATE SET
			volume_hm3 = excluded.volume_hm3,
			capacity_hm3 = excluded.capacity_hm3,
			fill_pct = excluded.fill_pct,
			fetched_at = excluded.fetched_at
	`)
	if err != nil {
		return fmt.Errorf("prepare upsert: %w", err)
	}
	defer upsertStmt.Close()

	var total int
	latestDates := make(map[string]string)
	for key, recs := range records {
		reservoirID, ok := reservoirMap[key]
		if !ok {
			continue
		}
		for _, rec := range recs {
			sourceID := sourceIDs[rec.SourceName]
			var capacity interface{}
			if rec.CapacityHM3 != nil {
				capacity = *rec.CapacityHM3
			} else {
				capacity = nil
			}
			if _, err := upsertStmt.Exec(reservoirID, sourceID, rec.ObservedAt, rec.VolumeHM3, capacity, rec.FillPct); err != nil {
				return fmt.Errorf("upsert reading: %w", err)
			}
			total++
			if rec.ObservedAt > latestDates[rec.SourceName] {
				latestDates[rec.SourceName] = rec.ObservedAt
			}
		}
	}

	// Update updater_state.
	for sourceName, latest := range latestDates {
		_, _ = tx.Exec(`
			INSERT INTO updater_state (source_name, last_fetch_date, last_fetch_status, records_count, updated_at)
			VALUES (?, ?, 'ok', ?, datetime('now'))
			ON CONFLICT(source_name) DO UPDATE SET
				last_fetch_date = excluded.last_fetch_date,
				last_fetch_status = excluded.last_fetch_status,
				records_count = excluded.records_count,
				updated_at = excluded.updated_at
		`, sourceName, latest, len(records))
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	log.Printf("Regional import complete: %d readings inserted/updated", total)
	return nil
}

func regionalSourceOrganism(name string) string {
	switch name {
	case "ACA":
		return "Agència Catalana de l'Aigua"
	case "CHD":
		return "Confederación Hidrográfica del Duero"
	case "CHJ":
		return "Confederación Hidrográfica del Júcar"
	}
	return name
}

func regionalSourceURL(name string) string {
	switch name {
	case "ACA":
		return "https://analisi.transparenciacatalunya.cat/"
	case "CHD":
		return "https://datos.chduero.es/"
	case "CHJ":
		return "https://saih.chj.es/"
	}
	return ""
}

func normalizeRegionalName(s string) string {
	s = removeAccents(strings.ToLower(s))
	// Strip parenthetical location hints (e.g. "Embassament de Foix (Castellet i la Gornal)")
	s = regexp.MustCompile(`\([^)]*\)`).ReplaceAllString(s, " ")
	s = strings.ReplaceAll(s, "embalse de ", "")
	s = strings.ReplaceAll(s, "embassament de ", "")
	s = strings.ReplaceAll(s, "embalse del ", "")
	s = strings.ReplaceAll(s, "pantano de ", "")
	s = strings.ReplaceAll(s, "presa de ", "")
	s = strings.ReplaceAll(s, "embalse-de-", "")
	s = strings.ReplaceAll(s, "(", "")
	s = strings.ReplaceAll(s, ")", "")
	s = strings.ReplaceAll(s, ",", "")
	// Collapse multiple spaces created by removed fragments.
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

func parseJSONNumber(v interface{}) *float64 {
	switch n := v.(type) {
	case float64:
		return &n
	case string:
		f, err := strconv.ParseFloat(strings.ReplaceAll(n, ",", "."), 64)
		if err != nil {
			return nil
		}
		return &f
	}
	return nil
}

// parseDecimalFloat parses numbers that use dot as decimal separator and may
// use comma as thousands separator (e.g. CHD's "2,258.462").
func parseDecimalFloat(s string) *float64 {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "n/d") {
		return nil
	}
	// If both separators are present, keep the last one as the decimal mark
	// and remove the other (thousands) separator.
	hasDot := strings.Contains(s, ".")
	hasComma := strings.Contains(s, ",")
	if hasDot && hasComma {
		lastDot := strings.LastIndex(s, ".")
		lastComma := strings.LastIndex(s, ",")
		if lastComma > lastDot {
			// European: 1.234,56
			s = strings.ReplaceAll(s, ".", "")
			s = strings.ReplaceAll(s, ",", ".")
		} else {
			// Anglo: 2,258.462
			s = strings.ReplaceAll(s, ",", "")
		}
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &f
}

func extractText(re *regexp.Regexp, s string) string {
	m := re.FindStringSubmatch(s)
	if len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return ""
}
