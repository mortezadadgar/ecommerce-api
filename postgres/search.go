package postgres

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mortezadadgar/ecommerce-api/domain"
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
	categories, err := fullSearchName[domain.Category](ctx, s.db, "categories", query)
	if err != nil {
		log.Fatal(err)
	}

	for i := range categories {
		results = append(results, domain.Search{Categories: &categories[i]})
	}

	products, err := fullSearchName[domain.Product](ctx, s.db, "products", query)
	if err != nil {
		log.Fatal(err)
	}

	for i := range products {
		results = append(results, domain.Search{Prodcuts: &products[i]})
	}

	if len(results) == 0 {
		return nil, domain.ErrNoSearchResult
	}

	return results, nil
}

func fullSearchName[T any](ctx context.Context, db *pgxpool.Pool, table string, query string) (results []T, err error) {
	sqlQuery := `
	SELECT * FROM ` + table + `
	WHERE to_tsvector('simple', name) @@ to_tsquery('simple', @query);
	`

	args := pgx.NamedArgs{
		"query": query,
	}

	rows, err := db.Query(ctx, sqlQuery, args)
	if err != nil {
		return nil, err
	}

	results, err = pgx.CollectRows(rows, pgx.RowToStructByName[T])
	if err != nil {
		return nil, err
	}

	return results, nil
}
