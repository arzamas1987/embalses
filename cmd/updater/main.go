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
		dbPath       = flag.String("db", "data/embalses.db", "Path to SQLite database")
		fullImport   = flag.Bool("full", false, "Run full historical import (one-time)")
		geoOnly      = flag.Bool("geo-only", false, "Import only GeoJSON fixtures (GPS + metadata)")
		seedReadings = flag.Bool("seed-readings", false, "Seed synthetic 6-month readings")
	)
	flag.Parse()

	log.Println("=== Embalses Updater ===")
	log.Printf("Database: %s", *dbPath)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
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
	if *fullImport {
		log.Println("--- Fetching MITECO historical data ---")
		if err := fetchMITECOHistorical(db.DB); err != nil {
			log.Printf("MITECO historical fetch (non-fatal): %v", err)
			log.Println("NOTE: MITECO ingestion requires manual download of BD-Embalses_1988-2022.zip")
			log.Println("      Place the unzipped Excel files in data/raw/miteco/ and re-run with -full")
		} else {
			log.Println("MITECO historical data imported.")
		}
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
