package http

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mortezadadgar/ecommerce-api/domain"
)

func (s *Server) registerProductsRoutes(r *chi.Mux) {
	r.Route("/products", func(r chi.Router) {
		r.Get("/{id}", s.getProductHandler)
		r.Get("/", s.listProductsHandler)
		r.With(requireAuth).Post("/", s.createProductHandler)
		r.With(requireAuth).Patch("/{id}", s.updateProductHandler)
		r.With(requireAuth).Delete("/{id}", s.deleteProductHandler)
		// bulk inserts data to db
	})
}

// @Summary      Get product by id
// @Tags 		 Products
// @Produce      json
// @Param        id             path        int  true "Product ID"
// @Success      200            {array}     domain.WrapProduct
// @Failure      400            {object}    http.WrapError
// @Failure      404            {object}    http.WrapError
// @Failure      500            {object}    http.WrapError
// @Router       /products/{id} [get]
func (s *Server) getProductHandler(w http.ResponseWriter, r *http.Request) {
	ID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		ErrorInvalidQuery(w, r)
		return
	}

	product, err := s.ProductsStore.GetByID(r.Context(), ID)
	if err != nil {
		if errors.Is(err, domain.ErrNoProductsFound) {
			Errorf(w, r, http.StatusNotFound, err.Error())
		} else {
			Errorf(w, r, http.StatusInternalServerError, err.Error())
		}
		return
	}

	err = ToJSON(w, domain.WrapProduct{Product: product}, http.StatusOK)
	if err != nil {
		Errorf(w, r, http.StatusInternalServerError, err.Error())
	}
}

// @Summary      List products
// @Tags 		 Products
// @Param        limit        query       string  false "Limit results"
// @Param        offset       query       string  false "Offset results"
// @Param        category_id  query       string  false "List by category id"
// @Param        sort         query       string  false "Sort by a column"
// @Success      200          {array}     domain.WrapProductList
// @Failure      400          {object}    http.WrapError
// @Failure      404          {object}    http.WrapError
// @Failure      500          {object}    http.WrapError
// @Router       /products/   [get]
func (s *Server) listProductsHandler(w http.ResponseWriter, r *http.Request) {
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

	category, err := ParseIntQuery(r, "category_id")
	if err != nil {
		ErrorInvalidQuery(w, r)
		return
	}

	filter := domain.ProductFilter{
		Sort:       r.URL.Query().Get("sort"),
		CategoryID: category,
		Limit:      limit,
		Offset:     offset,
	}

	products, err := s.ProductsStore.List(r.Context(), filter)
	if err != nil {
		if errors.Is(err, domain.ErrNoProductsFound) {
			Errorf(w, r, http.StatusNotFound, err.Error())
		} else {
			Errorf(w, r, http.StatusInternalServerError, err.Error())
		}
		return
	}

	err = ToJSON(w, domain.WrapProductList{Products: products}, http.StatusOK)
	if err != nil {
		Errorf(w, r, http.StatusInternalServerError, err.Error())
	}
}

// @Summary      Create product
// @Tags 		 Products
// @Security     Bearer
// @Produce      json
// @Accept       json
// @Param        product          body        domain.ProductCreate true "Create product"
// @Success      201              {array}     domain.WrapProduct
// @Failure      400              {object}    http.WrapError
// @Failure      403              {object}    http.WrapError
// @Failure      413              {object}    http.WrapError
// @Failure      500              {object}    http.WrapError
// @Router       /products/{id}   [post]
func (s *Server) createProductHandler(w http.ResponseWriter, r *http.Request) {
	input := domain.ProductCreate{}
	err := FromJSON(w, r, &input)
	if err != nil {
		Errorf(w, r, http.StatusBadRequest, err.Error())
		return
	}

	err = input.Validate()
	if err != nil {
		Errorf(w, r, http.StatusBadRequest, err.Error())
	}

	product := input.CreateModel()

	err = s.ProductsStore.Create(r.Context(), &product)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidProductCategory) ||
			errors.Is(err, domain.ErrDuplicatedProduct) {
			Errorf(w, r, http.StatusBadRequest, err.Error())
		} else {
			Errorf(w, r, http.StatusInternalServerError, err.Error())
		}
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/products/%d", product.ID))
	err = ToJSON(w, domain.WrapProduct{Product: product}, http.StatusCreated)
	if err != nil {
		Errorf(w, r, http.StatusInternalServerError, err.Error())
	}
}

// @Summary      Update product
// @Tags 		 Products
// @Security     Bearer
// @Produce      json
// @Accept       json
// @Param        product         body        domain.ProductUpdate true "Update product"
// @Success      200             {array}     domain.WrapProduct
// @Failure      400             {object}    http.WrapError
// @Failure      403             {object}    http.WrapError
// @Failure      413             {object}    http.WrapError
// @Failure      500             {object}    http.WrapError
// @Router       /products/{id}  [patch]
func (s *Server) updateProductHandler(w http.ResponseWriter, r *http.Request) {
	ID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		ErrorInvalidQuery(w, r)
		return
	}

	input := domain.ProductUpdate{}
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

	product, err := s.ProductsStore.GetByID(r.Context(), ID)
	if err != nil {
		if errors.Is(err, domain.ErrNoProductsFound) {
			Errorf(w, r, http.StatusNotFound, err.Error())
		} else {
			Errorf(w, r, http.StatusInternalServerError, err.Error())
		}
		return
	}

	input.UpdateModel(&product)

	err = s.ProductsStore.Update(r.Context(), &product)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidProductCategory) ||
			errors.Is(err, domain.ErrDuplicatedProduct) {
			Errorf(w, r, http.StatusBadRequest, err.Error())
		} else {
			Errorf(w, r, http.StatusInternalServerError, err.Error())
		}
		return
	}

	err = ToJSON(w, domain.WrapProduct{Product: product}, http.StatusOK)
	if err != nil {
		Errorf(w, r, http.StatusInternalServerError, err.Error())
	}
}

// @Summary      Delete product
// @Tags 		 Products
// @Security     Bearer
// @Param        id             path        int  true "Product ID"
// @Success      200
// @Failure      400            {object}    http.WrapError
// @Failure      403            {object}    http.WrapError
// @Failure      404            {object}    http.WrapError
// @Failure      500            {object}    http.WrapError
// @Router       /products/{id} [delete]
func (s *Server) deleteProductHandler(w http.ResponseWriter, r *http.Request) {
	ID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		Errorf(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	err = s.ProductsStore.Delete(r.Context(), ID)
	if err != nil {
		if errors.Is(err, domain.ErrNoProductsFound) {
			Errorf(w, r, http.StatusNotFound, err.Error())
		} else {
			Errorf(w, r, http.StatusInternalServerError, err.Error())
		}
	}
}
