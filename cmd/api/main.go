package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/arzamas1987/embalses/internal/api/v1"
	"github.com/arzamas1987/embalses/internal/config"
	"github.com/arzamas1987/embalses/internal/db"
	"github.com/arzamas1987/embalses/internal/health"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	addr := os.Getenv("API_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	pool, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer pool.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(middleware.AllowContentType("application/json"))

	// Public health endpoints
	r.Get("/healthz", health.Handler("api"))

	// API v1 routes (includes auth middleware)
	v1.RegisterRoutes(r, pool.Pool)

	log.Printf("API server starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("API server error: %v", err)
	}
}
