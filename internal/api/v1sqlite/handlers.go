package v1sqlite

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/arzamas1987/embalses/internal/api/v1"
	"github.com/arzamas1987/embalses/internal/planner"
	"github.com/arzamas1987/embalses/internal/storage/sqlite"
	"github.com/go-chi/chi/v5"
)

// Handler holds the API v1 handlers for SQLite backend.
type Handler struct {
	DB *sql.DB
}

// NewHandler creates a new handler instance.
func NewHandler(db *sql.DB) *Handler {
	return &Handler{DB: db}
}

// toV1Lineage converts sqlite.Lineage to v1.Lineage.
func toV1Lineage(l sqlite.Lineage) *v1.Lineage {
	return &v1.Lineage{
		Source:      l.Source,
		Licence:     l.Licence,
		Attribution: l.Attribution,
		FetchedAt:   l.FetchedAt,
	}
}

// Healthz returns the API health status.
func (h *Handler) Healthz(w http.ResponseWriter, r *http.Request) {
	v1.WriteJSON(w, http.StatusOK, v1.APIResponse{
		Data: map[string]string{"status": "ok", "service": "api", "backend": "sqlite"},
	})
}

// Readyz checks if the API is ready to serve requests.
func (h *Handler) Readyz(w http.ResponseWriter, r *http.Request) {
	if err := h.DB.Ping(); err != nil {
		v1.WriteError(w, http.StatusServiceUnavailable, "not_ready", "SQLite database unavailable")
		return
	}
	v1.WriteJSON(w, http.StatusOK, v1.APIResponse{
		Data: map[string]string{"status": "ready", "service": "api", "backend": "sqlite"},
	})
}

// ListSources returns all data sources with attribution.
func (h *Handler) ListSources(w http.ResponseWriter, r *http.Request) {
	sources, err := sqlite.QuerySources(h.DB)
	if err != nil {
		v1.WriteError(w, http.StatusInternalServerError, "query_error", err.Error())
		return
	}
	lineage, _ := sqlite.QueryLineage(h.DB, "IGN")
	v1.WriteJSON(w, http.StatusOK, v1.APIResponse{Data: sources, Lineage: toV1Lineage(lineage)})
}

// ListReservoirs returns a paginated list of reservoirs.
func (h *Handler) ListReservoirs(w http.ResponseWriter, r *http.Request) {
	page := v1IntQueryParam(r, "page", 1)
	perPage := v1IntQueryParam(r, "per_page", 20)

	offset, limit, _ := v1.Paginate(page, perPage, 0)
	reservoirs, total, err := sqlite.QueryReservoirs(h.DB, offset, limit)
	if err != nil {
		v1.WriteError(w, http.StatusInternalServerError, "query_error", err.Error())
		return
	}

	_, _, totalPages := v1.Paginate(page, perPage, total)
	meta := v1.Meta{Page: page, PerPage: perPage, Total: total, TotalPages: totalPages}

	lineage, _ := sqlite.QueryLineage(h.DB, "IGN")
	v1.WriteList(w, reservoirs, meta, toV1Lineage(lineage))
}

// GetReservoir returns a single reservoir by slug.
func (h *Handler) GetReservoir(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	reservoir, err := sqlite.QueryReservoirBySlug(h.DB, slug)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			v1.WriteError(w, http.StatusNotFound, "not_found", fmt.Sprintf("Reservoir '%s' not found", slug))
			return
		}
		v1.WriteError(w, http.StatusInternalServerError, "query_error", err.Error())
		return
	}
	lineage, _ := sqlite.QueryLineage(h.DB, "IGN")
	v1.WriteItem(w, reservoir, toV1Lineage(lineage))
}

// GetReservoirReadings returns time-series readings for a reservoir.
func (h *Handler) GetReservoirReadings(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	page := v1IntQueryParam(r, "page", 1)
	perPage := v1IntQueryParam(r, "per_page", 30)

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

	offset, limit, _ := v1.Paginate(page, perPage, 0)
	readings, total, err := sqlite.QueryReadings(h.DB, slug, since, until, offset, limit)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			v1.WriteError(w, http.StatusNotFound, "not_found", err.Error())
			return
		}
		v1.WriteError(w, http.StatusInternalServerError, "query_error", err.Error())
		return
	}

	_, _, totalPages := v1.Paginate(page, perPage, total)
	meta := v1.Meta{Page: page, PerPage: perPage, Total: total, TotalPages: totalPages}
	lineage, _ := sqlite.QueryLineage(h.DB, "MITECO")
	v1.WriteList(w, readings, meta, toV1Lineage(lineage))
}

