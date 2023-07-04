// revive:disable until there is a central package.
package domain

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
)

// WrapCategories wraps categories for user representation.
type WrapCategories struct {
	Category Categories `json:"category"`
}

// WrapCategoriesList wraps list of list categories for user representation.
type WrapCategoriesList struct {
	Categories []Categories `json:"categories"`
}

// Categories represents categories model.
type Categories struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"-" db:"created_at"`
	UpdatedAt   time.Time `json:"-" db:"updated_at"`
}

// CategoriesCreate represents categories model for POST requests.
type CategoriesCreate struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
}

// CategoriesUpdate represents categories model for PATCH requests.
type CategoriesUpdate struct {
	Name        *string `json:"name" validate:"omitempty,required"`
	Description *string `json:"description" validate:"omitempty,required"`
}

// CategoriesFilter represents filters passed to /categories requests.
type CategoriesFilter struct {
	Name string `json:"name"`
	Sort string `json:"sort"`

	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// CategoriesService represents a service for managing categories.
type CategoriesService interface {
	Create(ctx context.Context, category *Categories) error
	GetByID(ctx context.Context, id int) (Categories, error)
	Update(ctx context.Context, category *Categories) error
	Delete(ctx context.Context, ID int) error
	List(ctx context.Context, filter CategoriesFilter) ([]Categories, error)
}

// Validate validates catergories.
func (c CategoriesCreate) Validate() error {
	v := validator.New()
	return v.Struct(c)
}

// CreateModel set input values to a new struct and return a new instance.
func (c CategoriesCreate) CreateModel() Categories {
	return Categories{
		Name:        c.Name,
		Description: c.Description,
	}
}

// Validate validates catergories.
func (c CategoriesUpdate) Validate() error {
	v := validator.New()
	return v.Struct(c)
}

// UpdateModel checks whether input are not nil and set values.
func (c CategoriesUpdate) UpdateModel(category *Categories) {
	if c.Name != nil {
		category.Name = *c.Name
	}

	if c.Description != nil {
		category.Description = *c.Description
	}
}
