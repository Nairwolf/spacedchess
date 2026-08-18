// Package server implements REST JSON API
package server

import (
	"encoding/json"
	"net/http"

	"spacedchess/internal/store"
)

type apiError struct {
	Status  int    `json:"-"`
	Message string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func writeError(w http.ResponseWriter, err apiError) {
	writeJSON(w, err.Status, err)
}

func Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func addRoutes(mux *http.ServeMux, s *store.Store) {
	mux.Handle("GET /health", Health())

	// cards API
	mux.Handle("GET /api/cards", GetCardHandler(s))
}

func NewServer(s *store.Store) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, s)
	return mux
}
