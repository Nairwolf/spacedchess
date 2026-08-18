package store

import (
	"context"
	"os"
	"testing"
)

// Runs against the dev database from docker compose. Skipped when DATABASE_URL
// is unset so `go test ./...` passes without Docker running.
func TestNewAndPing(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set")
	}

	ctx := context.Background()
	s, err := New(ctx, databaseURL)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer s.Close()

	if err := s.Ping(ctx); err != nil {
		t.Errorf("Ping: %v", err)
	}
}

func TestNewRejectsUnreachableDatabase(t *testing.T) {
	// Port 1 is reserved and nothing listens there, so the ping fails immediately.
	_, err := New(context.Background(), "postgres://user:password@127.0.0.1:1/spacedchess")
	if err == nil {
		t.Fatal("expected an error for an unreachable database, got nil")
	}
}
