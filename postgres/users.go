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

type UsersStore struct {
	db *sql.DB
}

func NewUsersStore(db *sql.DB) *UsersStore {
	return &UsersStore{db: db}
}

func (u *UsersStore) Create(ctx context.Context, user *ecommerce.Users) error {
	tx, err := BeginTransaction(u.db)
	if err != nil {
		return err
	}

	query := `
	INSERT INTO users(email, password_hash) VALUES($1, $2)
	RETURNING id, created_at, updated_at
	`

	args := []any{
		&user.Email,
		&user.Password,
	}

	err = tx.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.UniqueViolation {
				return ErrDuplicatedEntries
			}
		}
		return fmt.Errorf("failed to insert into users: %v", err)
	}

	err = EndTransaction(tx)
	if err != nil {
		return err
	}

	return nil
}

func (u *UsersStore) GetByID(ctx context.Context, id int) (*ecommerce.Users, error) {
	tx, err := BeginTransaction(u.db)
	if err != nil {
		return nil, err
	}

	query := `
	SELECT id, email, created_at, updated_at FROM users
	WHERE id = $1
	`

	user := ecommerce.Users{}
	args := []any{
		&user.ID,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	}

	err = tx.QueryRowContext(ctx, query, id).Scan(args...)
	switch {
	case err == sql.ErrNoRows:
		return nil, sql.ErrNoRows
	case err != nil:
		return nil, fmt.Errorf("failed to query user: %v", err)
	}

	err = EndTransaction(tx)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
