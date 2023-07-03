package domain

import (
	"context"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
)

type WrapCategories struct {
	Category Categories `json:"category"`
}

type WrapCategoriesList struct {
	Categories []Categories `json:"categories"`
}

type Categories struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"-" db:"created_at"`
	UpdatedAt   time.Time `json:"-" db:"updated_at"`
}

type CategoriesCreate struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
}

type CategoriesUpdate struct {
	Name        *string `json:"name" validate:"omitempty,required"`
	Description *string `json:"description" validate:"omitempty,required"`
}

type CategoriesFilter struct {
	Name string `json:"name"`
	Sort string `json:"sort"`

	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type CategoriesService interface {
	Create(ctx context.Context, category *Categories) error
	GetByID(ctx context.Context, id int) (*Categories, error)
	Update(ctx context.Context, category *Categories) error
	Delete(ctx context.Context, ID int) error
	List(ctx context.Context, filter CategoriesFilter) (*[]Categories, error)
}

// Validate validates catergories.
func (c CategoriesCreate) Validate(r *http.Request) error {
	v := validator.New()
	return v.Struct(c)
}

// CreateModel set input values to a new struct and return a new instance.
func (p CategoriesCreate) CreateModel() Categories {
	return Categories{
		Name:        p.Name,
		Description: p.Description,
	}
}

// Validate validates catergories.
func (c CategoriesUpdate) Validate(r *http.Request) error {
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
