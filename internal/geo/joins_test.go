package geo

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/arzamas1987/embalses/internal/db"
	"github.com/arzamas1987/embalses/internal/geo/ign"
	"github.com/arzamas1987/embalses/internal/geo/snczi"
)

func testDB(t *testing.T) *db.Pool {
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/embalses?sslmode=disable"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := db.New(ctx, connStr)
	if err != nil {
		t.Skipf("database not available: %v", err)
	}
	return pool
}

func TestRunSpatialJoins(t *testing.T) {
	pool := testDB(t)
	defer pool.Close()
	ctx := context.Background()

	// Clean up tables from previous test runs
	_, _ = pool.Exec(ctx, `
		DROP TABLE IF EXISTS reservoirs CASCADE;
		DROP TABLE IF EXISTS dams CASCADE;
		DROP TABLE IF EXISTS provinces CASCADE;
		DROP TABLE IF EXISTS basins CASCADE;
		DROP TABLE IF EXISTS sources CASCADE;
	`)

	// Run the geo schema migration
	schemaSQL, err := os.ReadFile("../../migrations/000002_geo_schema.up.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	_, err = pool.Exec(ctx, string(schemaSQL))
	if err != nil {
		t.Fatalf("run migration: %v", err)
	}

	// Ingest SNCZI dams
	dams, err := snczi.Parse("../../test/fixtures/snczi_dams.geojson")
	if err != nil {
		t.Fatalf("parse snczi: %v", err)
	}
	if err := snczi.Ingest(ctx, pool.Pool, dams); err != nil {
		t.Fatalf("ingest snczi: %v", err)
	}

	// Ingest IGN data
	basins, err := ign.ParseBasins("../../test/fixtures/ign_basins.geojson")
	if err != nil {
		t.Fatalf("parse basins: %v", err)
	}
	provinces, err := ign.ParseProvinces("../../test/fixtures/ign_provinces.geojson")
	if err != nil {
		t.Fatalf("parse provinces: %v", err)
	}
	reservoirs, err := ign.ParseReservoirs("../../test/fixtures/ign_reservoirs.geojson")
	if err != nil {
		t.Fatalf("parse reservoirs: %v", err)
	}
	if err := ign.Ingest(ctx, pool.Pool, basins, provinces, reservoirs); err != nil {
		t.Fatalf("ingest ign: %v", err)
	}

	// Run spatial joins
	if err := RunSpatialJoins(ctx, pool.Pool); err != nil {
		t.Fatalf("spatial joins: %v", err)
	}

	// Verify: Mequinenza reservoir should be in Ebro basin and Zaragoza province
	info, err := GetReservoirWithSpatialInfo(ctx, pool.Pool, "Embalse de Mequinenza")
	if err != nil {
		t.Fatalf("get spatial info: %v", err)
	}
	if info.BasinName != "Ebro" {
		t.Errorf("expected basin 'Ebro', got '%s'", info.BasinName)
	}
	if info.ProvinceName != "Zaragoza" {
		t.Errorf("expected province 'Zaragoza', got '%s'", info.ProvinceName)
	}
	if info.SourceName != "IGN" {
		t.Errorf("expected source 'IGN', got '%s'", info.SourceName)
	}
	if info.Attribution == "" {
		t.Error("expected non-empty attribution")
	}
}

func TestSourceAttribution(t *testing.T) {
	pool := testDB(t)
	defer pool.Close()
	ctx := context.Background()

	// Clean up tables from previous test runs
	_, _ = pool.Exec(ctx, `
		DROP TABLE IF EXISTS reservoirs CASCADE;
		DROP TABLE IF EXISTS dams CASCADE;
		DROP TABLE IF EXISTS provinces CASCADE;
		DROP TABLE IF EXISTS basins CASCADE;
		DROP TABLE IF EXISTS sources CASCADE;
	`)

	// Run the geo schema migration (includes source seeding)
	schemaSQL, err := os.ReadFile("../../migrations/000002_geo_schema.up.sql")
	if err != nil {
		t.Fatalf("read migration: %v", err)
	}
	_, err = pool.Exec(ctx, string(schemaSQL))
	if err != nil {
		t.Fatalf("run migration: %v", err)
	}

	var attribution string
	err = pool.QueryRow(ctx, "SELECT attribution FROM sources WHERE name = 'IGN'").Scan(&attribution)
	if err != nil {
		t.Fatalf("query IGN attribution: %v", err)
	}
	if attribution == "" {
		t.Error("expected non-empty IGN attribution")
	}
	if attribution != "CC-BY 4.0 scne.es / © Instituto Geográfico Nacional" {
		t.Errorf("unexpected attribution: %s", attribution)
	}

	var sncziAttribution string
	err = pool.QueryRow(ctx, "SELECT attribution FROM sources WHERE name = 'SNCZI'").Scan(&sncziAttribution)
	if err != nil {
		t.Fatalf("query SNCZI attribution: %v", err)
	}
	if sncziAttribution == "" {
		t.Error("expected non-empty SNCZI attribution")
	}
}
