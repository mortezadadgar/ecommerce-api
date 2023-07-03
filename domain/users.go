package domain

import (
	"context"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
)

type WrapUsers struct {
	User Users `json:"user"`
}

type Users struct {
	ID        int       `json:"id"`
	Email     string    `json:"email"`
	Password  []byte    `json:"-" db:"-"`
	CreatedAt time.Time `json:"-" db:"created_at"`
	UpdatedAt time.Time `json:"-" db:"updated_at"`
}

type UsersCreate struct {
	Email    string `json:"email" validate:"required,email,lte=500"`
	Password string `json:"password" validate:"required,gte=8,lte=72"`
}

type UsersService interface {
	Create(ctx context.Context, user *Users) error
	GetByID(ctx context.Context, user int) (*Users, error)
}

// Validate validates create users.
func (u UsersCreate) Validate(r *http.Request) error {
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
