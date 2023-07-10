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
	"database/sql"
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

	// http-swagger
	_ "github.com/swaggo/http-swagger/example/go-chi/docs"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

// Server represents an HTTP server.
type Server struct {
	UsersStore      domain.UserService
	ProductsStore   domain.ProductService
	CategoriesStore domain.CategoryService
	TokensStore     domain.TokenService
	CartsStore      domain.CartService
	SearchStore     domain.Searcher
	store           Store

	*http.Server
	*chi.Mux
}

// Store used for common interaction with database from server.
type Store interface {
	Close() error
	Ping(ctx context.Context) error
}

// New returns a new instance of Server.
func New(store Store) *Server {
	s := Server{
		Server: &http.Server{
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		},

		Mux: chi.NewMux(),

		store: store,
	}

	s.Use(middleware.Logger)
	s.Use(middleware.Recoverer)
	s.Use(middleware.StripSlashes)
	s.Use(middleware.Timeout(5 * time.Second))
	s.Use(s.authentication)

	s.registerUsersRoutes()
	s.registerProductsRoutes()
	s.registerCategoriesRoutes()
	s.registerCartsRoutes()
	s.registerSearchRoutes()

	s.Get("/healthcheck", s.healthHandler)
	s.NotFound(s.notFoundHandler)
	s.MethodNotAllowed(s.methodNotAllowdHandler)

	fs := http.FileServer(http.Dir("./swagger"))
	s.With(requireAuth).Handle("/swagger/swagger.json", http.StripPrefix("/swagger", fs))

	s.With(requireAuth).Get("/docs/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/swagger.json"),
		httpSwagger.UIConfig(map[string]string{
			"defaultModelsExpandDepth": "-1",
		}),
	))

	s.Handle("/docs", http.RedirectHandler("/docs/index.html", http.StatusMovedPermanently))

	s.Handler = s.Mux

	return &s
}

// Start starts the server.
func (s *Server) Start() {
	l, err := net.Listen("tcp", os.Getenv("ADDRESS"))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Started listening on %s", os.Getenv("ADDRESS"))

	go s.Serve(l)
}

// Close closes the database connection.
func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	return s.Shutdown(ctx)
}

type healthResponse struct {
	DbStatus string `json:"db_status"`
	MemUsage string `json:"mem_usage"`
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	var dbStatus string
	if err := s.store.Ping(r.Context()); err != nil {
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
		Error(w, r, err)
	}
}

func (s *Server) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	Error(w, r, domain.Errorf(domain.ENOTFOUND, "the requested url not found"))
}

func (s *Server) methodNotAllowdHandler(w http.ResponseWriter, r *http.Request) {
	Error(w, r, domain.Errorf(domain.EINVALID, "the requested method not allowed"))
}

const maxBytesBodyRead = 1_048_576

// FromJSON decodes the giving struct.
//
// caller must pass v as pointer.
func FromJSON(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBytesBodyRead)

	err := json.NewDecoder(r.Body).Decode(&v)
	if err != nil {
		var MaxBytesError *http.MaxBytesError
		switch {
		case errors.As(err, &MaxBytesError):
			return domain.Errorf(domain.ETOOLARGE,
				"exceeded maximum of 1M request body size")
		default:
			return err
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

func (s *Server) authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if len(auth) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		plainToken := strings.TrimPrefix(auth, "Bearer ")
		if len(plainToken) == 0 {
			next.ServeHTTP(w, r)
			return
		}

		user, err := s.TokensStore.GetUserID(r.Context(), plainToken)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				Error(w, r, err)
			}
			next.ServeHTTP(w, r)
			return
		}

		r = r.WithContext(newUserContext(r.Context(), user))

		next.ServeHTTP(w, r)
	})
}

func requireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if user := userIDFromContext(r.Context()); user == 0 {
			Error(w, r, domain.Errorf(domain.EUNAUTHORIZED, "unauthorized access"))
			return
		}

		next.ServeHTTP(w, r)
	})

}
