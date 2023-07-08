// revive:disable until there is a central package.
package domain

// revive: enable

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
)

// WrapCategory wraps categories for user representation.
type WrapCategory struct {
	Category Category `json:"category"`
}

// WrapCategoryList wraps list of list categories for user representation.
type WrapCategoryList struct {
	Categories []Category `json:"categories"`
}

// Category represents categories model.
type Category struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"-" db:"created_at"`
	UpdatedAt   time.Time `json:"-" db:"updated_at"`
}

// CategoryCreate represents categories model for POST requests.
type CategoryCreate struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
}

// CategoryUpdate represents categories model for PATCH requests.
type CategoryUpdate struct {
	Name        *string `json:"name" validate:"omitempty,required"`
	Description *string `json:"description" validate:"omitempty,required"`
}

// CategoryFilter represents filters passed to List.
type CategoryFilter struct {
	ID   int    `json:"int"`
	Name string `json:"name"`
	Sort string `json:"sort"`

	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// CategoryService represents a service for managing categories.
type CategoryService interface {
	Create(ctx context.Context, category *Category) error
	GetByID(ctx context.Context, ID int) (Category, error)
	Update(ctx context.Context, category *Category) error
	Delete(ctx context.Context, ID int) error
	List(ctx context.Context, filter CategoryFilter) ([]Category, error)
}

// Validate validates catergories.
func (c CategoryCreate) Validate() error {
	v := validator.New()
	return v.Struct(c)
}

// CreateModel set input values to a new struct and return a new instance.
func (c CategoryCreate) CreateModel() Category {
	return Category{
		Name:        c.Name,
		Description: c.Description,
	}
}

// Validate validates catergories.
func (c CategoryUpdate) Validate() error {
	v := validator.New()
	return v.Struct(c)
}

// UpdateModel checks whether input are not nil and set values.
func (c CategoryUpdate) UpdateModel(category *Category) {
	if c.Name != nil {
		category.Name = *c.Name
	}

	if c.Description != nil {
		category.Description = *c.Description
	}
}
