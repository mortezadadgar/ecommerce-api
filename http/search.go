package http

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mortezadadgar/ecommerce-api/domain"
)

func (s *Server) registerSearchRoutes() {
	s.Route("/search", func(r chi.Router) {
		r.Get("/", s.searchHandler)
	})
}

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if len(query) == 0 {
		Error(w, r, domain.Errorf(domain.EINVALID, "search is not supported without query"))
		return
	}

	result, err := s.SearchStore.Search(r.Context(), query)
	if err != nil {
		Error(w, r, err)
		return
	}

	err = ToJSON(w, result, http.StatusOK)
	if err != nil {
		Error(w, r, err)
	}
}
