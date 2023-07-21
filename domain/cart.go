package domain

import (
	"context"
	"errors"
)

var (
	ErrCartInvalidUserID    = errors.New("invalid user id cart")
	ErrCartInvalidProductID = errors.New("invalid product id cart")
	ErrNoCartsFound         = errors.New("no carts found")

	errUserIDRequired    = errors.New("user_id is required")
	errProductIDRequired = errors.New("product_id is required")
)

// WrapCart wraps carts for user representation.
type WrapCart struct {
	Cart Cart `json:"cart"`
}

// WrapCartList wraps list of carts for user representation.
type WrapCartList struct {
	Carts []Cart `json:"carts"`
}

// Cart represents carts model.
type Cart struct {
	ID        int `json:"id"`
	ProductID int `json:"product_id" db:"product_id"`
	Quantity  int `json:"quantity" db:"quantity"`
	UserID    int `json:"user_id" db:"user_id"`
}

// CartCreate represents carts model for POST requests.
type CartCreate struct {
	UserID    int `json:"user_id"`
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

// CartUpdate represents carts model for PATCH requests.
type CartUpdate struct {
	ProductID *int `json:"product_id"`
	Quantity  *int `json:"quantity"`
}

// CartFilter represents filters passed to List.
type CartFilter struct {
	ID     int `json:"id"`
	UserID int `json:"user_id"`

	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Sort   string `json:"sort"`
}

// CartService represents a service for managing carts.
type CartService interface {
	GetByUser(ctx context.Context, userID int) ([]Cart, error)
	GetByID(ctx context.Context, ID int) (Cart, error)
	List(ctx context.Context, filter CartFilter) ([]Cart, error)
	Create(ctx context.Context, cart *Cart) error
	Update(ctx context.Context, ID int, cart CartUpdate) (Cart, error)
	Delete(ctx context.Context, userID int) error
}

// Validate validates update products.
func (c CartCreate) Validate() error {
	switch {
	case c.UserID == 0:
		return errUserIDRequired
	case c.ProductID == 0:
		return errProductIDRequired
	case c.Quantity == 0:
		return errQuantityRequired
	}
	return nil
}

// CreateModel set input values to a new struct and return a new instance.
func (c CartCreate) CreateModel() Cart {
	return Cart{
		ProductID: c.ProductID,
		Quantity:  c.Quantity,
		UserID:    c.UserID,
	}
}

// Validate validates update products.
func (c CartUpdate) Validate() error {
	switch {
	case c.ProductID != nil && *c.ProductID == 0:
		return errProductIDRequired
	case c.Quantity != nil && *c.Quantity == 0:
		return errQuantityRequired
	}
	return nil
}

// UpdateModel checks whether carts input are not nil and set values.
func (c CartUpdate) UpdateModel(cart *Cart) {
	if c.ProductID != nil {
		cart.ProductID = *c.ProductID
	}

	if c.Quantity != nil {
		cart.Quantity = *c.Quantity
	}
}
