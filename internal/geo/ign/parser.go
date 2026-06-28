package ign

import (
	"encoding/json"
	"fmt"
	"os"
)

// BasinFeature represents a parsed IGN basin.
type BasinFeature struct {
	Name string
	Code string
	WKT  string
}

// ProvinceFeature represents a parsed IGN province.
type ProvinceFeature struct {
	Name string
	Code string
	WKT  string
}

// ReservoirFeature represents a parsed IGN reservoir polygon.
type ReservoirFeature struct {
	Name       string
	ExternalID string
	WKT        string
}

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
	Type        string      `json:"type"`
	Coordinates interface{} `json:"coordinates"`
}

func ParseBasins(path string) ([]BasinFeature, error) {
	fc, err := readGeoJSON(path)
	if err != nil {
		return nil, err
	}
	var basins []BasinFeature
	for _, f := range fc.Features {
		wkt, err := geometryToWKT(f.Geometry)
		if err != nil {
			continue
		}
		basins = append(basins, BasinFeature{
			Name: stringProp(f.Properties, "name"),
			Code: stringProp(f.Properties, "code"),
			WKT:  wkt,
		})
	}
	return basins, nil
}

func ParseProvinces(path string) ([]ProvinceFeature, error) {
	fc, err := readGeoJSON(path)
	if err != nil {
		return nil, err
	}
	var provinces []ProvinceFeature
	for _, f := range fc.Features {
		wkt, err := geometryToWKT(f.Geometry)
		if err != nil {
			continue
		}
		provinces = append(provinces, ProvinceFeature{
			Name: stringProp(f.Properties, "name"),
			Code: stringProp(f.Properties, "code"),
			WKT:  wkt,
		})
	}
	return provinces, nil
}

func ParseReservoirs(path string) ([]ReservoirFeature, error) {
	fc, err := readGeoJSON(path)
	if err != nil {
		return nil, err
	}
	var reservoirs []ReservoirFeature
	for _, f := range fc.Features {
		wkt, err := geometryToWKT(f.Geometry)
		if err != nil {
			continue
		}
		reservoirs = append(reservoirs, ReservoirFeature{
			Name:       stringProp(f.Properties, "name"),
			ExternalID: stringProp(f.Properties, "external_id"),
			WKT:        wkt,
		})
	}
	return reservoirs, nil
}

func readGeoJSON(path string) (*geoJSONFeatureCollection, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	var fc geoJSONFeatureCollection
	if err := json.Unmarshal(data, &fc); err != nil {
		return nil, fmt.Errorf("unmarshal geojson: %w", err)
	}
	return &fc, nil
}

func stringProp(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func geometryToWKT(g geoJSONGeometry) (string, error) {
	coords := g.Coordinates
	if coords == nil {
		return "", fmt.Errorf("nil coordinates")
	}
	switch g.Type {
	case "Polygon":
		return polygonToWKT(coords)
	case "MultiPolygon":
		return multiPolygonToWKT(coords)
	default:
		return "", fmt.Errorf("unsupported geometry type: %s", g.Type)
	}
}

func polygonToWKT(coords interface{}) (string, error) {
	rings, ok := coords.([]interface{})
	if !ok {
		return "", fmt.Errorf("invalid polygon coordinates")
	}
	wkt := "POLYGON("
	for ri, ring := range rings {
		if ri > 0 {
			wkt += ","
		}
		points, ok := ring.([]interface{})
		if !ok {
			return "", fmt.Errorf("invalid ring")
		}
		wkt += "("
		for pi, p := range points {
			if pi > 0 {
				wkt += ","
			}
			pt, ok := p.([]interface{})
			if !ok || len(pt) < 2 {
				return "", fmt.Errorf("invalid point")
			}
			wkt += fmt.Sprintf("%v %v", pt[0], pt[1])
		}
		wkt += ")"
	}
	wkt += ")"
	return wkt, nil
}

func multiPolygonToWKT(coords interface{}) (string, error) {
	polys, ok := coords.([]interface{})
	if !ok {
		return "", fmt.Errorf("invalid multipolygon coordinates")
	}
	wkt := "MULTIPOLYGON("
	for pi, poly := range polys {
		if pi > 0 {
			wkt += ","
		}
		wkt += "("
		rings, ok := poly.([]interface{})
		if !ok {
			return "", fmt.Errorf("invalid polygon in multipolygon")
		}
		for ri, ring := range rings {
			if ri > 0 {
				wkt += ","
			}
			points, ok := ring.([]interface{})
			if !ok {
				return "", fmt.Errorf("invalid ring in multipolygon")
			}
			wkt += "("
			for pti, p := range points {
				if pti > 0 {
					wkt += ","
				}
				pt, ok := p.([]interface{})
				if !ok || len(pt) < 2 {
					return "", fmt.Errorf("invalid point in multipolygon")
				}
				wkt += fmt.Sprintf("%v %v", pt[0], pt[1])
			}
			wkt += ")"
		}
		wkt += ")"
	}
	wkt += ")"
	return wkt, nil
}
