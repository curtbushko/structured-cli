// Package cargo provides parsers for Cargo (Rust) command output.
// This package is in the adapters layer and implements parsers for
// converting raw cargo command output into structured domain types.
package cargo

// ClippyResult represents the structured output of 'cargo clippy --message-format=json'.
// It captures lint warnings and errors from Clippy, the Rust linter.
// Deprecated: Use ClippyResultCompact for new code.
type ClippyResult struct {
	// Success indicates whether clippy completed without errors.
	Success bool `json:"success"`

	// Warnings contains clippy lint warnings.
	Warnings []ClippyDiagnostic `json:"warnings"`

	// Errors contains clippy lint errors.
	Errors []ClippyDiagnostic `json:"errors"`
}

// ClippyResultCompact represents the compact output format for 'cargo clippy'.
// It provides a summarized view optimized for token efficiency.
type ClippyResultCompact struct {
	// TotalIssues is the total count of all lint issues.
	TotalIssues int `json:"total_issues"`

	// FilesWithIssues is the count of files that have at least one issue.
	FilesWithIssues int `json:"files_with_issues"`

	// SeverityCounts maps severity levels to their counts.
	SeverityCounts map[string]int `json:"severity_counts"`

	// Results contains grouped issues by file in compact tuple format.
	// Each ClippyFileIssueGroup is [filename, issue_count, issues].
	// Each issue tuple is [line, severity, message, rule_id].
	Results []ClippyFileIssueGroup `json:"results"`

	// Truncated indicates how many issues were omitted due to truncation limits.
	Truncated int `json:"truncated"`
}

// ClippyFileIssueGroup represents a group of issues for a single file.
// It is a tuple of [filename string, issue_count int, issues []ClippyIssueTuple].
type ClippyFileIssueGroup [3]interface{}

// ClippyIssueTuple represents a single lint issue in compact tuple format.
// It is a tuple of [line int, severity string, message string, rule_id string].
type ClippyIssueTuple [4]interface{}

// ClippyDiagnostic represents a single clippy diagnostic message.
type ClippyDiagnostic struct {
	// Message is the diagnostic message from clippy.
	Message string `json:"message"`

	// Code is the lint code (e.g., clippy::unwrap_used).
	Code string `json:"code"`

	// Level is the diagnostic level (warning, error).
	Level string `json:"level"`

	// File is the source file where the diagnostic occurred.
	File string `json:"file"`

	// Line is the line number where the diagnostic occurred.
	Line int `json:"line"`

	// Column is the column number where the diagnostic occurred.
	Column int `json:"column"`

	// Rendered is the human-readable rendered message.
	Rendered string `json:"rendered"`
}

// RunResult represents the structured output of 'cargo run'.
// It captures build and execution information.
type RunResult struct {
	// Success indicates whether the build and run succeeded.
	Success bool `json:"success"`

	// BuildSuccess indicates whether the build phase succeeded.
	BuildSuccess bool `json:"build_success"`

	// Executable is the path to the executed binary.
	Executable string `json:"executable"`

	// Errors contains any build errors.
	Errors []RunError `json:"errors"`

	// Output contains the stdout from the executed program.
	Output string `json:"output"`
}

// RunError represents a build error during cargo run.
type RunError struct {
	// Message is the error message.
	Message string `json:"message"`

	// Code is the error code.
	Code string `json:"code"`

	// File is the source file where the error occurred.
	File string `json:"file"`

	// Line is the line number where the error occurred.
	Line int `json:"line"`

	// Column is the column number where the error occurred.
	Column int `json:"column"`
}

// AddResult represents the structured output of 'cargo add'.
// It captures information about added dependencies.
type AddResult struct {
	// Success indicates whether the add operation succeeded.
	Success bool `json:"success"`

	// Dependencies contains the list of added dependencies.
	Dependencies []AddedDependency `json:"dependencies"`

	// Errors contains any error messages.
	Errors []string `json:"errors"`
}

