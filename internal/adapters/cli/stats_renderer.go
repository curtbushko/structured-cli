// Package cli provides the CLI adapter for structured-cli.
// This is an inbound adapter that handles user input via the command line.
package cli

import (
	"context"
	"io"
	"log"

	"github.com/curtbushko/structured-cli/internal/ports"
)

// renderPostExecutionStats retrieves the current stats summary from the tracker
// and renders them to the output writer using the configured StatsRenderer.
// Any errors are logged but do not fail the command execution.
func (h *Handler) renderPostExecutionStats(ctx context.Context, out io.Writer) {
	if h.statsRenderer == nil || h.tracker == nil {
		return
	}

	summary, err := h.tracker.Stats(ctx, ports.StatsOptions{})
	if err != nil {
		log.Printf("warning: failed to get stats for rendering: %v", err)
		return
	}

	if err := h.statsRenderer.RenderSummary(out, summary); err != nil {
		log.Printf("warning: failed to render stats: %v", err)
	}
}
