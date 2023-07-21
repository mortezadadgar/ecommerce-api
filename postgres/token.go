package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mortezadadgar/ecommerce-api/domain"
)

// tokenStore represents tokens database.
type tokenStore struct {
	db *pgxpool.Pool
}

// NewTokenStore returns a new instance of TokenStore
func NewTokenStore(db *pgxpool.Pool) tokenStore {
	return tokenStore{db: db}
}

// Create creates a new token in database.
func (t tokenStore) Create(ctx context.Context, token domain.Token) error {
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
		pgErr := pgError(err)
		if pgErr.Code == pgerrcode.ForeignKeyViolation {
			return domain.ErrInvalidToken
		}
		return err
	}

	return nil
}

// GetUserID get user by token and return ErrNoRows on expired tokens.
func (t tokenStore) GetUserID(ctx context.Context, plainToken string) (int, error) {
	query := `
	SELECT id FROM users
	INNER JOIN tokens ON users.id = tokens.user_id
	WHERE tokens.hashed = @hashed AND tokens.expiry > NOW()
	`

	hashedToken := domain.HashToken(plainToken)

	args := pgx.NamedArgs{
		"hashed": hashedToken,
	}

	var userID int
	err := t.db.QueryRow(ctx, query, args).Scan(&userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, domain.ErrNoTokenFound
		}
		return 0, err
	}

	return userID, nil
}
