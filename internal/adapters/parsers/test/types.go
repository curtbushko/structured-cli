// Package test provides parsers for test runner output.
// This package is in the adapters layer and implements parsers for
// converting raw test runner command output into structured domain types.
package test

// Package manager command constants used across test parsers.
const (
	cmdNpx  = "npx"
	cmdYarn = "yarn"
	cmdPnpm = "pnpm"
)

// PytestResult represents the structured output of 'pytest'.
// It captures test results from Python pytest runs.
type PytestResult struct {
	// Passed is the total number of tests that passed.
	Passed int `json:"passed"`

	// Failed is the total number of tests that failed.
	Failed int `json:"failed"`

	// Skipped is the total number of tests that were skipped.
	Skipped int `json:"skipped"`

	// Errors is the total number of tests that had errors.
	Errors int `json:"errors"`

	// Duration is the total time in seconds for the test run.
	Duration float64 `json:"duration"`

	// Tests contains individual test case results.
	Tests []PytestCase `json:"tests"`
}

// PytestCase represents a single pytest test case result.
type PytestCase struct {
	// Name is the name of the test function.
	Name string `json:"name"`

	// File is the test file containing this test.
	File string `json:"file"`

	// Outcome is the test outcome (passed, failed, skipped, error).
	Outcome string `json:"outcome"`

	// Duration is the time in seconds taken to run the test.
	Duration float64 `json:"duration"`

	// Message is the failure/error message if applicable.
	Message string `json:"message,omitempty"`
}

// JestResult represents the structured output of 'jest'.
// It captures test results from JavaScript jest runs.
type JestResult struct {
	// Passed is the total number of tests that passed.
	Passed int `json:"passed"`

	// Failed is the total number of tests that failed.
	Failed int `json:"failed"`

	// Skipped is the total number of tests that were skipped.
	Skipped int `json:"skipped"`

	// Total is the total number of tests.
	Total int `json:"total"`

	// Duration is the total time in seconds for the test run.
	Duration float64 `json:"duration"`

	// Suites contains test suite results.
	Suites []JestSuite `json:"suites"`
}

// JestSuite represents a single jest test suite.
type JestSuite struct {
	// Name is the name of the test suite (usually the file).
	Name string `json:"name"`

	// Passed indicates whether all tests in the suite passed.
	Passed bool `json:"passed"`

	// Tests contains individual test case results.
	Tests []JestCase `json:"tests"`
}

// JestCase represents a single jest test case result.
type JestCase struct {
	// Name is the name of the test.
	Name string `json:"name"`

	// Status is the test status (passed, failed, skipped).
	Status string `json:"status"`

	// Duration is the time in milliseconds taken to run the test.
	Duration float64 `json:"duration"`

	// Message is the failure message if applicable.
	Message string `json:"message,omitempty"`
}

// VitestResult represents the structured output of 'vitest'.
// It captures test results from vitest runs.
type VitestResult struct {
	// Passed is the total number of tests that passed.
	Passed int `json:"passed"`

	// Failed is the total number of tests that failed.
	Failed int `json:"failed"`

	// Skipped is the total number of tests that were skipped.
	Skipped int `json:"skipped"`

	// Duration is the total time in milliseconds for the test run.
	Duration float64 `json:"duration"`

	// Files contains test file results.
	Files []VitestFile `json:"files"`
}

// VitestFile represents test results for a single file.
type VitestFile struct {
	// Name is the name of the test file.
	Name string `json:"name"`

	// Passed indicates whether all tests in the file passed.
	Passed bool `json:"passed"`

	// Tests contains individual test case results.
	Tests []VitestCase `json:"tests"`
}

// VitestCase represents a single vitest test case result.
type VitestCase struct {
	// Name is the name of the test.
	Name string `json:"name"`

	// Status is the test status (pass, fail, skip).
	Status string `json:"status"`

	// Duration is the time in milliseconds taken to run the test.
	Duration float64 `json:"duration"`

	// Message is the failure message if applicable.
	Message string `json:"message,omitempty"`
}

// MochaResult represents the structured output of 'mocha'.
// It captures test results from mocha runs.
type MochaResult struct {
	// Passed is the total number of tests that passed.
	Passed int `json:"passed"`

	// Failed is the total number of tests that failed.
	Failed int `json:"failed"`

	// Pending is the total number of pending tests.
	Pending int `json:"pending"`

	// Duration is the total time in milliseconds for the test run.
	Duration float64 `json:"duration"`

	// Suites contains test suite results.
	Suites []MochaSuite `json:"suites"`
}

// MochaSuite represents a single mocha test suite.
type MochaSuite struct {
	// Title is the title of the test suite.
	Title string `json:"title"`

	// Passed indicates whether all tests in the suite passed.
	Passed bool `json:"passed"`

	// Tests contains individual test case results.
	Tests []MochaCase `json:"tests"`
}

// MochaCase represents a single mocha test case result.
type MochaCase struct {
	// Title is the title of the test.
	Title string `json:"title"`

	// State is the test state (passed, failed, pending).
	State string `json:"state"`

	// Duration is the time in milliseconds taken to run the test.
	Duration float64 `json:"duration"`

	// Error is the error message if the test failed.
	Error string `json:"error,omitempty"`
}

// CargoTestResult represents the structured output of 'cargo test'.
// It captures test results from Rust cargo test runs.
type CargoTestResult struct {
	// Passed is the total number of tests that passed.
	Passed int `json:"passed"`

	// Failed is the total number of tests that failed.
	Failed int `json:"failed"`

	// Ignored is the total number of tests that were ignored.
	Ignored int `json:"ignored"`

	// Measured is the number of benchmark tests measured.
	Measured int `json:"measured"`

	// Filtered is the number of tests filtered out.
	Filtered int `json:"filtered"`

	// Duration is the total time in seconds for the test run.
	Duration float64 `json:"duration"`

	// Tests contains individual test case results.
	Tests []CargoTestCase `json:"tests"`
}

// CargoTestCase represents a single cargo test case result.
type CargoTestCase struct {
	// Name is the full name of the test (module::test_name).
	Name string `json:"name"`

	// Status is the test status (ok, FAILED, ignored).
	Status string `json:"status"`

	// Duration is the time in seconds taken to run the test.
	Duration float64 `json:"duration"`

	// Message is the failure message if applicable.
	Message string `json:"message,omitempty"`
}
