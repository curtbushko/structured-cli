package build

// WebpackResult represents the structured output of 'webpack' bundler.
// It captures build success, errors, warnings, assets, chunks, modules, and duration.
type WebpackResult struct {
	// Success indicates whether the build completed without errors.
	Success bool `json:"success"`

	// Errors contains any build errors that occurred.
	Errors []WebpackError `json:"errors"`

	// Warnings contains any build warnings that occurred.
	Warnings []WebpackWarning `json:"warnings"`

	// Assets contains information about generated output files.
	Assets []WebpackAsset `json:"assets"`

	// Chunks contains information about webpack chunks.
	Chunks []WebpackChunk `json:"chunks"`

	// Modules is the number of modules processed during the build.
	Modules int `json:"modules"`

	// Duration is the build time in milliseconds.
	Duration float64 `json:"duration"`
}

// WebpackError represents a single webpack build error.
type WebpackError struct {
	// File is the source file where the error occurred.
	File string `json:"file,omitempty"`

	// Line is the line number where the error occurred.
	Line int `json:"line,omitempty"`

	// Column is the column number where the error occurred.
	Column int `json:"column,omitempty"`

	// Message is the error message from the bundler.
	Message string `json:"message"`
}

// WebpackWarning represents a single webpack build warning.
type WebpackWarning struct {
	// File is the source file where the warning occurred.
	File string `json:"file,omitempty"`

	// Line is the line number where the warning occurred.
	Line int `json:"line,omitempty"`

	// Column is the column number where the warning occurred.
	Column int `json:"column,omitempty"`

	// Message is the warning message from the bundler.
	Message string `json:"message"`
}

// WebpackAsset represents information about a generated output file.
type WebpackAsset struct {
	// Name is the name of the output file.
	Name string `json:"name"`

	// Size is the size of the output file in bytes.
	Size int64 `json:"size"`

	// Emitted indicates whether the file was actually written to disk.
	Emitted bool `json:"emitted"`

	// ChunkNames contains the names of chunks that contributed to this asset.
	ChunkNames []string `json:"chunkNames,omitempty"`
}

// WebpackChunk represents information about a webpack chunk.
type WebpackChunk struct {
	// Name is the name of the chunk file.
	Name string `json:"name"`

	// ChunkName is the logical name of the chunk (e.g., "main", "vendor").
	ChunkName string `json:"chunkName,omitempty"`

	// Size is the size of the chunk in bytes.
	Size int64 `json:"size"`

	// Entry indicates whether this is an entry chunk.
	Entry bool `json:"entry"`

	// Initial indicates whether this is an initial chunk.
	Initial bool `json:"initial"`

	// Rendered indicates whether the chunk was rendered.
	Rendered bool `json:"rendered"`
}
