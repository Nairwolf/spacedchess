package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"

	"spacedchess/internal/config"
	"spacedchess/internal/store"
)

const (
	sessionCookie = "session"
	sessionTTL    = 7 * 24 * time.Hour
	minPassword   = 8
)

type credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// startSession creates the session row and sets the cookie, reporting whether
// it succeeded; it writes the error response itself if it didn't.
func startSession(w http.ResponseWriter, r *http.Request, s *store.Store, cfg config.Config, userID int) bool {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	token, expiresAt, err := s.CreateSession(ctx, userID, sessionTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal error")
		return false
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookie,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
	return true
}

func Register(s *store.Store, cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var c credentials
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}
		if c.Username == "" {
			writeError(w, http.StatusBadRequest, "username is required")
			return
		}
		if len(c.Password) < minPassword {
			writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		hash, err := bcrypt.GenerateFromPassword([]byte(c.Password), bcrypt.DefaultCost)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		u, err := s.CreateUser(ctx, c.Username, string(hash))
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			writeError(w, http.StatusConflict, "username is taken")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}

		// Registering logs you in; otherwise the client holds a user it has no
		// cookie for and gets bounced on the next page load.
		if !startSession(w, r, s, cfg, u.ID) {
			return
		}
		writeJSON(w, http.StatusCreated, u)
	}
}

func Login(s *store.Store, cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var c credentials
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			writeError(w, http.StatusBadRequest, "invalid request body")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		u, hash, err := s.UserByUsername(ctx, c.Username)
		if errors.Is(err, pgx.ErrNoRows) {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		if err != nil {
			writeError(w, http.StatusInternalServerError, "internal error")
			return
		}
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(c.Password)) != nil {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}

		if !startSession(w, r, s, cfg, u.ID) {
			return
		}
		writeJSON(w, http.StatusOK, u)
	}
}

func Logout(s *store.Store, cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if c, err := r.Cookie(sessionCookie); err == nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()

			if err := s.DeleteSession(ctx, c.Value); err != nil {
				writeError(w, http.StatusInternalServerError, "internal error")
				return
			}
		}

		http.SetCookie(w, &http.Cookie{
			Name:     sessionCookie,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: true,
			Secure:   cfg.CookieSecure,
			SameSite: http.SameSiteLaxMode,
		})
		w.WriteHeader(http.StatusNoContent)
	}
}

func Me() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, userFromContext(r.Context()))
	}
}

type contextKey struct{}

var userKey contextKey

func requireAuth(s *store.Store, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(sessionCookie)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		u, err := s.SessionUser(ctx, c.Value)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userKey, u)))
	})
}

func userFromContext(ctx context.Context) store.User {
	u, _ := ctx.Value(userKey).(store.User)
	return u
}
