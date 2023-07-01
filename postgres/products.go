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
)

type ProductsStore struct {
	db *pgxpool.Pool
}

func NewProductsStore(db *pgxpool.Pool) *ProductsStore {
	return &ProductsStore{db: db}
}

func (p *ProductsStore) Create(ctx context.Context, product *domain.Products) error {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return ErrBeginTransaction
	}
	defer tx.Rollback(ctx)

	query := `
	 INSERT INTO products(name, description, category, price, quantity)
	 VALUES(@name, @description, @category, @price, @quantity)
	 RETURNING id, created_at, updated_at
	`

	args := pgx.NamedArgs{
		"name":        &product.Name,
		"description": &product.Description,
		"category":    &product.Category,
		"price":       &product.Price,
		"quantity":    &product.Quantity,
	}

	err = tx.QueryRow(ctx, query, args).Scan(
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

	err = tx.Commit(ctx)
	if err != nil {
		return ErrCommitTransaction
	}

	return nil
}

func (p *ProductsStore) GetByID(ctx context.Context, id int) (*domain.Products, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return nil, ErrBeginTransaction
	}
	defer tx.Rollback(ctx)

	query := `
	SELECT * FROM products
	WHERE id = @id
	`

	args := pgx.NamedArgs{
		"id": id,
	}

	rows, err := tx.Query(ctx, query, args)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	product, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Products])
	switch {
	case product.ID == 0:
		return nil, sql.ErrNoRows
	case err != nil:
		return nil, fmt.Errorf("failed to get product: %v", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, ErrCommitTransaction
	}

	return &product, nil
}

func (p *ProductsStore) List(ctx context.Context, filter domain.ProductsFilter) (*[]domain.Products, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return nil, ErrBeginTransaction
	}
	defer tx.Rollback(ctx)

	query := `
	SELECT * FROM products
	WHERE 1=1
	` + FormatSort(filter.Sort) + `
	` + FormatAndOp("category", filter.Category) + `
	` + FormatAndOp("name", filter.Name) + `
	` + FormatLimitOffset(filter.Limit, filter.Offset) + `
	`

	rows, err := tx.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query list products: %v", err)
	}
	defer rows.Close()

	products, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Products])
	if err != nil {
		return nil, fmt.Errorf("failed to scan rows of products: %v", err)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	if len(products) == 0 {
		return nil, sql.ErrNoRows
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, ErrCommitTransaction
	}

	return &products, nil
}
func (p *ProductsStore) Update(ctx context.Context, product *domain.Products) error {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return ErrBeginTransaction
	}
	defer tx.Rollback(ctx)

	query := `
	UPDATE products
	SET name = @name, description = @description, category = @category, price = @price, quantity = @quantity, updated_at = NOW()
	WHERE id = @id
	RETURNING updated_at
	`

	args := pgx.NamedArgs{
		"name":        &product.Name,
		"description": &product.Description,
		"category":    &product.Category,
		"price":       &product.Price,
		"quantity":    &product.Quantity,
		"id":          &product.ID,
	}

	err = tx.QueryRow(ctx, query, args).Scan(&product.UpdatedAt)
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

	err = tx.Commit(ctx)
	if err != nil {
		return ErrCommitTransaction
	}

	return nil
}

func (p *ProductsStore) Delete(ctx context.Context, id int) error {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return ErrBeginTransaction
	}
	defer tx.Rollback(ctx)

	query := `
	DELETE FROM products
	WHERE id = @id
	`

	args := pgx.NamedArgs{
		"id": id,
	}

	result, err := tx.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("failed to delete from products: %v", err)
	}

	rows := result.RowsAffected()
	if rows != 1 {
		return fmt.Errorf("expected to affect 1 rows, affected: %d", rows)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return ErrCommitTransaction
	}

	return nil
}
