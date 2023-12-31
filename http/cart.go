package http

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mortezadadgar/ecommerce-api/domain"
)

func (s *server) registerCartsRoutes(r *chi.Mux) {
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
// @Success      200  {array}   domain.WrapCart
// @Failure      400  {object}  http.WrapError
// @Failure      404  {object}  http.WrapError
// @Failure      500  {object}  http.WrapError
// @Router       /carts/{id}    [get]
func (s *server) getCartsHandler(w http.ResponseWriter, r *http.Request) {
	ID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		ErrorInvalidQuery(w, r)
		return
	}

	cart, err := s.CartsStore.GetByID(r.Context(), ID)
	if err != nil {
		if errors.Is(err, domain.ErrNoCartsFound) {
			Errorf(w, r, http.StatusNotFound, err.Error())
		} else {
			Errorf(w, r, http.StatusInternalServerError, err.Error())
		}
		return
	}

	err = ToJSON(w, domain.WrapCart{Cart: cart}, http.StatusOK)
	if err != nil {
		Errorf(w, r, http.StatusInternalServerError, err.Error())
	}
}

// @Summary      Get user carts
// @Tags 		 Carts
// @Security     Bearer
// @Produce      json
// @Param        id    path       int  true "User ID"
// @Success      200  {array}     domain.WrapCartList
// @Failure      400  {object}    http.WrapError
// @Failure      404  {object}    http.WrapError
// @Failure      500  {object}    http.WrapError
// @Router       /carts/user/{id} [get]
func (s *server) getUserCartsHandler(w http.ResponseWriter, r *http.Request) {
	UserID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		ErrorInvalidQuery(w, r)
		return
	}

	cart, err := s.CartsStore.GetByUser(r.Context(), UserID)
	if err != nil {
		if errors.Is(err, domain.ErrNoCartsFound) {
			Errorf(w, r, http.StatusNotFound, err.Error())
		} else {
			Errorf(w, r, http.StatusInternalServerError, err.Error())
		}
		return
	}

	err = ToJSON(w, domain.WrapCartList{Carts: cart}, http.StatusOK)
	if err != nil {
		Errorf(w, r, http.StatusInternalServerError, err.Error())
	}
}

// @Summary      List carts
// @Tags 		 Carts
// @Security     Bearer
// @Param        limit        query       string  false "Limit results"
// @Param        offset       query       string  false "Offset results"
// @Param        sort         query       string  false "Sort by a column"
// @Success      200  {array}   domain.WrapCartList
// @Failure      400  {object}  http.WrapError
// @Failure      404  {object}  http.WrapError
// @Failure      500  {object}  http.WrapError
// @Router       /carts/        [get]
func (s *server) listCartsHandler(w http.ResponseWriter, r *http.Request) {
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

	filter := domain.CartFilter{
		Sort:   r.URL.Query().Get("sort"),
		Limit:  limit,
		Offset: offset,
	}

	carts, err := s.CartsStore.List(r.Context(), filter)
	if err != nil {
		if errors.Is(err, domain.ErrNoCartsFound) {
			Errorf(w, r, http.StatusNotFound, err.Error())
		} else {
			Errorf(w, r, http.StatusInternalServerError, err.Error())
		}
		return
	}

	err = ToJSON(w, domain.WrapCartList{Carts: carts}, http.StatusOK)
	if err != nil {
		Errorf(w, r, http.StatusInternalServerError, err.Error())
	}
}

// @Summary      Create cart
// @Tags 		 Carts
// @Security     Bearer
// @Produce      json
// @Accept       json
// @Param        cart         body        domain.CartCreate true "Create cart"
// @Success      201          {array}     domain.WrapCart
// @Failure      400          {object}    http.WrapError
// @Failure      403          {object}    http.WrapError
// @Failure      404          {object}    http.WrapError
// @Failure      413          {object}    http.WrapError
// @Failure      500          {object}    http.WrapError
// @Router       /carts/{id}  [post]
func (s *server) postCartsHandler(w http.ResponseWriter, r *http.Request) {
	input := domain.CartCreate{}
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

	cart := input.CreateModel()
	err = s.CartsStore.Create(r.Context(), &cart)
	if err != nil {
		if errors.Is(err, domain.ErrCartInvalidUserID) ||
			errors.Is(err, domain.ErrCartInvalidProductID) {
			Errorf(w, r, http.StatusBadRequest, err.Error())
		} else {
			Errorf(w, r, http.StatusInternalServerError, err.Error())
		}
		return
	}

	err = ToJSON(w, domain.WrapCart{Cart: cart}, http.StatusCreated)
	if err != nil {
		Errorf(w, r, http.StatusInternalServerError, err.Error())
	}
}

// @Summary      Update cart
// @Tags 		 Carts
// @Security     Bearer
// @Produce      json
// @Accept       json
// @Param        cart         body        domain.CartUpdate true "Update cart"
// @Success      200          {array}     domain.WrapCart
// @Failure      400          {object}    http.WrapError
// @Failure      403          {object}    http.WrapError
// @Failure      413          {object}    http.WrapError
// @Failure      500          {object}    http.WrapError
// @Router       /carts/{id}  [patch]
func (s *server) updateCartshandler(w http.ResponseWriter, r *http.Request) {
	ID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		ErrorInvalidQuery(w, r)
		return
	}

	input := domain.CartUpdate{}
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

	cart, err := s.CartsStore.Update(r.Context(), ID, input)
	if err != nil {
		if errors.Is(err, domain.ErrCartInvalidProductID) {
			Errorf(w, r, http.StatusBadRequest, err.Error())
		} else if errors.Is(err, domain.ErrNoCartsFound) {
			Errorf(w, r, http.StatusNotFound, err.Error())
		} else {
			Errorf(w, r, http.StatusInternalServerError, err.Error())
		}
		return
	}

	err = ToJSON(w, domain.WrapCart{Cart: cart}, http.StatusOK)
	if err != nil {
		Errorf(w, r, http.StatusInternalServerError, err.Error())
	}
}

// @Summary      Delete carts
// @Tags 		 Carts
// @Security     Bearer
// @Param        id           path        int  true "cart ID"
// @Success      200          {array}     domain.WrapCart
// @Failure      400          {object}    http.WrapError
// @Failure      403          {object}    http.WrapError
// @Failure      404          {object}    http.WrapError
// @Failure      500          {object}    http.WrapError
// @Router       /carts/{id}  [delete]
func (s *server) deleteCartshandler(w http.ResponseWriter, r *http.Request) {
	ID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		ErrorInvalidQuery(w, r)
		return
	}

	err = s.CartsStore.Delete(r.Context(), ID)
	if err != nil {
		if errors.Is(err, domain.ErrNoCartsFound) {
			Errorf(w, r, http.StatusNotFound, err.Error())
		} else {
			Errorf(w, r, http.StatusInternalServerError, err.Error())
		}
	}
}
