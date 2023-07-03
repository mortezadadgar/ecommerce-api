package http

import (
	"database/sql"
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
		// r.Get("/me", s.getMeHandler)
		// r.Get("/login", s.loginAuthHandler)
		r.Post("/register", s.registerAuthHandler)
	})

	// admin only
	r.Route("/users", func(r chi.Router) {
		r.Get("/{id}", s.getUserHandler)
		// r.Get("/", s.listUsersHandler)
		// r.Put("/{id}", s.createUserHandler)
		// r.Patch("/{id}", s.updateUserHandler)
		// r.Delete("/{id}", s.deleteUserHandler)
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

func (s *Server) registerAuthHandler(w http.ResponseWriter, r *http.Request) {
	input := domain.UsersCreate{}
	err := FromJSON(w, r, &input)
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
		return
	}

	err = input.Validate(r)
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

	err = ToJSON(w, domain.WrapUsers{User: user}, http.StatusCreated)
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
	}

}
