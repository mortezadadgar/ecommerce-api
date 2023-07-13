package domain

import (
	"context"
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicatedUserEmail = errors.New("duplicated email")
	ErrNoUsersFound        = errors.New("no users to list")

	errEmailRequired    = errors.New("email is required")
	errEmailTooLong     = errors.New("email length too long")
	errEmailInvalid     = errors.New("email in not valid")
	errPasswordRequired = errors.New("password is required")
	errPasswrodTooSmall = errors.New("password length too small")
	errPasswordTooLong  = errors.New("password length too long")
)

// WrapUser wraps users for user representation.
type WrapUser struct {
	User User `json:"user"`
}

// WrapUserList wraps list of users for user representation.
type WrapUserList struct {
	Users []User `json:"users"`
}

// User represents users model.
type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Password []byte `json:"-" db:"password_hash"`
}

// UserCreate represents users model for POST requests.
type UserCreate struct {
	Email    string `json:"email" validate:"required,email,lte=500"`
	Password string `json:"password" validate:"required,gte=8,lte=72"`
}

// UserLogin represents users model for login requests.
type UserLogin struct {
	Email    string `json:"email" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// UserFilter represents filters passed to /users requests.
type UserFilter struct {
	Email string `json:"email"`
	ID    int    `json:"id"`

	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Sort   string `json:"sort"`
}

// UserService represents a service for managing users.
type UserService interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, ID int) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	Delete(ctx context.Context, ID int) error
	List(ctx context.Context, filter UserFilter) ([]User, error)
}

// Common erros used for users validations.
var ()

// Validate validates POST requests model.
func (u UserCreate) Validate() error {
	switch {
	case u.Email == "":
		return errEmailRequired
	case len(u.Email) >= 500:
		return errEmailTooLong
	case !strings.Contains(u.Email, "@"):
		return errEmailInvalid
	case u.Password == "":
		return errPasswordRequired
	case len(u.Password) <= 8:
		return errPasswrodTooSmall
	case len(u.Password) >= 72:
		return errPasswordTooLong
	}
	return nil
}

// CreateModel set input values and password to a new struct and return a new instance.
func (u UserCreate) CreateModel(password []byte) User {
	return User{
		Email:    u.Email,
		Password: password,
	}
}

// Validate validates PATCH requests model.
func (u UserLogin) Validate() error {
	switch {
	case u.Email == "":
		return errEmailRequired
	case u.Password == "":
		return errPasswordRequired
	}
	return nil
}

// GenerateHashedPassword generates hashed password.
func GenerateHashedPassword(password []byte) ([]byte, error) {
	return bcrypt.GenerateFromPassword(password, bcrypt.MinCost)
}

// CompareHashAndPassword compares hash and plaintext password.
func CompareHashAndPassword(hashedPassword []byte, password []byte) error {
	return bcrypt.CompareHashAndPassword(hashedPassword, password)
}
