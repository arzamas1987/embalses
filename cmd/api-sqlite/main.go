package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/arzamas1987/embalses/internal/api/v1sqlite"
	"github.com/arzamas1987/embalses/internal/health"
	"github.com/arzamas1987/embalses/internal/storage/sqlite"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	addr := os.Getenv("API_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	dbPath := os.Getenv("DATABASE_URL")
	if dbPath == "" {
		dbPath = "data/embalses.db"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := sqlite.Open(ctx, dbPath)
	if err != nil {
		log.Fatalf("SQLite connection failed: %v", err)
	}
	defer db.Close()

	// Ensure schema is created
	if err := sqlite.Migrate(db.DB); err != nil {
		log.Fatalf("SQLite migration failed: %v", err)
	}

	// Seed test API key if not present
	if err := sqlite.SeedTestKey(db.DB); err != nil {
		log.Printf("Test key seed warning: %v", err)
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(middleware.AllowContentType("application/json", "multipart/form-data"))

	// Public health endpoints
	r.Get("/healthz", health.Handler("api"))

	// API v1 routes (includes auth middleware)
	v1sqlite.RegisterRoutes(r, db.DB)

	log.Printf("API server (SQLite) starting on %s  [db=%s]", addr, dbPath)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("API server error: %v", err)
	}
}
