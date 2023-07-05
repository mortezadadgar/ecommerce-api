package domain

import (
	"context"

	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

// WrapUsers wraps users for user representation.
type WrapUsers struct {
	User Users `json:"user"`
}

// WrapUsersList wraps list of users for user representation.
type WrapUsersList struct {
	Users []Users `json:"users"`
}

// Users represents users model.
type Users struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password []byte `json:"-" db:"password_hash"`
}

// UsersCreate represents users model for POST requests.
type UsersCreate struct {
	Email    string `json:"email" validate:"required,email,lte=500"`
	Password string `json:"password" validate:"required,gte=8,lte=72"`
}

// UsersLogin represents users model for login requests.
type UsersLogin struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// UsersFilter represents filters passed to /users requests.
type UsersFilter struct {
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Sort   string `json:"sort"`
}

// UsersService represents a service for managing users.
type UsersService interface {
	Create(ctx context.Context, user *Users) error
	GetByID(ctx context.Context, ID int) (Users, error)
	GetByEmail(ctx context.Context, email string) (Users, error)
	Delete(ctx context.Context, ID int) error
	List(ctx context.Context, filter UsersFilter) ([]Users, error)
}

// Validate validates create users.
func (u UsersLogin) Validate() error {
	v := validator.New()
	return v.Struct(u)
}

// Validate validates create users.
func (u UsersCreate) Validate() error {
	v := validator.New()
	return v.Struct(u)
}

// CreateModel set input values and password to a new struct and return a new instance.
func (u UsersCreate) CreateModel(password []byte) Users {
	return Users{
		Email:    u.Email,
		Password: password,
	}
}

// GenerateHashedPassword generates hashed password.
func GenerateHashedPassword(password []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(password, bcrypt.MinCost)
}

// CompareHashAndPassword compares hash and plaintext password.
func CompareHashAndPassword(hashedPassword []byte, password []byte) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, password)
}
