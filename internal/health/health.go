package health

import (
	"encoding/json"
	"net/http"
)

// Response is the standard health check JSON response.
type Response struct {
	Status  string `json:"status"`
	Service string `json:"service"`
}

// Handler returns an HTTP handler that writes a health check response.
func Handler(service string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Response{
			Status:  "ok",
			Service: service,
		})
	}
}
