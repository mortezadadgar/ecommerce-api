package http

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mortezadadgar/ecommerce-api"
	"github.com/mortezadadgar/ecommerce-api/postgres"
	"golang.org/x/crypto/bcrypt"
)

func (s *Server) registerUsersRoutes(r *chi.Mux) {
	r.Route("/users", func(r chi.Router) {
		r.Get("/{id}", s.usersGetHandler)
		r.Post("/", s.usersPostHandler)
	})
}

func (s *Server) usersGetHandler(w http.ResponseWriter, r *http.Request) {
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

	err = ToJSON(w, ecommerce.WrapUsers{User: *user})
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
	}
}

func (s *Server) usersPostHandler(w http.ResponseWriter, r *http.Request) {
	input := ecommerce.UsersInput{}
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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), 8)
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
		return
	}

	user := ecommerce.Users{
		Email:    input.Email,
		Password: hashedPassword,
	}

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

	err = ToJSON(w, ecommerce.WrapUsers{User: user})
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
	}
}
