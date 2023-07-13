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

// ProductStore represents products database.
type ProductStore struct {
	db *pgxpool.Pool
}

// NewProductStore returns a new instance of ProductStore.
func NewProductStore(db *pgxpool.Pool) ProductStore {
	return ProductStore{db: db}
}

// Create creates a new product in database.
func (p ProductStore) Create(ctx context.Context, product *domain.Product) error {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", ErrBeginTransaction, err)
	}
	defer tx.Rollback(ctx)

	query := `
	 INSERT INTO products(name, description, category_id, price, quantity)
	 VALUES(@name, @description, @category, @price, @quantity)
	 RETURNING id, version
	`

	args := pgx.NamedArgs{
		"name":        &product.Name,
		"description": &product.Description,
		"category":    &product.CategoryID,
		"price":       &product.Price,
		"quantity":    &product.Quantity,
	}

	err = tx.QueryRow(ctx, query, args).Scan(&product.ID, &product.Version)
	if err != nil {
		pgErr := pgError(err)
		switch pgErr.Code {
		case pgerrcode.ForeignKeyViolation:
			if pgErr.ConstraintName == "products_category_id_fkey" {
				return domain.ErrInvalidProductCategory
			}
		case pgerrcode.UniqueViolation:
			if pgErr.ConstraintName == "products_name_key" {
				return domain.ErrDuplicatedProduct
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

// GetByID get product by id from database.
func (p ProductStore) GetByID(ctx context.Context, ID int) (domain.Product, error) {
	product, err := p.List(ctx, domain.ProductFilter{ID: ID})
	if err != nil {
		return domain.Product{}, err
	}

	return product[0], nil
}

// List lists products with optional filter.
func (p ProductStore) List(ctx context.Context, filter domain.ProductFilter) ([]domain.Product, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v: %v", ErrBeginTransaction, err)
	}
	defer tx.Rollback(ctx)

	query := `
	SELECT * FROM products
	WHERE 1=1
	` + FormatSort(filter.Sort) + `
	` + FormatAndInt("category_id", filter.CategoryID) + `
	` + FormatAndInt("id", filter.ID) + `
	` + FormatLimitOffset(filter.Limit, filter.Offset) + `
	`

	rows, err := tx.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query list products: %v", err)
	}

	products, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Product])
	if err != nil {
		return nil, fmt.Errorf("failed to scan rows of products: %v", err)
	}

	if len(products) == 0 {
		return nil, domain.ErrNoProductsFound
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v: %v", ErrCommitTransaction, err)
	}

	return products, nil
}

// Update updates a product by id in
func (p ProductStore) Update(ctx context.Context, product *domain.Product) error {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", ErrBeginTransaction, err)
	}
	defer tx.Rollback(ctx)

	query := `
	UPDATE products
	SET name = @name, description = @description, category_id = @category, price = @price, quantity = @quantity, updated_at = NOW(), version = version + 1
	WHERE id = @id AND version = @version
	RETURNING version
	`

	args := pgx.NamedArgs{
		"name":        &product.Name,
		"description": &product.Description,
		"category":    &product.CategoryID,
		"price":       &product.Price,
		"quantity":    &product.Quantity,
		"id":          &product.ID,
		"version":     &product.Version,
	}

	err = tx.QueryRow(ctx, query, args).Scan(&product.Version)
	if err != nil {
		pgErr := pgError(err)
		switch pgErr.Code {
		case pgerrcode.ForeignKeyViolation:
			if pgErr.ConstraintName == "products_category_id_fkey" {
				return domain.ErrInvalidProductCategory
			}
		case pgerrcode.UniqueViolation:
			if pgErr.ConstraintName == "products_name_key" {
				return domain.ErrDuplicatedProduct
			}
		}

		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrProductConflict
		}

		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", ErrCommitTransaction, err)
	}

	return nil
}

// Delete deletes a product by id from database.
func (p ProductStore) Delete(ctx context.Context, ID int) error {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", ErrBeginTransaction, err)
	}
	defer tx.Rollback(ctx)

	query := `
	DELETE FROM products
	WHERE id = @id
	`

	args := pgx.NamedArgs{
		"id": ID,
	}

	result, err := tx.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("failed to delete from products: %v", err)
	}

	if rows := result.RowsAffected(); rows != 1 {
		return domain.ErrNoProductsFound
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", ErrCommitTransaction, err)
	}

	return nil
}
