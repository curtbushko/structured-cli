package domain

import (
	"errors"
	"fmt"
	"testing"
)

func TestExitError_Error(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		err      error
		expected string
	}{
		{
			name:     "with wrapped error",
			code:     42,
			err:      errors.New("command failed"),
			expected: "exit code 42: command failed",
		},
		{
			name:     "without wrapped error",
			code:     1,
			err:      nil,
			expected: "exit code 1",
		},
		{
			name:     "zero exit code with error",
			code:     0,
			err:      errors.New("unexpected"),
			expected: "exit code 0: unexpected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exitErr := NewExitError(tt.code, tt.err)
			if got := exitErr.Error(); got != tt.expected {
				t.Errorf("ExitError.Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestExitError_Unwrap(t *testing.T) {
	wrappedErr := errors.New("inner error")
	exitErr := NewExitError(1, wrappedErr)

	// Use errors.Is to check the unwrapped error
	if !errors.Is(exitErr, wrappedErr) {
		t.Errorf("errors.Is(exitErr, wrappedErr) = false, want true")
	}

	// Test nil unwrap
	exitErrNil := NewExitError(1, nil)
	if exitErrNil.Unwrap() != nil {
		t.Errorf("ExitError.Unwrap() = %v, want nil", exitErrNil.Unwrap())
	}
}

func TestExitError_ErrorsAs(t *testing.T) {
	exitErr := NewExitError(42, errors.New("test"))
	wrappedErr := fmt.Errorf("wrapped: %w", exitErr)

	var target *ExitError
	if !errors.As(wrappedErr, &target) {
		t.Error("errors.As failed to find ExitError")
	}
	if target.Code != 42 {
		t.Errorf("target.Code = %d, want 42", target.Code)
	}
}

func TestExitCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: 0,
		},
		{
			name:     "ExitError with code 42",
			err:      NewExitError(42, nil),
			expected: 42,
		},
		{
			name:     "wrapped ExitError",
			err:      fmt.Errorf("wrapped: %w", NewExitError(5, nil)),
			expected: 5,
		},
		{
			name:     "non-exit error",
			err:      errors.New("regular error"),
			expected: 1,
		},
		{
			name:     "ExitError with code 0",
			err:      NewExitError(0, nil),
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExitCode(tt.err); got != tt.expected {
				t.Errorf("ExitCode() = %d, want %d", got, tt.expected)
			}
		})
	}
}
