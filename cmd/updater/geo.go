package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/arzamas1987/embalses/internal/geo/snczi"
)

// GeoJSONFeatureCollection mirrors the GeoJSON structure
type GeoJSONFeatureCollection struct {
	Type     string           `json:"type"`
	Features []GeoJSONFeature `json:"features"`
}

type GeoJSONFeature struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Geometry   struct {
		Type        string      `json:"type"`
		Coordinates interface{} `json:"coordinates"`
	} `json:"geometry"`
}

// importGeoJSONFixtures imports SNCZI and IGN GeoJSON into SQLite.
func importGeoJSONFixtures(db *sql.DB) error {
	// 1. Seed sources
	if err := seedSources(db); err != nil {
		return fmt.Errorf("seed sources: %w", err)
	}

	// 2. Import SNCZI dams (gives us GPS, basin, province, capacity, etc.)
	if err := importSNCZI(db); err != nil {
		return fmt.Errorf("import SNCZI: %w", err)
	}

	// 3. Import IGN reservoir polygons (gives us WGS84 polygons, used for centroid if needed)
	if err := importIGN(db); err != nil {
		return fmt.Errorf("import IGN: %w", err)
	}

	return nil
}

func seedSources(db *sql.DB) error {
	sources := []struct {
		name, organism, licence, attribution, url string
	}{
		{"MITECO", "Ministerio para la Transicion Ecologica y el Reto Demografico", "Ley 37/2007 + RD 1495/2011", "Fuente: MITECO", "https://www.miteco.gob.es"},
		{"SNCZI", "Subdireccion General de Dominio Publico Hidraulico e Infraestructuras", "Ley 37/2007 + RD 1495/2011", "SNCZI - Inventario de Presas y Embalses, MITECO", "https://www.miteco.gob.es/es/cartografia-y-sig/ide/descargas/agua/inventario-presas-embalses.html"},
		{"IGN", "Instituto Geografico Nacional (IGN)", "CC-BY 4.0", "CC-BY 4.0 scne.es / (c) Instituto Geografico Nacional", "https://www.ign.es"},
	}
	for _, s := range sources {
		_, err := db.Exec(`
			INSERT INTO sources (name, organism, licence, attribution, url)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT(name) DO UPDATE SET
				organism = excluded.organism,
				licence = excluded.licence,
				attribution = excluded.attribution,
				url = excluded.url
		`, s.name, s.organism, s.licence, s.attribution, s.url)
		if err != nil {
			return err
		}
	}
	return nil
}

func importSNCZI(db *sql.DB) error {
	// Clear existing test fixtures to avoid slug conflicts with real data
	if _, err := db.Exec(`DELETE FROM readings`); err != nil {
		return fmt.Errorf("clear readings: %w", err)
	}
	if _, err := db.Exec(`DELETE FROM reservoirs`); err != nil {
		return fmt.Errorf("clear reservoirs: %w", err)
	}
	if _, err := db.Exec(`DELETE FROM dams`); err != nil {
		return fmt.Errorf("clear dams: %w", err)
	}
	if _, err := db.Exec(`DELETE FROM basins`); err != nil {
		return fmt.Errorf("clear basins: %w", err)
	}
	if _, err := db.Exec(`DELETE FROM provinces`); err != nil {
		return fmt.Errorf("clear provinces: %w", err)
	}
	log.Println("Cleared existing test data.")

	features, err := snczi.Parse("data/snczi_reservoirs.geojson")
	if err != nil {
		return fmt.Errorf("parse SNCZI: %w", err)
	}

	if len(features) == 0 {
		return fmt.Errorf("no SNCZI features found")
	}

	for _, f := range features {
		// Derive comunidad autonoma from province
		ca := provinceToCA(f.Province)

		// Insert basin
		var basinID int
		err := db.QueryRow(`
			INSERT INTO basins (name, code) VALUES (?, ?)
			ON CONFLICT(name) DO UPDATE SET code = excluded.code
			RETURNING id
		`, f.Basin, f.Basin).Scan(&basinID)
		if err != nil {
			return fmt.Errorf("insert basin %s: %w", f.Basin, err)
		}

		// Insert province
		var provinceID int
		err = db.QueryRow(`
			INSERT INTO provinces (name, comunidad_autonoma) VALUES (?, ?)
			ON CONFLICT(name) DO UPDATE SET comunidad_autonoma = excluded.comunidad_autonoma
			RETURNING id
		`, f.Province, ca).Scan(&provinceID)
		if err != nil {
			return fmt.Errorf("insert province %s: %w", f.Province, err)
		}

		// Convert UTM30 to lat/lng (simplified: treat as already WGS84 for fixtures)
		// In production, real SNCZI data is in ETRS89 UTM30 and needs reprojection.
		lat, lon := f.Lat, f.Lon
		if lat > 90 || lon > 180 || lat == 0 && lon == 0 {
			// Likely UTM30 meters; use centroid from IGN polygon instead
			lat, lon = 0, 0
		}

		// Insert dam (avoid double prefixing)
		damName := f.Name
		if !strings.HasPrefix(damName, "Presa de ") {
			damName = "Presa de " + damName
		}

		_, err = db.Exec(`
			INSERT INTO dams (name, external_id, risk_category, river, municipality, province, basin,
				basin_id, province_id, basin_area_km2, capacity_hm3, nmn_elevation, dam_type, dam_height_m,
				latitude, longitude, source_id)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(name) DO UPDATE SET
				external_id = excluded.external_id,
				risk_category = excluded.risk_category,
				river = excluded.river,
				municipality = excluded.municipality,
				province = excluded.province,
				basin = excluded.basin,
				basin_id = excluded.basin_id,
				province_id = excluded.province_id,
				basin_area_km2 = excluded.basin_area_km2,
				capacity_hm3 = excluded.capacity_hm3,
				nmn_elevation = excluded.nmn_elevation,
				dam_type = excluded.dam_type,
				dam_height_m = excluded.dam_height_m,
				latitude = excluded.latitude,
				longitude = excluded.longitude
		`, damName, f.ExternalID, f.RiskCategory, f.River, f.Municipality, f.Province, f.Basin,
			basinID, provinceID, f.BasinAreaKm2, f.CapacityHM3, f.NMNElevation, f.DamType, f.DamHeightM,
			lat, lon, 2) // source_id 2 = SNCZI
		if err != nil {
			return fmt.Errorf("insert dam %s: %w", f.Name, err)
		}

		// Generate reservoir name (avoid double prefixing)
		reservoirName := f.Name
		if strings.HasPrefix(reservoirName, "Presa de ") {
			reservoirName = "Embalse de " + strings.TrimPrefix(reservoirName, "Presa de ")
		} else if !strings.HasPrefix(reservoirName, "Embalse de ") {
			reservoirName = "Embalse de " + reservoirName
		}
		reservoirSlug := slugify(reservoirName)

		// Insert reservoir (linked to dam)
		_, err = db.Exec(`
			INSERT INTO reservoirs (name, external_id, slug, basin_id, province_id, dam_id, capacity_hm3, latitude, longitude, source_id)
			SELECT ?, ?, ?, ?, ?, d.id, ?, COALESCE(?, d.latitude), COALESCE(?, d.longitude), ?
			FROM dams d WHERE d.name = ?
			ON CONFLICT(name) DO UPDATE SET
				external_id = excluded.external_id,
				slug = excluded.slug,
				basin_id = excluded.basin_id,
				province_id = excluded.province_id,
				dam_id = excluded.dam_id,
				capacity_hm3 = excluded.capacity_hm3,
				latitude = excluded.latitude,
				longitude = excluded.longitude
		`, reservoirName, f.ExternalID, reservoirSlug, basinID, provinceID,
			f.CapacityHM3, lat, lon, 2, damName)
		if err != nil {
			return fmt.Errorf("insert reservoir %s: %w", reservoirName, err)
		}
	}
	return nil
}

