package http

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mortezadadgar/ecommerce-api"
	"github.com/mortezadadgar/ecommerce-api/postgres"
)

func (s *Server) registerProductsRoutes(r *chi.Mux) {
	r.Route("/products", func(r chi.Router) {
		r.Get("/{id}", s.productGetHandler)
		r.Get("/", s.productsListHandler)
		r.Post("/", s.productCreateHandler)
		r.Patch("/{id}", s.productUpdateHandler)
		r.Delete("/{id}", s.productDeleteHandler)
	})

}

// @Summary      Get product
// @Tags 		 Products
// @Produce      json
// @Param        id             path        int  true "Product ID"
// @Success      200            {array}     ecommerce.WrapCategories
// @Failure      400            {object}    http.HTTPError
// @Failure      500            {object}    http.HTTPError
// @Router       /products/{id} [get]
func (s *Server) productGetHandler(w http.ResponseWriter, r *http.Request) {
	ID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, r, ErrInvalidQuery, http.StatusBadRequest)
		return
	}

	product, err := s.ProductsStore.GetByID(r.Context(), ID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			Error(w, r, ErrNotFound, http.StatusNotFound)
		default:
			Error(w, r, err, http.StatusInternalServerError)
		}

		return
	}

	err = ToJSON(w, ecommerce.WrapProducts{Product: *product})
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
	}
}

// @Summary      List products
// @Tags 		 Products
// @Param        limit        query       string  false "Limit results"
// @Param        offset       query       string  false "Offset results"
// @Param        name         query       string  false "List by name"
// @Param        category     query       string  false "List by category"
// @Param        sort         query       string  false "Sort by a column"
// @Success      200          {array}     ecommerce.WrapCategories
// @Failure      400          {object}    http.HTTPError
// @Failure      500          {object}    http.HTTPError
// @Router       /products/   [get]
func (s *Server) productsListHandler(w http.ResponseWriter, r *http.Request) {
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

	fmt.Println(limit, offset)

	filter := ecommerce.ProductsFilter{
		Name:     r.URL.Query().Get("name"),
		Sort:     r.URL.Query().Get("sort"),
		Category: r.URL.Query().Get("category"),
		Limit:    limit,
		Offset:   offset,
	}

	products, err := s.ProductsStore.List(r.Context(), filter)
	if err != nil {
		switch err {
		case sql.ErrNoRows, postgres.ErrInvalidColumn:
			Error(w, r, ErrNotFound, http.StatusNotFound)
		default:
			Error(w, r, err, http.StatusInternalServerError)
		}

		return
	}

	err = ToJSON(w, ecommerce.WrapProductsList{Products: *products})
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
	}
}

// @Summary      Create product
// @Tags 		 Products
// @Produce      json
// @Accept       json
// @Param        product      body         ecommerce.ProductsInput true "Update product"
// @Success      200          {array}     ecommerce.WrapProducts
// @Failure      400          {object}    http.HTTPError
// @Failure      500          {object}    http.HTTPError
// @Router       /products/   [post]
func (s *Server) productCreateHandler(w http.ResponseWriter, r *http.Request) {
	input := ecommerce.ProductsInput{}
	err := FromJSON(r, &input)
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
	}

	product := ecommerce.Products{}

	input.SetValuesTo(&product)

	err = product.Validate()
	if err != nil {
		Error(w, r, err, http.StatusBadRequest)
		return
	}

	err = s.ProductsStore.Create(r.Context(), &product)
	if err != nil {
		switch err {
		case postgres.ErrForeinKeyViolation, postgres.ErrDuplicatedEntries:
			Error(w, r, err, http.StatusBadRequest)
		default:
			Error(w, r, err, http.StatusInternalServerError)
		}

		return
	}

	w.Header().Set("Location", fmt.Sprintf("/products/%d", product.ID))
	err = ToJSON(w, ecommerce.WrapProducts{Product: product})
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
	}
}

// @Summary      Update product
// @Tags 		 Products
// @Produce      json
// @Accept       json
// @Param        product     body         ecommerce.CategoriesInput true "Update products"
// @Success      200          {array}     ecommerce.WrapProducts
// @Failure      400          {object}    http.HTTPError
// @Failure      413          {object}    http.HTTPError
// @Failure      500          {object}    http.HTTPError
// @Router       /products/ [patch]
func (s *Server) productUpdateHandler(w http.ResponseWriter, r *http.Request) {
	ID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, r, ErrInvalidQuery, http.StatusBadRequest)
		return
	}

	input := ecommerce.ProductsInput{}
	err = FromJSON(r, &input)
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
	}

	product, err := s.ProductsStore.GetByID(r.Context(), ID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			Error(w, r, ErrNotFound, http.StatusNotFound)
		default:
			Error(w, r, err, http.StatusInternalServerError)
		}

		return
	}

	input.SetValuesTo(product)

	err = product.Validate()
	if err != nil {
		Error(w, r, err, http.StatusBadRequest)
		return
	}

	err = s.ProductsStore.Update(r.Context(), product)
	if err != nil {
		switch err {
		case postgres.ErrForeinKeyViolation, postgres.ErrDuplicatedEntries:
			Error(w, r, err, http.StatusBadRequest)
		default:
			Error(w, r, err, http.StatusInternalServerError)
		}

		return
	}

	err = ToJSON(w, ecommerce.WrapProducts{Product: *product})
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
	}
}

// @Summary      Delete product
// @Tags 		 Products
// @Param        id           path        int  true "Product ID"
// @Success      200          {array}     ecommerce.WrapCategories
// @Failure      400          {object}    http.HTTPError
// @Failure      500          {object}    http.HTTPError
// @Router       /products/{id} [delete]
func (s *Server) productDeleteHandler(w http.ResponseWriter, r *http.Request) {
	ID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, r, ErrInvalidQuery, http.StatusBadRequest)
	}

	err = s.ProductsStore.Delete(r.Context(), ID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			Error(w, r, ErrNotFound, http.StatusNotFound)
		default:
			Error(w, r, err, http.StatusInternalServerError)
		}
	}
}
