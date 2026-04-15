package domain

import (
	"time"
)

// CommandRecord represents a single command execution record for usage tracking.
// It captures the command details, token metrics, and timing information.
type CommandRecord struct {
	// ID is the unique identifier for this record (set by repository).
	ID int64

	// Command is the base command name (e.g., "git", "kubectl").
	Command string

	// Subcommands are the subcommand chain (e.g., ["status"], ["config", "view"]).
	Subcommands []string

	// RawTokens is the estimated token count of the raw command output.
	RawTokens int

	// ParsedTokens is the estimated token count of the parsed JSON output.
	ParsedTokens int

	// TokensSaved is the difference between raw and parsed tokens.
	TokensSaved int

	// SavingsPercent is the percentage of tokens saved (0-100).
	SavingsPercent float64

	// ExecutionTime is how long the command took to execute.
	ExecutionTime time.Duration

	// Timestamp is when this command was executed.
	Timestamp time.Time

	// Project is the project/directory context where the command was run.
	Project string

	// FiltersApplied is the list of filter names that were applied to this command.
	// Empty if no filters were applied (e.g., "small", "success", "dedupe").
	FiltersApplied []string
}

// NewCommandRecord creates a new CommandRecord with computed token savings.
func NewCommandRecord(
	command string,
	subcommands []string,
	rawTokens int,
	parsedTokens int,
	execTime time.Duration,
	project string,
) CommandRecord {
	return NewCommandRecordWithFilters(command, subcommands, rawTokens, parsedTokens, execTime, project, nil)
}

// NewCommandRecordWithFilters creates a new CommandRecord with computed token savings
// and the list of filters that were applied to this command.
func NewCommandRecordWithFilters(
	command string,
	subcommands []string,
	rawTokens int,
	parsedTokens int,
	execTime time.Duration,
	project string,
	filters []string,
) CommandRecord {
	tokensSaved := rawTokens - parsedTokens
	var savingsPercent float64
	if rawTokens > 0 {
		savingsPercent = float64(tokensSaved) / float64(rawTokens) * 100
	}

	// Ensure filters is an empty slice, not nil, for consistent serialization
	if filters == nil {
		filters = []string{}
	}

	return CommandRecord{
		Command:        command,
		Subcommands:    subcommands,
		RawTokens:      rawTokens,
		ParsedTokens:   parsedTokens,
		TokensSaved:    tokensSaved,
		SavingsPercent: savingsPercent,
		ExecutionTime:  execTime,
		Timestamp:      time.Now(),
		Project:        project,
		FiltersApplied: filters,
	}
}

// EstimateTokens estimates the number of tokens in a string using the chars/4 heuristic,
// with additional counting for newlines. LLM tokenizers typically treat newlines as
// separate tokens, so this provides more accurate estimation for multi-line text.
// This is important for calculating token savings when comparing multi-line CLI output
// to compact JSON output.
func EstimateTokens(s string) int {
	// Base estimation: chars / 4
	baseTokens := len(s) / 4

	// Count newlines as additional tokens since LLM tokenizers typically
	// treat newlines as separate tokens
	newlineTokens := countNewlines(s)

	return baseTokens + newlineTokens
}

// countNewlines counts the number of newline characters in a string.
func countNewlines(s string) int {
	count := 0
	for _, c := range s {
		if c == '\n' {
			count++
		}
	}
	return count
}

// ParseFailure represents a failed parse attempt for tracking and debugging.
type ParseFailure struct {
	// ID is the unique identifier for this record (set by repository).
	ID int64

	// Command is the full command that failed to parse.
	Command string

	// ErrorMessage describes why parsing failed.
	ErrorMessage string

	// FallbackSuccess indicates if the fallback handler succeeded.
	FallbackSuccess bool

	// Timestamp is when this failure occurred.
	Timestamp time.Time
}

// NewParseFailure creates a new ParseFailure record.
func NewParseFailure(command string, errorMsg string, fallbackSuccess bool) ParseFailure {
	return ParseFailure{
		Command:         command,
		ErrorMessage:    errorMsg,
		FallbackSuccess: fallbackSuccess,
		Timestamp:       time.Now(),
	}
}
