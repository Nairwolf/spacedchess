package store

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

func testStore(t *testing.T) *Store {
	t.Helper()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set")
	}

	s, err := New(context.Background(), databaseURL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(s.Close)
	return s
}

// Unique per call so repeated runs against the same dev database don't collide.
func testUser(t *testing.T, s *Store) User {
	t.Helper()

	ctx := context.Background()
	u, err := s.CreateUser(ctx, fmt.Sprintf("test_%d", time.Now().UnixNano()), "hash")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	t.Cleanup(func() {
		if _, err := s.pool.Exec(ctx, `DELETE FROM users WHERE id = $1`, u.ID); err != nil {
			t.Errorf("cleanup user %d: %v", u.ID, err)
		}
	})
	return u
}

func TestCreateAndLookupUser(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	u := testUser(t, s)

	got, hash, err := s.UserByUsername(ctx, u.Username)
	if err != nil {
		t.Fatalf("UserByUsername: %v", err)
	}
	if got != u {
		t.Errorf("user = %+v, want %+v", got, u)
	}
	if hash != "hash" {
		t.Errorf("hash = %q, want %q", hash, "hash")
	}
}

func TestUserByUsernameUnknown(t *testing.T) {
	s := testStore(t)

	_, _, err := s.UserByUsername(context.Background(), "nobody_here")
	if !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("err = %v, want pgx.ErrNoRows", err)
	}
}

func TestSessionLifecycle(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()
	u := testUser(t, s)

	token, expiresAt, err := s.CreateSession(ctx, u.ID, time.Hour)
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if time.Until(expiresAt) <= 0 {
		t.Errorf("expiresAt = %v, want in the future", expiresAt)
	}

	got, err := s.SessionUser(ctx, token)
	if err != nil {
		t.Fatalf("SessionUser: %v", err)
	}
	if got != u {
		t.Errorf("user = %+v, want %+v", got, u)
	}

	if err := s.DeleteSession(ctx, token); err != nil {
		t.Fatalf("DeleteSession: %v", err)
	}
	if _, err := s.SessionUser(ctx, token); !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("err after delete = %v, want pgx.ErrNoRows", err)
	}
}

// The seed ships both rows precisely so this can be checked.
func TestSessionUserRejectsExpiredSeedSession(t *testing.T) {
	s := testStore(t)
	ctx := context.Background()

	if _, err := s.SessionUser(ctx, "dev_session_valid"); err != nil {
		t.Errorf("dev_session_valid: %v", err)
	}
	if _, err := s.SessionUser(ctx, "dev_session_expired"); !errors.Is(err, pgx.ErrNoRows) {
		t.Errorf("dev_session_expired err = %v, want pgx.ErrNoRows", err)
	}
}
