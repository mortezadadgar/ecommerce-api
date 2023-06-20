package ecommerce

import (
	"context"
	"fmt"
	"net/mail"
	"time"
)

type WrapUsers struct {
	User Users `json:"user"`
}

type Users struct {
	ID        int       `json:"-"`
	Email     string    `json:"email"`
	Password  []byte    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UsersInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UsersService interface {
	Create(ctx context.Context, user *Users) error
	GetByID(ctx context.Context, user int) (*Users, error)
}

// TODO: maybe use a library?
func (u *UsersInput) Validate() error {
	switch {
	case len(u.Password) < 8:
		return fmt.Errorf("password must be at least 8 bytes long")
	case len(u.Password) > 72:
		return fmt.Errorf("password must not be more than 72 bytes long")
	case u.Email == "":
		return fmt.Errorf("email must be provided")
	case len(u.Email) >= 500:
		return fmt.Errorf("email must not be more than 500 bytes long")
	case !u.isEmailValid():
		return fmt.Errorf("invalid email address")
	}

	return nil
}

func (u *UsersInput) isEmailValid() bool {
	_, err := mail.ParseAddress(u.Email)
	return err == nil
}
