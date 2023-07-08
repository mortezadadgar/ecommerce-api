package domain

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
)

// WrapProduct wraps products for user representation.
type WrapProduct struct {
	Product Product `json:"product"`
}

// WrapProductList wraps list of products for user representation.
type WrapProductList struct {
	Products []Product `json:"products"`
}

// Product represents products model.
type Product struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Price       int       `json:"price"`
	Quantity    int       `json:"quantity"`
	CreatedAt   time.Time `json:"-" db:"created_at"`
	UpdatedAt   time.Time `json:"-" db:"updated_at"`
}

// ProductCreate represents products model for POST requests.
type ProductCreate struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
	Category    string `json:"category" validate:"required"`
	Price       int    `json:"price" validate:"required"`
	Quantity    int    `json:"quantity" validate:"required"`
}

// ProductUpdate represents products model for PATCH requests.
type ProductUpdate struct {
	Name        *string `json:"name" validate:"omitempty,required"`
	Description *string `json:"description" validate:"omitempty,required"`
	Category    *string `json:"category" validate:"omitempty,required"`
	Price       *int    `json:"price" validate:"omitempty,required"`
	Quantity    *int    `json:"quantity" validate:"omitempty,required"`
}

// ProductFilter represents filters passed to List.
type ProductFilter struct {
	ID       int    `json:"id"`
	Category string `json:"category"`
	Name     string `json:"name"`

	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Sort   string `json:"sort"`
}

// ProductService represents a service for managing products.
type ProductService interface {
	Create(ctx context.Context, product *Product) error
	GetByID(ctx context.Context, ID int) (Product, error)
	Update(ctx context.Context, product *Product) error
	Delete(ctx context.Context, ID int) error
	List(ctx context.Context, filter ProductFilter) ([]Product, error)
}

// Validate validates create products.
func (p ProductCreate) Validate() error {
	v := validator.New()
	return v.Struct(p)
}

// CreateModel set input values to a new struct and return a new instance.
func (p ProductCreate) CreateModel() Product {
	return Product{
		Name:        p.Name,
		Description: p.Description,
		Category:    p.Category,
		Price:       p.Price,
		Quantity:    p.Quantity,
	}
}

// Validate validates update products.
func (p ProductUpdate) Validate() error {
	v := validator.New()
	return v.Struct(p)
}

// UpdateModel checks whether products input are not nil and set values.
func (p ProductUpdate) UpdateModel(product *Product) {
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