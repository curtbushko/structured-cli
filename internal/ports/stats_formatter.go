// Package ports defines the interfaces (contracts) for the structured-cli.
// This layer only imports from domain - never from adapters or application.
// Adapters implement these interfaces; application depends on them.
package ports

import (
	"github.com/curtbushko/structured-cli/internal/domain"
)

// StatsFormatter renders formatted statistics output with styled sections.
// Implementations produce terminal-aware output with borders, colors, and visual elements.
type StatsFormatter interface {
	// RenderHeader renders the styled header section with title.
	RenderHeader() string

	// RenderSummary renders the summary section with formatted stats.
	RenderSummary(summary domain.StatsSummary) string

	// RenderCommandTable renders the command table with impact bars.
	// Returns empty string if commands is empty.
	RenderCommandTable(commands []domain.AggregatedCommandStats) string
}
