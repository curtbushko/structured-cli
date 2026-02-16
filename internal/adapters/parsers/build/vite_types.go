package build

// ViteResult represents the structured output of 'vite build' command.
// It captures build success, errors, warnings, output files, duration, and module count.
type ViteResult struct {
	// Success indicates whether the build completed without errors.
	Success bool `json:"success"`

	// Errors contains any build errors that occurred.
	Errors []ViteError `json:"errors"`

	// Warnings contains any build warnings that occurred.
	Warnings []ViteWarning `json:"warnings"`

	// Outputs contains information about generated output files.
	Outputs []ViteOutput `json:"outputs"`

	// Duration is the build time in milliseconds.
	Duration float64 `json:"duration"`

	// Modules is the number of modules transformed during the build.
	Modules int `json:"modules"`
}

// ViteError represents a single Vite build error.
type ViteError struct {
	// Plugin is the Vite/Rollup plugin that produced the error (if applicable).
	Plugin string `json:"plugin,omitempty"`

	// File is the source file where the error occurred (if applicable).
	File string `json:"file,omitempty"`

	// Line is the line number where the error occurred (if applicable).
	Line int `json:"line,omitempty"`

	// Column is the column number where the error occurred (if applicable).
	Column int `json:"column,omitempty"`

	// Message is the error message from the bundler.
	Message string `json:"message"`
}

// ViteWarning represents a single Vite build warning.
type ViteWarning struct {
	// Plugin is the Vite/Rollup plugin that produced the warning (if applicable).
	Plugin string `json:"plugin,omitempty"`

	// File is the source file where the warning occurred (if applicable).
	File string `json:"file,omitempty"`

	// Line is the line number where the warning occurred (if applicable).
	Line int `json:"line,omitempty"`

	// Column is the column number where the warning occurred (if applicable).
	Column int `json:"column,omitempty"`

	// Message is the warning message from the bundler.
	Message string `json:"message"`
}

// ViteOutput represents information about a generated output file.
type ViteOutput struct {
	// Path is the path to the output file.
	Path string `json:"path"`

	// Size is the size of the output file in bytes.
	Size int64 `json:"size"`

	// GzipSize is the gzip-compressed size of the output file in bytes.
	GzipSize int64 `json:"gzipSize"`
}
