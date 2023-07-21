//	@title			e-commerce API
//	@license.name	MIT
//	@license.url	https://opensource.org/license/mit/
//  @BasePath	    /
//  @schemes        http
//  @securityDefinitions.apikey Bearer
//  @in header
//  @name Authorization
//  @description Type "Bearer" followed by a space and token.

// Package http handles HTTP requests.
package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mortezadadgar/ecommerce-api/domain"
)

// server represents an HTTP server.
type server struct {
	UsersStore      domain.UserService
	ProductsStore   domain.ProductService
	CategoriesStore domain.CategoryService
	TokensStore     domain.TokenService
	CartsStore      domain.CartService
	SearchStore     domain.Searcher
	Store           store

	*http.Server
}

// New returns a new instance of Server.
func New() *server {
	r := chi.NewMux()
	s := server{
		Server: &http.Server{
			Handler: r,
		},
	}

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Use(middleware.Timeout(5 * time.Second))
	r.Use(s.authentication)

	s.registerUsersRoutes(r)
	s.registerProductsRoutes(r)
	s.registerCategoriesRoutes(r)
	s.registerCartsRoutes(r)
	s.registerSearchRoutes(r)

	r.Get("/healthcheck", s.healthHandler)
	r.NotFound(s.notFoundHandler)
	r.MethodNotAllowed(s.methodNotAllowdHandler)

	return &s
}

// Start starts the server.
func (s *server) Start() error {
	l, err := net.Listen("tcp", os.Getenv("ADDRESS"))
	if err != nil {
		return err
	}

	log.Printf("Started listening on %s", os.Getenv("ADDRESS"))

	go s.Serve(l)
	return nil
}

// Close closes the database connection.
func (s *server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	return s.Shutdown(ctx)
}

type healthResponse struct {
	DbStatus string `json:"db_status"`
	MemUsage string `json:"mem_usage"`
}

type store interface {
	Ping(ctx context.Context) error
}

func (s *server) healthHandler(w http.ResponseWriter, r *http.Request) {
	var dbStatus string
	if err := s.Store.Ping(r.Context()); err != nil {
		dbStatus = "Connection Error"
	} else {
		dbStatus = "Connection OK"
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	memUsage := fmt.Sprintf("%dMiB", m.Alloc/1024/1024)

	h := healthResponse{
		DbStatus: dbStatus,
		MemUsage: memUsage,
	}

	err := ToJSON(w, h, http.StatusOK)
	if err != nil {
		Errorf(w, r, http.StatusInternalServerError, err.Error())
	}
}

func (s *server) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	Errorf(w, r, http.StatusNotFound, "the requested url not found")
}

func (s *server) methodNotAllowdHandler(w http.ResponseWriter, r *http.Request) {
	Errorf(w, r, http.StatusBadRequest, "the requested method not allowed")
}

const maxBytesBodyRead = 1_048_576

// FromJSON decodes to giving struct.
//
// caller must pass v as pointer.
func FromJSON(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytesBodyRead)

	err := json.NewDecoder(r.Body).Decode(v)
	if err != nil {
		var MaxBytesError *http.MaxBytesError
		switch {
		case errors.As(err, &MaxBytesError):
			return errors.New("exceeded maximum of 1M request body size")
		default:
			logError(r, err.Error())
			return errors.New("failed to parse json body")
		}
	}

	return nil
}

// ToJSON encodes to giving strcut.
func ToJSON(w http.ResponseWriter, v any, code int) error {
	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(v)
}

// ParseIntQuery parses integer url parameter, returns nothing when
// parameter is not provided.
func ParseIntQuery(r *http.Request, v string) (int, error) {
	if !r.URL.Query().Has(v) {
		return 0, nil
	}

	return strconv.Atoi(r.URL.Query().Get(v))
}

// ErrMalformedAuthHeader returned when authorization header is not
// formatted properly.
var ErrMalformedAuthHeader = errors.New("malformed authorization header")

// authentication a middleware that utilizes authorization header
// for token based authentications.
func (s *server) authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if len(auth) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		if !strings.Contains(auth, "Bearer") {
			Errorf(w, r, http.StatusBadRequest, ErrMalformedAuthHeader.Error())
			return
		}

		plainToken := strings.TrimPrefix(auth, "Bearer ")

		userID, err := s.TokensStore.GetUserID(r.Context(), plainToken)
		if err != nil {
			if errors.Is(err, domain.ErrInvalidToken) {
				Errorf(w, r, http.StatusBadRequest, domain.ErrInvalidToken.Error())
			} else {
				Errorf(w, r, http.StatusInternalServerError, err.Error())
			}

			return
		}

		r = r.WithContext(newUserContext(r.Context(), userID))

		next.ServeHTTP(w, r)
	})
}

func requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// if user := userIDFromContext(r.Context()); user == 0 {
		// Error(w, r, errors.New("unauthorized access"), http.StatusUnauthorized)
		// 	return
		// }

		next.ServeHTTP(w, r)
	})

}
