# structured-cli Developer Guide

Project-specific instructions for AI coding agents.

## Project Overview

**structured-cli** is a universal CLI wrapper that transforms raw CLI output into structured JSON. It follows hexagonal architecture with strict dependency rules enforced by go-arch-lint.

## Architecture

```
cmd/structured-cli/main.go     # Composition root - wires all dependencies
internal/
  domain/                      # Pure types, no external deps
    command.go                 # Command, CommandSpec
    result.go                  # ParseResult
    schema.go                  # Schema
    fallback.go                # FallbackResult for unsupported commands
    errors.go                  # ExitError for exit code propagation
  ports/                       # Interfaces only
    runner.go                  # CommandRunner interface
    parser.go                  # Parser, ParserRegistry interfaces
    writer.go                  # OutputWriter interface
  application/                 # Business logic, uses ports (never adapters)
    executor.go                # Main orchestration
    registry.go                # InMemoryParserRegistry
    matcher.go                 # CommandMatcher
    error_handler.go           # JSON error formatting
  adapters/                    # Implementations of ports
    cli/                       # Cobra CLI handler (inbound)
    runner/                    # os/exec wrapper (outbound)
    writer/                    # JSON and passthrough writers (outbound)
    parsers/                   # 81 parser implementations (outbound)
      git/
      npm/
      docker/
      golang/
      ...
schemas/                       # 71 embedded JSON Schema files
```

## Dependency Rules

**Enforced by go-arch-lint:**

1. `domain/` - knows nothing about other packages
2. `ports/` - imports only domain
3. `application/` - imports domain and ports (NEVER adapters)
4. `adapters/` - imports domain and ports (implements ports)
5. `cmd/` - composition root, wires everything

Run `go-arch-lint check` to verify architecture compliance.

## Adding a New Parser

### 1. Create the parser file

Location: `internal/adapters/parsers/{category}/{command}.go`

```go
package category

import (
    "io"
    "github.com/curtbushko/structured-cli/internal/domain"
)

// CommandResult holds parsed output
type CommandResult struct {
    Success bool   `json:"success"`
    // Add fields matching schema
}

// CommandParser parses 'tool command' output
type CommandParser struct {
    schema domain.Schema
}

func NewCommandParser() *CommandParser {
    return &CommandParser{
        schema: domain.NewSchema(
            "https://structured-cli.dev/schemas/tool-command.json",
            "Tool Command Output",
            "object",
            map[string]domain.PropertySchema{
                "success": {Type: "boolean", Description: "..."},
            },
            []string{"success"},
        ),
    }
}

func (p *CommandParser) Parse(r io.Reader) (domain.ParseResult, error) {
    // Parse raw output from r
    // Return domain.NewParseResult(result, raw, 0)
}

func (p *CommandParser) Schema() domain.Schema {
    return p.schema
}

func (p *CommandParser) Matches(cmd string, subcommands []string) bool {
    // Return true if this parser handles the command
}
```

### 2. Create the JSON schema

Location: `schemas/tool-command.json`

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://structured-cli.dev/schemas/tool-command.json",
  "title": "Tool Command Output",
  "type": "object",
  "properties": {
    "success": {
      "type": "boolean",
      "description": "Whether the command succeeded"
    }
  },
  "required": ["success"]
}
```

### 3. Write tests

Location: `internal/adapters/parsers/{category}/{command}_test.go`

Test requirements:
- Parse successful output
- Parse error output
- Matches() returns true for correct commands
- Matches() returns false for incorrect commands
- Schema() returns valid schema with required properties

### 4. Register the parser

In `cmd/structured-cli/main.go`:

```go
import "github.com/curtbushko/structured-cli/internal/adapters/parsers/category"

// In run() function:
registry.Register(category.NewCommandParser())
```

## Testing

```bash
# Run all tests with race detection
go test -race ./...

# Run tests for a specific package
go test -race ./internal/adapters/parsers/git/...

# Run with coverage
go test -race -cover ./...
```

## Linting

```bash
# Run golangci-lint
golangci-lint run

# Check architecture compliance
go-arch-lint check
```

## Quality Gates

Before committing:
- [ ] `go build ./...` succeeds
- [ ] `go test -race ./...` passes
- [ ] `golangci-lint run` has no issues
- [ ] `go-arch-lint check` passes

## Common Patterns

### Parser Output Structure

All parsers return a result struct with:
- `Success` bool - whether command succeeded
- Command-specific fields matching the schema

### Error Handling

Return errors via `domain.NewParseResultWithError(err, raw, exitCode)`:
- Parser errors are wrapped in the result
- Exit codes are propagated
- Raw output preserved for debugging

### Matching Commands

The `Matches(cmd, subcommands)` method:
- `cmd` is the base command (e.g., "git", "npm")
- `subcommands` is the slice of subcommands (e.g., ["status"], ["compose", "up"])

### Regex Patterns

Define patterns at package level for reuse:
```go
var (
    successPattern = regexp.MustCompile(`^Success:`)
    errorPattern   = regexp.MustCompile(`^ERROR:`)
)
```

## File Naming Conventions

- Parser: `{command}.go` (e.g., `status.go`, `build.go`)
- Tests: `{command}_test.go`
- Types file: `types.go` for shared types within a package
- Constants: `const.go` or at top of relevant file

## Schema Conventions

- Use snake_case for JSON property names
- Always include `$schema`, `$id`, `title`, `type`
- Document all properties with `description`
- Mark required fields in `required` array
