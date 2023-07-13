package domain

import (
	"context"
	"errors"
)

var (
	ErrNoSearchResult = errors.New("requested resource not found")
)

// Search represents search results to users.
type Search struct {
	Prodcuts   *Product  `json:"products,omitempty"`
	Categories *Category `json:"categories,omitempty"`
}

// Searcher searchs in database for asked query.
type Searcher interface {
	Search(ctx context.Context, query string) ([]Search, error)
}
