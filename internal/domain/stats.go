package domain

import (
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
