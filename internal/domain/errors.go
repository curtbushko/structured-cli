package domain

import (
	"errors"
	"fmt"
)

// Domain-specific error types.
// These errors represent business logic failures in the domain layer.

var (
	// ErrEmptyCommand is returned when attempting to parse an empty command.
	ErrEmptyCommand = errors.New("command cannot be empty")

	// ErrInvalidSchema is returned when a schema cannot be parsed.
	ErrInvalidSchema = errors.New("invalid schema")

	// ErrSchemaValidation is returned when data fails schema validation.
	ErrSchemaValidation = errors.New("schema validation failed")
)

// ExitError represents an error with an associated exit code.
// This allows propagating the exit code from executed commands to the CLI.
type ExitError struct {
	Code int
	Err  error
}

// NewExitError creates an ExitError with the given code and wrapped error.
// If err is nil, the error message will be based on the exit code alone.
func NewExitError(code int, err error) *ExitError {
	return &ExitError{Code: code, Err: err}
}

// Error implements the error interface.
func (e *ExitError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("exit code %d: %v", e.Code, e.Err)
	}
	return fmt.Sprintf("exit code %d", e.Code)
}

// Unwrap returns the underlying error for errors.Is and errors.As.
func (e *ExitError) Unwrap() error {
	return e.Err
}

// ExitCode returns the exit code from an error if it is or wraps an ExitError.
// Returns 1 if the error is not an ExitError and not nil, or 0 if err is nil.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *ExitError
	if errors.As(err, &exitErr) {
		return exitErr.Code
	}
	return 1
}
