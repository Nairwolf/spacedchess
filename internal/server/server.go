// Package server implements REST JSON API
package server

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"spacedchess/internal/config"
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

func addRoutes(mux *http.ServeMux, s *store.Store, cfg config.Config) {
	mux.Handle("GET /health", Health(s))
	mux.Handle("POST /auth/register", Register(s, cfg))
	mux.Handle("POST /auth/login", Login(s, cfg))
	mux.Handle("POST /auth/logout", Logout(s, cfg))
	mux.Handle("GET /auth/me", requireAuth(s, Me()))
}

func NewServer(s *store.Store, cfg config.Config) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, s, cfg)
	return mux
}
