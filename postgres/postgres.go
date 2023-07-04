// Package postgres handles all requests down to database.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	// postgres driver.
	"github.com/jackc/pgx/v5/pgxpool"
)

//revive:disable Ogh who is going to document these.
var ErrBeginTransaction = errors.New("failed to begin transaction")
var ErrInvalidColumn = errors.New("requested data field could not be found")
var ErrCommitTransaction = errors.New("failed to commit transaction")
var ErrDuplicatedEntries = errors.New("duplicated entries are not allowed")
var ErrUnableDeleteEntry = errors.New("unable to remove the entry")
var ErrForeinKeyViolation = errors.New("product category must matches a category entry")
var ErrNoRows = errors.New("no rows in results set")

type Postgres struct {
	DB *pgxpool.Pool
}

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
		return fmt.Sprintf(`LIMIT %d OFFSET %d`, limit, offset)
	case limit > 0:
		return fmt.Sprintf(`LIMIT %d`, limit)
	case offset > 0:
		return fmt.Sprintf(`OFFSET %d`, offset)
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

// FormatAndOp returns a SQL string for a giving category.
func FormatAndOp(c string, v string) string {
	if v != "" {
		return fmt.Sprintf("AND %s='%s'", c, v)
	}

	return ""
}
