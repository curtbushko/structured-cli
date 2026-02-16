// Package golang provides parsers for Go command output.
// This package is in the adapters layer and implements parsers for
// converting raw Go toolchain command output into structured domain types.
package golang

// Build represents the structured output of 'go build'.
// It captures build success, package list, and any compilation errors.
type Build struct {
	// Success indicates whether the build completed without errors.
	Success bool `json:"success"`

	// Packages contains the list of packages that were built.
	Packages []string `json:"packages"`

	// Errors contains any compilation errors that occurred.
	Errors []BuildError `json:"errors"`
}

// BuildError represents a single compilation error from go build.
type BuildError struct {
	// File is the source file where the error occurred.
	File string `json:"file"`

	// Line is the line number where the error occurred.
	Line int `json:"line"`

	// Column is the column number where the error occurred.
	Column int `json:"column"`

	// Message is the error message from the compiler.
	Message string `json:"message"`
}

// TestResult represents the structured output of 'go test'.
// It aggregates test results across all tested packages.
type TestResult struct {
	// Passed is the total number of tests that passed.
	Passed int `json:"passed"`

	// Failed is the total number of tests that failed.
	Failed int `json:"failed"`

	// Skipped is the total number of tests that were skipped.
	Skipped int `json:"skipped"`

	// Packages contains per-package test results.
	Packages []TestPackage `json:"packages"`

	// Coverage contains code coverage information when -cover flag is used.
	// This is nil when coverage is not collected.
	Coverage *Coverage `json:"coverage,omitempty"`
}

// TestPackage represents test results for a single package.
type TestPackage struct {
	// Package is the import path of the tested package.
	Package string `json:"package"`

	// Passed indicates whether all tests in the package passed.
	Passed bool `json:"passed"`

	// Elapsed is the time in seconds taken to run the package tests.
	Elapsed float64 `json:"elapsed"`

	// Tests contains individual test case results.
	Tests []TestCase `json:"tests"`
}

// TestCase represents a single test case result.
type TestCase struct {
	// Name is the name of the test function.
	Name string `json:"name"`

	// Package is the import path of the package containing the test.
	Package string `json:"package"`

	// Passed indicates whether the test passed.
	Passed bool `json:"passed"`

	// Duration is the time in seconds taken to run the test.
	Duration float64 `json:"duration"`

	// Output is the test output (stdout/stderr combined).
	Output string `json:"output"`
}

// Coverage represents the structured output of 'go test -cover'.
// It contains coverage percentages for the tested packages.
type Coverage struct {
	// Total is the overall coverage percentage across all packages.
	Total float64 `json:"total"`

	// Packages contains per-package coverage information.
	Packages []PackageCoverage `json:"packages"`
}

// PackageCoverage represents coverage information for a single package.
type PackageCoverage struct {
	// Package is the import path of the package.
	Package string `json:"package"`

	// Coverage is the coverage percentage for this package.
	Coverage float64 `json:"coverage"`
}

// VetResult represents the structured output of 'go vet'.
// It contains any issues found by the static analyzer.
type VetResult struct {
	// Issues contains the list of vet issues found.
	Issues []VetIssue `json:"issues"`
}

// VetIssue represents a single issue found by go vet.
type VetIssue struct {
	// File is the source file where the issue was found.
	File string `json:"file"`

	// Line is the line number where the issue was found.
	Line int `json:"line"`

	// Column is the column number where the issue was found.
	Column int `json:"column"`

	// Message is the description of the issue.
	Message string `json:"message"`
}

// RunResult represents the structured output of 'go run'.
// It captures the exit code and standard output/error streams.
type RunResult struct {
	// ExitCode is the exit code of the executed program.
	ExitCode int `json:"exitCode"`

	// Stdout is the standard output of the program.
	Stdout string `json:"stdout"`

	// Stderr is the standard error output of the program.
	Stderr string `json:"stderr"`
}

// ModTidyResult represents the structured output of 'go mod tidy'.
// It shows which dependencies were added or removed.
type ModTidyResult struct {
	// Added contains dependencies that were added.
	Added []string `json:"added"`

	// Removed contains dependencies that were removed.
	Removed []string `json:"removed"`
}

// FmtResult represents the structured output of 'go fmt' or 'gofmt'.
// It lists files that were not properly formatted.
type FmtResult struct {
	// Unformatted contains paths of files that need formatting.
	Unformatted []string `json:"unformatted"`
}

// GenerateResult represents the structured output of 'go generate'.
// It indicates success and lists generated files.
type GenerateResult struct {
	// Success indicates whether generation completed without errors.
	Success bool `json:"success"`

	// Generated contains paths of files that were generated.
	Generated []string `json:"generated"`
}
