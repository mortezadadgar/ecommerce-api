// Package store collects functions mainly used by databases.
package store

import (
	"errors"
	"fmt"
)

// ErrBeginTransaction returned when transaction fails from beginning.
var ErrBeginTransaction = errors.New("failed to begin transaction")

// ErrInvalidColumn returned when the requested column is invalid.
var ErrInvalidColumn = errors.New("requested data field could not be found")

// ErrCommitTransaction returned when commit transaction fails.
var ErrCommitTransaction = errors.New("failed to commit transaction")

// ErrDuplicatedEntries returned when driver unique constraints trigger error.
var ErrDuplicatedEntries = errors.New("duplicated entries are not allowed")

// ErrUnableDeleteEntry returned when driver is not able to remove a row.
var ErrUnableDeleteEntry = errors.New("unable to remove the entry")

// ErrForeinKeyViolation returned when a invalid category used for products.
var ErrForeinKeyViolation = errors.New("product category must matches a category entry")

// FormatLimitOffset returns a SQL string for a given limit & offset.
func FormatLimitOffset(limit int, offset int) string {
	switch {
	case limit > 0 && offset > 0:
		return fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)
	case limit > 0:
		return fmt.Sprintf("LIMIT %d", limit)
	case offset > 0:
		return fmt.Sprintf("OFFSET %d", offset)
	}

	return ""
}

// FormatSort returns a SQL string for a giving column.
func FormatSort(v string) string {
	if v != "" {
		return fmt.Sprintf("ORDER BY %s", v)
	}

	return ""
}

// FormatAndOp returns a SQL string for a giving column and string value.
func FormatAndOp(s string, v string) string {
	if v != "" {
		return fmt.Sprintf("AND %s='%s'", s, v)
	}

	return ""
}

// FormatAndIntOp returns a SQL string for a giving column and integer value.
func FormatAndIntOp(s string, v int) string {
	if v != 0 {
		return fmt.Sprintf("AND %s='%d'", s, v)
	}

	return ""
}
