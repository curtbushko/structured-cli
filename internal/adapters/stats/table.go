// Package stats provides rendering adapters for statistics display.
// This adapter package implements ports.StatsRenderer using formatted table output.
package stats

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// TableStatsRenderer renders statistics as formatted text tables.
// It implements ports.StatsRenderer for terminal output.
type TableStatsRenderer struct{}

// NewTableStatsRenderer creates a new TableStatsRenderer instance.
func NewTableStatsRenderer() *TableStatsRenderer {
	return &TableStatsRenderer{}
}

// RenderSummary writes a formatted summary of aggregated statistics.
func (r *TableStatsRenderer) RenderSummary(w io.Writer, summary domain.StatsSummary) error {
	_, err := fmt.Fprintf(w, "%-24s %d\n", "Total Commands:", summary.TotalCommands)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "%-24s %d\n", "Total Tokens Saved:", summary.TotalTokensSaved)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "%-24s %.1f%%\n", "Avg Savings:", summary.AvgSavingsPercent)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "%-24s %s\n", "Total Execution Time:", summary.TotalExecutionTime)
	if err != nil {
		return err
	}
	if summary.FilteredCount > 0 {
		_, err = fmt.Fprintf(w, "%-24s %d\n", "Filtered Commands:", summary.FilteredCount)
		if err != nil {
			return err
		}
	}
	return nil
}

// RenderHistory writes a formatted list of recent command records.
func (r *TableStatsRenderer) RenderHistory(w io.Writer, records []domain.CommandRecord) error {
	if len(records) == 0 {
		return nil
	}

	// Header
	_, err := fmt.Fprintf(w, "%-12s %-20s %10s %10s %8s %s\n",
		"Command", "Subcommands", "Raw", "Saved", "Savings", "Time")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "%s\n", strings.Repeat("-", 80))
	if err != nil {
		return err
	}

	for _, rec := range records {
		subcmds := strings.Join(rec.Subcommands, " ")
		_, err = fmt.Fprintf(w, "%-12s %-20s %10d %10d %7.1f%% %s\n",
			rec.Command, subcmds, rec.RawTokens, rec.TokensSaved,
			rec.SavingsPercent, rec.Timestamp.Format(time.DateTime))
		if err != nil {
			return err
		}
	}
	return nil
}

// RenderByParser writes statistics grouped by parser/command type.
func (r *TableStatsRenderer) RenderByParser(w io.Writer, stats []domain.CommandStats) error {
	if len(stats) == 0 {
		return nil
	}

	// Header
	_, err := fmt.Fprintf(w, "%-24s %12s %14s %14s\n",
		"Parser", "Invocations", "Tokens Saved", "Avg Time")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "%s\n", strings.Repeat("-", 68))
	if err != nil {
		return err
	}

	for _, s := range stats {
		_, err = fmt.Fprintf(w, "%-24s %12d %14d %14s\n",
			s.ParserName, s.InvocationCount, s.TotalTokensSaved, s.AvgExecutionTime)
		if err != nil {
			return err
		}
	}
	return nil
}

// RenderByFilter writes statistics grouped by filter type.
func (r *TableStatsRenderer) RenderByFilter(w io.Writer, stats []domain.FilterStats) error {
	if len(stats) == 0 {
		return nil
	}

	// Header
	_, err := fmt.Fprintf(w, "%-24s %14s %14s\n",
		"Filter", "Activations", "Tokens Saved")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "%s\n", strings.Repeat("-", 54))
	if err != nil {
		return err
	}

	for _, s := range stats {
		_, err = fmt.Fprintf(w, "%-24s %14d %14d\n",
			s.FilterName, s.ActivationCount, s.TotalTokensSaved)
		if err != nil {
			return err
		}
	}
	return nil
}
