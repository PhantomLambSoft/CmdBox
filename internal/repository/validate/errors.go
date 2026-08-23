package validate

import (
	"errors"
	"fmt"
)

// ErrValidation represents a generic validation error used across the application.
var ErrValidation = errors.New("validation error")

// validationErrorf formats a validation error message using the provided format and arguments and wraps it with ErrValidation.
func validationErrorf(format string, args ...any) error {
	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), ErrValidation)
}
