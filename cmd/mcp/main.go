package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/arzamas1987/embalses/internal/health"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	addr := os.Getenv("MCP_ADDR")
	if addr == "" {
		addr = ":8081"
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/healthz", health.Handler("mcp"))

	log.Printf("MCP server starting on %s", addr)
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}
