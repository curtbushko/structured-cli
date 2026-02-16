// Package application contains the business logic and use cases for structured-cli.
package application

import (
	"github.com/curtbushko/structured-cli/internal/domain"
)

// ErrorHandler formats errors for different output modes.
// It transforms ParseResult errors into structured JSON or passthrough formats
// depending on the output mode being used.
type ErrorHandler struct{}

// NewErrorHandler creates a new ErrorHandler.
func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{}
}

// FormatCommandError formats a command execution error for JSON output mode.
// Returns a ParseResult with Data containing:
//   - error: the error message
//   - exitCode: the exit code from the command
//
// This is used when the underlying command fails (non-zero exit code).
func (h *ErrorHandler) FormatCommandError(result domain.ParseResult) domain.ParseResult {
	errorMsg := h.FormatErrorMessage(result.Error)

	data := map[string]any{
		"error":    errorMsg,
		"exitCode": result.ExitCode,
	}

	return domain.ParseResult{
		Data:     data,
		Raw:      result.Raw,
		ExitCode: result.ExitCode,
		Error:    result.Error,
	}
}

// FormatPassthroughError formats an error for passthrough mode.
// Returns the result unchanged since passthrough mode outputs raw stderr/stdout.
// The Raw field contains the command's stderr output to be written directly.
func (h *ErrorHandler) FormatPassthroughError(result domain.ParseResult) domain.ParseResult {
	// In passthrough mode, we just return the result as-is
	// The raw output will be written directly
	return result
}

// FormatParserError formats a parser failure for JSON output mode.
// Returns a ParseResult with Data containing:
//   - error: the parser error message
//   - raw: the original command output that couldn't be parsed
//   - exitCode: the exit code from the command
//
// This is used when the command succeeds but parsing fails.
func (h *ErrorHandler) FormatParserError(result domain.ParseResult) domain.ParseResult {
	errorMsg := h.FormatErrorMessage(result.Error)

	data := map[string]any{
		"error":    errorMsg,
		"raw":      result.Raw,
		"exitCode": result.ExitCode,
	}

	return domain.ParseResult{
		Data:     data,
		Raw:      result.Raw,
		ExitCode: result.ExitCode,
		Error:    result.Error,
	}
}

// FormatErrorMessage extracts a user-friendly error message from an error.
// Returns an empty string if the error is nil.
func (h *ErrorHandler) FormatErrorMessage(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
