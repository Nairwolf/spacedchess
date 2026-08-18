package store

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"time"
)

// CreateSession generates an unguessable token, which is also the cookie value.
func (s *Store) CreateSession(ctx context.Context, userID int, ttl time.Duration) (string, time.Time, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", time.Time{}, err
	}
	token := base64.RawURLEncoding.EncodeToString(b)

	expiresAt := time.Now().Add(ttl)
	_, err := s.pool.Exec(ctx,
		`INSERT INTO sessions (id, user_id, expires_at) VALUES ($1, $2, $3)`,
		token, userID, expiresAt,
	)
	if err != nil {
		return "", time.Time{}, err
	}
	return token, expiresAt, nil
}

// SessionUser returns pgx.ErrNoRows for a token that is unknown or expired —
// the caller can't tell the difference, and doesn't need to.
func (s *Store) SessionUser(ctx context.Context, token string) (User, error) {
	var u User
	err := s.pool.QueryRow(ctx,
		`SELECT u.id, u.username
		 FROM sessions s JOIN users u ON u.id = s.user_id
		 WHERE s.id = $1 AND s.expires_at > now()`,
		token,
	).Scan(&u.ID, &u.Username)
	return u, err
}

func (s *Store) DeleteSession(ctx context.Context, token string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM sessions WHERE id = $1`, token)
	return err
}
