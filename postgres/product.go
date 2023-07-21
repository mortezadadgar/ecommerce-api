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

// productStore represents products database.
type productStore struct {
	db *pgxpool.Pool
}

// NewProductStore returns a new instance of ProductStore.
func NewProductStore(db *pgxpool.Pool) productStore {
	return productStore{db: db}
}

// Create creates a new product in database.
func (p productStore) Create(ctx context.Context, product *domain.Product) error {
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

	err := p.db.QueryRow(ctx, query, args).Scan(&product.ID, &product.Version)
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

	return nil
}

// GetByID get product by id from database.
func (p productStore) GetByID(ctx context.Context, ID int) (domain.Product, error) {
	product, err := p.List(ctx, domain.ProductFilter{ID: ID})
	if err != nil {
		return domain.Product{}, err
	}

	return product[0], nil
}

// List lists products with optional filter.
func (p productStore) List(ctx context.Context, filter domain.ProductFilter) ([]domain.Product, error) {
	query := `
	SELECT * FROM products
	WHERE 1=1
	` + FormatSort(filter.Sort) + `
	` + FormatAndInt("category_id", filter.CategoryID) + `
	` + FormatAndInt("id", filter.ID) + `
	` + FormatLimitOffset(filter.Limit, filter.Offset) + `
	`

	rows, err := p.db.Query(ctx, query)
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

	return products, nil
}

// Update updates a product by id in database.
func (p productStore) Update(ctx context.Context, ID int, input domain.ProductUpdate) (domain.Product, error) {
	query := `
	UPDATE products
	SET name        = COALESCE(@name,name),
		description = COALESCE(@description, description),
		category_id = COALESCE(@category, category_id),
		price       = COALESCE(@price, price),
		quantity    = COALESCE(@quantity, quantity),
		updated_at  = NOW(),
		version     = version + 1
	WHERE id = @id AND version = @version
	RETURNING *
	`

	args := pgx.NamedArgs{
		"name":        &input.Name,
		"description": &input.Description,
		"category":    &input.CategoryID,
		"price":       &input.Price,
		"quantity":    &input.Quantity,
		"version":     &input.Version,
		"id":          &ID,
	}

	row, err := p.db.Query(ctx, query, args)
	if err != nil {
		return domain.Product{}, fmt.Errorf("failed to query update product: %v", err)
	}

	product, err := pgx.CollectOneRow(row, pgx.RowToStructByName[domain.Product])
	if err != nil {
		pgErr := pgError(err)
		switch pgErr.Code {
		case pgerrcode.ForeignKeyViolation:
			if pgErr.ConstraintName == "products_category_id_fkey" {
				return domain.Product{}, domain.ErrInvalidProductCategory
			}
		case pgerrcode.UniqueViolation:
			if pgErr.ConstraintName == "products_name_key" {
				return domain.Product{}, domain.ErrDuplicatedProduct
			}
		}

		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Product{}, domain.ErrProductConflict
		}

		return domain.Product{}, fmt.Errorf("failed to scan rows of product: %v", err)
	}

	return product, nil
}

// Delete deletes a product by id from database.
func (p productStore) Delete(ctx context.Context, ID int) error {
	query := `
	DELETE FROM products
	WHERE id = @id
	`

	args := pgx.NamedArgs{
		"id": ID,
	}

	result, err := p.db.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("failed to delete from products: %v", err)
	}

	if rows := result.RowsAffected(); rows != 1 {
		return domain.ErrNoProductsFound
	}

	return nil
}
