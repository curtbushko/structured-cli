// Package domain contains the core domain types for structured-cli.
package domain

// SuccessFilterConfig holds configuration for success filtering.
// It controls whether filtering of passed/successful items is enabled.
type SuccessFilterConfig struct {
	// Enabled indicates whether success filtering is active.
	Enabled bool
}

// NewSuccessFilterConfig creates a new SuccessFilterConfig with default values.
// By default, success filtering is enabled.
func NewSuccessFilterConfig() SuccessFilterConfig {
	return SuccessFilterConfig{
		Enabled: true,
	}
}

// SuccessFilterResult represents the result of a success filtering operation.
// It contains counts of items categorized by their status.
type SuccessFilterResult struct {
	// Total is the total number of items before filtering.
	Total int `json:"total"`

	// Passed is the number of items that passed/succeeded.
	Passed int `json:"passed"`

	// Failed is the number of items that failed.
	Failed int `json:"failed"`

	// Skipped is the number of items that were skipped.
	Skipped int `json:"skipped"`

	// Removed is the number of items removed by the filter.
	Removed int `json:"removed"`

	// Kept is the number of items kept after filtering.
	Kept int `json:"kept"`
}

// NewSuccessFilterResult creates a new SuccessFilterResult with the given values.
func NewSuccessFilterResult(total, passed, failed, skipped, removed, kept int) SuccessFilterResult {
	return SuccessFilterResult{
		Total:   total,
		Passed:  passed,
		Failed:  failed,
		Skipped: skipped,
		Removed: removed,
		Kept:    kept,
	}
}

// FilterRule defines how to detect pass/fail status for a specific output type.
// It maps status field names to their pass and fail values.
type FilterRule struct {
	// StatusField is the name of the field to check (e.g., "status", "outcome", "state").
	StatusField string

	// PassValues are values that indicate a passing/successful status.
	PassValues []string

	// FailValues are values that indicate a failing status.
	FailValues []string
}

// NewFilterRule creates a new FilterRule with the given configuration.
func NewFilterRule(statusField string, passValues, failValues []string) FilterRule {
	return FilterRule{
		StatusField: statusField,
		PassValues:  passValues,
		FailValues:  failValues,
	}
}

// MatchesPass returns true if the given value matches any pass value.
func (r FilterRule) MatchesPass(value string) bool {
	for _, v := range r.PassValues {
		if v == value {
			return true
		}
	}
	return false
}

// MatchesFail returns true if the given value matches any fail value.
func (r FilterRule) MatchesFail(value string) bool {
	for _, v := range r.FailValues {
		if v == value {
			return true
		}
	}
	return false
}
