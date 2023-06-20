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

type ProductsStore struct {
	db *sql.DB
}

func NewProductsStore(db *sql.DB) *ProductsStore {
	return &ProductsStore{db: db}
}

func (p *ProductsStore) Create(ctx context.Context, product *ecommerce.Products) error {
	tx, err := BeginTransaction(p.db)
	if err != nil {
		return err
	}

	query := `
	 INSERT INTO products(name, description, category, price, quantity)
	 VALUES($1, $2, $3, $4, $5)
	 RETURNING id, created_at, updated_at
	`

	args := []any{
		&product.Name,
		&product.Description,
		&product.Category,
		&product.Price,
		&product.Quantity,
	}

	err = tx.QueryRowContext(ctx, query, args...).Scan(
		&product.ID,
		&product.CreatedAt,
		&product.UpdatedAt,
	)
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
		return fmt.Errorf("failed to insert into products: %v", err)
	}

	err = EndTransaction(ctx, tx)
	if err != nil {
		return err
	}

	return nil
}

func (p *ProductsStore) GetByID(ctx context.Context, id int) (*ecommerce.Products, error) {
	tx, err := BeginTransaction(p.db)
	if err != nil {
		return nil, err
	}

	query := `
	SELECT * FROM products
	WHERE id = $1
	`

	product := &ecommerce.Products{}
	args := []any{
		&product.ID,
		&product.Name,
		&product.Description,
		&product.Category,
		&product.Price,
		&product.Quantity,
		&product.CreatedAt,
		&product.UpdatedAt,
	}

	err = tx.QueryRowContext(ctx, query, id).Scan(args...)
	if err != nil {
		switch err {
		case sql.ErrNoRows:
			return nil, sql.ErrNoRows
		default:
			return nil, fmt.Errorf("failed to query product: %v", err)
		}
	}

	err = EndTransaction(ctx, tx)
	if err != nil {
		return nil, err
	}

	return product, nil
}

func (p *ProductsStore) List(ctx context.Context, filter ecommerce.ProductsFilter) (*[]ecommerce.Products, error) {
	tx, err := BeginTransaction(p.db)
	if err != nil {
		return nil, err
	}

	query := `
	SELECT * FROM products
	WHERE 1=1
	` + FormatSort(filter.Sort) + `
	` + FormatAndOp("category", filter.Category) + `
	` + FormatAndOp("name", filter.Name) + `
	` + FormatLimitOffset(filter.Limit, filter.Offset) + `
	`

	rows, err := tx.QueryContext(ctx, query)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UndefinedColumn {
			return nil, ErrInvalidColumn
		}
		return nil, fmt.Errorf("failed to query list products: %v", err)
	}
	defer rows.Close()

	products := []ecommerce.Products{}
	for rows.Next() {
		product := ecommerce.Products{}
		args := []any{
			&product.ID,
			&product.Name,
			&product.Description,
			&product.Category,
			&product.Price,
			&product.Quantity,
			&product.CreatedAt,
			&product.UpdatedAt,
		}

		err := rows.Scan(args...)
		if err != nil {
			return nil, fmt.Errorf("failed to scan rows of products: %v", err)
		}

		products = append(products, product)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	if len(products) == 0 {
		return nil, sql.ErrNoRows
	}

	err = EndTransaction(ctx, tx)
	if err != nil {
		return nil, err
	}

	return &products, nil
}
func (p *ProductsStore) Update(ctx context.Context, product *ecommerce.Products) error {
	tx, err := BeginTransaction(p.db)
	if err != nil {
		return err
	}

	query := `
	UPDATE products
	SET name = $1, description = $2, category = $3, price = $4, quantity = $5, updated_at = NOW()
	WHERE id = $6
	RETURNING updated_at
	`

	args := []any{
		&product.Name,
		&product.Description,
		&product.Category,
		&product.Price,
		&product.Quantity,
		&product.ID,
	}

	err = tx.QueryRowContext(ctx, query, args...).Scan(&product.UpdatedAt)
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
		return fmt.Errorf("failed to update product: %v", err)
	}

	err = EndTransaction(ctx, tx)
	if err != nil {
		return err
	}

	return nil
}

func (p *ProductsStore) Delete(ctx context.Context, id int) error {
	tx, err := BeginTransaction(p.db)
	if err != nil {
		return err
	}

	query := `
	DELETE FROM products
	WHERE id = $1
	`

	result, err := tx.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete from products: %v", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows != 1 {
		return fmt.Errorf("expected to affect 1 rows, affected: %d", rows)
	}

	err = EndTransaction(ctx, tx)
	if err != nil {
		return err
	}

	return nil
}
