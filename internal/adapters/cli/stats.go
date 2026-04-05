package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

// statsFlags holds the flag values for the stats command.
type statsFlags struct {
	json     bool
	history  int // 0 = disabled, >0 = limit
	byParser bool
	project  bool // if true, filter to current working directory
}

// statsJSONSummary is the JSON output format for summary stats.
type statsJSONSummary struct {
	TotalCommands      int     `json:"total_commands"`
	TotalTokensSaved   int     `json:"total_tokens_saved"`
	AvgSavingsPercent  float64 `json:"avg_savings_percent"`
	TotalExecutionTime string  `json:"total_execution_time"`
}

// statsJSONHistoryItem is the JSON output format for a history item.
type statsJSONHistoryItem struct {
	Command        string   `json:"command"`
	Subcommands    []string `json:"subcommands"`
	TokensSaved    int      `json:"tokens_saved"`
	SavingsPercent float64  `json:"savings_percent"`
	Timestamp      string   `json:"timestamp"`
}

// statsJSONParserItem is the JSON output format for per-parser stats.
type statsJSONParserItem struct {
	ParserName       string `json:"parser_name"`
	InvocationCount  int    `json:"invocation_count"`
	TotalTokensSaved int    `json:"total_tokens_saved"`
	AvgExecutionTime string `json:"avg_execution_time"`
}

