package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mortezadadgar/ecommerce-api/domain"
	"github.com/mortezadadgar/ecommerce-api/store"
)

// SearchStore represents search in database.
type SearchStore struct {
	db *pgxpool.Pool
}

// NewSearchStore returns a new instance of SearchStore.
func NewSearchStore(db *pgxpool.Pool) SearchStore {
	return SearchStore{db: db}
}

// Search full searches database.
func (s SearchStore) Search(ctx context.Context, query string) (results []domain.Search, err error) {
	categories, err := fullSearch[domain.Category](ctx, s.db, "categories", query)
	if err != nil {
		log.Fatal(err)
	}

	for i := range categories {
		results = append(results, domain.Search{Categories: &categories[i]})
	}

	products, err := fullSearch[domain.Product](ctx, s.db, "products", query)
	if err != nil {
		log.Fatal(err)
	}

	for i := range products {
		results = append(results, domain.Search{Prodcuts: &products[i]})
	}

	if len(results) == 0 {
		return nil, sql.ErrNoRows
	}

	return results, nil
}

func fullSearch[T any](ctx context.Context, db *pgxpool.Pool, table string, query string) (results []T, err error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v: %v", store.ErrBeginTransaction, err)
	}
	defer tx.Rollback(ctx)

	sqlQuery := `
	SELECT * FROM ` + table + `
	WHERE to_tsvector('simple', name) @@ to_tsquery('simple', @query);
	`

	args := pgx.NamedArgs{
		"query": query,
	}

	rows, err := tx.Query(ctx, sqlQuery, args)
	if err != nil {
		return nil, err
	}

	results, err = pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return results, nil
}
