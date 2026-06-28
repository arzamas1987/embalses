package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/arzamas1987/embalses/internal/config"
	"github.com/arzamas1987/embalses/internal/db"
	"github.com/arzamas1987/embalses/internal/geo/ign"
	"github.com/arzamas1987/embalses/internal/geo/snczi"
)

func main() {
	fs := flag.NewFlagSet("embalses-ingest", flag.ExitOnError)
	source := fs.String("source", "", "Source to ingest (miteco, snczi, ign)")
	dryRun := fs.Bool("dry-run", false, "Parse and validate without writing to DB")
	fixture := fs.String("fixture", "", "Path to fixture file (for testing)")
	help := fs.Bool("help", false, "Show help")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) < 2 || *help || (len(os.Args) == 2 && os.Args[1] == "--help") {
		printHelp(fs)
		os.Exit(0)
	}

	cfg := config.Load()

	var err error
	switch strings.ToLower(*source) {
	case "snczi":
		err = ingestSNCZI(cfg, *dryRun, *fixture)
	case "ign":
		err = ingestIGN(cfg, *dryRun, *fixture)
	case "miteco":
		fmt.Println("embalses-ingest: MITECO ingestion not yet implemented (Phase 1)")
	default:
		fmt.Fprintf(os.Stderr, "Unknown source: %s\n", *source)
		printHelp(fs)
		os.Exit(1)
	}

	if err != nil {
		log.Fatalf("Ingestion failed: %v", err)
	}
}

func printHelp(fs *flag.FlagSet) {
	fmt.Println("embalses-ingest — ingestion CLI for reservoir data")
	fmt.Println()
	fmt.Println("Usage: embalses-ingest [options]")
	fmt.Println()
	fmt.Println("Options:")
	fs.PrintDefaults()
	fmt.Println()
	fmt.Println("Sources: miteco, snczi, ign")
}

func ingestSNCZI(cfg config.Config, dryRun bool, fixturePath string) error {
	fmt.Println("=== SNCZI Ingestion ===")
	if dryRun {
		fmt.Println("(dry-run mode: parsing only)")
	}

	path := fixturePath
	if path == "" {
		path = "test/fixtures/snczi_dams.geojson"
	}

	features, err := snczi.Parse(path)
	if err != nil {
		return fmt.Errorf("parse SNCZI fixture: %w", err)
	}
	fmt.Printf("Parsed %d dam features from %s\n", len(features), path)

	if dryRun {
		for _, f := range features {
			fmt.Printf("  - %s (%.4f, %.4f) province=%s basin=%s\n",
				f.Name, f.Lon, f.Lat, f.Province, f.Basin)
		}
		return nil
	}

	ctx := context.Background()
	pool, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect to db: %w", err)
	}
	defer pool.Close()

	if err := snczi.Ingest(ctx, pool.Pool, features); err != nil {
		return fmt.Errorf("ingest SNCZI: %w", err)
	}
	fmt.Printf("Ingested %d dams into database\n", len(features))
	return nil
}

func ingestIGN(cfg config.Config, dryRun bool, fixturePath string) error {
	fmt.Println("=== IGN Ingestion ===")
	if dryRun {
		fmt.Println("(dry-run mode: parsing only)")
	}

	basinPath := "test/fixtures/ign_basins.geojson"
	provincePath := "test/fixtures/ign_provinces.geojson"
	reservoirPath := "test/fixtures/ign_reservoirs.geojson"
	if fixturePath != "" {
		basinPath = fixturePath
		provincePath = ""
		reservoirPath = ""
	}

	basins, err := ign.ParseBasins(basinPath)
	if err != nil {
		return fmt.Errorf("parse IGN basins: %w", err)
	}
	fmt.Printf("Parsed %d basin features\n", len(basins))

	var provinces []ign.ProvinceFeature
	if provincePath != "" {
		provinces, err = ign.ParseProvinces(provincePath)
		if err != nil {
			return fmt.Errorf("parse IGN provinces: %w", err)
		}
		fmt.Printf("Parsed %d province features\n", len(provinces))
	}

	var reservoirs []ign.ReservoirFeature
	if reservoirPath != "" {
		reservoirs, err = ign.ParseReservoirs(reservoirPath)
		if err != nil {
			return fmt.Errorf("parse IGN reservoirs: %w", err)
		}
		fmt.Printf("Parsed %d reservoir features\n", len(reservoirs))
	}

	if dryRun {
		for _, b := range basins {
			fmt.Printf("  Basin: %s\n", b.Name)
		}
		for _, p := range provinces {
			fmt.Printf("  Province: %s (code=%s)\n", p.Name, p.Code)
		}
		for _, r := range reservoirs {
			fmt.Printf("  Reservoir: %s\n", r.Name)
		}
		return nil
	}

	ctx := context.Background()
	pool, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect to db: %w", err)
	}
	defer pool.Close()

	if err := ign.Ingest(ctx, pool.Pool, basins, provinces, reservoirs); err != nil {
		return fmt.Errorf("ingest IGN: %w", err)
	}
	fmt.Printf("Ingested %d basins, %d provinces, %d reservoirs\n", len(basins), len(provinces), len(reservoirs))
	return nil
}
