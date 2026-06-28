package v1sqlite

import (
	"database/sql"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes mounts all API v1 routes on the given router.
func RegisterRoutes(r chi.Router, db *sql.DB) {
	h := NewHandler(db)

	// Public health endpoints (no auth required)
	r.Get("/healthz", h.Healthz)
	r.Get("/readyz", h.Readyz)

	// API v1 routes (auth required)
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(Middleware(db))

		r.Get("/sources", h.ListSources)

		r.Get("/reservoirs", h.ListReservoirs)
		r.Get("/reservoirs/{slug}", h.GetReservoir)
		r.Get("/reservoirs/{slug}/readings", h.GetReservoirReadings)

		r.Get("/basins", h.ListBasins)
		r.Get("/basins/{slug}", h.GetBasin)

		r.Get("/rankings/reservoirs", h.GetRankings)
		r.Get("/compare", h.CompareReservoirs)
		r.Get("/data-quality", h.GetDataQuality)
		r.Post("/query", h.Query)
	})
}
