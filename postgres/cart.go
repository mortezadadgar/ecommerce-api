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

// cartStore represents carts database.
type cartStore struct {
	db *pgxpool.Pool
}

// NewCartStore returns a new instance of CartStore.
func NewCartStore(db *pgxpool.Pool) cartStore {
	return cartStore{db: db}
}

// Create creates a new cart in database.
func (c cartStore) Create(ctx context.Context, cart *domain.Cart) error {
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

	err := c.db.QueryRow(ctx, query, args).Scan(&cart.ID)
	if err != nil {
		pgErr := pgError(err)
		if pgErr.Code == pgerrcode.ForeignKeyViolation {
			if pgErr.ConstraintName == "carts_user_id_fkey" {
				return domain.ErrCartInvalidUserID
			}
			if pgErr.ConstraintName == "carts_product_id_fkey" {
				return domain.ErrCartInvalidProductID
			}
		}
		return err
	}

	return nil
}

// GetByUser get user's cart by id from database.
func (c cartStore) GetByUser(ctx context.Context, userID int) ([]domain.Cart, error) {
	cart, err := c.List(ctx, domain.CartFilter{UserID: userID})
	if err != nil {
		return nil, err
	}

	return cart, nil
}

// GetByID get cart by id from database.
func (c cartStore) GetByID(ctx context.Context, ID int) (domain.Cart, error) {
	cart, err := c.List(ctx, domain.CartFilter{ID: ID})
	if err != nil {
		return domain.Cart{}, err
	}

	return cart[0], nil
}

// List lists carts with optional filter.
func (c cartStore) List(ctx context.Context, filter domain.CartFilter) ([]domain.Cart, error) {
	query := `
	SELECT * FROM carts
	WHERE 1=1
	` + FormatSort(filter.Sort) + `
	` + FormatAndInt("id", filter.ID) + `
	` + FormatAndInt("user_id", filter.UserID) + `
	` + FormatLimitOffset(filter.Limit, filter.Offset) + `
	`

	rows, err := c.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query list carts: %v", err)
	}

	carts, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.Cart])
	if err != nil {
		return nil, fmt.Errorf("failed to scan rows of carts: %v", err)
	}

	if len(carts) == 0 {
		return nil, domain.ErrNoCartsFound
	}

	return carts, nil
}

// Update updates a cart by id in database.
func (c cartStore) Update(ctx context.Context, ID int, input domain.CartUpdate) (domain.Cart, error) {
	query := `
	UPDATE carts
	SET product_id = COALESCE(@product_id, product_id),
		quantity   = COALESCE(@quantity, quantity)
	WHERE id = @id
	RETURNING *
	`

	args := pgx.NamedArgs{
		"product_id": &input.ProductID,
		"quantity":   &input.Quantity,
		"id":         &ID,
	}

	row, err := c.db.Query(ctx, query, args)
	if err != nil {
		return domain.Cart{}, err
	}

	cart, err := pgx.CollectOneRow(row, pgx.RowToStructByName[domain.Cart])
	if err != nil {
		pgErr := pgError(err)
		if pgErr.Code == pgerrcode.ForeignKeyViolation {
			if pgErr.ConstraintName == "carts_product_id_fkey" {
				return domain.Cart{}, domain.ErrCartInvalidProductID
			}
		}

		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Cart{}, domain.ErrNoCartsFound
		}

		return domain.Cart{}, fmt.Errorf("failed to scan rows of cart: %v", err)
	}

	return cart, nil
}

// Delete deletes a cart by id from database.
func (c cartStore) Delete(ctx context.Context, ID int) error {
	query := `
	DELETE FROM carts
	WHERE id = @id
	`

	args := pgx.NamedArgs{
		"id": ID,
	}

	result, err := c.db.Exec(ctx, query, args)
	if err != nil {
		return fmt.Errorf("failed to delete from users: %v", err)
	}

	if rows := result.RowsAffected(); rows != 1 {
		return domain.ErrNoCartsFound
	}

	return nil
}
