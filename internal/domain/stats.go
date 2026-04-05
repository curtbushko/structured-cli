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
