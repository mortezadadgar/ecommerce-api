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

// UsersStore represents users database.
type UsersStore struct {
	db *pgxpool.Pool
}

// NewUsersStore returns a new instance of UsersStore.
func NewUsersStore(db *pgxpool.Pool) *UsersStore {
	return &UsersStore{db: db}
}

// Create creates a new user in store.
func (u *UsersStore) Create(ctx context.Context, user *domain.Users) error {
	tx, err := u.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", ErrBeginTransaction, err)
	}
	defer tx.Rollback(ctx)

	query := `
	INSERT INTO users(email, password_hash) VALUES(@email, @password_hash)
	RETURNING id, created_at, updated_at
	`

	args := pgx.NamedArgs{
		"email":         &user.Email,
		"password_hash": &user.Password,
	}

	err = tx.QueryRow(ctx, query, args).Scan(
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

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", ErrCommitTransaction, err)
	}

	return nil
}

// GetByID get user by id from store.
func (u *UsersStore) GetByID(ctx context.Context, id int) (*domain.Users, error) {
	tx, err := u.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v: %v", ErrBeginTransaction, err)
	}
	defer tx.Rollback(ctx)

	query := `
	SELECT id, email, created_at, updated_at FROM users
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

	user, err := pgx.CollectOneRow(rows, pgx.RowToStructByName[domain.Users])
	switch {
	// serial type starts from 1
	case user.ID == 0:
		return nil, sql.ErrNoRows
	case err != nil:
		return nil, fmt.Errorf("failed to query user: %v", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v: %v", ErrCommitTransaction, err)
	}

	return &user, nil
}
