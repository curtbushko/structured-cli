// Package ports defines the interfaces (contracts) for the structured-cli.
// This layer only imports from domain - never from adapters or application.
// Adapters implement these interfaces; application depends on them.
package ports

import (
	"io"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// StatsRenderer renders statistics and history to an output stream.
// Implementations format output for different mediums (terminal, JSON, etc.).
type StatsRenderer interface {
	// RenderSummary writes a formatted summary of aggregated statistics.
	RenderSummary(w io.Writer, summary domain.StatsSummary) error

	// RenderHistory writes a formatted list of recent command records.
	RenderHistory(w io.Writer, records []domain.CommandRecord) error

	// RenderByParser writes statistics grouped by parser/command type.
	RenderByParser(w io.Writer, stats []domain.CommandStats) error

	// RenderByFilter writes statistics grouped by filter type.
	RenderByFilter(w io.Writer, stats []domain.FilterStats) error
}
