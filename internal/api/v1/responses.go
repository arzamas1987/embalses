package v1

import (
	"encoding/json"
	"net/http"
	"time"
)

// APIResponse is the standard envelope for all API responses.
type APIResponse struct {
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Lineage *Lineage    `json:"lineage,omitempty"`
}

// Meta holds pagination and request metadata.
type Meta struct {
	Page       int `json:"page"`
	PerPage    int `json:"per_page"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// APIError is the standard error response shape.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Lineage holds data provenance information.
type Lineage struct {
	Source      string    `json:"source"`
	Licence     string    `json:"licence"`
	Attribution string    `json:"attribution"`
	FetchedAt   time.Time `json:"fetched_at,omitempty"`
}

// WriteJSON writes a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

// WriteError writes a standardized error response.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	WriteJSON(w, status, APIResponse{
		Error: &APIError{Code: code, Message: message},
	})
}

// WriteList writes a paginated list response with lineage.
func WriteList(w http.ResponseWriter, data interface{}, meta Meta, lineage *Lineage) {
	WriteJSON(w, http.StatusOK, APIResponse{
		Data:    data,
		Meta:    &meta,
		Lineage: lineage,
	})
}

// WriteItem writes a single item response with lineage.
func WriteItem(w http.ResponseWriter, data interface{}, lineage *Lineage) {
	WriteJSON(w, http.StatusOK, APIResponse{
		Data:    data,
		Lineage: lineage,
	})
}

// Paginate computes pagination bounds.
func Paginate(page, perPage, total int) (offset, limit, totalPages int) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 10000 {
		perPage = 10000
	}
	totalPages = (total + perPage - 1) / perPage
	if totalPages < 1 {
		totalPages = 1
	}
	offset = (page - 1) * perPage
	limit = perPage
	return
}
