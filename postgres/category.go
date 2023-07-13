package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mortezadadgar/ecommerce-api/domain"
)

// CategoryStore represents categories database.
type CategoryStore struct {
	db *pgxpool.Pool
}

// NewCategoryStore returns a new instance of CategoriesStore.
func NewCategoryStore(db *pgxpool.Pool) CategoryStore {
	return CategoryStore{db: db}
}

// Create creates a new category in database.
func (c CategoryStore) Create(ctx context.Context, category *domain.Category) error {
	tx, err := c.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", ErrBeginTransaction, err)
	}
	defer tx.Rollback(ctx)

	query := `
	 INSERT INTO categories(name, description)
	 VALUES(@name, @description)
	 RETURNING id, version
	`

	args := pgx.NamedArgs{
		"name":        &category.Name,
		"description": &category.Description,
	}

	err = tx.QueryRow(ctx, query, args).Scan(&category.ID, &category.Version)
	if err != nil {
		pgErr := pgError(err)
		if pgErr.Code == pgerrcode.UniqueViolation {
			if pgErr.ConstraintName == "categories_name_key" {
				return domain.ErrDuplicatedCategory
			}
		}
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", ErrCommitTransaction, err)
	}

	return nil
}

// GetByID get category by id from database.
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
		return nil, fmt.Errorf("%v: %v", ErrBeginTransaction, err)
	}
	defer tx.Rollback(ctx)

	query := `
	SELECT * FROM categories
	WHERE 1=1
	` + FormatSort(filter.Sort) + `
	` + FormatAnd("name", filter.Name) + `
	` + FormatAndInt("id", filter.ID) + `
	` + FormatLimitOffset(filter.Limit, filter.Offset) + `
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
		return nil, domain.ErrNoCategoryFound
	}

	return categories, nil
}

// Update updates a category by id in
func (c CategoryStore) Update(ctx context.Context, category *domain.Category) error {
	tx, err := c.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", ErrBeginTransaction, err)
	}
	defer tx.Rollback(ctx)

	query := `
	UPDATE categories
	SET name = @name, description = @description, updated_at = NOW(), version = version + 1
	WHERE id = @id AND version = @version
	RETURNING version
	`

	args := pgx.NamedArgs{
		"name":        &category.Name,
		"description": &category.Description,
		"id":          &category.ID,
		"version":     &category.Version,
	}

	err = tx.QueryRow(ctx, query, args).Scan(&category.Version)
	if err != nil {
		pgErr := pgError(err)
		if pgErr.Code == pgerrcode.UniqueViolation {
			if pgErr.ConstraintName == "categories_name_key" {
				return domain.ErrDuplicatedCategory
			}
		}

		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrCategoryConflict
		}

		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", ErrCommitTransaction, err)
	}

	return nil
}

// Delete deletes a category by id from database.
func (c CategoryStore) Delete(ctx context.Context, ID int) error {
	tx, err := c.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", ErrBeginTransaction, err)
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
		return err
	}

	if rows := result.RowsAffected(); rows != 1 {
		return domain.ErrNoCategoryFound
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", ErrCommitTransaction, err)
	}

	return nil
}
