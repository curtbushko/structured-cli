package application

import (
	"github.com/curtbushko/structured-cli/internal/ports"
)

// InMemoryParserRegistry is an in-memory implementation of ports.ParserRegistry.
// It stores parsers in a slice and searches through them to find matches.
type InMemoryParserRegistry struct {
	parsers []ports.Parser
}

// NewInMemoryParserRegistry creates a new empty InMemoryParserRegistry.
func NewInMemoryParserRegistry() *InMemoryParserRegistry {
	return &InMemoryParserRegistry{
		parsers: make([]ports.Parser, 0),
	}
}

// Find returns the parser that matches the given command and subcommands.
// It iterates through registered parsers and returns the first match.
// Returns the parser and true if found, or nil and false if no parser matches.
func (r *InMemoryParserRegistry) Find(cmd string, subcommands []string) (ports.Parser, bool) {
	for _, p := range r.parsers {
		if p.Matches(cmd, subcommands) {
			return p, true
		}
	}
	return nil, false
}

// Register adds a parser to the registry.
// The parser's Matches method determines which commands it handles.
func (r *InMemoryParserRegistry) Register(parser ports.Parser) {
	r.parsers = append(r.parsers, parser)
}

// All returns all registered parsers.
// Useful for listing supported commands or iterating over parsers.
func (r *InMemoryParserRegistry) All() []ports.Parser {
	// Return a copy to prevent external modification
	result := make([]ports.Parser, len(r.parsers))
	copy(result, r.parsers)
	return result
}