// AddedDependency represents a dependency that was added.
type AddedDependency struct {
	// Name is the crate name.
	Name string `json:"name"`

	// Version is the version that was added.
	Version string `json:"version"`

	// Features contains the enabled features.
	Features []string `json:"features"`

	// Dev indicates if this is a dev dependency.
	Dev bool `json:"dev"`

	// Build indicates if this is a build dependency.
	Build bool `json:"build"`
}

// RemoveResult represents the structured output of 'cargo remove'.
// It captures information about removed dependencies.
type RemoveResult struct {
	// Success indicates whether the remove operation succeeded.
	Success bool `json:"success"`

	// Removed contains the names of removed dependencies.
	Removed []string `json:"removed"`

	// Errors contains any error messages.
	Errors []string `json:"errors"`
}

// FmtResult represents the structured output of 'cargo fmt --check'.
// It captures formatting issues found by rustfmt.
type FmtResult struct {
	// Success indicates whether all files are correctly formatted.
	Success bool `json:"success"`

	// Files contains files with formatting issues.
	Files []FmtFile `json:"files"`
}

// FmtFile represents a file with formatting issues.
type FmtFile struct {
	// Path is the path to the file.
	Path string `json:"path"`

	// Diff contains the formatting diff if available.
	Diff string `json:"diff,omitempty"`
}

// DocResult represents the structured output of 'cargo doc --message-format=json'.
// It captures documentation generation information.
type DocResult struct {
	// Success indicates whether doc generation succeeded.
	Success bool `json:"success"`

	// Warnings contains any documentation warnings.
	Warnings []DocWarning `json:"warnings"`

	// Errors contains any documentation errors.
	Errors []DocError `json:"errors"`

	// GeneratedDocs contains paths to generated documentation.
	GeneratedDocs []string `json:"generated_docs"`
}

// DocWarning represents a documentation warning.
type DocWarning struct {
	// Message is the warning message.
	Message string `json:"message"`

	// File is the source file.
	File string `json:"file"`

	// Line is the line number.
	Line int `json:"line"`
}

// DocError represents a documentation error.
type DocError struct {
	// Message is the error message.
	Message string `json:"message"`

	// Code is the error code.
	Code string `json:"code"`

	// File is the source file.
	File string `json:"file"`

	// Line is the line number.
	Line int `json:"line"`
}

// CheckResult represents the structured output of 'cargo check --message-format=json'.
// It captures type checking results without producing artifacts.
type CheckResult struct {
	// Success indicates whether the check succeeded.
	Success bool `json:"success"`

	// Errors contains compilation errors.
	Errors []CheckError `json:"errors"`

	// Warnings contains compilation warnings.
	Warnings []CheckWarning `json:"warnings"`
}

// CheckError represents a type check error.
type CheckError struct {
	// Message is the error message.
	Message string `json:"message"`

	// Code is the error code.
	Code string `json:"code"`

	// File is the source file.
	File string `json:"file"`

	// Line is the line number.
	Line int `json:"line"`

	// Column is the column number.
	Column int `json:"column"`

	// Rendered is the human-readable rendered message.
	Rendered string `json:"rendered"`
}

// CheckWarning represents a type check warning.
type CheckWarning struct {
	// Message is the warning message.
	Message string `json:"message"`

	// Code is the warning code.
	Code string `json:"code"`

	// File is the source file.
	File string `json:"file"`

	// Line is the line number.
	Line int `json:"line"`

	// Column is the column number.
	Column int `json:"column"`

	// Rendered is the human-readable rendered message.
	Rendered string `json:"rendered"`
}

// Common constants for cargo command parsing.
const (
	// cmdCargo is the cargo command name.
	cmdCargo = "cargo"

	// reasonBuildFinished indicates build completion in cargo JSON output.
	reasonBuildFinished = "build-finished"

	// reasonCompilerMessage indicates a compiler diagnostic in cargo JSON output.
	reasonCompilerMessage = "compiler-message"

	// levelError indicates an error diagnostic level.
	levelError = "error"

	// levelErrorICE indicates an internal compiler error diagnostic level.
	levelErrorICE = "error: internal compiler error"

	// levelWarning indicates a warning diagnostic level.
	levelWarning = "warning"
)
