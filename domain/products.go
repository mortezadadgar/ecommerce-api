package domain

import (
	"context"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
)

type WrapProducts struct {
	Product Products `json:"product"`
}

type WrapProductsList struct {
	Products []Products `json:"products"`
}

type Products struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Price       int       `json:"price"`
	Quantity    int       `json:"quantity"`
	CreatedAt   time.Time `json:"-" db:"created_at"`
	UpdatedAt   time.Time `json:"-" db:"updated_at"`
}

type ProductsCreate struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
	Category    string `json:"category" validate:"required"`
	Price       int    `json:"price" validate:"required"`
	Quantity    int    `json:"quantity" validate:"required"`
}

type ProductsUpdate struct {
	Name        *string `json:"name" validate:"omitempty,required"`
	Description *string `json:"description" validate:"omitempty,required"`
	Category    *string `json:"category" validate:"omitempty,required"`
	Price       *int    `json:"price" validate:"omitempty,required"`
	Quantity    *int    `json:"quantity" validate:"omitempty,required"`
}

type ProductsFilter struct {
	Category string `json:"category"`
	Name     string `json:"name"`

	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Sort   string `json:"sort"`
}

type ProductsService interface {
	Create(ctx context.Context, product *Products) error
	GetByID(ctx context.Context, id int) (*Products, error)
	Update(ctx context.Context, product *Products) error
	Delete(ctx context.Context, ID int) error
	List(ctx context.Context, filter ProductsFilter) (*[]Products, error)
}

// Validate validates create products.
func (p ProductsCreate) Validate(r *http.Request) error {
	v := validator.New()
	return v.Struct(p)
}

// CreateModel set input values to a new struct and return a new instance.
func (p ProductsCreate) CreateModel() Products {
	return Products{
		Name:        p.Name,
		Description: p.Description,
		Category:    p.Category,
		Price:       p.Price,
		Quantity:    p.Quantity,
	}
}

// Validate validates update products.
func (p ProductsUpdate) Validate(r *http.Request) error {
	v := validator.New()
	return v.Struct(p)
}

// UpdateModel checks whether products input are not nil and set values.
func (p ProductsUpdate) UpdateModel(product *Products) {
	if p.Name != nil {
		product.Name = *p.Name
	}

	if p.Description != nil {
		product.Description = *p.Description
	}

	if p.Category != nil {
		product.Category = *p.Category
	}

	if p.Price != nil {
		product.Price = *p.Price
	}

	if p.Quantity != nil {
		product.Quantity = *p.Quantity
	}
}
