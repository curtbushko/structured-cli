// Package application contains the business logic and use cases for structured-cli.
// This layer orchestrates domain logic and depends on ports (interfaces) - never adapters.
package application

import (
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Status constants for filter results.
const (
	statusPassed  = "passed"
	statusFailed  = "failed"
	statusSkipped = "skipped"
	statusUnknown = "unknown"
)

// knownTestCommands maps base command to subcommands that indicate test execution.
var knownTestCommands = map[string][]string{
	"npm":    {"test", "run test"},
	"npx":    {"jest", "vitest", "mocha"},
	"go":     {"test"},
	"cargo":  {"test"},
	"pytest": {""}, // pytest itself is the command
	"jest":   {""},
	"vitest": {""},
	"mocha":  {""},
}

// knownLintCommands maps base command to subcommands that indicate lint execution.
var knownLintCommands = map[string][]string{
	"eslint":        {""},
	"golangci-lint": {"run"},
	"ruff":          {"check"},
	"mypy":          {""},
	"tsc":           {""},
	"prettier":      {"--check"},
	"biome":         {"lint", "check"},
}

// statusFieldRules defines which fields to check for pass/fail status.
// Key is the field name, value contains pass and fail values.
var statusFieldRules = []struct {
	field      string
	passValues []string
	failValues []string
	skipValues []string
}{
	// Jest, cargo test, vitest
	{field: "status", passValues: []string{"passed", "ok", "pass"}, failValues: []string{"failed", "FAILED", "error", "fail"}, skipValues: []string{"skipped", "pending", "skip", "ignored"}},
	// Pytest
	{field: "outcome", passValues: []string{"passed"}, failValues: []string{"failed", "error"}, skipValues: []string{"skipped"}},
	// Go test
	{field: "action", passValues: []string{"pass"}, failValues: []string{"fail"}, skipValues: []string{"skip"}},
	// Vitest, mocha
	{field: "state", passValues: []string{"pass", "passed"}, failValues: []string{"fail", "failed"}, skipValues: []string{"skipped", "pending"}},
	// ESLint, golangci-lint severity
	{field: "severity", passValues: []string{"warning"}, failValues: []string{"error"}, skipValues: []string{}},
}

// arrayFields are the common field names that contain test/lint results.
var arrayFields = []string{"tests", "results", "issues", "errors", "messages", "items", "suites", "files"}

// SuccessFilterer removes passing/successful items from parsed output.
// It implements the ports.SuccessFilter interface.
type SuccessFilterer struct {
	config domain.SuccessFilterConfig
}

// NewSuccessFilterer creates a new SuccessFilterer with default configuration.
// The filterer is enabled by default.
func NewSuccessFilterer() *SuccessFilterer {
	return &SuccessFilterer{
		config: domain.NewSuccessFilterConfig(),
	}
}

// NewSuccessFiltererWithConfig creates a new SuccessFilterer with custom configuration.
func NewSuccessFiltererWithConfig(config domain.SuccessFilterConfig) *SuccessFilterer {
	return &SuccessFilterer{
		config: config,
	}
}

// ShouldFilter returns true if this filter applies to the given command.
// It checks if the command is a known test runner or linter.
func (f *SuccessFilterer) ShouldFilter(cmd string, subcmds []string) bool {
	if !f.config.Enabled {
		return false
	}

	// Check test commands
	if f.matchesCommand(cmd, subcmds, knownTestCommands) {
		return true
	}

	// Check lint commands
	return f.matchesCommand(cmd, subcmds, knownLintCommands)
}

// matchesCommand checks if cmd+subcmds match any entry in the command map.
func (f *SuccessFilterer) matchesCommand(cmd string, subcmds []string, commands map[string][]string) bool {
	validSubcmds, exists := commands[cmd]
	if !exists {
		return false
	}

	// If command has empty string as valid subcmd, it matches with any subcmds
	for _, valid := range validSubcmds {
		if valid == "" {
			return true
		}
	}

	// Build the subcmd string for matching
	subcmdStr := strings.Join(subcmds, " ")
	for _, valid := range validSubcmds {
		if subcmdStr == valid || strings.HasPrefix(subcmdStr, valid+" ") || strings.HasPrefix(subcmdStr, valid) {
			return true
		}
	}

	return false
}

// Filter removes passing items from data, returning filtered data and statistics.
func (f *SuccessFilterer) Filter(data any) (any, domain.SuccessFilterResult) {
	// Handle disabled filter
	if !f.config.Enabled {
		return data, domain.NewSuccessFilterResult(0, 0, 0, 0, 0, 0)
	}

	// Handle nil data
	if data == nil {
		return nil, domain.NewSuccessFilterResult(0, 0, 0, 0, 0, 0)
	}

	// Handle top-level array
	if arr, ok := data.([]any); ok {
		return f.filterArray(arr)
	}

	// Handle map with nested arrays
	if m, ok := data.(map[string]any); ok {
		return f.filterMap(m)
	}

	// Pass through other data unchanged
	return data, domain.NewSuccessFilterResult(0, 0, 0, 0, 0, 0)
}

// filterArray filters a top-level array of test/lint results.
func (f *SuccessFilterer) filterArray(arr []any) ([]any, domain.SuccessFilterResult) {
	if len(arr) == 0 {
		return arr, domain.NewSuccessFilterResult(0, 0, 0, 0, 0, 0)
	}

	var filtered []any
	var passed, failed, skipped int

	for _, item := range arr {
		itemMap, ok := item.(map[string]any)
		if !ok {
			filtered = append(filtered, item)
			continue
		}

		status := f.getItemStatus(itemMap)
		switch status {
		case statusPassed:
			passed++
		case statusFailed:
			failed++
			filtered = append(filtered, item)
		case statusSkipped:
			skipped++
			filtered = append(filtered, item)
		default:
			// Unknown status - keep the item
			filtered = append(filtered, item)
		}
	}

	total := len(arr)
	removed := passed
	kept := len(filtered)

	return filtered, domain.NewSuccessFilterResult(total, passed, failed, skipped, removed, kept)
}

// filterMap processes a map looking for arrays to filter.
func (f *SuccessFilterer) filterMap(m map[string]any) (map[string]any, domain.SuccessFilterResult) {
	result := make(map[string]any)
	var totalStats domain.SuccessFilterResult

	for k, v := range m {
		switch val := v.(type) {
		case []any:
			if f.isFilterableArray(k) {
				// Filter array items and recursively process nested maps
				filtered, stats := f.filterArrayWithNesting(val)
				result[k] = filtered
				totalStats = f.mergeStats(totalStats, stats)
			} else {
				result[k] = v
			}
		case map[string]any:
			// Recursively filter nested maps
			filtered, stats := f.filterMap(val)
			result[k] = filtered
			totalStats = f.mergeStats(totalStats, stats)
		default:
			result[k] = v
		}
	}

	return result, totalStats
}

// filterArrayWithNesting filters array items and recursively processes nested structures.
func (f *SuccessFilterer) filterArrayWithNesting(arr []any) ([]any, domain.SuccessFilterResult) {
	if len(arr) == 0 {
		return arr, domain.NewSuccessFilterResult(0, 0, 0, 0, 0, 0)
	}

	var filtered []any
	var totalStats domain.SuccessFilterResult

	for _, item := range arr {
		itemMap, ok := item.(map[string]any)
		if !ok {
			filtered = append(filtered, item)
			continue
		}

		// First, recursively filter any nested arrays in this item
		processedItem, nestedStats := f.filterMap(itemMap)
		totalStats = f.mergeStats(totalStats, nestedStats)

		// Then determine the status of this item itself
		status := f.getItemStatus(itemMap)
		switch status {
		case statusPassed:
			totalStats.Total++
			totalStats.Passed++
			totalStats.Removed++
			// Don't add to filtered - removing passed items
		case statusFailed:
			totalStats.Total++
			totalStats.Failed++
			totalStats.Kept++
			filtered = append(filtered, processedItem)
		case statusSkipped:
			totalStats.Total++
			totalStats.Skipped++
			totalStats.Kept++
			filtered = append(filtered, processedItem)
		default:
			// Unknown status - keep the item only if it has nested content or is a container
			// Check if this is a container (has nested arrays that were processed)
			if nestedStats.Total > 0 {
				// This is a container, keep it
				filtered = append(filtered, processedItem)
			} else {
				// No status and no nested content - keep it
				totalStats.Total++
				totalStats.Kept++
				filtered = append(filtered, processedItem)
			}
		}
	}

	return filtered, totalStats
}

// isFilterableArray checks if the array field name is a known results array.
func (f *SuccessFilterer) isFilterableArray(fieldName string) bool {
	for _, name := range arrayFields {
		if fieldName == name {
			return true
		}
	}
	return false
}

// getItemStatus determines if an item passed, failed, or was skipped.
func (f *SuccessFilterer) getItemStatus(item map[string]any) string {
	for _, rule := range statusFieldRules {
		val, exists := item[rule.field]
		if !exists {
			continue
		}

		// Handle string values
		if strVal, ok := val.(string); ok {
			for _, passVal := range rule.passValues {
				if strVal == passVal {
					return statusPassed
				}
			}
			for _, failVal := range rule.failValues {
				if strVal == failVal {
					return statusFailed
				}
			}
			for _, skipVal := range rule.skipValues {
				if strVal == skipVal {
					return statusSkipped
				}
			}
		}
	}

	return statusUnknown
}

// mergeStats combines two SuccessFilterResult statistics.
func (f *SuccessFilterer) mergeStats(a, b domain.SuccessFilterResult) domain.SuccessFilterResult {
	return domain.NewSuccessFilterResult(
		a.Total+b.Total,
		a.Passed+b.Passed,
		a.Failed+b.Failed,
		a.Skipped+b.Skipped,
		a.Removed+b.Removed,
		a.Kept+b.Kept,
	)
}
