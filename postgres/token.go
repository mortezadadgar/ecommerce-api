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

// TokenStore represents tokens database.
type TokenStore struct {
	db *pgxpool.Pool
}

// NewTokenStore returns a new instance of TokenStore
func NewTokenStore(db *pgxpool.Pool) TokenStore {
	return TokenStore{db: db}
}

// Create creates a new token in store.
func (t TokenStore) Create(ctx context.Context, token domain.Token) error {
	query := `
	INSERT INTO tokens(hashed, user_id, expiry)
	VALUES(@hashed, @user_id, @expiry)
	`

	args := pgx.NamedArgs{
		"hashed":  &token.Hashed,
		"user_id": &token.UserID,
		"expiry":  &token.Expiry,
	}

	_, err := t.db.Exec(ctx, query, args)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case pgerrcode.ForeignKeyViolation:
				return store.ErrForeinKeyViolation
			}
		}

		return fmt.Errorf("failed to insert into tokens: %v", err)
	}

	return nil
}

// GetUser get user by token and returns store.ErrNoRows on expired tokens.
func (t TokenStore) GetUser(ctx context.Context, plainToken string) (domain.User, error) {
	query := `
	SELECT id, email, password_hash FROM users
	INNER JOIN tokens ON users.id = tokens.user_id
	WHERE tokens.hashed = @hashed
	AND tokens.expiry > NOW()
	`

	hashedToken := domain.HashToken(plainToken)

	args := pgx.NamedArgs{
		"hashed": hashedToken,
	}

	rows, err := t.db.Query(ctx, query, args)
	if err != nil {
		return domain.User{}, err
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.User])
	switch {
	case err != nil:
		return domain.User{}, fmt.Errorf("failed to query user: %v", err)
	case user.ID == 0:
		return domain.User{}, sql.ErrNoRows
	}

	return user, nil
}
