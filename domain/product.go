package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidProductCategory = errors.New("invalid category")
	ErrDuplicatedProduct      = errors.New("duplicated product")
	ErrNoProductsFound        = errors.New("products not found")
	ErrProductConflict        = errors.New("update conflict error")

	errProductNameRequired        = errors.New("name is required")
	errProductDescriptionRequired = errors.New("description is required")
	errQuantityRequired           = errors.New("quantity is required")
	errPriceRequired              = errors.New("price is required")
	errCategoryIDRequired         = errors.New("CategoryID is required")
	errVersionRequired            = errors.New("version is required")
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
	CategoryID  int       `json:"category_id" db:"category_id"`
	Price       int       `json:"price"`
	Quantity    int       `json:"quantity"`
	CreatedAt   time.Time `json:"-" db:"created_at"`
	UpdatedAt   time.Time `json:"-" db:"updated_at"`
	Version     int       `json:"version"`
}

// ProductCreate represents products model for POST requests.
type ProductCreate struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	CategoryID  int    `json:"category_id"`
	Price       int    `json:"price"`
	Quantity    int    `json:"quantity"`
}

// ProductUpdate represents products model for PATCH requests.
type ProductUpdate struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	CategoryID  *int    `json:"category_id"`
	Price       *int    `json:"price"`
	Quantity    *int    `json:"quantity"`
	Version     int     `json:"version"`
}

// ProductFilter represents filters passed to List.
type ProductFilter struct {
	ID         int `json:"id"`
	CategoryID int `json:"category"`

	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
	Sort   string `json:"sort"`
}

// ProductService represents a service for managing products.
type ProductService interface {
	Create(ctx context.Context, product *Product) error
	GetByID(ctx context.Context, ID int) (Product, error)
	Update(ctx context.Context, ID int, product ProductUpdate) (Product, error)
	Delete(ctx context.Context, ID int) error
	List(ctx context.Context, filter ProductFilter) ([]Product, error)
}

// Validate validates POST requests model.
func (p ProductCreate) Validate() error {
	switch {
	case p.Name == "":
		return errProductNameRequired
	case p.Description == "":
		return errProductDescriptionRequired
	case p.Quantity == 0:
		return errQuantityRequired
	case p.Price == 0:
		return errPriceRequired
	case p.CategoryID == 0:
		return errCategoryIDRequired
	}
	return nil
}

// CreateModel set input values to a new struct and return a new instance.
func (p ProductCreate) CreateModel() Product {
	return Product{
		Name:        p.Name,
		Description: p.Description,
		CategoryID:  p.CategoryID,
		Price:       p.Price,
		Quantity:    p.Quantity,
	}
}

// Validate validates PATCH requests model.
func (p ProductUpdate) Validate() error {
	switch {
	case p.Name != nil && *p.Name == "":
		return errProductNameRequired
	case p.Description != nil && *p.Description == "":
		return errProductDescriptionRequired
	case p.Quantity != nil && *p.Quantity == 0:
		return errQuantityRequired
	case p.Price != nil && *p.Price == 0:
		return errPriceRequired
	case p.CategoryID != nil && *p.CategoryID == 0:
		return errCategoryIDRequired
	case p.Version == 0:
		return errVersionRequired
	}
	return nil
}

// UpdateModel checks whether products input are not nil and set values.
func (p ProductUpdate) UpdateModel(product *Product) {
	if p.Name != nil {
		product.Name = *p.Name
	}

	if p.Description != nil {
		product.Description = *p.Description
	}

	if p.CategoryID != nil {
		product.CategoryID = *p.CategoryID
	}

	if p.Price != nil {
		product.Price = *p.Price
	}

	if p.Quantity != nil {
		product.Quantity = *p.Quantity
	}

	product.Version = p.Version
}