func importIGN(db *sql.DB) error {
	data, err := os.ReadFile("test/fixtures/ign_reservoirs.geojson")
	if err != nil {
		return fmt.Errorf("read IGN: %w", err)
	}

	var fc GeoJSONFeatureCollection
	if err := json.Unmarshal(data, &fc); err != nil {
		return fmt.Errorf("parse IGN: %w", err)
	}

	for _, f := range fc.Features {
		name := stringProp(f.Properties, "name")
		if name == "" {
			continue
		}

		// Compute centroid from polygon
		lat, lon := centroidFromPolygon(f.Geometry.Coordinates)
		if lat == 0 && lon == 0 {
			continue
		}

		// Update reservoir with IGN coordinates if present
		reservoirName := name
		_, _ = db.Exec(`
			UPDATE reservoirs
			SET latitude = ?, longitude = ?, source_id = 2
			WHERE name = ?
		`, lat, lon, reservoirName)
	}
	return nil
}

func slugify(name string) string {
	s := strings.ToLower(name)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "ñ", "n")
	s = strings.ReplaceAll(s, "á", "a")
	s = strings.ReplaceAll(s, "é", "e")
	s = strings.ReplaceAll(s, "í", "i")
	s = strings.ReplaceAll(s, "ó", "o")
	s = strings.ReplaceAll(s, "ú", "u")
	s = strings.ReplaceAll(s, "ü", "u")
	s = strings.ReplaceAll(s, "ç", "c")
	// Remove any non-alphanumeric except dash
	var result []rune
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result = append(result, r)
		}
	}
	return string(result)
}

func stringProp(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func centroidFromPolygon(coords interface{}) (lat, lon float64) {
	// Simple polygon centroid from first ring
	rings, ok := coords.([]interface{})
	if !ok || len(rings) == 0 {
		return 0, 0
	}
	firstRing, ok := rings[0].([]interface{})
	if !ok || len(firstRing) == 0 {
		return 0, 0
	}
	var sumLon, sumLat float64
	var count int
	for _, pt := range firstRing {
		pair, ok := pt.([]interface{})
		if !ok || len(pair) < 2 {
			continue
		}
		lon, ok1 := toFloat64(pair[0])
		lat, ok2 := toFloat64(pair[1])
		if ok1 && ok2 {
			sumLon += lon
			sumLat += lat
			count++
		}
	}
	if count > 0 {
		return sumLat / float64(count), sumLon / float64(count)
	}
	return 0, 0
}

func toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case json.Number:
		f, err := val.Float64()
		return f, err == nil
	}
	return 0, false
}
