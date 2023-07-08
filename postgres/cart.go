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

// CartStore represents carts database.
type CartStore struct {
	db *pgxpool.Pool
}

// NewCartStore returns a new instance of CartStore.
func NewCartStore(db *pgxpool.Pool) CartStore {
	return CartStore{db: db}
}

// Create creates a new cart in store.
func (c CartStore) Create(ctx context.Context, cart *domain.Cart) error {
	tx, err := c.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", store.ErrBeginTransaction, err)

	}
	defer tx.Rollback(ctx)

	query := `
	INSERT INTO carts(product_id, quantity, user_id)
	VALUES(@product_id, @quantity, @user_id)
	RETURNING id
	`

	args := pgx.NamedArgs{
		"product_id": cart.ProductID,
		"quantity":   cart.Quantity,
		"user_id":    cart.UserID,
	}

	err = tx.QueryRow(ctx, query, args).Scan(&cart.ID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.ForeignKeyViolation:
				return store.ErrForeinKeyViolation
			}
		}

		return fmt.Errorf("failed to insert into carts: %v", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", store.ErrCommitTransaction, err)
	}

	return nil
}

// GetByUser get user's cart by id from store.
func (c CartStore) GetByUser(ctx context.Context, userID int) ([]domain.Cart, error) {
	cart, err := c.List(ctx, domain.CartFilter{UserID: userID})
	if err != nil {
		return nil, err
	}

	return cart, nil
}

// GetByID get cart by id from store.
func (c CartStore) GetByID(ctx context.Context, ID int) (domain.Cart, error) {
	cart, err := c.List(ctx, domain.CartFilter{ID: ID})
	if err != nil {
		return domain.Cart{}, err
	}

	return cart[0], nil
}

// List lists carts with optional filter.
func (c CartStore) List(ctx context.Context, filter domain.CartFilter) ([]domain.Cart, error) {
	tx, err := c.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v: %v", store.ErrBeginTransaction, err)
	}
	defer tx.Rollback(ctx)

	query := `
	SELECT * FROM carts
	WHERE 1=1
	` + store.FormatSort(filter.Sort) + `
	` + store.FormatAndIntOp("id", filter.ID) + `
	` + store.FormatAndIntOp("user_id", filter.UserID) + `
	` + store.FormatLimitOffset(filter.Limit, filter.Offset) + `
	`

	rows, err := tx.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query list carts: %v", err)
	}

	carts, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Cart])
	if err != nil {
		return nil, fmt.Errorf("failed to scan rows of carts: %v", err)
	}

	if len(carts) == 0 {
		return nil, sql.ErrNoRows
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v: %v", store.ErrCommitTransaction, err)
	}

	return carts, nil
}

// Update updates a cart by id in store.
func (c CartStore) Update(ctx context.Context, cart *domain.Cart) error {
	tx, err := c.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", store.ErrBeginTransaction, err)
	}
	defer tx.Rollback(ctx)

	query := `
	UPDATE carts
	SET product_id = @product_id, quantity = @quantity
	WHERE id = @id
	`

	args := pgx.NamedArgs{
		"product_id": &cart.ProductID,
		"quantity":   &cart.Quantity,
		"id":         &cart.ID,
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

		return fmt.Errorf("failed to update cart: %v", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", store.ErrCommitTransaction, err)
	}

	return nil
}

// Delete deletes a cart by id from store.
func (c CartStore) Delete(ctx context.Context, ID int) error {
	tx, err := c.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", store.ErrBeginTransaction, err)
	}
	defer tx.Rollback(ctx)

	query := `
	DELETE FROM carts
	WHERE id = @id
	`

	args := pgx.NamedArgs{
		"id": ID,
	}

	result, err := tx.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("failed to delete from users: %v", err)
	}

	if rows := result.RowsAffected(); rows != 1 {
		return sql.ErrNoRows
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", store.ErrCommitTransaction, err)
	}

	return nil
}
