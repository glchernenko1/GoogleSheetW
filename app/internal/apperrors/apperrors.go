package apperrors

import (
	"errors"
	"fmt"
)

var (
	ErrCacheNotFound = errors.New("not found")
)

type ValidationError struct {
	Field string
	Msg   string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed: field '%s': %s", e.Field, e.Msg)
}
