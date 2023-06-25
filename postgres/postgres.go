package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	// postgres driver.
	_ "github.com/jackc/pgx/v5/stdlib"
)

var ErrBeginTransaction = errors.New("failed to begin transaction")
var ErrInvalidColumn = errors.New("requested data field could not be found")
var ErrCommitTransaction = errors.New("failed to commit transaction")
var ErrDuplicatedEntries = errors.New("duplicated entries are not allowed")
var ErrUnableDeleteEntry = errors.New("unable to remove the entry")
var ErrForeinKeyViolation = errors.New("product category must matches a category entry")

// Connect initialize a new postgresql driver.
func Connect() (*sql.DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := sql.Open("pgx", os.Getenv("DSN"))
	if err != nil {
		return nil, fmt.Errorf("failed to open postgresql: %v", err)
	}

	err = db.PingContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to make connection to pg: %v", err)
	}

	return db, nil
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

func BeginTransaction(db *sql.DB) (*sql.Tx, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, ErrBeginTransaction
	}

	return tx, nil
}

func EndTransaction(tx *sql.Tx) error {
	defer tx.Rollback()

	err := tx.Commit()
	if err != nil {
		return ErrCommitTransaction
	}

	return nil
}
