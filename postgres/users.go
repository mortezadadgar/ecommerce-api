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

type UsersStore struct {
	db *pgxpool.Pool
}

func NewUsersStore(db *pgxpool.Pool) *UsersStore {
	return &UsersStore{db: db}
}

func (u *UsersStore) Create(ctx context.Context, user *domain.Users) error {
	tx, err := u.db.Begin(ctx)
	if err != nil {
		return ErrBeginTransaction
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
		return ErrCommitTransaction
	}

	return nil
}

func (u *UsersStore) GetByID(ctx context.Context, id int) (*domain.Users, error) {
	tx, err := u.db.Begin(ctx)
	if err != nil {
		return nil, ErrBeginTransaction
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
		return nil, ErrCommitTransaction
	}

	return &user, nil
}
