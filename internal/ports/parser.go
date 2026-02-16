package ports

import (
	"io"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Parser transforms raw CLI output into structured data.
// Each parser handles a specific command/subcommand combination
// and produces output conforming to its associated schema.
//
// Implementations live in adapters/parsers/ (e.g., git/status.go).
type Parser interface {
	// Parse reads raw CLI output and returns structured data.
	// The reader typically contains stdout from the executed command.
	// Returns a ParseResult containing the structured data and metadata.
	Parse(r io.Reader) (domain.ParseResult, error)

	// Schema returns the JSON Schema that describes this parser's output format.
	// The schema is used for validation and documentation.
	Schema() domain.Schema

	// Matches returns true if this parser handles the given command.
	// The cmd parameter is the base command (e.g., "git").
	// The subcommands parameter is the subcommand chain (e.g., ["status"]).
	Matches(cmd string, subcommands []string) bool
}

// ParserRegistry manages parser discovery and registration.
// It provides a way to find the appropriate parser for a given command.
type ParserRegistry interface {
	// Find returns the parser that matches the given command and subcommands.
	// Returns the parser and true if found, or nil and false if no parser matches.
	Find(cmd string, subcommands []string) (Parser, bool)

	// Register adds a parser to the registry.
	// The parser's Matches method determines which commands it handles.
	Register(parser Parser)

	// All returns all registered parsers.
	// Useful for listing supported commands or iterating over parsers.
	All() []Parser
}
