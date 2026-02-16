// Package lint provides parsers for linter tool output.
// This package is in the adapters layer and implements parsers for
// converting raw linter command output into structured domain types.
package lint

// ESLintResult represents the structured output of 'eslint'.
// It captures lint success and any issues found.
type ESLintResult struct {
	// Success indicates whether linting completed without issues.
	Success bool `json:"success"`

	// Issues contains the list of lint issues found.
	Issues []ESLintIssue `json:"issues"`
}

// ESLintIssue represents a single ESLint issue.
type ESLintIssue struct {
	// File is the source file where the issue was found.
	File string `json:"file"`

	// Line is the line number where the issue was found.
	Line int `json:"line"`

	// Column is the column number where the issue was found.
	Column int `json:"column"`

	// Severity is the severity level (error or warning).
	Severity string `json:"severity"`

	// Message is the description of the issue.
	Message string `json:"message"`

	// Rule is the ESLint rule that was violated.
	Rule string `json:"rule"`
}

// PrettierResult represents the structured output of 'prettier --check'.
// It lists files that are not properly formatted.
type PrettierResult struct {
	// Success indicates whether all files are properly formatted.
	Success bool `json:"success"`

	// Unformatted contains paths of files that need formatting.
	Unformatted []string `json:"unformatted"`
}

// BiomeResult represents the structured output of 'biome check'.
// It captures lint and format issues found.
type BiomeResult struct {
	// Success indicates whether the check completed without issues.
	Success bool `json:"success"`

	// Issues contains the list of issues found.
	Issues []BiomeIssue `json:"issues"`
}

// BiomeIssue represents a single Biome issue.
type BiomeIssue struct {
	// File is the source file where the issue was found.
	File string `json:"file"`

	// Line is the line number where the issue was found.
	Line int `json:"line"`

	// Column is the column number where the issue was found.
	Column int `json:"column"`

	// Severity is the severity level (error, warning, info).
	Severity string `json:"severity"`

	// Category is the issue category (lint, format, parse).
	Category string `json:"category"`

	// Message is the description of the issue.
	Message string `json:"message"`

	// Rule is the Biome rule that was violated.
	Rule string `json:"rule"`
}

// GolangCILintResult represents the structured output of 'golangci-lint run'.
// It captures lint success and any issues found.
type GolangCILintResult struct {
	// Success indicates whether linting completed without issues.
	Success bool `json:"success"`

	// Issues contains the list of lint issues found.
	Issues []GolangCILintIssue `json:"issues"`
}

// GolangCILintIssue represents a single golangci-lint issue.
type GolangCILintIssue struct {
	// File is the source file where the issue was found.
	File string `json:"file"`

	// Line is the line number where the issue was found.
	Line int `json:"line"`

	// Column is the column number where the issue was found.
	Column int `json:"column"`

	// Message is the description of the issue.
	Message string `json:"message"`

	// Linter is the linter that found the issue (e.g., "errcheck", "govet").
	Linter string `json:"linter"`

	// Severity is the severity level.
	Severity string `json:"severity"`
}

// RuffResult represents the structured output of 'ruff check'.
// It captures lint success and any issues found.
type RuffResult struct {
	// Success indicates whether linting completed without issues.
	Success bool `json:"success"`

	// Issues contains the list of lint issues found.
	Issues []RuffIssue `json:"issues"`
}

// RuffIssue represents a single Ruff issue.
type RuffIssue struct {
	// File is the source file where the issue was found.
	File string `json:"file"`

	// Line is the line number where the issue was found.
	Line int `json:"line"`

	// Column is the column number where the issue was found.
	Column int `json:"column"`

	// Code is the Ruff rule code (e.g., "E501", "F401").
	Code string `json:"code"`

	// Message is the description of the issue.
	Message string `json:"message"`
}

// MypyResult represents the structured output of 'mypy'.
// It captures type check success and any errors found.
type MypyResult struct {
	// Success indicates whether type checking completed without errors.
	Success bool `json:"success"`

	// Errors contains the list of type errors found.
	Errors []MypyError `json:"errors"`

	// Summary contains the summary line from mypy output.
	Summary string `json:"summary"`
}

// MypyError represents a single mypy type error.
type MypyError struct {
	// File is the source file where the error was found.
	File string `json:"file"`

	// Line is the line number where the error was found.
	Line int `json:"line"`

	// Severity is the severity level (error, warning, note).
	Severity string `json:"severity"`

	// Message is the description of the error.
	Message string `json:"message"`

	// Code is the optional mypy error code (e.g., "arg-type", "return-value").
	Code string `json:"code,omitempty"`
}
