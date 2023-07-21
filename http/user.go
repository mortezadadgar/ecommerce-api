package http

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mortezadadgar/ecommerce-api/domain"
	"golang.org/x/crypto/bcrypt"
)

func (s *server) registerUsersRoutes(r *chi.Mux) {
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", s.loginAuthHandler)
		// sign_up
		// log_out
		// forget_password
	})

	r.With(requireAuth).Route("/users", func(r chi.Router) {
		r.Get("/{id}", s.getUserHandler)
		r.Get("/", s.listUsersHandler)
		r.Post("/", s.createUserHandler)
		r.Delete("/{id}", s.deleteUserHandler)
	})
}

// @Summary      Get user by id
// @Tags 		 Users
// @Security     Bearer
// @Produce      json
// @Param        id             path        int  true "User ID"
// @Success      200            {array}     domain.WrapUser
// @Failure      400            {object}    http.WrapError
// @Failure      403            {object}    http.WrapError
// @Failure      404            {object}    http.WrapError
// @Failure      500            {object}    http.WrapError
// @Router       /users/{id}    [get]
func (s *server) getUserHandler(w http.ResponseWriter, r *http.Request) {
	ID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		ErrorInvalidQuery(w, r)
		return
	}

	user, err := s.UsersStore.GetByID(r.Context(), ID)
	if err != nil {
		if errors.Is(err, domain.ErrNoUsersFound) {
			Errorf(w, r, http.StatusNotFound, err.Error())
		} else {
			Errorf(w, r, http.StatusInternalServerError, err.Error())
		}
		return
	}

	err = ToJSON(w, domain.WrapUser{User: user}, http.StatusOK)
	if err != nil {
		Errorf(w, r, http.StatusInternalServerError, err.Error())
	}
}

// @Summary      List users
// @Tags 		 Users
// @Security     Bearer
// @Produce      json
// @Param        limit          query       string  false "Limit results"
// @Param        offset         query       string  false "Offset results"
// @Param        sort           query       string  false "Sort by a column"
// @Success      200            {array}     domain.WrapUserList
// @Failure      400            {object}    http.WrapError
// @Failure      403            {object}    http.WrapError
// @Failure      404            {object}    http.WrapError
// @Failure      500            {object}    http.WrapError
// @Router       /users/        [get]
func (s *server) listUsersHandler(w http.ResponseWriter, r *http.Request) {
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

	filter := domain.UserFilter{
		Sort:   r.URL.Query().Get("sort"),
		Limit:  limit,
		Offset: offset,
	}

	users, err := s.UsersStore.List(r.Context(), filter)
	if err != nil {
		if errors.Is(err, domain.ErrNoUsersFound) {
			Errorf(w, r, http.StatusNotFound, err.Error())
		} else {
			Errorf(w, r, http.StatusInternalServerError, err.Error())
		}
		return
	}

	err = ToJSON(w, domain.WrapUserList{Users: users}, http.StatusOK)
	if err != nil {
		Errorf(w, r, http.StatusInternalServerError, err.Error())
	}
}

// @Summary      Create user
// @Tags 		 Users
// @Security     Bearer
// @Produce      json
// @Accept       json
// @Param        user         body        domain.UserCreate true "Create User"
// @Success      201          {array}     domain.WrapUser
// @Failure      400          {object}    http.WrapError
// @Failure      403          {object}    http.WrapError
// @Failure      413          {object}    http.WrapError
// @Failure      500          {object}    http.WrapError
// @Router       /users/      [post]
func (s *server) createUserHandler(w http.ResponseWriter, r *http.Request) {
	input := domain.UserCreate{}
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

	hashedPassword, err := domain.GenerateHashedPassword([]byte(input.Password))
	if err != nil {
		Errorf(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	user := input.CreateModel(hashedPassword)

	err = s.UsersStore.Create(r.Context(), &user)
	if err != nil {
		if errors.Is(err, domain.ErrDuplicatedUserEmail) {
			Errorf(w, r, http.StatusBadRequest, err.Error())
		} else {
			Errorf(w, r, http.StatusInternalServerError, err.Error())
		}
		return
	}

	w.Header().Set("Location", fmt.Sprintf("/users/%d", user.ID))
	err = ToJSON(w, domain.WrapUser{User: user}, http.StatusCreated)
	if err != nil {
		Errorf(w, r, http.StatusInternalServerError, err.Error())
	}
}

// @Summary      Delete User
// @Tags 		 Users
// @Security     Bearer
// @Param        id             path        int  true "User ID"
// @Success      200
// @Failure      400            {object}    http.WrapError
// @Failure      403            {object}    http.WrapError
// @Failure      404            {object}    http.WrapError
// @Failure      500            {object}    http.WrapError
// @Router       /users/{id}    [delete]
func (s *server) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	ID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		ErrorInvalidQuery(w, r)
		return
	}

	err = s.UsersStore.Delete(r.Context(), ID)
	if err != nil {
		if errors.Is(err, domain.ErrNoTokenFound) {
			Errorf(w, r, http.StatusNotFound, err.Error())
		} else {
			Errorf(w, r, http.StatusInternalServerError, err.Error())
		}
	}
}

// @Summary      User login
// @Tags 		 Users
// @Produce      json
// @Accept       json
// @Param        user         body        domain.UserLogin true "Create User"
// @Success      200          {array}     domain.WrapUser
// @Failure      400          {object}    http.WrapError
// @Failure      404          {object}    http.WrapError
// @Failure      413          {object}    http.WrapError
// @Failure      500          {object}    http.WrapError
// @Router       /auth/login  [post]
func (s *server) loginAuthHandler(w http.ResponseWriter, r *http.Request) {
	input := domain.UserLogin{}
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

	user, err := s.UsersStore.GetByEmail(r.Context(), input.Email)
	if err != nil {
		if errors.Is(err, domain.ErrNoUsersFound) {
			Errorf(w, r, http.StatusNotFound, err.Error())
		} else {
			Errorf(w, r, http.StatusInternalServerError, err.Error())
		}
		return
	}

	err = domain.CompareHashAndPassword(user.Password, []byte(input.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			Errorf(w, r, http.StatusUnauthorized, "unauthorized access")
		} else {
			Errorf(w, r, http.StatusInternalServerError, err.Error())
		}
		return
	}

	token, err := domain.GenerateToken(user.ID, 16, 24*3*time.Hour)
	if err != nil {
		Errorf(w, r, http.StatusInternalServerError, err.Error())
		return
	}

	err = s.TokensStore.Create(r.Context(), token)
	if err != nil {
		if errors.Is(err, domain.ErrNoTokenFound) {
			Errorf(w, r, http.StatusNotFound, err.Error())
		} else {
			Errorf(w, r, http.StatusInternalServerError, err.Error())
		}
		return
	}

	err = ToJSON(w, domain.WrapToken{Token: token}, http.StatusOK)
	if err != nil {
		Errorf(w, r, http.StatusInternalServerError, err.Error())
	}
}
