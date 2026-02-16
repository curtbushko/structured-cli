package application

import (
	"io"
	"testing"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// matchingParser is a test double for ports.Parser that matches specific cmd/subcommands.
type matchingParser struct {
	cmd         string
	subcommands []string
}

func (m *matchingParser) Parse(_ io.Reader) (domain.ParseResult, error) {
	return domain.ParseResult{}, nil
}

func (m *matchingParser) Schema() domain.Schema {
	return domain.Schema{}
}

func (m *matchingParser) Matches(cmd string, subcommands []string) bool {
	if cmd != m.cmd {
		return false
	}
	if len(subcommands) != len(m.subcommands) {
		return false
	}
	for i, sub := range subcommands {
		if sub != m.subcommands[i] {
			return false
		}
	}
	return true
}

func TestRegistry_Find_Exists(t *testing.T) {
	// Arrange: Registry with registered parser for 'git status'
	registry := NewInMemoryParserRegistry()
	parser := &matchingParser{cmd: "git", subcommands: []string{"status"}}
	registry.Register(parser)

	// Act: Find('git', ['status'])
	found, ok := registry.Find("git", []string{"status"})

	// Assert: Returns parser, true
	if !ok {
		t.Fatal("expected parser to be found")
	}
	if found != parser {
		t.Errorf("expected to find the registered parser, got %v", found)
	}
}

func TestRegistry_Find_NotExists(t *testing.T) {
	// Arrange: Empty registry
	registry := NewInMemoryParserRegistry()

	// Act: Find('git', ['status'])
	found, ok := registry.Find("git", []string{"status"})

	// Assert: Returns nil, false
	if ok {
		t.Fatal("expected parser not to be found")
	}
	if found != nil {
		t.Errorf("expected nil parser, got %v", found)
	}
}

func TestRegistry_Register(t *testing.T) {
	// Arrange: Empty registry
	registry := NewInMemoryParserRegistry()
	parser := &matchingParser{cmd: "npm", subcommands: []string{"install"}}

	// Act: Register parser, then Find
	registry.Register(parser)
	found, ok := registry.Find("npm", []string{"install"})

	// Assert: Parser is found
	if !ok {
		t.Fatal("expected parser to be found after registration")
	}
	if found != parser {
		t.Errorf("expected to find the registered parser, got %v", found)
	}
}

func TestRegistry_All(t *testing.T) {
	// Arrange: Registry with 3 parsers
	registry := NewInMemoryParserRegistry()
	parsers := []*matchingParser{
		{cmd: "git", subcommands: []string{"status"}},
		{cmd: "git", subcommands: []string{"log"}},
		{cmd: "npm", subcommands: []string{"install"}},
	}
	for _, p := range parsers {
		registry.Register(p)
	}

	// Act: Call All()
	all := registry.All()

	// Assert: Returns slice of 3 parsers
	if len(all) != 3 {
		t.Errorf("expected 3 parsers, got %d", len(all))
	}
}

func TestRegistry_Find_NoSubcommands(t *testing.T) {
	// Arrange: Registry with parser for command without subcommands
	registry := NewInMemoryParserRegistry()
	parser := &matchingParser{cmd: "ls", subcommands: nil}
	registry.Register(parser)

	// Act: Find('ls', [])
	found, ok := registry.Find("ls", nil)

	// Assert: Returns parser, true
	if !ok {
		t.Fatal("expected parser to be found")
	}
	if found != parser {
		t.Errorf("expected to find the registered parser, got %v", found)
	}
}

func TestRegistry_Find_NestedSubcommands(t *testing.T) {
	// Arrange: Registry with parser for nested subcommands
	registry := NewInMemoryParserRegistry()
	parser := &matchingParser{cmd: "docker", subcommands: []string{"compose", "up"}}
	registry.Register(parser)

	// Act: Find('docker', ['compose', 'up'])
	found, ok := registry.Find("docker", []string{"compose", "up"})

	// Assert: Returns parser, true
	if !ok {
		t.Fatal("expected parser to be found")
	}
	if found != parser {
		t.Errorf("expected to find the registered parser, got %v", found)
	}
}

func TestRegistry_Find_PartialSubcommandMatch(t *testing.T) {
	// Arrange: Registry with parser for 'docker compose up'
	registry := NewInMemoryParserRegistry()
	parser := &matchingParser{cmd: "docker", subcommands: []string{"compose", "up"}}
	registry.Register(parser)

	// Act: Find('docker', ['compose']) - partial match should not find
	_, ok := registry.Find("docker", []string{"compose"})

	// Assert: Returns false since we need exact subcommand match
	if ok {
		t.Fatal("expected parser not to be found for partial subcommand match")
	}
}
