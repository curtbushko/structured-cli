// Package domain contains the core domain types for structured-cli.
package domain

// MinTokenThreshold is the minimum number of tokens below which output is
// considered "small" and may be processed differently (e.g., simplified).
const MinTokenThreshold = 25

// SmallOutputConfig holds configuration for small output filtering.
// It controls whether small output filtering is enabled and the threshold
// for determining what constitutes "small" output.
type SmallOutputConfig struct {
	// Enabled indicates whether small output filtering is active.
	Enabled bool

	// TokenThreshold is the number of tokens below which output is
	// considered small. Output with fewer tokens than this threshold
	// may be simplified or processed differently.
	TokenThreshold int
}

// NewSmallOutputConfig creates a new SmallOutputConfig with default values.
// By default, filtering is enabled and the threshold is set to MinTokenThreshold.
func NewSmallOutputConfig() SmallOutputConfig {
	return SmallOutputConfig{
		Enabled:        true,
		TokenThreshold: MinTokenThreshold,
	}
}

// SmallOutputResult represents a simplified result for small output.
// It contains a status indicator and a human-readable summary.
type SmallOutputResult struct {
	// Status indicates the state of the command result (e.g., "clean", "modified").
	Status string `json:"status"`

	// Summary provides a brief description of the result.
	Summary string `json:"summary"`
}

// NewSmallOutputResult creates a new SmallOutputResult with the given status
// and summary.
func NewSmallOutputResult(status, summary string) SmallOutputResult {
	return SmallOutputResult{
		Status:  status,
		Summary: summary,
	}
}
