// Package ports defines the interfaces (contracts) for the structured-cli.
// This layer only imports from domain - never from adapters or services.
// Adapters implement these interfaces; services depend on them.
package ports

import (
	"github.com/curtbushko/structured-cli/internal/domain"
)

// SuccessFilter removes passing/successful items from parsed output,
// keeping only failures and a summary.
//
// Implementations handle:
//   - Test runners: remove tests where status/outcome = passed
//   - Linters: optionally remove warnings, keep errors
type SuccessFilter interface {
	// ShouldFilter returns true if this filter applies to the given command.
	// It checks if the command is a known test runner or linter.
	//
	// Parameters:
	//   - cmd: the base command (e.g., "npm", "go", "cargo")
	//   - subcmds: the subcommand chain (e.g., ["test"], ["run", "test"])
	ShouldFilter(cmd string, subcmds []string) bool

	// Filter removes passing items from data, returning filtered data
	// and statistics about what was removed.
	//
	// Parameters:
	//   - data: the parsed result data (typically from ParseResult.Data)
	//
	// Returns:
	//   - filtered data with passing items removed
	//   - statistics (total, passed, failed, skipped, removed, kept)
	Filter(data any) (any, domain.SuccessFilterResult)
}