// buildStatsCommand creates the stats subcommand for the CLI.
func buildStatsCommand(reader ports.TrackingReader) *cobra.Command {
	flags := statsFlags{}

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show usage statistics for structured-cli",
		Long: `Display usage statistics including total commands executed,
tokens saved, and performance metrics.

By default, shows a summary of all statistics.
Use flags to customize the output.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			return executeStatsCommand(cmd.Context(), reader, flags, cmd.OutOrStdout())
		},
	}

	cmd.Flags().BoolVar(&flags.json, "json", false, "Output in JSON format")
	// --history accepts optional limit: --history (default 10) or --history=N
	cmd.Flags().IntVar(&flags.history, "history", 0, "Show recent command history (optionally specify limit, default 10)")
	// Set NoOptDefVal so --history without argument defaults to 10
	cmd.Flags().Lookup("history").NoOptDefVal = "10"
	cmd.Flags().BoolVar(&flags.byParser, "by-parser", false, "Show statistics grouped by parser")
	cmd.Flags().BoolVar(&flags.project, "project", false, "Filter to current working directory")

	return cmd
}

// executeStatsCommand runs the stats command with the given flags.
func executeStatsCommand(ctx context.Context, reader ports.TrackingReader, flags statsFlags, out io.Writer) error {
	// Handle history output (history > 0 means enabled)
	if flags.history > 0 {
		return executeHistoryStats(ctx, reader, flags, out)
	}

	// Handle by-parser output
	if flags.byParser {
		return executeByParserStats(ctx, reader, flags, out)
	}

	// Default: summary stats
	return executeSummaryStats(ctx, reader, flags, out)
}

// executeSummaryStats outputs the summary statistics.
func executeSummaryStats(ctx context.Context, reader ports.TrackingReader, flags statsFlags, out io.Writer) error {
	opts := ports.StatsOptions{}

	// If --project flag is set, use current working directory
	if flags.project {
		cwd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
		opts.Project = cwd
	}

	summary, err := reader.Stats(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	if flags.json {
		return writeSummaryJSON(out, summary)
	}

	return writeSummaryText(out, summary)
}

// writeSummaryJSON writes the summary as JSON.
func writeSummaryJSON(out io.Writer, summary domain.StatsSummary) error {
	output := statsJSONSummary{
		TotalCommands:      summary.TotalCommands,
		TotalTokensSaved:   summary.TotalTokensSaved,
		AvgSavingsPercent:  summary.AvgSavingsPercent,
		TotalExecutionTime: summary.TotalExecutionTime.String(),
	}

	enc := json.NewEncoder(out)
	return enc.Encode(output)
}

// writeSummaryText writes the summary as formatted text.
func writeSummaryText(out io.Writer, summary domain.StatsSummary) error {
	_, err := fmt.Fprintf(out, "Structured CLI Usage Statistics\n")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "================================\n\n")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "Total Commands:      %d\n", summary.TotalCommands)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "Total Tokens Saved:  %d\n", summary.TotalTokensSaved)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "Avg Savings:         %.1f%%\n", summary.AvgSavingsPercent)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "Total Exec Time:     %s\n", summary.TotalExecutionTime)
	return err
}

// executeHistoryStats outputs the command history.
func executeHistoryStats(ctx context.Context, reader ports.TrackingReader, flags statsFlags, out io.Writer) error {
	// flags.history contains the limit directly
	history, err := reader.History(ctx, flags.history)
	if err != nil {
		return fmt.Errorf("failed to get history: %w", err)
	}

	if flags.json {
		return writeHistoryJSON(out, history)
	}

	return writeHistoryText(out, history)
}

// writeHistoryJSON writes the history as JSON.
func writeHistoryJSON(out io.Writer, history []domain.CommandRecord) error {
	items := make([]statsJSONHistoryItem, 0, len(history))
	for _, record := range history {
		items = append(items, statsJSONHistoryItem{
			Command:        record.Command,
			Subcommands:    record.Subcommands,
			TokensSaved:    record.TokensSaved,
			SavingsPercent: record.SavingsPercent,
			Timestamp:      record.Timestamp.Format(time.RFC3339),
		})
	}

	enc := json.NewEncoder(out)
	return enc.Encode(items)
}

// writeHistoryText writes the history as formatted text.
func writeHistoryText(out io.Writer, history []domain.CommandRecord) error {
	if len(history) == 0 {
		_, err := fmt.Fprintln(out, "No command history found.")
		return err
	}

	_, err := fmt.Fprintf(out, "Recent Command History\n")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "======================\n\n")
	if err != nil {
		return err
	}

	// Header
	_, err = fmt.Fprintf(out, "%-20s %-30s %10s %10s\n", "TIMESTAMP", "COMMAND", "SAVED", "SAVINGS")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "%s\n", strings.Repeat("-", 75))
	if err != nil {
		return err
	}

	for _, record := range history {
		cmdStr := formatCommandString(record.Command, record.Subcommands)
		timestamp := record.Timestamp.Format("2006-01-02 15:04")
		_, err = fmt.Fprintf(out, "%-20s %-30s %10d %9.1f%%\n",
			timestamp, cmdStr, record.TokensSaved, record.SavingsPercent)
		if err != nil {
			return err
		}
	}

	return nil
}

// executeByParserStats outputs the per-parser statistics.
func executeByParserStats(ctx context.Context, reader ports.TrackingReader, flags statsFlags, out io.Writer) error {
	stats, err := reader.StatsByParser(ctx)
	if err != nil {
		return fmt.Errorf("failed to get parser stats: %w", err)
	}

	if flags.json {
		return writeByParserJSON(out, stats)
	}

	return writeByParserText(out, stats)
}

// writeByParserJSON writes the per-parser stats as JSON.
func writeByParserJSON(out io.Writer, stats []domain.CommandStats) error {
	items := make([]statsJSONParserItem, 0, len(stats))
	for _, stat := range stats {
		items = append(items, statsJSONParserItem{
			ParserName:       stat.ParserName,
			InvocationCount:  stat.InvocationCount,
			TotalTokensSaved: stat.TotalTokensSaved,
			AvgExecutionTime: stat.AvgExecutionTime.String(),
		})
	}

	enc := json.NewEncoder(out)
	return enc.Encode(items)
}

// writeByParserText writes the per-parser stats as formatted text.
func writeByParserText(out io.Writer, stats []domain.CommandStats) error {
	if len(stats) == 0 {
		_, err := fmt.Fprintln(out, "No parser statistics found.")
		return err
	}

	_, err := fmt.Fprintf(out, "Statistics by Parser\n")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "====================\n\n")
	if err != nil {
		return err
	}

	// Header
	_, err = fmt.Fprintf(out, "%-25s %10s %15s %15s\n", "PARSER", "COUNT", "TOKENS SAVED", "AVG TIME")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(out, "%s\n", strings.Repeat("-", 70))
	if err != nil {
		return err
	}

	for _, stat := range stats {
		_, err = fmt.Fprintf(out, "%-25s %10d %15d %15s\n",
			stat.ParserName, stat.InvocationCount, stat.TotalTokensSaved, stat.AvgExecutionTime)
		if err != nil {
			return err
		}
	}

	return nil
}

// formatCommandString builds a command string from the command and subcommands.
func formatCommandString(command string, subcommands []string) string {
	if len(subcommands) == 0 {
		return command
	}
	parts := make([]string, 0, len(subcommands)+1)
	parts = append(parts, command)
	parts = append(parts, subcommands...)
	return strings.Join(parts, " ")
}
