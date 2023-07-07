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

// CartsStore represents carts database.
type CartsStore struct {
	db *pgxpool.Pool
}

// NewCartsStore returns a new instance of CartsStore.
func NewCartsStore(db *pgxpool.Pool) CartsStore {
	return CartsStore{db: db}
}

// Create creates a new cart in store.
func (c CartsStore) Create(ctx context.Context, cart *domain.Carts) error {
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
func (c CartsStore) GetByUser(ctx context.Context, userID int) ([]domain.Carts, error) {
	cart, err := c.List(ctx, domain.CartsFilter{UserID: userID})
	if err != nil {
		return nil, err
	}

	return cart, nil
}

// GetByID get cart by id from store.
func (c CartsStore) GetByID(ctx context.Context, id int) (domain.Carts, error) {
	cart, err := c.List(ctx, domain.CartsFilter{ID: id})
	if err != nil {
		return domain.Carts{}, err
	}

	return cart[0], nil
}

// List lists carts with optional filter.
func (c CartsStore) List(ctx context.Context, filter domain.CartsFilter) ([]domain.Carts, error) {
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

	carts, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Carts])
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
func (c CartsStore) Update(ctx context.Context, cart *domain.Carts) error {
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
func (c CartsStore) Delete(ctx context.Context, id int) error {
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
		"id": id,
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
