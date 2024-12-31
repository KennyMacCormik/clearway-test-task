package errors

import (
	"errors"
	"fmt"
)

// ErrNotFound is a sentinel error to indicate resource not found.
var ErrNotFound = errors.New("resource not found")

// NewNotFoundError creates a formatted not-found error.
func NewNotFoundError(resource string) error {
	return fmt.Errorf("%w: %s", ErrNotFound, resource)
}
