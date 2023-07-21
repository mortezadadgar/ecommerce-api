// revive:disable until there is a central package.
package domain

// revive: enable

import (
	"context"
	"errors"
	"time"
)

var (
	ErrDuplicatedCategory = errors.New("duplicated category")
	ErrNoCategoryFound    = errors.New("no categories found")
	ErrCategoryConflict   = errors.New("update conflict error")

	errCategoryNameRequired        = errors.New("name is required")
	errCategoryDescriptionRequired = errors.New("description is required")
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
	Version     int       `json:"version"`
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
	Version     int     `json:"version" validate:"required"`
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
	Update(ctx context.Context, ID int, category CategoryUpdate) (Category, error)
	Delete(ctx context.Context, ID int) error
	List(ctx context.Context, filter CategoryFilter) ([]Category, error)
}

// Validate validates POST requests model.
func (c CategoryCreate) Validate() error {
	switch {
	case c.Name == "":
		return errCategoryNameRequired
	case c.Description == "":
		return errCategoryDescriptionRequired
	}
	return nil
}

// CreateModel set input values to a new struct and return a new instance.
func (c CategoryCreate) CreateModel() Category {
	return Category{
		Name:        c.Name,
		Description: c.Description,
	}
}

// Validate validates PATCH requests model.
func (c CategoryUpdate) Validate() error {
	switch {
	case c.Name != nil && *c.Name == "":
		return errCategoryNameRequired
	case c.Description != nil && *c.Description == "":
		return errCategoryDescriptionRequired
	}
	return nil
}

// UpdateModel checks whether input are not nil and set values.
func (c CategoryUpdate) UpdateModel(category *Category) {
	if c.Name != nil {
		category.Name = *c.Name
	}

	if c.Description != nil {
		category.Description = *c.Description
	}

	category.Version = c.Version
}
