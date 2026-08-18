package server

import (
	"net/http"
	"strconv"

	"spacedchess/internal/store"
)

func GetCardHandler(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := r.PathValue("id")
		idInt, err := strconv.Atoi(idStr)
		if err != nil {
			writeError(w, apiError{http.StatusBadRequest, "id invalid"})
			return
		}

		s.GetCardByID(idInt)
	}
}