// ListBasins returns all basins.
func (h *Handler) ListBasins(w http.ResponseWriter, r *http.Request) {
	basins, err := sqlite.QueryBasins(h.DB)
	if err != nil {
		v1.WriteError(w, http.StatusInternalServerError, "query_error", err.Error())
		return
	}
	lineage, _ := sqlite.QueryLineage(h.DB, "IGN")
	v1.WriteJSON(w, http.StatusOK, v1.APIResponse{Data: basins, Lineage: toV1Lineage(lineage)})
}

// GetBasin returns a single basin by slug.
func (h *Handler) GetBasin(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	basin, err := sqlite.QueryBasinBySlug(h.DB, slug)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			v1.WriteError(w, http.StatusNotFound, "not_found", fmt.Sprintf("Basin '%s' not found", slug))
			return
		}
		v1.WriteError(w, http.StatusInternalServerError, "query_error", err.Error())
		return
	}
	lineage, _ := sqlite.QueryLineage(h.DB, "IGN")
	v1.WriteItem(w, basin, toV1Lineage(lineage))
}

// GetRankings returns reservoir rankings.
func (h *Handler) GetRankings(w http.ResponseWriter, r *http.Request) {
	metric := r.URL.Query().Get("metric")
	if metric == "" {
		metric = "fullest"
	}
	limit := v1IntQueryParam(r, "limit", 10)
	if limit > 50 {
		limit = 50
	}

	items, err := sqlite.QueryRankings(h.DB, metric, limit)
	if err != nil {
		v1.WriteError(w, http.StatusInternalServerError, "query_error", err.Error())
		return
	}
	lineage, _ := sqlite.QueryLineage(h.DB, "MITECO")
	v1.WriteJSON(w, http.StatusOK, v1.APIResponse{Data: items, Lineage: toV1Lineage(lineage)})
}

// CompareReservoirs returns aligned readings for multiple reservoirs.
func (h *Handler) CompareReservoirs(w http.ResponseWriter, r *http.Request) {
	names := r.URL.Query()["reservoir"]
	if len(names) == 0 {
		v1.WriteError(w, http.StatusBadRequest, "bad_request", "At least one reservoir required (query param: reservoir)")
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

	result, err := sqlite.QueryComparator(h.DB, names, since, until)
	if err != nil {
		v1.WriteError(w, http.StatusBadRequest, "bad_request", err.Error())
		return
	}
	lineage, _ := sqlite.QueryLineage(h.DB, "MITECO")
	v1.WriteJSON(w, http.StatusOK, v1.APIResponse{Data: result, Lineage: toV1Lineage(lineage)})
}

// GetDataQuality returns a data quality report.
func (h *Handler) GetDataQuality(w http.ResponseWriter, r *http.Request) {
	report, err := sqlite.QueryDataQuality(h.DB)
	if err != nil {
		v1.WriteError(w, http.StatusInternalServerError, "query_error", err.Error())
		return
	}
	lineage, _ := sqlite.QueryLineage(h.DB, "MITECO")
	v1.WriteJSON(w, http.StatusOK, v1.APIResponse{Data: report, Lineage: toV1Lineage(lineage)})
}

// Query accepts a Query Intent JSON, validates, compiles, executes.
func (h *Handler) Query(w http.ResponseWriter, r *http.Request) {
	var intent planner.QueryIntent
	if err := json.NewDecoder(r.Body).Decode(&intent); err != nil {
		v1.WriteError(w, http.StatusBadRequest, "invalid_json", err.Error())
		return
	}
	if err := planner.ValidateIntent(intent); err != nil {
		v1.WriteError(w, http.StatusBadRequest, "invalid_intent", err.Error())
		return
	}
	plan, err := planner.CompilePlan(intent)
	if err != nil {
		v1.WriteError(w, http.StatusBadRequest, "compile_error", err.Error())
		return
	}
	result, err := planner.ExecutePlan(r.Context(), nil, plan) // NOTE: ExecutePlan expects pgx pool; needs adaptation for SQLite
	if err != nil {
		v1.WriteError(w, http.StatusInternalServerError, "execution_error", err.Error())
		return
	}
	lineage, _ := sqlite.QueryLineage(h.DB, "MITECO")
	v1.WriteJSON(w, http.StatusOK, v1.APIResponse{Data: result, Lineage: toV1Lineage(lineage)})
}

func v1IntQueryParam(r *http.Request, name string, fallback int) int {
	v := r.URL.Query().Get(name)
	if v == "" {
		return fallback
	}
	i, err := strconv.Atoi(v)
	if err != nil || i < 1 {
		return fallback
	}
	return i
}
