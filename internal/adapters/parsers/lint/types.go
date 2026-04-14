// Package lint provides parsers for linter tool output.
// This package is in the adapters layer and implements parsers for
// converting raw linter command output into structured domain types.
package lint

// ESLintResult represents the structured output of 'eslint'.
// It captures lint success and any issues found.
// Deprecated: Use ESLintResultCompact for new code.
type ESLintResult struct {
	// Success indicates whether linting completed without issues.
	Success bool `json:"success"`

	// Issues contains the list of lint issues found.
	Issues []ESLintIssue `json:"issues"`
}

// ESLintResultCompact represents the compact output format for 'eslint'.
// It provides a summarized view optimized for token efficiency.
type ESLintResultCompact struct {
	// TotalIssues is the total count of all lint issues.
	TotalIssues int `json:"total_issues"`

	// FilesWithIssues is the count of files that have at least one issue.
	FilesWithIssues int `json:"files_with_issues"`

	// SeverityCounts maps severity levels to their counts.
	SeverityCounts map[string]int `json:"severity_counts"`

	// Results contains grouped issues by file in compact tuple format.
	Results []FileIssueGroup `json:"results"`

	// Truncated indicates how many issues were omitted due to truncation limits.
	Truncated int `json:"truncated"`
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
// Deprecated: Use PrettierResultCompact for new code.
type PrettierResult struct {
	// Success indicates whether all files are properly formatted.
	Success bool `json:"success"`

	// Unformatted contains paths of files that need formatting.
	Unformatted []string `json:"unformatted"`
}

// PrettierResultCompact represents the compact output format for 'prettier --check'.
// It provides a simple summary since Prettier is binary (formatted vs needs formatting).
type PrettierResultCompact struct {
	// Success indicates whether all files are properly formatted.
	Success bool `json:"success"`

	// TotalChecked is the total number of files checked.
	TotalChecked int `json:"total_checked"`

	// NeedFormatting is the count of files that need formatting.
	NeedFormatting int `json:"need_formatting"`

	// Files contains paths of files that need formatting.
	Files []string `json:"files"`
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
// Deprecated: Use GolangCILintResultCompact for new code.
type GolangCILintResult struct {
	// Success indicates whether linting completed without issues.
	Success bool `json:"success"`

	// Issues contains the list of lint issues found.
	Issues []GolangCILintIssue `json:"issues"`
}

// GolangCILintResultCompact represents the compact output format for 'golangci-lint run'.
// It provides a summarized view optimized for token efficiency.
type GolangCILintResultCompact struct {
	// TotalIssues is the total count of all lint issues.
	TotalIssues int `json:"total_issues"`

	// FilesWithIssues is the count of files that have at least one issue.
	FilesWithIssues int `json:"files_with_issues"`

	// SeverityCounts maps severity levels to their counts.
	SeverityCounts map[string]int `json:"severity_counts"`

	// Results contains grouped issues by file in compact tuple format.
	Results []FileIssueGroup `json:"results"`

	// Truncated indicates how many issues were omitted due to truncation limits.
	Truncated int `json:"truncated"`
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
// Deprecated: Use RuffResultCompact for new code.
type RuffResult struct {
	// Success indicates whether linting completed without issues.
	Success bool `json:"success"`

	// Issues contains the list of lint issues found.
	Issues []RuffIssue `json:"issues"`
}

// RuffResultCompact represents the compact output format for 'ruff check'.
// It provides a summarized view optimized for token efficiency.
type RuffResultCompact struct {
	// TotalIssues is the total count of all lint issues.
	TotalIssues int `json:"total_issues"`

	// FilesWithIssues is the count of files that have at least one issue.
	FilesWithIssues int `json:"files_with_issues"`

	// SeverityCounts maps severity levels to their counts.
	SeverityCounts map[string]int `json:"severity_counts"`

	// Results contains grouped issues by file in compact tuple format.
	Results []FileIssueGroup `json:"results"`

	// Truncated indicates how many issues were omitted due to truncation limits.
	Truncated int `json:"truncated"`
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
// Deprecated: Use MypyResultCompact for new code.
type MypyResult struct {
	// Success indicates whether type checking completed without errors.
	Success bool `json:"success"`

	// Errors contains the list of type errors found.
	Errors []MypyError `json:"errors"`

	// Summary contains the summary line from mypy output.
	Summary string `json:"summary"`
}

// MypyResultCompact represents the compact output format for 'mypy'.
// It provides a summarized view optimized for token efficiency.
type MypyResultCompact struct {
	// TotalIssues is the total count of all type errors.
	TotalIssues int `json:"total_issues"`

	// FilesWithIssues is the count of files that have at least one error.
	FilesWithIssues int `json:"files_with_issues"`

	// SeverityCounts maps severity levels to their counts.
	SeverityCounts map[string]int `json:"severity_counts"`

	// Results contains grouped errors by file in compact tuple format.
	Results []FileIssueGroup `json:"results"`

	// Truncated indicates how many errors were omitted due to truncation limits.
	Truncated int `json:"truncated"`

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

// Severity level constants for linter issues.
const (
	// SeverityError indicates an error-level issue that must be fixed.
	SeverityError = "error"

	// SeverityWarning indicates a warning-level issue that should be addressed.
	SeverityWarning = "warning"

	// SeverityInfo indicates an informational issue.
	SeverityInfo = "info"

	// SeverityStyle indicates a style-related issue.
	SeverityStyle = "style"
)

// Truncation limit constants for compact output format.
const (
	// MaxTotalIssues is the maximum number of total issues to include in compact output.
	MaxTotalIssues = 200

	// MaxIssuesPerFile is the maximum number of issues per file to include in compact output.
	MaxIssuesPerFile = 20
)

// OutputCompact represents a compact format for linter output.
// It provides a summarized view of lint issues optimized for token efficiency.
type OutputCompact struct {
	// TotalIssues is the total count of all lint issues.
	TotalIssues int `json:"total_issues"`

	// FilesWithIssues is the count of files that have at least one issue.
	FilesWithIssues int `json:"files_with_issues"`

	// SeverityCounts maps severity levels to their counts.
	SeverityCounts map[string]int `json:"severity_counts"`

	// Results contains grouped issues by file in compact tuple format.
	Results []FileIssueGroup `json:"results"`

	// Truncated indicates how many issues were omitted due to truncation limits.
	Truncated int `json:"truncated"`
}

// FileIssueGroup represents a group of issues for a single file.
// It is a tuple of [filename string, issue_count int, issues []IssueTuple].
type FileIssueGroup [3]interface{}

// IssueTuple represents a single lint issue in compact tuple format.
// It is a tuple of [line int, severity string, message string, rule_id string].
type IssueTuple [4]interface{}
