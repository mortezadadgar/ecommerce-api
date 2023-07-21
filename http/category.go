package http

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mortezadadgar/ecommerce-api/domain"
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
// @Param        id    path       int  true "Category ID"
// @Success      200  {array}     domain.WrapCategory
// @Failure      400  {object}    http.WrapError
// @Failure      404  {object}    http.WrapError
// @Failure      500  {object}    http.WrapError
// @Router       /categories/{id} [get]
func (s *Server) getCategoryHandler(w http.ResponseWriter, r *http.Request) {
	ID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		ErrorInvalidQuery(w, r)
		return
	}

	category, err := s.CategoriesStore.GetByID(r.Context(), ID)
	if err != nil {
		if errors.Is(err, domain.ErrNoCategoryFound) {
			Errorf(w, r, http.StatusNotFound, err.Error())
		} else {
			Errorf(w, r, http.StatusInternalServerError, err.Error())
		}
		return
	}

	err = ToJSON(w, domain.WrapCategory{Category: category}, http.StatusOK)
	if err != nil {
		Errorf(w, r, http.StatusInternalServerError, err.Error())
	}
}

// @Summary      List categories
// @Tags 		 Categories
// @Param        limit        query       string  false "Limit results"
// @Param        offset       query       string  false "Offset results"
// @Param        name         query       string  false "List by name"
// @Param        sort         query       string  false "Sort by a column"
// @Success      200  {array}   domain.WrapCategoryList
// @Failure      400  {object}  http.WrapError
// @Failure      404  {object}  http.WrapError
// @Failure      500  {object}  http.WrapError
// @Router       /categories/   [get]
func (s *Server) listCategoriesHandler(w http.ResponseWriter, r *http.Request) {
	limit, err := ParseIntQuery(r, "limit")
	if err != nil {
		ErrorInvalidQuery(w, r)
		return
	}

	offset, err := ParseIntQuery(r, "offset")
	if err != nil {
		ErrorInvalidQuery(w, r)
		return
	}

	filter := domain.CategoryFilter{
		Name:   r.URL.Query().Get("name"),
		Sort:   r.URL.Query().Get("sort"),
		Limit:  limit,
		Offset: offset,
	}

	categories, err := s.CategoriesStore.List(r.Context(), filter)
	if err != nil {
		if errors.Is(err, domain.ErrNoCategoryFound) {
			Errorf(w, r, http.StatusNotFound, err.Error())
		} else {
			Errorf(w, r, http.StatusInternalServerError, err.Error())
		}
		return
	}

	err = ToJSON(w, domain.WrapCategoryList{Categories: categories}, http.StatusOK)
	if err != nil {
		Errorf(w, r, http.StatusInternalServerError, err.Error())
	}
}

// @Summary      Create category
// @Tags 		 Categories
// @Security     Bearer
// @Produce      json
// @Accept       json
// @Param        category         body        domain.CategoryCreate true "Create category"
// @Success      201              {array}     domain.WrapCategory
// @Failure      400              {object}    http.WrapError
// @Failure      403              {object}    http.WrapError
// @Failure      404              {object}    http.WrapError
// @Failure      413              {object}    http.WrapError
// @Failure      500              {object}    http.WrapError
// @Router       /categories/{id} [post]
func (s *Server) createCategoryHandler(w http.ResponseWriter, r *http.Request) {
	input := domain.CategoryCreate{}

	err := FromJSON(w, r, &input)
	if err != nil {
		Errorf(w, r, http.StatusBadRequest, err.Error())
		return
	}

	err = input.Validate()
	if err != nil {
		Errorf(w, r, http.StatusBadRequest, err.Error())
		return
	}

	category := input.CreateModel()

	err = s.CategoriesStore.Create(r.Context(), &category)
	if err != nil {
		if errors.Is(err, domain.ErrDuplicatedCategory) {
			Errorf(w, r, http.StatusBadRequest, err.Error())
		} else {
			Errorf(w, r, http.StatusInternalServerError, err.Error())
		}
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/products/%d", category.ID))
	err = ToJSON(w, domain.WrapCategory{Category: category}, http.StatusCreated)
	if err != nil {
		Errorf(w, r, http.StatusInternalServerError, err.Error())
	}
}

// @Summary      Update category
// @Tags 		 Categories
// @Security     Bearer
// @Produce      json
// @Accept       json
// @Param        category         body        domain.CategoryUpdate true "Update category"
// @Success      200              {array}     domain.WrapCategory
// @Failure      400              {object}    http.WrapError
// @Failure      403              {object}    http.WrapError
// @Failure      413              {object}    http.WrapError
// @Failure      500              {object}    http.WrapError
// @Router       /categories/{id} [patch]
func (s *Server) updateCategoryHandler(w http.ResponseWriter, r *http.Request) {
	ID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		ErrorInvalidQuery(w, r)
		return
	}

	input := domain.CategoryUpdate{}
	err = FromJSON(w, r, &input)
	if err != nil {
		Errorf(w, r, http.StatusBadRequest, err.Error())
		return
	}

	err = input.Validate()
	if err != nil {
		Errorf(w, r, http.StatusBadRequest, err.Error())
		return
	}

	category, err := s.CategoriesStore.Update(r.Context(), ID, input)
	if err != nil {
		if errors.Is(err, domain.ErrDuplicatedCategory) {
			Errorf(w, r, http.StatusBadRequest, err.Error())
		} else {
			Errorf(w, r, http.StatusInternalServerError, err.Error())
		}
		return
	}

	err = ToJSON(w, domain.WrapCategory{Category: category}, http.StatusOK)
	if err != nil {
		Errorf(w, r, http.StatusInternalServerError, err.Error())
	}
}

// @Summary      Delete category
// @Tags 		 Categories
// @Security     Bearer
// @Param        id           path        int  true "Category ID"
// @Success      200          {array}     domain.WrapCategory
// @Failure      400          {object}    http.WrapError
// @Failure      403          {object}    http.WrapError
// @Failure      404          {object}    http.WrapError
// @Failure      500          {object}    http.WrapError
// @Router       /categories/{id} [delete]
func (s *Server) deleteCategoryHandler(w http.ResponseWriter, r *http.Request) {
	ID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		ErrorInvalidQuery(w, r)
		return
	}

	err = s.CategoriesStore.Delete(r.Context(), ID)
	if err != nil {
		if errors.Is(err, domain.ErrNoCategoryFound) {
			Errorf(w, r, http.StatusNotFound, err.Error())
		} else {
			Errorf(w, r, http.StatusInternalServerError, err.Error())
		}
	}
}
