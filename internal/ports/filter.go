// Package ports defines the interfaces (contracts) for the structured-cli.
// This layer only imports from domain - never from adapters or services.
// Adapters implement these interfaces; services depend on them.
package ports

import (
	"regexp"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// MinimalPattern represents a parser-specific detection pattern for identifying
// small output that can be simplified. Each MinimalPattern targets a specific
// command/subcommand combination and uses a regex to detect known patterns.
type MinimalPattern struct {
	// Command is the base command (e.g., "git", "npm", "docker").
	Command string

	// Subcommand is the subcommand being targeted (e.g., "status", "install").
	Subcommand string

	// Pattern is the compiled regex pattern to match against raw output.
	Pattern *regexp.Regexp
}

// Matches returns true if the pattern matches the given raw output.
// Returns false for empty output or when the pattern does not match.
func (p MinimalPattern) Matches(raw string) bool {
	if raw == "" || p.Pattern == nil {
		return false
	}
	return p.Pattern.MatchString(raw)
}

// SmallOutputFilter determines whether command output should be filtered
// into a simplified format and performs that filtering.
//
// Implementations decide based on factors like token count, command type,
// and output patterns whether the output qualifies as "small" and can be
// reduced to a simple status and summary.
type SmallOutputFilter interface {
	// ShouldFilter returns true if the given output should be filtered
	// into a simplified format.
	//
	// Parameters:
	//   - raw: the raw command output string
	//   - tokenCount: estimated number of tokens in the output
	//   - cmd: the base command (e.g., "git")
	//   - subcmds: the subcommand chain (e.g., ["status"])
	ShouldFilter(raw string, tokenCount int, cmd string, subcmds []string) bool

	// Filter transforms raw output into a simplified SmallOutputResult.
	// Should only be called after ShouldFilter returns true.
	//
	// Parameters:
	//   - raw: the raw command output string
	//
	// Returns a SmallOutputResult with status and summary fields.
	Filter(raw string) domain.SmallOutputResult
}
