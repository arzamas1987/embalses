package main

import (
	"context"
	"flag"
	"log"
	"os"
	"time"

	"github.com/arzamas1987/embalses/internal/storage/sqlite"
)

func main() {
	var (
		dbPath         = flag.String("db", "data/embalses.db", "Path to SQLite database")
		fullImport     = flag.Bool("full", false, "Run full historical import (one-time)")
		mitecoImport   = flag.Bool("miteco", false, "Import MITECO historical reservoir data (BD-Embales.zip)")
		regionalImport = flag.Bool("regional", false, "Import regional open-data sources (ACA, CHD, CHJ)")
		geoOnly        = flag.Bool("geo-only", false, "Import only GeoJSON fixtures (GPS + metadata)")
		seedReadings   = flag.Bool("seed-readings", false, "Seed synthetic 6-month readings")
	)
	flag.Parse()

	log.Println("=== Embalses Updater ===")
	log.Printf("Database: %s", *dbPath)

	timeout := 30 * time.Minute
	if *mitecoImport || *fullImport || *regionalImport {
		timeout = 2 * time.Hour
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Open or create database
	db, err := sqlite.Open(ctx, *dbPath)
	if err != nil {
		log.Fatalf("Open database: %v", err)
	}
	defer db.Close()

	// Migrate schema
	if err := sqlite.Migrate(db.DB); err != nil {
		log.Fatalf("Migrate: %v", err)
	}
	log.Println("Schema ready.")

	// ── Import GeoJSON fixtures (SNCZI + IGN) ──
	if *geoOnly || *fullImport {
		log.Println("--- Importing GeoJSON fixtures ---")
		if err := importGeoJSONFixtures(db.DB); err != nil {
			log.Fatalf("GeoJSON import: %v", err)
		}
		log.Println("GeoJSON fixtures imported.")
	}

	// ── Seed synthetic readings (for UI testing without real data) ──
	if *seedReadings {
		log.Println("--- Seeding synthetic 6-month readings ---")
		if err := seedSyntheticReadings(db.DB); err != nil {
			log.Fatalf("Synthetic seed: %v", err)
		}
		log.Println("Synthetic readings seeded.")
	}

	// ── MITECO Historical Data (one-time full import) ──
	if *mitecoImport || *fullImport {
		log.Println("--- Fetching MITECO historical data ---")
		if err := fetchMITECOHistorical(db.DB); err != nil {
			log.Fatalf("MITECO historical fetch: %v", err)
		}
		log.Println("MITECO historical data imported.")
	}

	// ── Regional open-data sources (ACA, CHD, CHJ) ──
	if *regionalImport || *fullImport {
		log.Println("--- Fetching regional open-data sources ---")
		if err := fetchRegional(db.DB); err != nil {
			log.Fatalf("Regional fetch: %v", err)
		}
		log.Println("Regional data imported.")
	}

	// ── MITECO Weekly Updates (incremental) ──
	log.Println("--- Checking MITECO weekly bulletin ---")
	if err := fetchMITECOWeekly(db.DB); err != nil {
		log.Printf("MITECO weekly fetch (non-fatal): %v", err)
		log.Println("NOTE: Weekly bulletin ingestion requires URL discovery and PDF parsing.")
	}

	// ── SAIH Real-time Updates (incremental, per basin) ──
	log.Println("--- Checking SAIH updates ---")
	if err := fetchSAIHAll(db.DB); err != nil {
		log.Printf("SAIH fetch (non-fatal): %v", err)
		log.Println("NOTE: SAIH ingestion requires per-basin portal registration.")
	}

	log.Println("=== Updater finished ===")
}

// ensureDir creates a directory if it doesn't exist.
func ensureDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}
