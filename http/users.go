package http

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mortezadadgar/ecommerce-api/domain"
	"github.com/mortezadadgar/ecommerce-api/postgres"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) registerUsersRoutes(r *chi.Mux) {
	// authentication
	r.Route("/auth", func(r chi.Router) {
		r.Post("/login", s.loginAuthHandler)
	})

	// admin only
	r.Route("/users", func(r chi.Router) {
		r.Get("/{id}", s.getUserHandler)
		r.Get("/", s.listUsersHandler)
		r.Post("/", s.createUserHandler)
		r.Delete("/{id}", s.deleteUserHandler)
	})
}

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

	err = ToJSON(w, domain.WrapUsers{User: *user}, http.StatusOK)
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
	}
}

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

	filter := domain.UsersFilter{
		Sort:   r.URL.Query().Get("sort"),
		Limit:  limit,
		Offset: offset,
	}

	users, err := s.UsersStore.List(r.Context(), filter)
	if err != nil {
		switch err {
		case sql.ErrNoRows, postgres.ErrInvalidColumn:
			Error(w, r, ErrNotFound, http.StatusNotFound)
		default:
			Error(w, r, err, http.StatusInternalServerError)
		}

		return
	}

	err = ToJSON(w, domain.WrapUsersList{Users: *users}, http.StatusOK)
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
	}
}

func (s *Server) createUserHandler(w http.ResponseWriter, r *http.Request) {
	input := domain.UsersCreate{}
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.MinCost)
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
		return
	}

	user := input.CreateModel(hashedPassword)

	err = s.UsersStore.Create(r.Context(), &user)
	if err != nil {
		switch err {
		case postgres.ErrForeinKeyViolation, postgres.ErrDuplicatedEntries:
			Error(w, r, err, http.StatusBadRequest)
		default:
			Error(w, r, err, http.StatusInternalServerError)
		}

		return
	}

	w.Header().Set("Location", fmt.Sprintf("/products/%d", user.ID))
	err = ToJSON(w, domain.WrapUsers{User: user}, http.StatusCreated)
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
	}
}

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

func (s *Server) loginAuthHandler(w http.ResponseWriter, r *http.Request) {
	input := domain.UsersLogin{}
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

	fmt.Println(input.Email)
	user, err := s.UsersStore.GetByEmail(r.Context(), input.Email)
	if err != nil {
		switch err {
		case sql.ErrNoRows, postgres.ErrInvalidColumn:
			Error(w, r, ErrNotFound, http.StatusNotFound)
		default:
			Error(w, r, err, http.StatusInternalServerError)
		}

		return
	}

	fmt.Printf("%+v\n", user)
	err = bcrypt.CompareHashAndPassword(user.Password, []byte(input.Password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			Error(w, r, err, http.StatusUnauthorized)
		} else {
			Error(w, r, err, http.StatusInternalServerError)
		}

		return
	}
}
