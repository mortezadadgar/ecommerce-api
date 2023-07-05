// Package postgres handles all requests down to database.
package postgres

import (
	"context"
	"fmt"
	"os"
	"time"

	// postgres driver.
	"github.com/jackc/pgx/v5/pgxpool"
)

// Postgres represents postgres connection pool.
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
