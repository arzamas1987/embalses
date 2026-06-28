package snczi

import (
	"encoding/json"
	"fmt"
	"os"
)

// DamFeature represents a parsed SNCZI dam record.
type DamFeature struct {
	Name         string
	ExternalID   string
	Province     string
	Basin        string
	RiskCategory string
	River        string
	Municipality string
	Lat          float64
	Lon          float64
	BasinAreaKm2 float64
	CapacityHM3  float64
	NMNElevation float64
	DamType      string
	DamHeightM   float64
}

// GeoJSON structures for parsing.
type geoJSONFeatureCollection struct {
	Type     string           `json:"type"`
	Features []geoJSONFeature `json:"features"`
}

type geoJSONFeature struct {
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
	Geometry   geoJSONGeometry        `json:"geometry"`
}

type geoJSONGeometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

// Parse reads a GeoJSON file and returns dam features.
func Parse(path string) ([]DamFeature, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var fc geoJSONFeatureCollection
	if err := json.Unmarshal(data, &fc); err != nil {
		return nil, fmt.Errorf("unmarshal geojson: %w", err)
	}

	var features []DamFeature
	for _, f := range fc.Features {
		props := f.Properties
		coords := f.Geometry.Coordinates
		if len(coords) < 2 {
			continue
		}

		feat := DamFeature{
			Name:         stringProp(props, "name"),
			ExternalID:   stringProp(props, "external_id"),
			Province:     stringProp(props, "province"),
			Basin:        stringProp(props, "basin"),
			RiskCategory: stringProp(props, "risk_category"),
			River:        stringProp(props, "river"),
			Municipality: stringProp(props, "municipality"),
			Lon:          coords[0],
			Lat:          coords[1],
			BasinAreaKm2: floatProp(props, "basin_area_km2"),
			CapacityHM3:  floatProp(props, "capacity_hm3"),
			NMNElevation: floatProp(props, "nmn_elevation"),
			DamType:      stringProp(props, "dam_type"),
			DamHeightM:   floatProp(props, "dam_height_m"),
		}
		features = append(features, feat)
	}

	return features, nil
}

func stringProp(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func floatProp(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case int:
			return float64(val)
		case int64:
			return float64(val)
		}
	}
	return 0
}
