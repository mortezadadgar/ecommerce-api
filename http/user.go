package http

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mortezadadgar/ecommerce-api/domain"
	"github.com/mortezadadgar/ecommerce-api/store"
	"golang.org/x/crypto/bcrypt"
)

// ErrUnauthorizedAccess returns when user has invalid authentication token
// or is not provided.
var ErrUnauthorizedAccess = errors.New("unauthorized access")

func (s *Server) registerUsersRoutes(r *chi.Mux) {
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
// @Failure      400            {object}    http.HTTPError
// @Failure      403            {object}    http.HTTPError
// @Failure      404            {object}    http.HTTPError
// @Failure      500            {object}    http.HTTPError
// @Router       /users/{id}    [get]
func (s *Server) getUserHandler(w http.ResponseWriter, r *http.Request) {
	ID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, r, ErrInvalidQuery, http.StatusBadRequest)
		return
	}

	user, err := s.UsersStore.GetByID(r.Context(), ID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			Error(w, r, ErrNotFound, http.StatusNotFound)
		default:
			Error(w, r, err, http.StatusInternalServerError)
		}

		return
	}

	err = ToJSON(w, domain.WrapUser{User: user}, http.StatusOK)
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
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
// @Failure      400            {object}    http.HTTPError
// @Failure      403            {object}    http.HTTPError
// @Failure      404            {object}    http.HTTPError
// @Failure      500            {object}    http.HTTPError
// @Router       /users/        [get]
func (s *Server) listUsersHandler(w http.ResponseWriter, r *http.Request) {
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

	filter := domain.UserFilter{
		Sort:   r.URL.Query().Get("sort"),
		Limit:  limit,
		Offset: offset,
	}

	users, err := s.UsersStore.List(r.Context(), filter)
	if err != nil {
		switch err {
		case sql.ErrNoRows, store.ErrInvalidColumn:
			Error(w, r, ErrNotFound, http.StatusNotFound)
		default:
			Error(w, r, err, http.StatusInternalServerError)
		}

		return
	}

	err = ToJSON(w, domain.WrapUserList{Users: users}, http.StatusOK)
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
	}
}

// @Summary      Create user
// @Tags 		 Users
// @Security     Bearer
// @Produce      json
// @Accept       json
// @Param        user         body        domain.UserCreate true "Create User"
// @Success      201          {array}     domain.WrapUser
// @Failure      400          {object}    http.HTTPError
// @Failure      403          {object}    http.HTTPError
// @Failure      413          {object}    http.HTTPError
// @Failure      500          {object}    http.HTTPError
// @Router       /users/      [post]
func (s *Server) createUserHandler(w http.ResponseWriter, r *http.Request) {
	input := domain.UserCreate{}
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

	hashedPassword, err := domain.GenerateHashedPassword([]byte(input.Password))
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
		return
	}

	user := input.CreateModel(hashedPassword)

	err = s.UsersStore.Create(r.Context(), &user)
	if err != nil {
		switch err {
		case store.ErrForeinKeyViolation, store.ErrDuplicatedEntries:
			Error(w, r, err, http.StatusBadRequest)
		default:
			Error(w, r, err, http.StatusInternalServerError)
		}

		return
	}

	w.Header().Set("Location", fmt.Sprintf("/users/%d", user.ID))
	err = ToJSON(w, domain.WrapUser{User: user}, http.StatusCreated)
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
	}
}

// @Summary      Delete User
// @Tags 		 Users
// @Security     Bearer
// @Param        id             path        int  true "User ID"
// @Success      200
// @Failure      400            {object}    http.HTTPError
// @Failure      403            {object}    http.HTTPError
// @Failure      404            {object}    http.HTTPError
// @Failure      500            {object}    http.HTTPError
// @Router       /users/{id}    [delete]
func (s *Server) deleteUserHandler(w http.ResponseWriter, r *http.Request) {
	ID, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		Error(w, r, ErrInvalidQuery, http.StatusBadRequest)
		return
	}

	err = s.UsersStore.Delete(r.Context(), ID)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			Error(w, r, ErrNotFound, http.StatusNotFound)
		default:
			Error(w, r, err, http.StatusInternalServerError)
		}
	}
}

// @Summary      User login
// @Tags 		 Users
// @Produce      json
// @Accept       json
// @Param        user         body        domain.UserLogin true "Create User"
// @Success      200          {array}     domain.WrapUser
// @Failure      400          {object}    http.HTTPError
// @Failure      404          {object}    http.HTTPError
// @Failure      413          {object}    http.HTTPError
// @Failure      500          {object}    http.HTTPError
// @Router       /auth/login  [post]
func (s *Server) loginAuthHandler(w http.ResponseWriter, r *http.Request) {
	input := domain.UserLogin{}
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

	user, err := s.UsersStore.GetByEmail(r.Context(), input.Email)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			Error(w, r, ErrNotFound, http.StatusNotFound)
		default:
			Error(w, r, err, http.StatusInternalServerError)
		}

		return
	}

	err = domain.CompareHashAndPassword(user.Password, []byte(input.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			Error(w, r, ErrUnauthorizedAccess, http.StatusUnauthorized)
		} else {
			Error(w, r, err, http.StatusInternalServerError)
		}

		return
	}

	token, err := domain.GenerateToken(user.ID, 16, 24*3*time.Hour)
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
		return
	}

	err = s.TokensStore.Create(r.Context(), token)
	if err != nil {
		switch err {
		case store.ErrForeinKeyViolation:
			Error(w, r, err, http.StatusBadRequest)
		default:
			Error(w, r, err, http.StatusInternalServerError)
		}

		return
	}

	err = ToJSON(w, domain.WrapToken{Token: token}, http.StatusOK)
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
	}
}