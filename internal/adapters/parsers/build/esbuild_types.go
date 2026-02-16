package build

// ESBuildResult represents the structured output of 'esbuild' bundler.
// It captures build success, errors, warnings, output files, and duration.
type ESBuildResult struct {
	// Success indicates whether the build completed without errors.
	Success bool `json:"success"`

	// Errors contains any build errors that occurred.
	Errors []ESBuildError `json:"errors"`

	// Warnings contains any build warnings that occurred.
	Warnings []ESBuildWarning `json:"warnings"`

	// Outputs contains information about generated output files.
	Outputs []ESBuildOutput `json:"outputs"`

	// Duration is the build time in milliseconds.
	Duration float64 `json:"duration"`
}

// ESBuildError represents a single esbuild error.
type ESBuildError struct {
	// File is the source file where the error occurred.
	File string `json:"file"`

	// Line is the line number where the error occurred.
	Line int `json:"line"`

	// Column is the column number where the error occurred.
	Column int `json:"column"`

	// Message is the error message from the bundler.
	Message string `json:"message"`
}

// ESBuildWarning represents a single esbuild warning.
type ESBuildWarning struct {
	// File is the source file where the warning occurred.
	File string `json:"file"`

	// Line is the line number where the warning occurred.
	Line int `json:"line"`

	// Column is the column number where the warning occurred.
	Column int `json:"column"`

	// Message is the warning message from the bundler.
	Message string `json:"message"`
}

// ESBuildOutput represents information about a generated output file.
type ESBuildOutput struct {
	// Path is the path to the output file.
	Path string `json:"path"`

	// Size is the size of the output file in bytes.
	Size int64 `json:"size"`
}
