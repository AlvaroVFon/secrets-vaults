// Package httpx provides small HTTP helpers shared across handlers.
package httpx

import (
	"encoding/json"
	"net/http"
)

// Response is the standard API response body.
type Response struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// WriteResponse writes a JSON response with the given status code.
func WriteResponse(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
}
