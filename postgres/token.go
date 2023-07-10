package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mortezadadgar/ecommerce-api/domain"
)

// TokenStore represents tokens database.
type TokenStore struct {
	db *pgxpool.Pool
}

// NewTokenStore returns a new instance of TokenStore
func NewTokenStore(db *pgxpool.Pool) TokenStore {
	return TokenStore{db: db}
}

// Create creates a new token in database.
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
		return FormatError(err)
	}

	return nil
}

// GetUserID get user by token and return ErrNoRows on expired tokens.
func (t TokenStore) GetUserID(ctx context.Context, plainToken string) (int, error) {
	// TODO: create a struct for returned values
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
		return 0, err
	}

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.User])
	if err != nil {
		return 0, FormatError(err)
	}

	if user.ID == 0 {
		return 0, domain.Errorf(domain.ENOTFOUND, "requested user by token not found")
	}

	return user.ID, nil
}
