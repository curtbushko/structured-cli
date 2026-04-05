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
	tokensSaved := rawTokens - parsedTokens
	var savingsPercent float64
	if rawTokens > 0 {
		savingsPercent = float64(tokensSaved) / float64(rawTokens) * 100
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
	}
}

// EstimateTokens estimates the number of tokens in a string using the chars/4 heuristic.
// This is a rough approximation commonly used for LLM token estimation.
func EstimateTokens(s string) int {
	return len(s) / 4
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
