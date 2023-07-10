package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mortezadadgar/ecommerce-api/domain"
)

// UserStore represents users database.
type UserStore struct {
	db *pgxpool.Pool
}

// NewUserStore returns a new instance of UsersStore.
func NewUserStore(db *pgxpool.Pool) UserStore {
	return UserStore{db: db}
}

// Create creates a new user in database.
func (u UserStore) Create(ctx context.Context, user *domain.User) error {
	tx, err := u.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", ErrBeginTransaction, err)
	}
	defer tx.Rollback(ctx)

	query := `
	INSERT INTO users(email, password_hash)
	VALUES(@email, @password_hash)
	RETURNING id
	`

	args := pgx.NamedArgs{
		"email":         &user.Email,
		"password_hash": &user.Password,
	}

	err = tx.QueryRow(ctx, query, args).Scan(&user.ID)
	if err != nil {
		return FormatError(err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", ErrCommitTransaction, err)
	}

	return nil
}

// GetByID get user by id from database.
func (u UserStore) GetByID(ctx context.Context, ID int) (domain.User, error) {
	user, err := u.List(ctx, domain.UserFilter{ID: ID})
	if err != nil {
		return domain.User{}, err
	}

	return user[0], nil
}

// GetByEmail get user by email from database.
func (u UserStore) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	user, err := u.List(ctx, domain.UserFilter{Email: email})
	if err != nil {
		return domain.User{}, err
	}

	return user[0], nil
}

// List lists users with optional filter.
func (u UserStore) List(ctx context.Context, filter domain.UserFilter) ([]domain.User, error) {
	tx, err := u.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v: %v", ErrBeginTransaction, err)
	}
	defer tx.Rollback(ctx)

	query := `
	SELECT id, email, password_hash FROM users
	WHERE 1=1
	` + FormatAndOp("email", filter.Email) + `
	` + FormatAndIntOp("id", filter.ID) + `
	` + FormatSort(filter.Sort) + `
	` + FormatLimitOffset(filter.Limit, filter.Offset) + `
	`

	rows, err := tx.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query list users: %v", err)
	}

	users, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.User])
	if err != nil {
		return nil, fmt.Errorf("failed to scan rows of users: %v", err)
	}

	if len(users) == 0 {
		return nil, domain.Errorf(domain.ENOTFOUND, "there is no user to list")
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("%v: %v", ErrCommitTransaction, err)
	}

	return users, nil
}

// Delete deletes a user by id from database.
func (u UserStore) Delete(ctx context.Context, ID int) error {
	tx, err := u.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", ErrBeginTransaction, err)
	}
	defer tx.Rollback(ctx)

	query := `
	DELETE FROM users
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
		return domain.Errorf(domain.ENOTFOUND, "requested user not found")
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("%v: %v", ErrCommitTransaction, err)
	}

	return nil
}
