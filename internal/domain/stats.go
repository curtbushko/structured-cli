package domain

import (
	"strings"
	"time"
)

// StatsSummary represents aggregated usage statistics.
type StatsSummary struct {
	// TotalCommands is the total number of commands executed.
	TotalCommands int

	// TotalTokensSaved is the cumulative tokens saved across all commands.
	TotalTokensSaved int

	// AvgSavingsPercent is the average percentage of tokens saved.
	AvgSavingsPercent float64

	// TotalExecutionTime is the cumulative execution time.
	TotalExecutionTime time.Duration

	// FilteredCount is the number of commands that had filters applied.
	FilteredCount int
}

// NewStatsSummary creates a new StatsSummary with the given values.
func NewStatsSummary(
	totalCommands int,
	totalTokensSaved int,
	avgSavingsPercent float64,
	totalExecTime time.Duration,
) StatsSummary {
	return StatsSummary{
		TotalCommands:      totalCommands,
		TotalTokensSaved:   totalTokensSaved,
		AvgSavingsPercent:  avgSavingsPercent,
		TotalExecutionTime: totalExecTime,
		FilteredCount:      0,
	}
}

// NewStatsSummaryWithFiltered creates a new StatsSummary with filtered count.
func NewStatsSummaryWithFiltered(
	totalCommands int,
	totalTokensSaved int,
	avgSavingsPercent float64,
	totalExecTime time.Duration,
	filteredCount int,
) StatsSummary {
	return StatsSummary{
		TotalCommands:      totalCommands,
		TotalTokensSaved:   totalTokensSaved,
		AvgSavingsPercent:  avgSavingsPercent,
		TotalExecutionTime: totalExecTime,
		FilteredCount:      filteredCount,
	}
}

// CommandStats represents statistics for a specific parser/command type.
type CommandStats struct {
	// ParserName identifies the parser (e.g., "git-status", "kubectl-get").
	ParserName string

	// InvocationCount is how many times this parser was used.
	InvocationCount int

	// TotalTokensSaved is cumulative tokens saved by this parser.
	TotalTokensSaved int

	// AvgExecutionTime is the average execution time for this command.
	AvgExecutionTime time.Duration
}

// NewCommandStats creates a new CommandStats with the given values.
func NewCommandStats(
	parserName string,
	invocationCount int,
	totalTokensSaved int,
	avgExecTime time.Duration,
) CommandStats {
	return CommandStats{
		ParserName:       parserName,
		InvocationCount:  invocationCount,
		TotalTokensSaved: totalTokensSaved,
		AvgExecutionTime: avgExecTime,
	}
}

// AggregatedCommandStats represents aggregated statistics for a command type,
// grouping multiple invocations of the same normalized command.
type AggregatedCommandStats struct {
	// CommandName is the normalized command name (e.g., "git status").
	CommandName string

	// Count is the number of times this command was invoked.
	Count int

	// TotalTokensSaved is the cumulative tokens saved by this command.
	TotalTokensSaved int

	// AvgSavingsPercent is the average percentage of tokens saved.
	AvgSavingsPercent float64

	// AvgExecutionTime is the average execution time for this command.
	AvgExecutionTime time.Duration

	// ImpactPercent is the relative impact (0-100%) based on token savings.
	ImpactPercent float64
}

// NewAggregatedCommandStats creates a new AggregatedCommandStats with the given values.
func NewAggregatedCommandStats(
	commandName string,
	count int,
	totalTokensSaved int,
	avgSavingsPercent float64,
	avgExecTime time.Duration,
) AggregatedCommandStats {
	return AggregatedCommandStats{
		CommandName:       commandName,
		Count:             count,
		TotalTokensSaved:  totalTokensSaved,
		AvgSavingsPercent: avgSavingsPercent,
		AvgExecutionTime:  avgExecTime,
	}
}

// NormalizeCommandName strips variable parts (paths, file arguments) from a
// command string, returning just the base command and subcommands.
func NormalizeCommandName(cmd string) string {
	if cmd == "" {
		return ""
	}

	parts := strings.Fields(cmd)
	var normalized []string
	for _, part := range parts {
		// Skip parts that look like paths or file arguments
		if strings.HasPrefix(part, "/") ||
			strings.HasPrefix(part, "./") ||
			strings.HasPrefix(part, "../") ||
			strings.Contains(part, ".") {
			continue
		}
		normalized = append(normalized, part)
	}

	if len(normalized) == 0 {
		return parts[0]
	}

	return strings.Join(normalized, " ")
}

// CalculateImpact computes relative impact percentages (0-100%) for each
// command based on token savings. The command with the highest savings gets
// 100%, and others are scaled proportionally.
func CalculateImpact(stats []AggregatedCommandStats) []AggregatedCommandStats {
	if len(stats) == 0 {
		return nil
	}

	// Find the maximum tokens saved.
	maxSaved := 0
	for _, s := range stats {
		if s.TotalTokensSaved > maxSaved {
			maxSaved = s.TotalTokensSaved
		}
	}

	result := make([]AggregatedCommandStats, len(stats))
	copy(result, stats)

	for i := range result {
		if maxSaved == 0 {
			result[i].ImpactPercent = 0.0
		} else {
			result[i].ImpactPercent = (float64(result[i].TotalTokensSaved) / float64(maxSaved)) * 100.0
		}
	}

	return result
}

// FilterStats represents statistics for a specific filter type.
type FilterStats struct {
	// FilterName identifies the filter (e.g., "small", "success", "dedupe").
	FilterName string

	// ActivationCount is how many times this filter was applied.
	ActivationCount int

	// TotalTokensSaved is cumulative tokens saved when this filter was active.
	TotalTokensSaved int
}

// NewFilterStats creates a new FilterStats with the given values.
func NewFilterStats(
	filterName string,
	activationCount int,
	totalTokensSaved int,
) FilterStats {
	return FilterStats{
		FilterName:       filterName,
		ActivationCount:  activationCount,
		TotalTokensSaved: totalTokensSaved,
	}
}
