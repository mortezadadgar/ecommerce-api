package ecommerce

import (
	"context"
	"fmt"
	"time"
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
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CategoriesInput struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
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
func (p *Categories) Validate() error {
	switch {
	case p.Name == "":
		return fmt.Errorf("name is required")
	case p.Description == "":
		return fmt.Errorf("description is required")
	}
	return nil
}

// SetValuesTo checks whether input are not nil and set values.
func (p *CategoriesInput) SetValuesTo(category *Categories) {
	if p.Name != nil {
		category.Name = *p.Name
	}

	if p.Description != nil {
		category.Description = *p.Description
	}
}

func (p *CategoriesFilter) Validate() error {
	return nil
}
