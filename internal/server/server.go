// Package server implements REST JSON API
package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"spacedchess/internal/store"
)

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		_ = json.NewEncoder(w).Encode(v)
	}
}

func Health(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := s.Ping(ctx); err != nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "unavailable"})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}
}

func addRoutes(mux *http.ServeMux, s *store.Store) {
	mux.Handle("GET /health", Health(s))
}

func NewServer(s *store.Store) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, s)
	return mux
}
