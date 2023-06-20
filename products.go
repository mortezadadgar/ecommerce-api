package ecommerce

import (
	"context"
	"fmt"
	"time"
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
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProductsInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Category    *string `json:"category"`
	Price       *int    `json:"price"`
	Quantity    *int    `json:"quantity"`
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

// Validate validates products.
func (p *Products) Validate() error {
	switch {
	case p.Name == "":
		return fmt.Errorf("name is required")
	case p.Category == "":
		return fmt.Errorf("invalid category value")
	case p.Price < 0:
		return fmt.Errorf("invalid price value")
	case p.Quantity < 0:
		return fmt.Errorf("invalid quantity value")
	}

	return nil
}

// SetValuesTo checks whether products input are not nil and set values.
func (p *ProductsInput) SetValuesTo(product *Products) {
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

func (p *ProductsFilter) Validate() error {
	return nil
}
