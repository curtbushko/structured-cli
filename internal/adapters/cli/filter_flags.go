// Package cli provides the CLI adapter for structured-cli.
// This is an inbound adapter that handles user input via the command line.
package cli

import (
	"strings"
)

const (
	// DisableFilterFlag is the flag used to disable specific output filters.
	DisableFilterFlag = "--disable-filter="

	// EnvDisableFilterKey is the environment variable that controls filter disabling.
	EnvDisableFilterKey = "STRUCTURED_CLI_DISABLE_FILTER"

	// FilterNameSmall is the name of the small output filter.
	FilterNameSmall = "small"

	// FilterNameSuccess is the name of the success filter.
	FilterNameSuccess = "success"

	// FilterNameDedupe is the name of the dedupe filter.
	FilterNameDedupe = "dedupe"

	// FilterNameAll disables all filters.
	FilterNameAll = "all"
)

// ExtractDisableFilter scans args for the --disable-filter=name flag, removes it,
// and returns the filter names along with the remaining arguments.
//
// The flag value can be comma-separated to disable multiple filters:
//
//	--disable-filter=small,success,dedupe
//
// Example:
//
//	ExtractDisableFilter([]string{"git", "--disable-filter=small", "status"})
//	// Returns: []string{"small"}, []string{"git", "status"}
func ExtractDisableFilter(args []string) (filters []string, remaining []string) {
	remaining = make([]string, 0, len(args))

	for _, arg := range args {
		if strings.HasPrefix(arg, DisableFilterFlag) {
			// Extract filter names after the equals sign
			value := strings.TrimPrefix(arg, DisableFilterFlag)
			if value != "" {
				filters = strings.Split(value, ",")
			}
		} else {
			remaining = append(remaining, arg)
		}
	}

	return filters, remaining
}

// ShouldDisableFilter determines whether a specific filter should be disabled
// based on the --disable-filter flag values and STRUCTURED_CLI_DISABLE_FILTER
// environment variable.
//
// Parameters:
//   - filterName: the name of the filter to check (e.g., "small", "success", "dedupe")
//   - filters: the filter names from the --disable-filter flag
//   - envValue: the value of the STRUCTURED_CLI_DISABLE_FILTER environment variable
//
// Returns true if the filter should be disabled.
//
// The "all" filter name disables all filters when present in either the flag
// or environment variable.
func ShouldDisableFilter(filterName string, filters []string, envValue string) bool {
	// Check flag values
	for _, f := range filters {
		if f == filterName || f == FilterNameAll {
			return true
		}
	}

	// Check environment variable
	if envValue == "" {
		return false
	}

	envFilters := strings.Split(envValue, ",")
	for _, f := range envFilters {
		if f == filterName || f == FilterNameAll {
			return true
		}
	}

	return false
}
