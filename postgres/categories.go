package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mortezadadgar/ecommerce-api"
)

type CategoriesStore struct {
	db *sql.DB
}

func NewCategoriesStore(db *sql.DB) *CategoriesStore {
	return &CategoriesStore{db: db}
}

// Create creates a new category in store.
func (c *CategoriesStore) Create(ctx context.Context, category *ecommerce.Categories) error {
	tx, err := BeginTransaction(c.db)
	if err != nil {
		return err
	}

	query := `
	 INSERT INTO categories(name, description)
	 VALUES($1, $2)
	 RETURNING id, created_at, updated_at
	`

	args := []any{
		&category.Name,
		&category.Description,
	}

	err = tx.QueryRowContext(ctx, query, args...).Scan(
		&category.ID,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return ErrDuplicatedEntries
		}

		return fmt.Errorf("failed to insert into categories: %v", err)
	}

	err = EndTransaction(tx)
	if err != nil {
		return err
	}

	return nil
}

// GetByID gets category by id from store.
func (c *CategoriesStore) GetByID(ctx context.Context, id int) (*ecommerce.Categories, error) {
	tx, err := BeginTransaction(c.db)
	if err != nil {
		return nil, err
	}

	query := `
	SELECT * FROM categories
	WHERE id = $1
	`

	category := ecommerce.Categories{}
	args := []any{
		&category.ID,
		&category.Name,
		&category.Description,
		&category.CreatedAt,
		&category.UpdatedAt,
	}

	err = tx.QueryRowContext(ctx, query, id).Scan(args...)
	switch {
	case err == sql.ErrNoRows:
		return nil, sql.ErrNoRows
	case err != nil:
		return nil, fmt.Errorf("failed to query category: %v", err)
	}

	err = EndTransaction(tx)
	if err != nil {
		return nil, err
	}

	return &category, nil

}

// List lists categories with optional filter.
func (c *CategoriesStore) List(ctx context.Context, filter ecommerce.CategoriesFilter) (*[]ecommerce.Categories, error) {
	tx, err := BeginTransaction(c.db)
	if err != nil {
		return nil, ErrBeginTransaction
	}

	query := `
	SELECT * FROM products
	WHERE 1=1
	` + FormatSort(filter.Sort) + `
	` + FormatAndOp("name", filter.Name) + `
	` + FormatLimitOffset(filter.Limit, filter.Offset) + `
	`

	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %v", err)
	}
	defer rows.Close()

	categories := []ecommerce.Categories{}

	for rows.Next() {
		category := ecommerce.Categories{}
		args := []any{
			&category.ID,
			&category.Name,
			&category.Description,
			&category.CreatedAt,
			&category.UpdatedAt,
		}

		if err := rows.Scan(args...); err != nil {
			return nil, fmt.Errorf("failed to scan rows of categories: %v", err)
		}

		categories = append(categories, category)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	err = EndTransaction(tx)
	if err != nil {
		return nil, err
	}

	return &categories, nil
}

// Update updates a category by id in store.
func (c *CategoriesStore) Update(ctx context.Context, category *ecommerce.Categories) error {
	tx, err := BeginTransaction(c.db)
	if err != nil {
		return err
	}

	query := `
	UPDATE categories
	SET name = $1, description = $2, updated_at = NOW()
	WHERE id = $3
	RETURNING updated_at
	`

	args := []any{
		&category.Name,
		&category.Description,
		&category.ID,
	}

	err = tx.QueryRowContext(ctx, query, args...).Scan(&category.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return ErrDuplicatedEntries
			case pgerrcode.ForeignKeyViolation:
				return ErrForeinKeyViolation
			}
		}
		return fmt.Errorf("failed to update category: %v", err)
	}

	err = EndTransaction(tx)
	if err != nil {
		return err
	}

	return nil

}

// Delete deletes a category by id from store.
func (c *CategoriesStore) Delete(ctx context.Context, id int) error {
	tx, err := BeginTransaction(c.db)
	if err != nil {
		return err
	}

	query := `
	DELETE FROM categories
	WHERE id = $1
	`

	result, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete from categories: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows != 1 {
		return fmt.Errorf("expected to affect 1 rows, affected: %d", rows)
	}

	err = EndTransaction(tx)
	if err != nil {
		return err
	}

	return nil
}
