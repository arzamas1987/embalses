package snczi

import (
	"testing"
)

func TestParse(t *testing.T) {
	features, err := Parse("../../../test/fixtures/snczi_dams.geojson")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}
	if len(features) != 2 {
		t.Fatalf("expected 2 features, got %d", len(features))
	}

	// Check first dam (Mequinenza)
	f := features[0]
	if f.Name != "Presa de Mequinenza" {
		t.Errorf("expected 'Presa de Mequinenza', got %s", f.Name)
	}
	if f.Province != "Zaragoza" {
		t.Errorf("expected province Zaragoza, got %s", f.Province)
	}
	if f.Basin != "Ebro" {
		t.Errorf("expected basin Ebro, got %s", f.Basin)
	}
	if f.CapacityHM3 != 1534 {
		t.Errorf("expected capacity 1534, got %f", f.CapacityHM3)
	}
	if f.Lon == 0 || f.Lat == 0 {
		t.Error("expected non-zero coordinates")
	}

	// Check second dam (Sau)
	f2 := features[1]
	if f2.Name != "Presa de Sau" {
		t.Errorf("expected 'Presa de Sau', got %s", f2.Name)
	}
	if f2.Province != "Barcelona" {
		t.Errorf("expected province Barcelona, got %s", f2.Province)
	}
}

func TestParse_FileNotFound(t *testing.T) {
	_, err := Parse("nonexistent.geojson")
	if err == nil {
		t.Error("expected error for missing file")
	}
}
