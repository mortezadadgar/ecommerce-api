package http

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mortezadadgar/ecommerce-api/domain"
)

func (s *Server) registerProductsRoutes() {
	s.Route("/products", func(r chi.Router) {
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
// @Failure      400            {object}    domain.Error
// @Failure      404            {object}    domain.Error
// @Failure      500            {object}    domain.Error
// @Router       /products/{id} [get]
func (s *Server) getProductHandler(w http.ResponseWriter, r *http.Request) {
	ID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		ErrorInvalidQuery(w, r)
		return
	}

	product, err := s.ProductsStore.GetByID(r.Context(), ID)
	if err != nil {
		Error(w, r, err)
		return
	}

	err = ToJSON(w, domain.WrapProduct{Product: product}, http.StatusOK)
	if err != nil {
		Error(w, r, err)
	}
}

// @Summary      List products
// @Tags 		 Products
// @Param        limit        query       string  false "Limit results"
// @Param        offset       query       string  false "Offset results"
// @Param        name         query       string  false "List by name"
// @Param        category     query       string  false "List by category"
// @Param        sort         query       string  false "Sort by a column"
// @Success      200          {array}     domain.WrapProductList
// @Failure      400          {object}    domain.Error
// @Failure      404          {object}    domain.Error
// @Failure      500          {object}    domain.Error
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

	filter := domain.ProductFilter{
		Name:     r.URL.Query().Get("name"),
		Sort:     r.URL.Query().Get("sort"),
		Category: r.URL.Query().Get("category"),
		Limit:    limit,
		Offset:   offset,
	}

	products, err := s.ProductsStore.List(r.Context(), filter)
	if err != nil {
		Error(w, r, err)
		return
	}

	err = ToJSON(w, domain.WrapProductList{Products: products}, http.StatusOK)
	if err != nil {
		Error(w, r, err)
	}
}

// @Summary      Create product
// @Tags 		 Products
// @Security     Bearer
// @Produce      json
// @Accept       json
// @Param        product          body        domain.ProductCreate true "Create product"
// @Success      201              {array}     domain.WrapProduct
// @Failure      400              {object}    domain.Error
// @Failure      403              {object}    domain.Error
// @Failure      413              {object}    domain.Error
// @Failure      500              {object}    domain.Error
// @Router       /products/{id}   [post]
func (s *Server) createProductHandler(w http.ResponseWriter, r *http.Request) {
	input := domain.ProductCreate{}
	err := FromJSON(w, r, &input)
	if err != nil {
		Error(w, r, err)
		return
	}

	err = input.Validate()
	if err != nil {
		Error(w, r, domain.Errorf(domain.EINVALID, err.Error()))
		return
	}

	product := input.CreateModel()

	err = s.ProductsStore.Create(r.Context(), &product)
	if err != nil {
		Error(w, r, err)
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/products/%d", product.ID))
	err = ToJSON(w, domain.WrapProduct{Product: product}, http.StatusCreated)
	if err != nil {
		Error(w, r, err)
	}
}

// @Summary      Update product
// @Tags 		 Products
// @Security     Bearer
// @Produce      json
// @Accept       json
// @Param        product         body        domain.ProductUpdate true "Update product"
// @Success      200             {array}     domain.WrapProduct
// @Failure      400             {object}    domain.Error
// @Failure      403             {object}    domain.Error
// @Failure      413             {object}    domain.Error
// @Failure      500             {object}    domain.Error
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
		Error(w, r, err)
		return
	}

	err = input.Validate()
	if err != nil {
		Error(w, r, domain.Errorf(domain.EINVALID, err.Error()))
		return
	}

	product, err := s.ProductsStore.GetByID(r.Context(), ID)
	if err != nil {
		Error(w, r, err)
		return
	}

	input.UpdateModel(&product)

	err = s.ProductsStore.Update(r.Context(), &product)
	if err != nil {
		Error(w, r, err)
		return
	}

	err = ToJSON(w, domain.WrapProduct{Product: product}, http.StatusOK)
	if err != nil {
		Error(w, r, err)
	}
}

// @Summary      Delete product
// @Tags 		 Products
// @Security     Bearer
// @Param        id             path        int  true "Product ID"
// @Success      200
// @Failure      400            {object}    domain.Error
// @Failure      403            {object}    domain.Error
// @Failure      404            {object}    domain.Error
// @Failure      500            {object}    domain.Error
// @Router       /products/{id} [delete]
func (s *Server) deleteProductHandler(w http.ResponseWriter, r *http.Request) {
	ID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, r, domain.Errorf(domain.EINVALID, err.Error()))
		return
	}

	err = s.ProductsStore.Delete(r.Context(), ID)
	if err != nil {
		Error(w, r, err)
	}
}
