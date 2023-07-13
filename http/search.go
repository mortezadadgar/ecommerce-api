package http

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mortezadadgar/ecommerce-api/domain"
)

func (s *Server) registerSearchRoutes(r *chi.Mux) {
	r.Route("/search", func(r chi.Router) {
		r.Get("/", s.searchHandler)
	})
}

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if len(query) == 0 {
		Errorf(w, r, http.StatusBadRequest, "search is not supported without query")
		return
	}

	result, err := s.SearchStore.Search(r.Context(), query)
	if err != nil {
		if errors.Is(err, domain.ErrNoSearchResult) {
			Errorf(w, r, http.StatusNotFound, err.Error())
		} else {
			Errorf(w, r, http.StatusInternalServerError, err.Error())
		}
		return
	}

	err = ToJSON(w, result, http.StatusOK)
	if err != nil {
		Errorf(w, r, http.StatusInternalServerError, err.Error())
	}
}
