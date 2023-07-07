package http

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mortezadadgar/ecommerce-api/domain"
	"github.com/mortezadadgar/ecommerce-api/store"
)

func (s *Server) registerCartsRoutes(r *chi.Mux) {
	r.Route("/carts", func(r chi.Router) {
		r.Get("/", s.listCartsHandler)
		r.Get("/{id}", s.getCartsHandler)
		r.Post("/", s.postCartsHandler)
		r.Patch("/{id}", s.updateCartshandler)
		r.Delete("/{id}", s.deleteCartshandler)

		r.Route("/user", func(r chi.Router) {
			r.Get("/{id}", s.getUserCartsHandler)
		})
	})
}

// @Summary      Get cart
// @Tags 		 Carts
// @Security     Bearer
// @Produce      json
// @Param        id    path     int  true "Cart ID"
// @Success      200  {array}   domain.WrapCarts
// @Failure      400  {object}  http.HTTPError
// @Failure      404  {object}  http.HTTPError
// @Failure      500  {object}  http.HTTPError
// @Router       /carts/{id}    [get]
func (s *Server) getCartsHandler(w http.ResponseWriter, r *http.Request) {
	ID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, r, ErrInvalidQuery, http.StatusBadRequest)
		return
	}

	cart, err := s.CartsStore.GetByID(r.Context(), ID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			Error(w, r, ErrNotFound, http.StatusNotFound)
		default:
			Error(w, r, err, http.StatusInternalServerError)
		}

		return
	}

	err = ToJSON(w, domain.WrapCarts{Carts: cart}, http.StatusOK)
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
	}
}

// @Summary      Get user carts
// @Tags 		 Carts
// @Security     Bearer
// @Produce      json
// @Param        id    path       int  true "User ID"
// @Success      200  {array}     domain.WrapCartsList
// @Failure      400  {object}    http.HTTPError
// @Failure      404  {object}    http.HTTPError
// @Failure      500  {object}    http.HTTPError
// @Router       /carts/user/{id} [get]
func (s *Server) getUserCartsHandler(w http.ResponseWriter, r *http.Request) {
	UserID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, r, ErrInvalidQuery, http.StatusBadRequest)
		return
	}

	cart, err := s.CartsStore.GetByUser(r.Context(), UserID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			Error(w, r, ErrNotFound, http.StatusNotFound)
		default:
			Error(w, r, err, http.StatusInternalServerError)
		}

		return
	}

	err = ToJSON(w, domain.WrapCartsList{Carts: cart}, http.StatusOK)
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
	}
}

// @Summary      List carts
// @Tags 		 Carts
// @Security     Bearer
// @Param        limit        query       string  false "Limit results"
// @Param        offset       query       string  false "Offset results"
// @Param        sort         query       string  false "Sort by a column"
// @Success      200  {array}   domain.WrapCartsList
// @Failure      400  {object}  http.HTTPError
// @Failure      404  {object}  http.HTTPError
// @Failure      500  {object}  http.HTTPError
// @Router       /carts/        [get]
func (s *Server) listCartsHandler(w http.ResponseWriter, r *http.Request) {
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

	filter := domain.CartsFilter{
		Sort:   r.URL.Query().Get("sort"),
		Limit:  limit,
		Offset: offset,
	}

	carts, err := s.CartsStore.List(r.Context(), filter)
	if err != nil {
		switch err {
		case sql.ErrNoRows, store.ErrInvalidColumn:
			Error(w, r, ErrNotFound, http.StatusNotFound)
		default:
			Error(w, r, err, http.StatusInternalServerError)
		}

		return
	}

	err = ToJSON(w, domain.WrapCartsList{Carts: carts}, http.StatusOK)
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
	}
}

// @Summary      Create cart
// @Tags 		 Carts
// @Security     Bearer
// @Produce      json
// @Accept       json
// @Param        cart         body        domain.CartsCreate true "Create cart"
// @Success      201          {array}     domain.WrapCarts
// @Failure      400          {object}    http.HTTPError
// @Failure      403          {object}    http.HTTPError
// @Failure      404          {object}    http.HTTPError
// @Failure      413          {object}    http.HTTPError
// @Failure      500          {object}    http.HTTPError
// @Router       /carts/{id}  [post]
func (s *Server) postCartsHandler(w http.ResponseWriter, r *http.Request) {
	input := domain.CartsCreate{}
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

	cart := input.CreateModel()
	err = s.CartsStore.Create(r.Context(), &cart)
	if err != nil {
		switch err {
		case store.ErrForeinKeyViolation:
			Error(w, r, err, http.StatusBadRequest)
		default:
			Error(w, r, err, http.StatusInternalServerError)
		}

		return
	}

	err = ToJSON(w, domain.WrapCarts{Carts: cart}, http.StatusCreated)
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
	}
}

// @Summary      Update cart
// @Tags 		 Carts
// @Security     Bearer
// @Produce      json
// @Accept       json
// @Param        cart         body        domain.CartsUpdate true "Update carts"
// @Success      200          {array}     domain.WrapCarts
// @Failure      400          {object}    http.HTTPError
// @Failure      403          {object}    http.HTTPError
// @Failure      413          {object}    http.HTTPError
// @Failure      500          {object}    http.HTTPError
// @Router       /carts/{id}  [patch]
func (s *Server) updateCartshandler(w http.ResponseWriter, r *http.Request) {
	ID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, r, ErrInvalidQuery, http.StatusBadRequest)
		return
	}

	input := domain.CartsUpdate{}
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

	cart, err := s.CartsStore.GetByID(r.Context(), ID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			Error(w, r, ErrNotFound, http.StatusNotFound)
		default:
			Error(w, r, err, http.StatusInternalServerError)
		}

		return
	}

	input.UpdateModel(&cart)

	err = s.CartsStore.Update(r.Context(), &cart)
	if err != nil {
		switch err {
		case store.ErrForeinKeyViolation, store.ErrDuplicatedEntries:
			Error(w, r, err, http.StatusBadRequest)
		default:
			Error(w, r, err, http.StatusInternalServerError)
		}

		return
	}

	err = ToJSON(w, domain.WrapCarts{Carts: cart}, http.StatusOK)
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
	}

}

// @Summary      Delete carts
// @Tags 		 Carts
// @Security     Bearer
// @Param        id           path        int  true "cart ID"
// @Success      200          {array}     domain.WrapCarts
// @Failure      400          {object}    http.HTTPError
// @Failure      403          {object}    http.HTTPError
// @Failure      404          {object}    http.HTTPError
// @Failure      500          {object}    http.HTTPError
// @Router       /carts/{id}  [delete]
func (s *Server) deleteCartshandler(w http.ResponseWriter, r *http.Request) {
	ID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, r, ErrInvalidQuery, http.StatusBadRequest)
		return
	}

	err = s.CartsStore.Delete(r.Context(), ID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			Error(w, r, ErrNotFound, http.StatusNotFound)
		default:
			Error(w, r, err, http.StatusInternalServerError)
		}
	}
}
