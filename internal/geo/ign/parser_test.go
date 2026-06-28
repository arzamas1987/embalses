package ign

import (
	"testing"
)

func TestParseBasins(t *testing.T) {
	basins, err := ParseBasins("../../../test/fixtures/ign_basins.geojson")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(basins) != 2 {
		t.Fatalf("expected 2 basins, got %d", len(basins))
	}

	if basins[0].Name != "Ebro" {
		t.Errorf("expected 'Ebro', got %s", basins[0].Name)
	}
	if basins[0].Code != "ES-EBRO" {
		t.Errorf("expected code ES-EBRO, got %s", basins[0].Code)
	}
	if basins[0].WKT == "" {
		t.Error("expected non-empty WKT")
	}
}

func TestParseProvinces(t *testing.T) {
	provinces, err := ParseProvinces("../../../test/fixtures/ign_provinces.geojson")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(provinces) != 2 {
		t.Fatalf("expected 2 provinces, got %d", len(provinces))
	}

	if provinces[0].Name != "Zaragoza" {
		t.Errorf("expected 'Zaragoza', got %s", provinces[0].Name)
	}
	if provinces[0].Code != "50" {
		t.Errorf("expected code 50, got %s", provinces[0].Code)
	}
}

func TestParseReservoirs(t *testing.T) {
	reservoirs, err := ParseReservoirs("../../../test/fixtures/ign_reservoirs.geojson")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(reservoirs) != 2 {
		t.Fatalf("expected 2 reservoirs, got %d", len(reservoirs))
	}

	if reservoirs[0].Name != "Embalse de Mequinenza" {
		t.Errorf("expected 'Embalse de Mequinenza', got %s", reservoirs[0].Name)
	}
	if reservoirs[0].ExternalID != "IGN-MEQ-001" {
		t.Errorf("expected external_id IGN-MEQ-001, got %s", reservoirs[0].ExternalID)
	}
	if reservoirs[0].WKT == "" {
		t.Error("expected non-empty WKT")
	}
}
