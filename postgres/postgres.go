// Package postgres handles all requests down to database.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	// postgres driver.
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mortezadadgar/ecommerce-api/domain"
)

// Postgres represents postgres connection pool.
type Postgres struct {
	DB *pgxpool.Pool
}

// ErrBeginTransaction returned when transaction fails from beginning.
var ErrBeginTransaction = errors.New("failed to begin transaction")

// ErrCommitTransaction returned when commit transaction fails.
var ErrCommitTransaction = errors.New("failed to commit transaction")

// New returns a new instance of postgres and connect as well.
func New() (Postgres, error) {
	db, err := connect()
	if err != nil {
		return Postgres{}, err
	}

	return Postgres{
		DB: db,
	}, nil
}

func connect() (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pgxpool.New(ctx, os.Getenv("DSN"))
	if err != nil {
		return nil, fmt.Errorf("failed to open postgresql: %v", err)
	}

	err = db.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to make connection to pg: %v", err)
	}

	return db, nil
}

// Close closes postgres connection.
func (p Postgres) Close() error {
	p.DB.Close()
	return nil
}

// Ping test postgres connection.
func (p Postgres) Ping(ctx context.Context) error {
	return p.DB.Ping(ctx)
}

// FormatLimitOffset returns a SQL string for a given limit & offset.
func FormatLimitOffset(limit int, offset int) string {
	switch {
	case limit > 0 && offset > 0:
		return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
	case limit > 0:
		return fmt.Sprintf("LIMIT %d", limit)
	case offset > 0:
		return fmt.Sprintf("OFFSET %d", offset)
	}

	return ""
}

// FormatSort returns a SQL string for a giving column.
func FormatSort(v string) string {
	if v != "" {
		return fmt.Sprintf("ORDER BY %s", v)
	}

	return ""
}

// FormatAndOp returns a SQL string for a giving column and string value.
func FormatAndOp(s string, v string) string {
	if v != "" {
		return fmt.Sprintf("AND %s='%s'", s, v)
	}

	return ""
}

// FormatAndIntOp returns a SQL string for a giving column and integer value.
func FormatAndIntOp(s string, v int) string {
	if v != 0 {
		return fmt.Sprintf("AND %s='%d'", s, v)
	}

	return ""
}

// FormatError formats common errors returned from postgres driver.
func FormatError(err error) error {
	message := err.Error()
	switch {
	case strings.Contains(message, "carts_user_id_fkey"):
		return domain.Errorf(domain.EINVALID, "carts's user must matches a valid user")
	case strings.Contains(message, "carts_product_id_fkey"):
		return domain.Errorf(domain.EINVALID, "carts's product must matches a valid product")
	case strings.Contains(message, "products_name_key"):
		return domain.Errorf(domain.ECONFLICT, "duplicated products are not allowed")
	case strings.Contains(message, "products_category_fkey"):
		return domain.Errorf(domain.EINVALID, "product's category must matches a valid category")
	case strings.Contains(message, "users_email_key"):
		return domain.Errorf(domain.ECONFLICT, "duplicated email are not allowed")
	case strings.Contains(message, "categories_name_key"):
		return domain.Errorf(domain.ECONFLICT, "duplicated categories are not allowed")
	default:
		return err
	}
}
