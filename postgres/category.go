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

// categoryStore represents categories database.
type categoryStore struct {
	db *pgxpool.Pool
}

// NewCategoryStore returns a new instance of CategoriesStore.
func NewCategoryStore(db *pgxpool.Pool) categoryStore {
	return categoryStore{db: db}
}

// Create creates a new category in database.
func (c categoryStore) Create(ctx context.Context, category *domain.Category) error {
	query := `
	 INSERT INTO categories(name, description)
	 VALUES(@name, @description)
	 RETURNING id, version
	`

	args := pgx.NamedArgs{
		"name":        &category.Name,
		"description": &category.Description,
	}

	err := c.db.QueryRow(ctx, query, args).Scan(&category.ID, &category.Version)
	if err != nil {
		pgErr := pgError(err)
		if pgErr.Code == pgerrcode.UniqueViolation {
			if pgErr.ConstraintName == "categories_name_key" {
				return domain.ErrDuplicatedCategory
			}
		}
		return err
	}

	return nil
}

// GetByID get category by id from database.
func (c categoryStore) GetByID(ctx context.Context, ID int) (domain.Category, error) {
	category, err := c.List(ctx, domain.CategoryFilter{ID: ID})
	if err != nil {
		return domain.Category{}, err
	}

	return category[0], nil
}

// List lists categories with optional filter.
func (c categoryStore) List(ctx context.Context, filter domain.CategoryFilter) ([]domain.Category, error) {
	query := `
	SELECT * FROM categories
	WHERE 1=1
	` + FormatSort(filter.Sort) + `
	` + FormatAnd("name", filter.Name) + `
	` + FormatAndInt("id", filter.ID) + `
	` + FormatLimitOffset(filter.Limit, filter.Offset) + `
	`

	rows, err := c.db.Query(ctx, query)
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
func (c categoryStore) Update(ctx context.Context, ID int, input domain.CategoryUpdate) (domain.Category, error) {
	query := `
	UPDATE categories
	SET name = COALESCE(@name, name),
		description = COALESCE(@description, description),
		updated_at = NOW(),
		version = version + 1
	WHERE id = @id AND version = @version
	RETURNING *
	`

	args := pgx.NamedArgs{
		"name":        &input.Name,
		"description": &input.Description,
		"version":     &input.Version,
		"id":          ID,
	}

	row, err := c.db.Query(ctx, query, args)
	if err != nil {
		return domain.Category{}, err
	}

	category, err := pgx.CollectOneRow(row, pgx.RowToStructByName[domain.Category])
	if err != nil {
		pgErr := pgError(err)
		if pgErr.Code == pgerrcode.UniqueViolation {
			if pgErr.ConstraintName == "categories_name_key" {
				return domain.Category{}, domain.ErrDuplicatedCategory
			}
		}

		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Category{}, domain.ErrCategoryConflict
		}

		return domain.Category{}, fmt.Errorf("failed to scan row of category: %v", err)
	}

	return category, nil
}

// Delete deletes a category by id from database.
func (c categoryStore) Delete(ctx context.Context, ID int) error {
	query := `
	DELETE FROM categories
	WHERE id = @id
	`

	args := pgx.NamedArgs{
		"id": ID,
	}

	result, err := c.db.Exec(ctx, query, args)
	if err != nil {
		return err
	}

	if rows := result.RowsAffected(); rows != 1 {
		return domain.ErrNoCategoryFound
	}

	return nil
}
