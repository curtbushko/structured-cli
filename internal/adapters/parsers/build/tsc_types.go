// Package build provides parsers for build tool output.
// This package is in the adapters layer and implements parsers for
// converting raw build tool command output into structured domain types.
package build

// TSCResult represents the structured output of 'tsc' (TypeScript compiler).
// It captures compilation success and any errors encountered.
type TSCResult struct {
	// Success indicates whether the compilation completed without errors.
	Success bool `json:"success"`

	// Errors contains any compilation errors that occurred.
	Errors []TSCError `json:"errors"`
}

// TSCError represents a single TypeScript compilation error.
type TSCError struct {
	// File is the source file where the error occurred.
	File string `json:"file"`

	// Line is the line number where the error occurred.
	Line int `json:"line"`

	// Column is the column number where the error occurred.
	Column int `json:"column"`

	// Code is the TypeScript error code (e.g., "TS2322").
	Code string `json:"code"`

	// Message is the error message from the compiler.
	Message string `json:"message"`
}
