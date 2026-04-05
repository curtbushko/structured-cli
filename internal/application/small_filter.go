// Package application contains the business logic and use cases for structured-cli.
// This layer orchestrates domain logic and depends on ports (interfaces) - never adapters.
package application

import (
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

// Status constants for SmallOutputResult.
const (
	StatusClean   = "clean"
	StatusEmpty   = "empty"
	StatusSuccess = "success"
	StatusOK      = "ok"
)

// SmallFilter determines whether command output should be filtered into a
// simplified format based on token count and pattern matching.
//
// It implements the ports.SmallOutputFilter interface and provides:
// - Token count threshold checking
// - Command/subcommand pattern matching
// - Compact status extraction from raw output
type SmallFilter struct {
	config   domain.SmallOutputConfig
	patterns []ports.MinimalPattern
}

// NewSmallFilter creates a new SmallFilter with default configuration.
// The filter is enabled by default with a token threshold of 25.
func NewSmallFilter() *SmallFilter {
	return &SmallFilter{
		config:   domain.NewSmallOutputConfig(),
		patterns: make([]ports.MinimalPattern, 0),
	}
}

// NewSmallFilterWithConfig creates a new SmallFilter with custom configuration.
func NewSmallFilterWithConfig(config domain.SmallOutputConfig) *SmallFilter {
	return &SmallFilter{
		config:   config,
		patterns: make([]ports.MinimalPattern, 0),
	}
}

// RegisterPattern adds a MinimalPattern to the filter.
// Patterns are checked in order when determining if output should be filtered.
func (f *SmallFilter) RegisterPattern(pattern ports.MinimalPattern) {
	f.patterns = append(f.patterns, pattern)
}

// ShouldFilter returns true if the given output should be filtered into a
// simplified format.
//
// The filter triggers when ALL of the following conditions are met:
// - The filter is enabled
// - The token count is below the threshold
// - A registered pattern matches the command/subcommand AND the raw output
//
// Parameters:
//   - raw: the raw command output string
//   - tokenCount: estimated number of tokens in the output
//   - cmd: the base command (e.g., "git")
//   - subcmds: the subcommand chain (e.g., ["status"])
func (f *SmallFilter) ShouldFilter(raw string, tokenCount int, cmd string, subcmds []string) bool {
	// Check if filter is enabled
	if !f.config.Enabled {
		return false
	}

	// Check token threshold (except for empty output which should always pass)
	if tokenCount >= f.config.TokenThreshold && raw != "" {
		return false
	}

	// Check if any pattern matches the command/subcommand and raw output
	return f.matchesPattern(raw, cmd, subcmds)
}

// matchesPattern checks if any registered pattern matches the given command
// and raw output.
func (f *SmallFilter) matchesPattern(raw string, cmd string, subcmds []string) bool {
	subcommand := ""
	if len(subcmds) > 0 {
		subcommand = subcmds[0]
	}

	for _, p := range f.patterns {
		// Check command matches
		if p.Command != cmd {
			continue
		}

		// Check subcommand matches (empty pattern subcommand matches any)
		if p.Subcommand != "" && p.Subcommand != subcommand {
			continue
		}

		// Check pattern matches raw output
		if p.Matches(raw) {
			return true
		}

		// Special case: empty pattern regex (^$) should match empty raw
		if raw == "" && p.Pattern != nil && p.Pattern.MatchString("") {
			return true
		}
	}

	return false
}

// Filter transforms raw output into a simplified SmallOutputResult.
// Should only be called after ShouldFilter returns true.
//
// The method extracts status from common patterns:
// - Empty output -> "empty"
// - Contains "nothing to commit" -> "clean"
// - Contains "success" or "succeeded" -> "success"
// - Default -> "ok"
//
// Parameters:
//   - raw: the raw command output string
//
// Returns a SmallOutputResult with status and summary fields.
func (f *SmallFilter) Filter(raw string) domain.SmallOutputResult {
	status := f.extractStatus(raw)
	summary := raw
	if summary == "" {
		summary = "no output"
	}

	return domain.NewSmallOutputResult(status, summary)
}

// extractStatus determines the status string from raw output.
func (f *SmallFilter) extractStatus(raw string) string {
	// Handle empty output
	if raw == "" {
		return StatusEmpty
	}

	rawLower := strings.ToLower(raw)

	// Check for common status patterns
	if strings.Contains(rawLower, "nothing to commit") {
		return StatusClean
	}

	if strings.Contains(rawLower, "success") || strings.Contains(rawLower, "succeeded") {
		return StatusSuccess
	}

	if strings.Contains(rawLower, "no changes") || strings.Contains(rawLower, "up to date") {
		return StatusClean
	}

	if strings.Contains(rawLower, "found 0 vulnerabilities") {
		return StatusClean
	}

	return StatusOK
}
