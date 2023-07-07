package domain

import (
	"context"

	"github.com/go-playground/validator/v10"
)

// WrapCarts wraps carts for user representation.
type WrapCarts struct {
	Carts Carts `json:"cart"`
}

// WrapCartsList wraps list of carts for user representation.
type WrapCartsList struct {
	Carts []Carts `json:"carts"`
}

// Carts represents carts model.
type Carts struct {
	ID        int `json:"id"`
	ProductID int `json:"product_id" db:"product_id"`
	Quantity  int `json:"quantity" db:"quantity"`
	UserID    int `json:"user_id" db:"user_id"`
}

// CartsCreate represents carts model for POST requests.
type CartsCreate struct {
	UserID    int `json:"user_id" validate:"required"`
	ProductID int `json:"product_id" validate:"required"`
	Quantity  int `json:"quantity" validate:"required"`
}

// CartsUpdate represents carts model for PATCH requests.
type CartsUpdate struct {
	ProductID *int `json:"product_id" validate:"omitempty,required"`
	Quantity  *int `json:"quantity" validate:"omitempty,required"`
}

// CartsFilter represents filters passed to List.
type CartsFilter struct {
	ID     int `json:"id"`
	UserID int `json:"user_id"`

	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Sort   string `json:"sort"`
}

// CartsService represents a service for managing carts.
type CartsService interface {
	GetByUser(ctx context.Context, userID int) ([]Carts, error)
	GetByID(ctx context.Context, id int) (Carts, error)
	List(ctx context.Context, filter CartsFilter) ([]Carts, error)
	Create(ctx context.Context, cart *Carts) error
	Update(ctx context.Context, cart *Carts) error
	Delete(ctx context.Context, userID int) error
}

// Validate validates create products.
func (c CartsCreate) Validate() error {
	v := validator.New()
	return v.Struct(c)
}

// CreateModel set input values to a new struct and return a new instance.
func (c CartsCreate) CreateModel() Carts {
	return Carts{
		ProductID: c.ProductID,
		Quantity:  c.Quantity,
		UserID:    c.UserID,
	}
}

// Validate validates update products.
func (c CartsUpdate) Validate() error {
	v := validator.New()
	return v.Struct(c)
}

// UpdateModel checks whether carts input are not nil and set values.
func (c CartsUpdate) UpdateModel(cart *Carts) {
	if c.ProductID != nil {
		cart.ProductID = *c.ProductID
	}

	if c.Quantity != nil {
		cart.Quantity = *c.Quantity
	}
}
