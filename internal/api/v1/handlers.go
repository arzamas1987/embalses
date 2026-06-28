package v1

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Handler holds the API v1 handlers.
type Handler struct {
	Pool *pgxpool.Pool
}

// NewHandler creates a new handler instance.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{Pool: pool}
}

// Healthz returns the API health status.
func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, APIResponse{
		Data: map[string]string{"status": "ok", "service": "api"},
	})
}

// Readyz checks if the API is ready to serve requests (DB available).
func (h *Handler) Readyz(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := h.Pool.Ping(ctx); err != nil {
		WriteError(w, http.StatusServiceUnavailable, "not_ready", "Database unavailable")
		return
	}
	WriteJSON(w, http.StatusOK, APIResponse{
		Data: map[string]string{"status": "ready", "service": "api"},
	})
}

// ListSources returns all data sources with attribution.
func (h *Handler) ListSources(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sources, err := QuerySources(ctx, h.Pool)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "query_error", err.Error())
		return
	}
	lineage, _ := QueryLineage(ctx, h.Pool, "IGN")
	WriteJSON(w, http.StatusOK, APIResponse{Data: sources, Lineage: &lineage})
}

// ListReservoirs returns a paginated list of reservoirs.
func (h *Handler) ListReservoirs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	page := intQueryParam(r, "page", 1)
	perPage := intQueryParam(r, "per_page", 20)

	offset, limit, _ := Paginate(page, perPage, 0)
	reservoirs, total, err := QueryReservoirs(ctx, h.Pool, offset, limit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "query_error", err.Error())
		return
	}

	_, _, totalPages := Paginate(page, perPage, total)
	meta := Meta{Page: page, PerPage: perPage, Total: total, TotalPages: totalPages}

	lineage, _ := QueryLineage(ctx, h.Pool, "IGN")
	WriteList(w, reservoirs, meta, &lineage)
}

// GetReservoir returns a single reservoir by slug (name).
func (h *Handler) GetReservoir(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")
	reservoir, err := QueryReservoirBySlug(ctx, h.Pool, slug)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			WriteError(w, http.StatusNotFound, "not_found", fmt.Sprintf("Reservoir '%s' not found", slug))
			return
		}
		WriteError(w, http.StatusInternalServerError, "query_error", err.Error())
		return
	}
	lineage, _ := QueryLineage(ctx, h.Pool, "IGN")
	WriteItem(w, reservoir, &lineage)
}

// GetReservoirReadings returns time-series readings for a reservoir.
func (h *Handler) GetReservoirReadings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")
	page := intQueryParam(r, "page", 1)
	perPage := intQueryParam(r, "per_page", 30)

	// Default to last 30 days
	since := time.Now().AddDate(0, 0, -30)
	until := time.Now()
	if s := r.URL.Query().Get("since"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			since = t
		}
	}
	if u := r.URL.Query().Get("until"); u != "" {
		if t, err := time.Parse("2006-01-02", u); err == nil {
			until = t
		}
	}

	offset, limit, _ := Paginate(page, perPage, 0)
	readings, total, err := QueryReadings(ctx, h.Pool, slug, since, until, offset, limit)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			WriteError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		WriteError(w, http.StatusInternalServerError, "query_error", err.Error())
		return
	}

	_, _, totalPages := Paginate(page, perPage, total)
	meta := Meta{Page: page, PerPage: perPage, Total: total, TotalPages: totalPages}
	lineage, _ := QueryLineage(ctx, h.Pool, "MITECO")
	WriteList(w, readings, meta, &lineage)
}

// ListBasins returns all basins.
func (h *Handler) ListBasins(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	basins, err := QueryBasins(ctx, h.Pool)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "query_error", err.Error())
		return
	}
	lineage, _ := QueryLineage(ctx, h.Pool, "IGN")
	WriteJSON(w, http.StatusOK, APIResponse{Data: basins, Lineage: &lineage})
}

// GetBasin returns a single basin by slug.
func (h *Handler) GetBasin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	slug := chi.URLParam(r, "slug")
	basin, err := QueryBasinBySlug(ctx, h.Pool, slug)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			WriteError(w, http.StatusNotFound, "not_found", fmt.Sprintf("Basin '%s' not found", slug))
			return
		}
		WriteError(w, http.StatusInternalServerError, "query_error", err.Error())
		return
	}
	lineage, _ := QueryLineage(ctx, h.Pool, "IGN")
	WriteItem(w, basin, &lineage)
}

// GetRankings returns reservoir rankings.
func (h *Handler) GetRankings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	metric := r.URL.Query().Get("metric")
	if metric == "" {
		metric = "fullest"
	}
	limit := intQueryParam(r, "limit", 10)
	if limit > 50 {
		limit = 50
	}

	items, err := QueryRankings(ctx, h.Pool, metric, limit)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "query_error", err.Error())
		return
	}
	lineage, _ := QueryLineage(ctx, h.Pool, "MITECO")
	WriteJSON(w, http.StatusOK, APIResponse{Data: items, Lineage: &lineage})
}

// CompareReservoirs returns aligned readings for multiple reservoirs.
func (h *Handler) CompareReservoirs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	names := r.URL.Query()["reservoir"]
	if len(names) == 0 {
		WriteError(w, http.StatusBadRequest, "bad_request", "At least one reservoir required (query param: reservoir)")
		return
	}

	since := time.Now().AddDate(0, 0, -30)
	until := time.Now()
	if s := r.URL.Query().Get("since"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			since = t
		}
	}
	if u := r.URL.Query().Get("until"); u != "" {
		if t, err := time.Parse("2006-01-02", u); err == nil {
			until = t
		}
	}

	result, err := QueryComparator(ctx, h.Pool, names, since, until)
	if err != nil {
		WriteError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	lineage, _ := QueryLineage(ctx, h.Pool, "MITECO")
	WriteJSON(w, http.StatusOK, APIResponse{Data: result, Lineage: &lineage})
}

// GetDataQuality returns a data quality report.
func (h *Handler) GetDataQuality(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	report, err := QueryDataQuality(ctx, h.Pool)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "query_error", err.Error())
		return
	}
	lineage, _ := QueryLineage(ctx, h.Pool, "MITECO")
	WriteJSON(w, http.StatusOK, APIResponse{Data: report, Lineage: &lineage})
}
