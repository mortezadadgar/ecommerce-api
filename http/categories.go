package http

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mortezadadgar/ecommerce-api/domain"
	"github.com/mortezadadgar/ecommerce-api/postgres"
)

func (s *Server) registerCategoriesRoutes(r *chi.Mux) {
	r.Route("/categories", func(r chi.Router) {
		r.Get("/{id}", s.getCategoryHandler)
		r.Get("/", s.listCategoriesHandler)
		r.With(requireAuth).Post("/", s.createCategoryHandler)
		r.With(requireAuth).Patch("/{id}", s.updateCategoryHandler)
		r.With(requireAuth).Delete("/{id}", s.deleteCategoryHandler)
	})
}

// @Summary      Get category
// @Tags 		 Categories
// @Produce      json
// @Param        id    path     int  true "Category ID"
// @Success      200  {array}   domain.WrapCategories
// @Failure      400  {object}  http.HTTPError
// @Failure      500  {object}  http.HTTPError
// @Router       /categories/{id} [get]
func (s *Server) getCategoryHandler(w http.ResponseWriter, r *http.Request) {
	ID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, r, ErrInvalidQuery, http.StatusBadRequest)
		return
	}

	category, err := s.CategoriesStore.GetByID(r.Context(), ID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			Error(w, r, ErrNotFound, http.StatusNotFound)
		default:
			Error(w, r, err, http.StatusInternalServerError)
		}

		return
	}

	err = ToJSON(w, domain.WrapCategories{Category: category}, http.StatusOK)
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
	}
}

// @Summary      List categories
// @Tags 		 Categories
// @Produce      json
// @Success      200  {array}   domain.WrapCategoriesList
// @Failure      400  {object}  http.HTTPError
// @Failure      500  {object}  http.HTTPError
// @Router       /categories/   [get]
func (s *Server) listCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	limit, err := ParseIntQuery(r, "limit")
	if err != nil {
		Error(w, r, ErrInvalidQuery, http.StatusBadRequest)
		return
	}

	offset, err := ParseIntQuery(r, "offset")
	if err != nil {
		Error(w, r, ErrInvalidQuery, http.StatusBadRequest)
		return
	}

	filter := domain.CategoriesFilter{
		Name:   r.URL.Query().Get("name"),
		Sort:   r.URL.Query().Get("sort"),
		Limit:  limit,
		Offset: offset,
	}

	categories, err := s.CategoriesStore.List(r.Context(), filter)
	if err != nil {
		switch err {
		case sql.ErrNoRows, postgres.ErrInvalidColumn:
			Error(w, r, ErrNotFound, http.StatusNotFound)
		default:
			Error(w, r, err, http.StatusInternalServerError)
		}

		return
	}

	err = ToJSON(w, domain.WrapCategoriesList{Categories: categories}, http.StatusOK)
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
	}
}

// @Summary      Create category
// @Tags 		 Categories
// @Produce      json
// @Accept       json
// @Param        category     body        domain.CategoriesCreate true "Create category"
// @Success      200          {array}     domain.WrapCategories
// @Failure      400          {object}    http.HTTPError
// @Failure      413          {object}    http.HTTPError
// @Failure      500          {object}    http.HTTPError
// @Router       /categories/ [post]
func (s *Server) createCategoryHandler(w http.ResponseWriter, r *http.Request) {
	input := domain.CategoriesCreate{}

	err := FromJSON(w, r, &input)
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
		return
	}

	err = input.Validate()
	if err != nil {
		Error(w, r, err, http.StatusBadRequest)
		return
	}

	category := input.CreateModel()

	err = s.CategoriesStore.Create(r.Context(), &category)
	if err != nil {
		switch err {
		case postgres.ErrDuplicatedEntries:
			Error(w, r, err, http.StatusBadRequest)
		default:
			Error(w, r, err, http.StatusInternalServerError)
		}

		return
	}

	w.Header().Set("Location", fmt.Sprintf("/products/%d", category.ID))
	err = ToJSON(w, domain.WrapCategories{Category: category}, http.StatusCreated)
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
	}
}

// @Summary      Update category
// @Tags 		 Categories
// @Produce      json
// @Accept       json
// @Param        category     body        domain.CategoriesUpdate true "Update category"
// @Success      200          {array}     domain.WrapCategories
// @Failure      400          {object}    http.HTTPError
// @Failure      413          {object}    http.HTTPError
// @Failure      500          {object}    http.HTTPError
// @Router       /categories/ [patch]
func (s *Server) updateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	ID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, r, ErrInvalidQuery, http.StatusBadRequest)
		return
	}

	input := domain.CategoriesUpdate{}
	err = FromJSON(w, r, &input)
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
		return
	}

	err = input.Validate()
	if err != nil {
		Error(w, r, err, http.StatusBadRequest)
		return
	}

	category, err := s.CategoriesStore.GetByID(r.Context(), ID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			Error(w, r, ErrNotFound, http.StatusNotFound)
		default:
			Error(w, r, err, http.StatusInternalServerError)
		}

		return
	}

	input.UpdateModel(&category)

	err = s.CategoriesStore.Update(r.Context(), &category)
	if err != nil {
		switch err {
		case postgres.ErrForeinKeyViolation, postgres.ErrDuplicatedEntries:
			Error(w, r, err, http.StatusBadRequest)
		default:
			Error(w, r, err, http.StatusInternalServerError)
		}

		return
	}

	err = ToJSON(w, domain.WrapCategories{Category: category}, http.StatusOK)
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
	}
}

// @Summary      Delete category
// @Tags 		 Categories
// @Param        id           path        int  true "Category ID"
// @Success      200          {array}     domain.WrapCategories
// @Failure      400          {object}    http.HTTPError
// @Failure      500          {object}    http.HTTPError
// @Router       /categories/{id} [delete]
func (s *Server) deleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	ID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, r, ErrInvalidQuery, http.StatusBadRequest)
		return
	}

	err = s.CategoriesStore.Delete(r.Context(), ID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			Error(w, r, ErrNotFound, http.StatusNotFound)
		case postgres.ErrForeinKeyViolation:
			Error(w, r, postgres.ErrUnableDeleteEntry, http.StatusBadRequest)
		default:
			Error(w, r, err, http.StatusInternalServerError)
		}
	}
}
