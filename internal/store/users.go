package store

import "context"

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}

func (s *Store) CreateUser(ctx context.Context, username, passwordHash string) (User, error) {
	var u User
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id, username`,
		username, passwordHash,
	).Scan(&u.ID, &u.Username)
	return u, err
}

// UserByUsername also returns the password hash, which User deliberately does
// not carry: User is serialized straight into responses.
func (s *Store) UserByUsername(ctx context.Context, username string) (User, string, error) {
	var (
		u    User
		hash string
	)
	err := s.pool.QueryRow(ctx,
		`SELECT id, username, password_hash FROM users WHERE username = $1`,
		username,
	).Scan(&u.ID, &u.Username, &hash)
	return u, hash, err
}
