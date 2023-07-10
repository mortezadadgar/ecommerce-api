package http

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (s *Server) registerSearchRoutes() {
	s.Route("/search", func(r chi.Router) {
		r.Get("/", s.searchHandler)
	})
}

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if len(query) == 0 {
		Error(w, r, fmt.Errorf("search is not supported without query"), http.StatusBadRequest)
		return
	}

	result, err := s.SearchStore.Search(r.Context(), query)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			Error(w, r, ErrNotFound, http.StatusNotFound)
		default:
			Error(w, r, err, http.StatusInternalServerError)
		}

		return
	}

	err = ToJSON(w, result, http.StatusOK)
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
	}
}
