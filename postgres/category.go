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

// CategoryStore represents categories database.
type CategoryStore struct {
	db *pgxpool.Pool
}

// NewCategoryStore returns a new instance of CategoriesStore.
func NewCategoryStore(db *pgxpool.Pool) CategoryStore {
	return CategoryStore{db: db}
}

// Create creates a new category in store.
func (c CategoryStore) Create(ctx context.Context, category *domain.Category) error {
	tx, err := c.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", store.ErrBeginTransaction, err)
	}
	defer tx.Rollback(ctx)

	query := `
	 INSERT INTO categories(name, description)
	 VALUES(@name, @description)
	 RETURNING id
	`

	args := pgx.NamedArgs{
		"name":        &category.Name,
		"description": &category.Description,
	}

	err = tx.QueryRow(ctx, query, args).Scan(&category.ID)
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
func (c CategoryStore) GetByID(ctx context.Context, ID int) (domain.Category, error) {
	category, err := c.List(ctx, domain.CategoryFilter{ID: ID})
	if err != nil {
		return domain.Category{}, err
	}

	return category[0], nil
}

// List lists categories with optional filter.
func (c CategoryStore) List(ctx context.Context, filter domain.CategoryFilter) ([]domain.Category, error) {
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

	categories, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Category])
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
func (c CategoryStore) Update(ctx context.Context, category *domain.Category) error {
	tx, err := c.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", store.ErrBeginTransaction, err)
	}
	defer tx.Rollback(ctx)

	query := `
	UPDATE categories
	SET name = @name, description = @description, updated_at = NOW()
	WHERE id = @id
	`

	args := pgx.NamedArgs{
		"name":        &category.Name,
		"description": &category.Description,
		"id":          &category.ID,
	}

	_, err = tx.Exec(ctx, query, args)
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
func (c CategoryStore) Delete(ctx context.Context, ID int) error {
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
		"id": ID,
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
