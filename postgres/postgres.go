// Package postgres handles all requests down to database.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	// postgres driver.
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	// ErrBeginTransaction returned when transaction fails from beginning.
	ErrBeginTransaction = errors.New("failed to begin transaction")

	// ErrCommitTransaction returned when commit transaction fails.
	ErrCommitTransaction = errors.New("failed to commit transaction")
)

// Postgres represents Postgres connection pool.
type Postgres struct {
	DB *pgxpool.Pool
}

// New returns a new instace of postgres.
func New(dsn string) (Postgres, error) {
	db, err := Connect(dsn)
	if err != nil {
		return Postgres{}, err
	}
	return Postgres{DB: db}, nil
}

// Connect connet to postgres driver with giving dsn.
func Connect(dsn string) (*pgxpool.Pool, error) {
	var (
		once sync.Once
		err  error
		db   *pgxpool.Pool
	)

	once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		db, err = pgxpool.New(ctx, dsn)
		err = db.Ping(ctx)
	})

	if err != nil || db == nil {
		return nil, err
	}

	return db, nil
}

// Close closes postgres connection.
func (p *Postgres) Close() error {
	p.DB.Close()
	return nil
}

// Ping test postgres connection.
func (p *Postgres) Ping(ctx context.Context) error {
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
	default:
		return ""
	}
}

// FormatSort returns a SQL string for a giving column.
func FormatSort(column string) string {
	if column != "" {
		return fmt.Sprintf("ORDER BY %s", column)
	}

	return ""
}

// FormatAnd returns a SQL string for a giving column and string value.
func FormatAnd(column string, value string) string {
	if value != "" {
		return fmt.Sprintf("AND %s='%s'", column, value)
	}

	return ""
}

// FormatAndInt returns a SQL string for a giving column and integer value.
func FormatAndInt(column string, value int) string {
	if value != 0 {
		return fmt.Sprintf("AND %s='%d'", column, value)
	}

	return ""
}

// pgError returns postgres error type.
func pgError(err error) pgconn.PgError {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return *pgErr
	}

	return pgconn.PgError{}
}
