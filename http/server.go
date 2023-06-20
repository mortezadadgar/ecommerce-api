//	@title			e-commerce API
//	@license.name	MIT
//	@license.url	https://opensource.org/license/mit/
//  @BasePath	    /
//  @schemes        http

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
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mortezadadgar/ecommerce-api"
	_ "github.com/swaggo/http-swagger/example/go-chi/docs"
	httpSwagger "github.com/swaggo/http-swagger/v2"
)

type Server struct {
	UsersStore      ecommerce.UsersService
	ProductsStore   ecommerce.ProductsService
	CategoriesStore ecommerce.CategoriesService
	db              *sql.DB
	*http.Server
}

var ErrNotFound = errors.New("the requested resource could not be found")
var ErrMaxBytes = errors.New("exceeded maximum of 1M request body size")
var ErrInvalidQuery = errors.New("invalid url query")

// New returns a new instance of Server.
func New(db *sql.DB) *Server {
	s := Server{
		Server: &http.Server{
			Addr:         os.Getenv("ADDRESS"),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
		db: db,
	}

	r := chi.NewMux()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)
	r.Use(middlewareRequestSize(1_048_576))
	r.Use(middleware.Timeout(5 * time.Second))
	r.Use(middleware.SetHeader("Content-type", "application/json"))

	s.registerUsersRoutes(r)
	s.registerProductsRoutes(r)
	s.registerCategoriesRoutes(r)

	r.Get("/healthcheck", s.healthHandler)
	r.NotFound(s.notFoundHandler)
	r.MethodNotAllowed(s.methodNotAllowdHandler)

	fs := http.FileServer(http.Dir("./swagger"))
	r.Handle("/swagger/swagger.json", http.StripPrefix("/swagger", fs))

	r.Get("/docs/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/swagger.json"),
		httpSwagger.UIConfig(map[string]string{
			"defaultModelsExpandDepth": "-1",
		}),
	))

	r.Handle("/docs", http.RedirectHandler("/docs/index.html", http.StatusMovedPermanently))

	s.Handler = r

	return &s
}

// Start starts the server.
func (s *Server) Start() {
	l, err := net.Listen("tcp", os.Getenv("ADDRESS"))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Started listening on %s", os.Getenv("ADDRESS"))

	s.Serve(l)
}

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
	if err := s.db.PingContext(r.Context()); err != nil {
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

	err := ToJSON(w, h)
	if err != nil {
		Error(w, r, err, http.StatusInternalServerError)
	}
}

func (s *Server) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	Error(w, r, fmt.Errorf("the requested url not found"), http.StatusNotFound)
}

func (s *Server) methodNotAllowdHandler(w http.ResponseWriter, r *http.Request) {
	Error(w, r, fmt.Errorf("the requested method not allowed"), http.StatusMethodNotAllowed)
}

// FromJSON decodes the giving struct.
//
// caller must pass v as pointer.
func FromJSON(r *http.Request, v any) error {
	err := json.NewDecoder(r.Body).Decode(&v)
	if err != nil {
		var MaxBytesError *http.MaxBytesError
		switch {
		case errors.As(err, &MaxBytesError):
			return ErrMaxBytes
		default:
			return err
		}
	}
	defer r.Body.Close()

	return nil
}

// ToJSON encodes to giving strcut.
func ToJSON(w http.ResponseWriter, v any) error {
	return json.NewEncoder(w).Encode(v)
}

func ParseIntQuery(r *http.Request, v string) (int, error) {
	if !r.URL.Query().Has(v) {
		return 0, nil
	}

	return strconv.Atoi(r.URL.Query().Get(v))
}

// use chi's middleware when released.
func middlewareRequestSize(bytes int64) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, bytes)
			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
