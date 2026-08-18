// Package store handles Store layer
package store

// Store handles storing layer
type Store struct{}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) GetCardByID(id int) {
	_ = id
}
