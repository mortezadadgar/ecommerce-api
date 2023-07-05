package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mortezadadgar/ecommerce-api/domain"
	"github.com/mortezadadgar/ecommerce-api/store"
)

// CategoriesStore represents categories database.
type CategoriesStore struct {
	db *pgxpool.Pool
}

// NewCategoriesStore returns a new instance of CategoriesStore.
func NewCategoriesStore(db *pgxpool.Pool) CategoriesStore {
	return CategoriesStore{db: db}
}

// Create creates a new category in store.
func (c CategoriesStore) Create(ctx context.Context, category *domain.Categories) error {
	tx, err := c.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", store.ErrBeginTransaction, err)
	}
	defer tx.Rollback(ctx)

	query := `
	 INSERT INTO categories(name, description)
	 VALUES(@name, @description)
	 RETURNING id, created_at, updated_at
	`

	args := pgx.NamedArgs{
		"name":        &category.Name,
		"description": &category.Description,
	}

	err = tx.QueryRow(ctx, query, args).Scan(
		&category.ID,
		&category.CreatedAt,
		&category.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return store.ErrDuplicatedEntries
			}
		}

		return fmt.Errorf("failed to insert into categories: %v", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", store.ErrCommitTransaction, err)
	}

	return nil
}

// GetByID get category by id from store.
func (c CategoriesStore) GetByID(ctx context.Context, id int) (domain.Categories, error) {
	category, err := c.List(ctx, domain.CategoriesFilter{ID: id})
	if err != nil {
		return domain.Categories{}, err
	}

	return category[0], nil
}

// List lists categories with optional filter.
func (c CategoriesStore) List(ctx context.Context, filter domain.CategoriesFilter) ([]domain.Categories, error) {
	tx, err := c.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v: %v", store.ErrBeginTransaction, err)
	}
	defer tx.Rollback(ctx)

	query := `
	SELECT * FROM categories
	WHERE 1=1
	` + store.FormatSort(filter.Sort) + `
	` + store.FormatAndOp("name", filter.Name) + `
	` + store.FormatAndIntOp("id", filter.ID) + `
	` + store.FormatLimitOffset(filter.Limit, filter.Offset) + `
	`

	rows, err := tx.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %v", err)
	}

	categories, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Categories])
	if err != nil {
		return nil, fmt.Errorf("failed to scan rows of categories: %v", err)
	}

	if len(categories) == 0 {
		return nil, sql.ErrNoRows
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v: %v", store.ErrCommitTransaction, err)
	}

	return categories, nil
}

// Update updates a category by id in store.
func (c CategoriesStore) Update(ctx context.Context, category *domain.Categories) error {
	tx, err := c.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", store.ErrBeginTransaction, err)
	}
	defer tx.Rollback(ctx)

	query := `
	UPDATE categories
	SET name = @name, description = @description, updated_at = NOW()
	WHERE id = @id
	RETURNING updated_at
	`

	args := pgx.NamedArgs{
		"name":        &category.Name,
		"description": &category.Description,
		"id":          &category.ID,
	}

	err = tx.QueryRow(ctx, query, args).Scan(&category.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.UniqueViolation:
				return store.ErrDuplicatedEntries
			case pgerrcode.ForeignKeyViolation:
				return store.ErrForeinKeyViolation
			}
		}

		return fmt.Errorf("failed to update category: %v", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", store.ErrCommitTransaction, err)
	}

	return nil
}

// Delete deletes a category by id from store.
func (c CategoriesStore) Delete(ctx context.Context, id int) error {
	tx, err := c.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", store.ErrBeginTransaction, err)
	}
	defer tx.Rollback(ctx)

	query := `
	DELETE FROM categories
	WHERE id = @id
	`

	args := pgx.NamedArgs{
		"id": id,
	}

	result, err := tx.Exec(ctx, query, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.ForeignKeyViolation {
				return store.ErrForeinKeyViolation
			}
		}

		return fmt.Errorf("failed to delete from category: %v", err)
	}

	rows := result.RowsAffected()
	if rows != 1 {
		return sql.ErrNoRows
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", store.ErrCommitTransaction, err)
	}

	return nil
}
