# structured-cli: Technical Design Plan

A Go CLI wrapper that transforms raw CLI output into structured JSON.

## Executive Summary

**Goal**: Create a universal CLI wrapper (`structured-cli`) that intercepts commands like `git`, `npm`, `docker`, etc., parses their output, and returns structured JSON conforming to documented schemas.

**Key insight from research**: The real value of structured output is:
1. **Token savings** (up to 95% reduction)
2. **Consistent structure** (no regex parsing needed by AI)
3. **Schema validation** (guaranteed conformance)

---

## Implementation Checklist

### Phase 1: Project Setup
- [x] Initialize Go module (`github.com/yourorg/structured-cli`)
- [x] Create directory structure (`cmd/`, `internal/domain/`, `internal/ports/`, `internal/services/`, `internal/adapters/`)
- [x] Add `go.mod` with dependencies (cobra, jsonschema)
- [x] Create basic `main.go` composition root

### Phase 2: Core Domain & Ports
- [x] Define `domain/command.go` (Command, CommandSpec types)
- [x] Define `domain/result.go` (ParseResult type)
- [x] Define `domain/schema.go` (Schema type)
- [x] Define `ports/runner.go` (CommandRunner interface)
- [x] Define `ports/parser.go` (Parser, ParserRegistry interfaces)
- [x] Define `ports/writer.go` (OutputWriter interface)
- [x] Define `ports/schema.go` (SchemaRepository interface)

### Phase 3: Core Services
- [x] Implement `services/executor.go` (main orchestration)
- [x] Implement `services/matcher.go` (command → parser matching)
- [ ] Implement `services/validator.go` (schema validation)

### Phase 4: Adapters - Infrastructure
- [x] Implement `adapters/runner/exec.go` (os/exec wrapper)
- [x] Implement `adapters/schema/embedded.go` (embed.FS loader)
- [x] Implement `adapters/writer/json.go` (JSON output)
- [x] Implement `adapters/writer/passthrough.go` (raw output)

### Phase 5: Adapters - CLI Handler
- [x] Implement `adapters/cli/handler.go` (cobra setup)
- [x] Implement `adapters/cli/flags.go` (--json flag, env var)
- [x] Wire up composition root in `main.go`
- [x] Test passthrough mode works
- [x] Test `--json` flag triggers JSON output

### Phase 6: Git Parsers
- [ ] `git status` parser + schema
- [ ] `git log` parser + schema
- [ ] `git diff` parser + schema
- [ ] `git branch` parser + schema
- [ ] `git show` parser + schema
- [ ] `git add` parser + schema
- [ ] `git commit` parser + schema
- [ ] `git push` parser + schema
- [ ] `git pull` parser + schema
- [ ] `git checkout` parser + schema
- [ ] `git blame` parser + schema
- [ ] `git reflog` parser + schema

### Phase 7: Go Parsers
- [ ] `go build` parser + schema
- [ ] `go test` parser + schema
- [ ] `go test -cover` parser + schema
- [ ] `go vet` parser + schema
- [ ] `go run` parser + schema
- [ ] `go mod tidy` parser + schema
- [ ] `gofmt` parser + schema
- [ ] `go generate` parser + schema

### Phase 8: Build Tool Parsers
- [ ] `tsc` parser + schema
- [ ] `esbuild` parser + schema
- [ ] `vite build` parser + schema
- [ ] `webpack` parser + schema
- [ ] `cargo build` parser + schema

### Phase 9: Linter Parsers
- [ ] `eslint` parser + schema
- [ ] `prettier --check` parser + schema
- [ ] `biome check` parser + schema
- [ ] `golangci-lint` parser + schema
- [ ] `ruff check` parser + schema
- [ ] `mypy` parser + schema

### Phase 10: Test Runner Parsers
- [ ] `pytest` parser + schema
- [ ] `jest` parser + schema
- [ ] `vitest` parser + schema
- [ ] `mocha` parser + schema
- [ ] `cargo test` parser + schema

### Phase 11: Schema Validation & Error Handling
- [ ] Implement schema validation service
- [x] Implement error JSON output (`{"error": "...", "exitCode": N}`)
- [x] Implement unsupported command fallback (`{"raw": "...", "parsed": false}`)
- [ ] Implement native JSON passthrough with validation
- [ ] Implement streaming command buffering

### Phase 12: NPM Parsers
- [ ] `npm install` parser + schema
- [ ] `npm audit` parser + schema
- [ ] `npm outdated` parser + schema
- [ ] `npm list` parser + schema
- [ ] `npm run` parser + schema
- [ ] `npm test` parser + schema
- [ ] `npm init` parser + schema

### Phase 13: Docker Parsers
- [ ] `docker ps` parser + schema
- [ ] `docker build` parser + schema
- [ ] `docker logs` parser + schema
- [ ] `docker images` parser + schema
- [ ] `docker run` parser + schema
- [ ] `docker exec` parser + schema
- [ ] `docker pull` parser + schema
- [ ] `docker compose up` parser + schema
- [ ] `docker compose down` parser + schema
- [ ] `docker compose ps` parser + schema

### Phase 14: Python Parsers
- [ ] `pip install` parser + schema
- [ ] `pip-audit` parser + schema
- [ ] `uv pip install` parser + schema
- [ ] `uv run` parser + schema
- [ ] `black --check` parser + schema

### Phase 15: Cargo (Rust) Parsers
- [ ] `cargo build` parser + schema
- [ ] `cargo test` parser + schema
- [ ] `cargo clippy` parser + schema
- [ ] `cargo run` parser + schema
- [ ] `cargo add` parser + schema
- [ ] `cargo remove` parser + schema
- [ ] `cargo fmt` parser + schema
- [ ] `cargo doc` parser + schema
- [ ] `cargo check` parser + schema

### Phase 16: Make/Just Parsers
- [ ] `make` parser + schema
- [ ] `just` parser + schema

### Phase 17: File Operations Parsers
- [ ] `ls` parser + schema
- [ ] `find` parser + schema
- [ ] `grep` parser + schema
- [ ] `ripgrep (rg)` parser + schema
- [ ] `fd` parser + schema
- [ ] `cat` parser + schema
- [ ] `head` / `tail` parser + schema
- [ ] `wc` parser + schema
- [ ] `du` parser + schema
- [ ] `df` parser + schema

### Phase 18: Polish & Release
- [ ] Comprehensive test coverage (>80%)
- [ ] Documentation (README, CLAUDE.md)

---

## Research Findings

### Schema Sources

#### 1. SchemaStore.org
- **What it is**: Central registry of JSON schemas for config files (tsconfig.json, package.json, etc.)
- **API endpoint**: `https://www.schemastore.org/api/json/catalog.json`
- **Limitation**: Focuses on *config file formats*, not CLI output schemas
- **Verdict**: Limited utility for our use case—these are input schemas, not output schemas

#### 2. Pare (@paretools/*)
- **What it is**: Dev tools with hand-crafted output schemas
- **Repository**: https://github.com/Dave-London/Pare
- **Schema approach**: Custom TypeScript Zod schemas defining CLI output structure
- **Coverage**: git, npm, docker, test runners, build tools, linters, Python tools, Cargo, Go
- **Verdict**: Primary source for output schemas—extract and port to JSON Schema

#### 3. Native CLI JSON flags
Many CLIs already support JSON output:
- `gh pr list --json` (GitHub CLI)
- `npm ls --json`
- `docker inspect`
- `kubectl get -o json`
- `cargo metadata --format-version=1`

### Background: MCP structuredContent

For reference, MCP (Model Context Protocol) uses this format for tool responses:

```json
{
  "structuredContent": { "temperature": 22.5 },
  "content": [{ "type": "text", "text": "{\"temperature\": 22.5}" }]
}
```

**Key insight**: At the CLI level, AI doesn't see `structuredContent` differently than regular JSON text. When Claude Code runs `structured-cli git status --json`, stdout goes into the conversation context as plain text. Claude parses JSON like any other text—there's no special treatment.

Simply outputting valid JSON achieves the same benefit. Schemas can be published separately for documentation/validation.

---

## Architecture

### Hexagonal Architecture (Ports & Adapters)

The architecture follows the classic hexagonal/ports-and-adapters pattern where:
- **domain/** contains pure business types with no external dependencies (the hexagon center)
- **ports/** defines interfaces (contracts) for how the domain interacts with the outside world
- **adapters/** implements those ports, connecting to external systems
- **services/** orchestrates domain logic (application layer)

```
                            ┌─────────────────────────────────────┐
                            │           cmd/main.go               │
                            │         (Composition Root)          │
                            │   Wire up adapters → ports → domain │
                            └──────────────┬──────────────────────┘
                                           │
        ┌──────────────────────────────────┼──────────────────────────────────┐
        │                                  │                                  │
        ▼                                  ▼                                  ▼
┌───────────────────┐            ┌─────────────────────┐            ┌───────────────────┐
│  INBOUND ADAPTER  │            │    THE HEXAGON      │            │ OUTBOUND ADAPTERS │
│    adapters/cli   │            │                     │            │  adapters/runner  │
│                   │            │  ┌───────────────┐  │            │  adapters/schema  │
├───────────────────┤            │  │    domain/    │  │            │  adapters/writer  │
│                   │            │  │               │  │            │  adapters/parsers │
│  CLI Handler      │───────────▶│  │ - Command     │  │◀───────────├───────────────────┤
│  (cobra/flags)    │            │  │ - ParseResult │  │            │                   │
│                   │            │  │ - Schema      │  │            │  Exec Runner      │
│  --json flag      │            │  └───────────────┘  │            │  (os/exec)        │
│  STRUCTURED_CLI_  │            │         │          │            │                   │
│  JSON env var     │            │         ▼          │            │  Schema Loader    │
│                   │            │  ┌───────────────┐  │            │  (embed.FS)       │
└───────────────────┘            │  │    ports/     │  │            │                   │
                                 │  │  (Interfaces) │  │            │  Output Writers   │
                                 │  │               │  │            │  (json/passthru)  │
                                 │  │ - Runner      │  │            │                   │
                                 │  │ - Parser      │  │            │  Parsers          │
                                 │  │ - Writer      │  │            │  (git/npm/docker) │
                                 │  │ - SchemaRepo  │  │            │                   │
                                 │  └───────────────┘  │            └───────────────────┘
                                 │         │          │
                                 │         ▼          │
                                 │  ┌───────────────┐  │
                                 │  │   services/   │  │
                                 │  │               │  │
                                 │  │ - Executor    │  │
                                 │  │ - Matcher     │  │
                                 │  │ - Validator   │  │
                                 │  └───────────────┘  │
                                 │                     │
                                 └─────────────────────┘
```

### Directory Layout

```
internal/
├── domain/      # Pure types (no imports from other internal packages)
├── ports/       # Interfaces only (imports domain)
├── services/    # Business logic (imports domain, ports)
└── adapters/    # Implementations (imports domain, ports)
    ├── cli/
    ├── runner/
    ├── schema/
    ├── writer/
    └── parsers/
```

### Dependency Rules

Following the Dependency Inversion Principle:

1. **domain/** → knows nothing about other packages
2. **ports/** → imports only domain
3. **services/** → imports domain and ports (never adapters)
4. **adapters/** → imports domain and ports (implements ports)
5. **cmd/** → composition root, wires everything together

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              DEPENDENCY FLOW                                 │
│                                                                             │
│   cmd/main.go ──────▶ adapters ──────▶ services ──────▶ ports ──────▶ domain│
│       │                   │               │              │            │     │
│       │                   │               │              │            │     │
│       │                   ▼               ▼              ▼            ▼     │
│       │              implements       uses ports     interfaces    types    │
│       │                                (never imports                       │
│       └── wires adapters to ports        adapters)                          │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### Parser Adapters

Each CLI tool gets a parser adapter that implements the `ports.Parser` interface:

```
adapters/parsers/
├── git/
│   ├── status.go
│   ├── log.go
│   ├── diff.go
│   └── ...
├── npm/
│   ├── install.go
│   ├── audit.go
│   └── ...
├── docker/
│   ├── ps.go
│   ├── build.go
│   └── ...
└── ...
```

### Ports (Interfaces)

```go
// internal/ports/runner.go
package ports

import (
    "context"
    "io"
)

// CommandRunner executes shell commands
// Implemented by: adapters/runner/exec.go
type CommandRunner interface {
    Run(ctx context.Context, cmd string, args []string) (stdout, stderr io.Reader, exitCode int, err error)
}
```

```go
// internal/ports/parser.go
package ports

import (
    "io"
    "github.com/yourorg/structured-cli/internal/domain"
)

// Parser transforms raw CLI output into structured data
// Implemented by: adapters/parsers/*
type Parser interface {
    // Parse reads raw CLI output and returns structured data
    Parse(r io.Reader) (domain.ParseResult, error)
    
    // Schema returns the JSON Schema for this parser's output
    Schema() domain.Schema
    
    // Matches returns true if this parser handles the given command
    Matches(cmd string, subcommands []string) bool
}

// ParserRegistry finds the right parser for a command
type ParserRegistry interface {
    Find(cmd string, subcommands []string) (Parser, bool)
    Register(parser Parser)
    All() []Parser
}
```

```go
// internal/ports/writer.go
package ports

import (
    "io"
    "github.com/yourorg/structured-cli/internal/domain"
)

// OutputWriter formats and writes results
// Implemented by: adapters/writer/*.go
type OutputWriter interface {
    Write(w io.Writer, result domain.ParseResult, schema domain.Schema) error
}
```

```go
// internal/ports/schema.go
package ports

import "github.com/yourorg/structured-cli/internal/domain"

// SchemaRepository provides access to JSON schemas
// Implemented by: adapters/schema/embedded.go
type SchemaRepository interface {
    Get(name string) (domain.Schema, error)
    List() []string
}
```

### Domain Types

```go
// internal/domain/command.go
package domain

// Command represents a CLI command to execute
type Command struct {
    Name        string   // e.g., "git"
    Subcommands []string // e.g., ["status"]
    Args        []string // remaining arguments
}

// CommandSpec describes a supported command pattern
type CommandSpec struct {
    Name        string
    Subcommand  string
    Description string
}
```

```go
// internal/domain/result.go
package domain

// ParseResult holds the structured output from parsing
type ParseResult struct {
    Data     any    // The structured data (matches schema)
    Raw      string // Original raw output (for passthrough)
    ExitCode int
    Error    error
}
```

```go
// internal/domain/schema.go
package domain

import "encoding/json"

// Schema wraps a JSON Schema
type Schema struct {
    ID         string          `json:"$id"`
    Title      string          `json:"title"`
    Type       string          `json:"type"`
    Properties json.RawMessage `json:"properties"`
    Required   []string        `json:"required"`
    raw        []byte          // Original JSON for validation
}

func (s Schema) Raw() []byte { return s.raw }
```

### Decorator/Wrapper Pattern

```go
// internal/services/validator.go
package services

import (
    "fmt"
    "io"
    "github.com/yourorg/structured-cli/internal/domain"
    "github.com/yourorg/structured-cli/internal/ports"
)

// ValidatingParser wraps a parser with schema validation
type ValidatingParser struct {
    inner     ports.Parser
    validator SchemaValidator
}

func NewValidatingParser(inner ports.Parser, validator SchemaValidator) *ValidatingParser {
    return &ValidatingParser{inner: inner, validator: validator}
}

func (p *ValidatingParser) Parse(r io.Reader) (domain.ParseResult, error) {
    result, err := p.inner.Parse(r)
    if err != nil {
        return result, err
    }
    
    schema := p.inner.Schema()
    if err := p.validator.Validate(result.Data, schema); err != nil {
        return result, fmt.Errorf("schema validation failed: %w", err)
    }
    
    return result, nil
}

func (p *ValidatingParser) Schema() domain.Schema {
    return p.inner.Schema()
}

func (p *ValidatingParser) Matches(cmd string, subcommands []string) bool {
    return p.inner.Matches(cmd, subcommands)
}
```

```go
// internal/adapters/writer/json.go
package writer

import (
    "encoding/json"
    "io"
    "github.com/yourorg/structured-cli/internal/domain"
)

// JSONWriter outputs structured JSON
type JSONWriter struct {
    Indent bool
}

func (w *JSONWriter) Write(out io.Writer, result domain.ParseResult, schema domain.Schema) error {
    enc := json.NewEncoder(out)
    if w.Indent {
        enc.SetIndent("", "  ")
    }
    return enc.Encode(result.Data)
}
```

```go
// internal/adapters/writer/passthrough.go
package writer

import (
    "io"
    "github.com/yourorg/structured-cli/internal/domain"
)

// PassthroughWriter outputs raw command output unchanged
type PassthroughWriter struct{}

func (w *PassthroughWriter) Write(out io.Writer, result domain.ParseResult, schema domain.Schema) error {
    _, err := io.WriteString(out, result.Raw)
    return err
}
```

---

## Command Matching Strategy

### Pattern Matching

```go
type CommandMatcher struct {
    patterns []CommandPattern
}

type CommandPattern struct {
    Command    string   // "git"
    Subcommand string   // "status"
    ArgPattern []string // optional positional arg patterns
    Parser     OutputParser
}

// Match "structured-cli git status" → GitStatusParser
// Match "structured-cli git log --oneline -n 5" → GitLogParser
// Match "structured-cli npm install" → NpmInstallParser
```

### Subcommand Detection

```go
func (m *CommandMatcher) Match(args []string) (OutputParser, []string, error) {
    if len(args) < 1 {
        return nil, nil, ErrNoCommand
    }
    
    cmd := args[0]  // "git"
    subcommands := detectSubcommands(cmd, args[1:])  // ["status"]
    remaining := args[len(subcommands)+1:]           // any trailing args
    
    for _, pattern := range m.patterns {
        if pattern.Matches(cmd, subcommands) {
            return pattern.Parser, remaining, nil
        }
    }
    
    // No structured parser available - passthrough mode
    return nil, args, nil
}
```

---

## Schema Strategy

### Priority Order

1. **Native JSON output** (if CLI supports it)
   - Use `gh --json`, `npm --json`, `docker inspect`, etc.
   - Pass through with optional schema validation

2. **Pare schemas** (primary source)
   - Extract from Pare's TypeScript definitions
   - Convert to JSON Schema
   - Embed in binary

3. **SchemaStore** (config files only)
   - Use for validating config file content
   - Not applicable to CLI output

### Schema Registry

```go
//go:embed schemas/*.json
var schemaFS embed.FS

type SchemaRegistry struct {
    schemas map[string]*jsonschema.Schema
}

func NewSchemaRegistry() *SchemaRegistry {
    r := &SchemaRegistry{schemas: make(map[string]*jsonschema.Schema)}
    
    // Load embedded schemas
    entries, _ := schemaFS.ReadDir("schemas")
    for _, e := range entries {
        data, _ := schemaFS.ReadFile("schemas/" + e.Name())
        schema, _ := jsonschema.CompileString(e.Name(), string(data))
        r.schemas[strings.TrimSuffix(e.Name(), ".json")] = schema
    }
    
    return r
}

func (r *SchemaRegistry) Get(name string) *jsonschema.Schema {
    return r.schemas[name]
}
```

### Example Schema (git status)

Derived from Pare's output:

```json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://structured-cli.dev/schemas/git-status.json",
  "title": "Git Status Output",
  "type": "object",
  "properties": {
    "branch": { "type": "string" },
    "upstream": { "type": ["string", "null"] },
    "ahead": { "type": "integer" },
    "behind": { "type": "integer" },
    "staged": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "file": { "type": "string" },
          "status": { "type": "string", "enum": ["added", "modified", "deleted", "renamed", "copied"] }
        },
        "required": ["file", "status"]
      }
    },
    "modified": { "type": "array", "items": { "type": "string" } },
    "deleted": { "type": "array", "items": { "type": "string" } },
    "untracked": { "type": "array", "items": { "type": "string" } },
    "conflicts": { "type": "array", "items": { "type": "string" } },
    "clean": { "type": "boolean" }
  },
  "required": ["branch", "staged", "modified", "deleted", "untracked", "conflicts", "clean"]
}
```

---

## CLI Interface

### Basic Usage

```bash
# Default: passthrough (STRUCTURED_CLI_JSON=false)
structured-cli git status
# Output: raw git status

# JSON mode via flag
structured-cli git status --json
structured-cli --json git status

# JSON mode via env var
STRUCTURED_CLI_JSON=true structured-cli git status

# Alias usage
alias git="structured-cli git"
git status --json
```

### Flag Handling

```go
func main() {
    var jsonOutput bool
    
    // Check env var first
    if os.Getenv("STRUCTURED_CLI_JSON") == "true" {
        jsonOutput = true
    }
    
    // Parse our flags before the subcommand
    args := os.Args[1:]
    args, jsonOutput = extractJSONFlag(args, jsonOutput)
    
    // Now args contains only the wrapped command
    // e.g., ["git", "status"] or ["npm", "install", "lodash"]
}

func extractJSONFlag(args []string, current bool) ([]string, bool) {
    var filtered []string
    jsonOutput := current
    
    for _, arg := range args {
        if arg == "--json" {
            jsonOutput = true
        } else {
            filtered = append(filtered, arg)
        }
    }
    
    return filtered, jsonOutput
}
```

### Output Modes

```go
type OutputMode int

const (
    Passthrough OutputMode = iota  // STRUCTURED_CLI_JSON=false (default)
    JSON                           // --json flag or STRUCTURED_CLI_JSON=true  
)

func selectWriter(mode OutputMode) ports.OutputWriter {
    switch mode {
    case JSON:
        return &writer.JSONWriter{}
    default:
        return &writer.PassthroughWriter{}
    }
}
```

---

## Native JSON Passthrough

For CLIs that already support JSON output:

```go
type NativeJSONAdapter struct {
    cmd       string
    jsonFlag  string  // "--json", "-o json", etc.
    transform func([]byte) (any, error)  // optional transformation
}

var nativeJSONAdapters = map[string]NativeJSONAdapter{
    "gh":      {cmd: "gh", jsonFlag: "--json"},
    "kubectl": {cmd: "kubectl", jsonFlag: "-o=json"},
    "docker":  {cmd: "docker", jsonFlag: "--format={{json .}}"},
    "npm":     {cmd: "npm", jsonFlag: "--json"},
    "cargo":   {cmd: "cargo", jsonFlag: "--message-format=json"},
}

func (a *NativeJSONAdapter) Run(ctx context.Context, args []string) (any, error) {
    // Inject JSON flag if not present
    args = a.injectJSONFlag(args)
    
    output, err := exec.CommandContext(ctx, a.cmd, args...).Output()
    if err != nil {
        return nil, err
    }
    
    if a.transform != nil {
        return a.transform(output)
    }
    
    var result any
    return result, json.Unmarshal(output, &result)
}
```

---

## Implementation Phases

### Phase 1: Core Framework (Week 1-2)
- [ ] Project structure with hexagonal architecture
- [ ] CommandMatcher and pattern registration
- [ ] Reader/Writer interfaces
- [ ] Basic CLI parsing (--json flag, env var)
- [ ] Passthrough mode for unsupported commands
- [ ] Schema registry with embedded schemas

### Phase 2: Git Adapter (Week 2-3)
- [ ] Port Pare's git schemas to JSON Schema
- [ ] Implement parsers for: status, log, diff, branch, show
- [ ] Add parsers for: add, commit, push, pull, checkout
- [ ] Add parsers for: stash, blame, reflog

### Phase 3: NPM/Node Adapters (Week 3-4)
- [ ] npm install, audit, outdated, list
- [ ] npm run, test, info, search
- [ ] nvm support

### Phase 4: Docker Adapter (Week 4-5)
- [ ] ps, build, logs, images
- [ ] run, exec, pull, inspect
- [ ] compose-up, compose-down, compose-ps
- [ ] network-ls, volume-ls, stats

### Phase 5: Test/Build/Lint Adapters (Week 5-6)
- [ ] Test: pytest, jest, vitest, mocha, playwright
- [ ] Build: tsc, esbuild, vite, webpack, turbo, nx
- [ ] Lint: eslint, prettier, biome, stylelint

### Phase 6: Language-specific Adapters (Week 6-7)
- [ ] Python: pip, mypy, ruff, uv, black
- [ ] Rust: cargo build/test/clippy/fmt/audit
- [ ] Go: build, test, vet, fmt, golangci-lint

### Phase 7: Remaining Adapters (Week 7-8)
- [ ] GitHub CLI (gh)
- [ ] Search (ripgrep, fd, jq)
- [ ] HTTP (curl wrapper)
- [ ] Make/just
- [ ] K8s (kubectl, helm)
- [ ] Security (trivy, semgrep, gitleaks)

### Phase 8: Polish (Week 8-9)
- [ ] Comprehensive test coverage
- [ ] Documentation
- [ ] Homebrew formula
- [ ] GitHub releases with binaries

---

## File Structure

```
structured-cli/
├── cmd/
│   └── structured-cli/
│       └── main.go                    # Composition root: wire adapters to services
├── internal/
│   ├── domain/                        # Pure domain types (no dependencies)
│   │   ├── command.go                 # Command, Subcommand types
│   │   ├── result.go                  # ParseResult, Error types
│   │   └── schema.go                  # Schema type
│   ├── ports/                         # Interfaces (contracts)
│   │   ├── runner.go                  # CommandRunner interface
│   │   ├── parser.go                  # Parser interface
│   │   ├── writer.go                  # OutputWriter interface
│   │   └── schema.go                  # SchemaRepository interface
│   ├── services/                      # Application/business logic (uses ports only)
│   │   ├── executor.go                # Main execution orchestrator
│   │   ├── matcher.go                 # Command → Parser matching
│   │   └── validator.go               # Schema validation service
│   └── adapters/                      # Implementations of ports
│       ├── cli/                       # CLI handler (inbound)
│       │   ├── handler.go             # Cobra command setup
│       │   └── flags.go               # Flag parsing (--json, etc.)
│       ├── runner/                    # Command execution (outbound)
│       │   └── exec.go                # os/exec implementation
│       ├── schema/                    # Schema loading (outbound)
│       │   └── embedded.go            # embed.FS implementation
│       ├── writer/                    # Output writers (outbound)
│       │   ├── json.go                # JSON output
│       │   └── passthrough.go         # Raw passthrough
│       └── parsers/                   # CLI output parsers (outbound)
│           ├── git/
│           │   ├── status.go
│           │   ├── log.go
│           │   ├── diff.go
│           │   ├── branch.go
│           │   ├── show.go
│           │   ├── blame.go
│           │   └── reflog.go
│           ├── npm/
│           │   ├── install.go
│           │   ├── audit.go
│           │   ├── outdated.go
│           │   └── list.go
│           ├── docker/
│           │   ├── ps.go
│           │   ├── build.go
│           │   ├── logs.go
│           │   └── compose.go
│           ├── test/
│           │   ├── jest.go
│           │   ├── pytest.go
│           │   ├── vitest.go
│           │   └── mocha.go
│           ├── build/
│           │   ├── tsc.go
│           │   ├── esbuild.go
│           │   └── webpack.go
│           ├── lint/
│           │   ├── eslint.go
│           │   ├── prettier.go
│           │   └── biome.go
│           ├── python/
│           │   ├── pip.go
│           │   ├── mypy.go
│           │   ├── ruff.go
│           │   └── uv.go
│           ├── cargo/
│           │   ├── build.go
│           │   ├── test.go
│           │   └── clippy.go
│           ├── golang/
│           │   ├── build.go
│           │   ├── test.go
│           │   └── vet.go
│           ├── github/
│           │   ├── pr.go
│           │   └── issue.go
│           ├── search/
│           │   ├── ripgrep.go
│           │   └── fd.go
│           ├── http/
│           │   └── curl.go
│           ├── make/
│           │   └── make.go
│           ├── k8s/
│           │   └── kubectl.go
│           └── security/
│               ├── trivy.go
│               └── semgrep.go
├── schemas/                           # Embedded JSON schemas
│   ├── git-status.json
│   ├── git-log.json
│   ├── npm-install.json
│   └── ...
├── .go-arch-lint.yml                  # Architecture enforcement
├── go.mod
├── go.sum
└── README.md
```

### go-arch-lint Configuration

```yaml
# .go-arch-lint.yml
version: 3
workdir: internal

components:
  domain:    { in: domain }
  ports:     { in: ports }
  services:  { in: services }
  cli:       { in: adapters/cli }
  runner:    { in: adapters/runner }
  schema:    { in: adapters/schema }
  writer:    { in: adapters/writer }
  parsers:   { in: adapters/parsers/** }

commonComponents:
  - domain

deps:
  # Domain knows nothing
  domain:
    mayDependOn: []
  
  # Ports only know domain
  ports:
    mayDependOn:
      - domain
  
  # Services know ports (interfaces), not adapters
  services:
    mayDependOn:
      - domain
      - ports
  
  # Inbound adapter uses services and ports
  cli:
    mayDependOn:
      - domain
      - ports
      - services
  
  # Outbound adapters implement ports
  runner:
    mayDependOn:
      - domain
      - ports
  
  schema:
    mayDependOn:
      - domain
      - ports
  
  writer:
    mayDependOn:
      - domain
      - ports
  
  parsers:
    mayDependOn:
      - domain
      - ports
```

---

## Dependencies

```go
// go.mod
module github.com/yourorg/structured-cli

go 1.22

require (
    github.com/santhosh-tekuri/jsonschema/v6  // JSON Schema validation
    github.com/spf13/cobra                     // CLI framework
)
```

---

## Testing Strategy

### Unit Tests
- Parser output for each subcommand
- Schema validation
- Flag parsing

### Integration Tests
- End-to-end CLI invocation
- Real command execution (with mocks for CI)

### Snapshot Tests
- Golden file comparisons for parser output
- Schema conformance

```go
func TestGitStatusParser(t *testing.T) {
    input := `On branch main
Your branch is ahead of 'origin/main' by 2 commits.

Changes to be committed:
        modified:   src/index.ts
        new file:   src/utils.ts

Changes not staged for commit:
        modified:   README.md

Untracked files:
        temp.log
`
    
    parser := NewGitStatusParser()
    result, err := parser.Parse(strings.NewReader(input))
    require.NoError(t, err)
    
    status := result.(*GitStatus)
    assert.Equal(t, "main", status.Branch)
    assert.Equal(t, 2, status.Ahead)
    assert.Len(t, status.Staged, 2)
    assert.Len(t, status.Modified, 1)
    assert.Len(t, status.Untracked, 1)
    assert.False(t, status.Clean)
}
```

---

## BDD Features (Gherkin)

### Feature: Git Commands

```gherkin
Feature: Git Commands
  As an AI coding agent
  I want structured output from all git commands
  So that I can understand repository state without parsing text

  Scenario: git status - clean repository
    Given I have a git repository with no changes
    When I run "structured-cli git status --json"
    Then the JSON should contain "clean" equal to true
    And the JSON should contain "branch" as a string
    And the JSON should contain "staged" as an empty array
    And the JSON should contain "modified" as an empty array
    And the JSON should contain "untracked" as an empty array

  Scenario: git status - with changes
    Given I have a git repository with staged, modified, and untracked files
    When I run "structured-cli git status --json"
    Then the JSON should contain "clean" equal to false
    And each item in "staged" should have "file" and "status"
    And "modified" should be an array of file paths
    And "untracked" should be an array of file paths

  Scenario: git status - tracking upstream
    Given I have a git repository tracking "origin/main"
    And I am 3 commits ahead and 1 commit behind
    When I run "structured-cli git status --json"
    Then the JSON should contain "upstream" equal to "origin/main"
    And the JSON should contain "ahead" equal to 3
    And the JSON should contain "behind" equal to 1

  Scenario: git log - basic
    Given I have a git repository with 10 commits
    When I run "structured-cli git log -n 5 --json"
    Then the JSON should contain "commits" as an array with 5 items
    And each commit should have "hash", "author", "email", "date", "message"

  Scenario: git log - with stats
    Given I have a git repository with commits that modify files
    When I run "structured-cli git log --stat -n 1 --json"
    Then the JSON "commits[0]" should contain "files" as an array
    And each file should have "path", "additions", "deletions"

  Scenario: git diff - unstaged changes
    Given I have modified "README.md" without staging
    When I run "structured-cli git diff --json"
    Then the JSON should contain "files" as an array
    And each file should have "path", "additions", "deletions", "hunks"

  Scenario: git diff - staged changes
    Given I have staged changes to "src/index.ts"
    When I run "structured-cli git diff --staged --json"
    Then the JSON should contain "files" as an array with 1 item
    And the file should have "path" equal to "src/index.ts"

  Scenario: git branch - list branches
    Given I have branches "main", "feature/auth", "bugfix/login"
    When I run "structured-cli git branch --json"
    Then the JSON should contain "branches" as an array with 3 items
    And each branch should have "name", "current", "upstream"
    And one branch should have "current" equal to true

  Scenario: git branch - verbose with tracking
    Given I have branch "feature" tracking "origin/feature"
    When I run "structured-cli git branch -vv --json"
    Then the JSON branch entry should have "upstream", "ahead", "behind"

  Scenario: git show - commit details
    Given I have a commit with hash "abc123"
    When I run "structured-cli git show abc123 --json"
    Then the JSON should contain "hash", "author", "date", "message"
    And the JSON should contain "diff" with file changes

  Scenario: git add - stage files
    Given I have modified files "a.txt" and "b.txt"
    When I run "structured-cli git add a.txt --json"
    Then the JSON should contain "staged" as an array with "a.txt"

  Scenario: git commit - create commit
    Given I have staged changes
    When I run "structured-cli git commit -m 'test commit' --json"
    Then the JSON should contain "hash" as a string
    And the JSON should contain "message" equal to "test commit"
    And the JSON should contain "filesChanged" as an integer

  Scenario: git push - push to remote
    Given I have commits to push to "origin/main"
    When I run "structured-cli git push --json"
    Then the JSON should contain "success" as a boolean
    And the JSON should contain "remote", "branch", "commits"

  Scenario: git pull - pull from remote
    Given remote "origin/main" has new commits
    When I run "structured-cli git pull --json"
    Then the JSON should contain "success" as a boolean
    And the JSON should contain "commits" as an array
    And the JSON should contain "filesChanged"

  Scenario: git checkout - switch branch
    Given I have branch "feature/auth"
    When I run "structured-cli git checkout feature/auth --json"
    Then the JSON should contain "branch" equal to "feature/auth"
    And the JSON should contain "success" equal to true

  Scenario: git blame - file attribution
    Given I have a file "src/main.go" with 50 lines
    When I run "structured-cli git blame src/main.go --json"
    Then the JSON should contain "lines" as an array
    And each line should have "lineNumber", "hash", "author", "date", "content"

  Scenario: git reflog - reference log
    When I run "structured-cli git reflog -n 10 --json"
    Then the JSON should contain "entries" as an array
    And each entry should have "hash", "action", "message", "date"
```

### Feature: Go Commands

```gherkin
Feature: Go Commands
  As an AI coding agent
  I want structured output from Go commands
  So that I can manage Go projects programmatically

  Scenario: go build - successful build
    Given I have a valid Go project
    When I run "structured-cli go build ./... --json"
    Then the JSON should contain "success" equal to true
    And the JSON should contain "packages" as an array

  Scenario: go build - with errors
    Given I have a Go project with compilation errors
    When I run "structured-cli go build ./... --json"
    Then the JSON should contain "success" equal to false
    And the JSON should contain "errors" as an array
    And each error should have "file", "line", "column", "message"

  Scenario: go test - run tests
    Given I have a Go project with tests
    When I run "structured-cli go test ./... --json"
    Then the JSON should contain "passed", "failed", "skipped"
    And the JSON should contain "packages" as an array

  Scenario: go test - with coverage
    Given I have a Go project with tests
    When I run "structured-cli go test -cover ./... --json"
    Then the JSON should contain "coverage" as an object
    And coverage should have percentage per package

  Scenario: go vet - static analysis
    Given I have a Go project
    When I run "structured-cli go vet ./... --json"
    Then the JSON should contain "issues" as an array
    And each issue should have "file", "line", "message"

  Scenario: go run - execute program
    Given I have a Go file with main function
    When I run "structured-cli go run main.go --json"
    Then the JSON should contain "exitCode" as an integer
    And the JSON should contain "stdout" as a string

  Scenario: go mod tidy - clean dependencies
    Given I have a Go project with unused dependencies
    When I run "structured-cli go mod tidy --json"
    Then the JSON should contain "added" as an array
    And the JSON should contain "removed" as an array

  Scenario: go fmt - format check
    Given I have a Go project
    When I run "structured-cli gofmt -l . --json"
    Then the JSON should contain "unformatted" as an array of file paths

  Scenario: go generate - run generators
    Given I have a Go project with generate directives
    When I run "structured-cli go generate ./... --json"
    Then the JSON should contain "success" equal to true
    And the JSON should contain "generated" as an array
```

### Feature: Build Tool Commands

```gherkin
Feature: Build Tool Commands
  As an AI coding agent
  I want structured output from build tools
  So that I can analyze build results and errors

  Scenario: tsc - successful compilation
    Given I have a TypeScript project with no errors
    When I run "structured-cli tsc --json"
    Then the JSON should contain "success" equal to true
    And the JSON should contain "errors" as an empty array

  Scenario: tsc - compilation errors
    Given I have a TypeScript project with type errors
    When I run "structured-cli tsc --json"
    Then the JSON should contain "success" equal to false
    And the JSON should contain "errors" as an array
    And each error should have "file", "line", "column", "code", "message"

  Scenario: esbuild - successful build
    Given I have a valid esbuild configuration
    When I run "structured-cli esbuild src/index.ts --bundle --json"
    Then the JSON should contain "success" equal to true
    And the JSON should contain "outputs" as an array
    And the JSON should contain "duration" as a number

  Scenario: esbuild - build errors
    Given I have esbuild with import errors
    When I run "structured-cli esbuild src/index.ts --bundle --json"
    Then the JSON should contain "success" equal to false
    And the JSON should contain "errors" as an array

  Scenario: vite build - production build
    Given I have a Vite project
    When I run "structured-cli vite build --json"
    Then the JSON should contain "success" equal to true
    And the JSON should contain "outputs" with file sizes
    And the JSON should contain "duration" as a number

  Scenario: webpack - successful build
    Given I have a webpack configuration
    When I run "structured-cli webpack --json"
    Then the JSON should contain "success" equal to true
    And the JSON should contain "assets" as an array
    And each asset should have "name", "size"

  Scenario: go build - successful build
    Given I have a Go project
    When I run "structured-cli go build ./... --json"
    Then the JSON should contain "success" equal to true
    And the JSON should contain "packages" as an array

  Scenario: go build - compilation errors
    Given I have a Go project with syntax errors
    When I run "structured-cli go build ./... --json"
    Then the JSON should contain "success" equal to false
    And the JSON should contain "errors" as an array
    And each error should have "file", "line", "column", "message"

  Scenario: cargo build - successful build
    Given I have a Rust project
    When I run "structured-cli cargo build --json"
    Then the JSON should contain "success" equal to true
    And the JSON should contain "target" as a string

  Scenario: cargo build - compilation errors
    Given I have a Rust project with errors
    When I run "structured-cli cargo build --json"
    Then the JSON should contain "success" equal to false
    And the JSON should contain "errors" as an array
    And each error should have "level", "message", "file", "line"
```

### Feature: Linter Commands

```gherkin
Feature: Linter Commands
  As an AI coding agent
  I want structured output from linters
  So that I can fix code issues programmatically

  Scenario: eslint - no issues
    Given I have a JavaScript project with no lint errors
    When I run "structured-cli eslint src/ --json"
    Then the JSON should contain "errorCount" equal to 0
    And the JSON should contain "warningCount" equal to 0
    And the JSON should contain "files" as an array

  Scenario: eslint - with issues
    Given I have a JavaScript project with lint errors
    When I run "structured-cli eslint src/ --json"
    Then the JSON should contain "errorCount" greater than 0
    And the JSON should contain "issues" as an array
    And each issue should have "file", "line", "column", "rule", "message", "severity"

  Scenario: prettier - format check pass
    Given I have properly formatted files
    When I run "structured-cli prettier --check src/ --json"
    Then the JSON should contain "formatted" equal to true
    And the JSON should contain "files" as an array

  Scenario: prettier - format check fail
    Given I have unformatted files
    When I run "structured-cli prettier --check src/ --json"
    Then the JSON should contain "formatted" equal to false
    And the JSON should contain "unformatted" as an array of file paths

  Scenario: biome check - lint and format
    Given I have a project using Biome
    When I run "structured-cli biome check src/ --json"
    Then the JSON should contain "errors", "warnings"
    And the JSON should contain "issues" as an array

  Scenario: golangci-lint - Go linting
    Given I have a Go project
    When I run "structured-cli golangci-lint run --json"
    Then the JSON should contain "issues" as an array
    And each issue should have "file", "line", "linter", "message"

  Scenario: ruff check - Python linting
    Given I have a Python project
    When I run "structured-cli ruff check . --json"
    Then the JSON should contain "issues" as an array
    And each issue should have "file", "line", "code", "message"

  Scenario: mypy - Python type checking
    Given I have a Python project with type annotations
    When I run "structured-cli mypy src/ --json"
    Then the JSON should contain "errors" as an array
    And each error should have "file", "line", "severity", "message"
```

### Feature: Test Runner Commands

```gherkin
Feature: Test Runner Commands
  As an AI coding agent
  I want structured output from test runners
  So that I can analyze test results programmatically

  Scenario: pytest - all tests pass
    Given I have a Python project with 20 passing tests
    When I run "structured-cli pytest --json"
    Then the JSON should contain "passed" equal to 20
    And the JSON should contain "failed" equal to 0
    And the JSON should contain "success" equal to true
    And the JSON should contain "duration" as a number

  Scenario: pytest - some tests fail
    Given I have a Python project with 18 passing and 2 failing tests
    When I run "structured-cli pytest --json"
    Then the JSON should contain "passed" equal to 18
    And the JSON should contain "failed" equal to 2
    And the JSON should contain "failures" as an array with 2 items
    And each failure should have "test", "file", "line", "message"

  Scenario: pytest - with coverage
    Given I have a Python project with pytest-cov configured
    When I run "structured-cli pytest --cov --json"
    Then the JSON should contain "coverage" as an object
    And coverage should have "total" as a percentage
    And coverage should have "files" as an array

  Scenario: jest - all tests pass
    Given I have a JavaScript project with 30 passing tests
    When I run "structured-cli jest --json"
    Then the JSON should contain "passed" equal to 30
    And the JSON should contain "failed" equal to 0
    And the JSON should contain "suites" as an array

  Scenario: jest - some tests fail
    Given I have failing Jest tests
    When I run "structured-cli jest --json"
    Then the JSON should contain "failures" as an array
    And each failure should have "suite", "test", "message", "stack"

  Scenario: vitest - test results
    Given I have a Vite project with tests
    When I run "structured-cli vitest run --json"
    Then the JSON should contain "passed", "failed", "skipped"
    And the JSON should contain "files" as an array of test files

  Scenario: mocha - test results
    Given I have a Node.js project with Mocha tests
    When I run "structured-cli mocha --json"
    Then the JSON should contain "passed", "failed", "pending"
    And the JSON should contain "suites" as an array

  Scenario: go test - all tests pass
    Given I have a Go project with 15 passing tests
    When I run "structured-cli go test ./... --json"
    Then the JSON should contain "passed" equal to 15
    And the JSON should contain "packages" as an array

  Scenario: go test - with coverage
    Given I have a Go project with tests
    When I run "structured-cli go test -cover ./... --json"
    Then the JSON should contain "coverage" with percentage per package

  Scenario: cargo test - Rust tests
    Given I have a Rust project with tests
    When I run "structured-cli cargo test --json"
    Then the JSON should contain "passed", "failed", "ignored"
    And the JSON should contain "tests" as an array
```

### Feature: Schema Validation

```gherkin
Feature: Schema Validation
  As a developer
  I want parser output validated against schemas
  So that I can trust the JSON structure

  Scenario: Valid output passes schema validation
    Given I have a git repository with changes
    When I run "structured-cli git status --json"
    Then the output should conform to "schemas/git-status.json"

  Scenario: Schema validation catches parser bugs
    Given a parser returns invalid data for git status
    When schema validation runs
    Then an error should be raised
    And the error should indicate which field failed validation

  Scenario: All parsers have corresponding schemas
    Given I list all registered parsers
    Then each parser should have a schema file in "schemas/"
    And each schema should be valid JSON Schema draft 2020-12
```

### Feature: Passthrough Mode

```gherkin
Feature: Passthrough Mode
  As a developer
  I want structured-cli to pass through raw output by default
  So that I can use it as a drop-in replacement for CLI tools

  Scenario: Default passthrough without --json flag
    Given I have a git repository with uncommitted changes
    When I run "structured-cli git status"
    Then the output should match raw "git status" output exactly
    And the exit code should match the underlying command

  Scenario: Passthrough with STRUCTURED_CLI_JSON=false
    Given I have set STRUCTURED_CLI_JSON=false
    When I run "structured-cli git status"
    Then the output should match raw "git status" output exactly

  Scenario: Alias usage in passthrough mode
    Given I have aliased git to "structured-cli git"
    When I run "git status"
    Then the output should match raw "git status" output exactly
```

### Feature: JSON Output Mode

```gherkin
Feature: JSON Output Mode
  As an AI coding agent
  I want structured JSON output from CLI commands
  So that I can parse results without regex

  Scenario: JSON output with --json flag
    Given I have a git repository on branch "main" with staged file "src/index.ts"
    When I run "structured-cli git status --json"
    Then the output should be valid JSON
    And the JSON should contain "branch" equal to "main"
    And the JSON should contain "staged" as an array with 1 item

  Scenario: JSON output with environment variable
    Given I have set STRUCTURED_CLI_JSON=true
    When I run "structured-cli git status"
    Then the output should be valid JSON

  Scenario: --json flag overrides STRUCTURED_CLI_JSON=false
    Given I have set STRUCTURED_CLI_JSON=false
    When I run "structured-cli git status --json"
    Then the output should be valid JSON
```

### Feature: Error Handling

```gherkin
Feature: Error Handling
  As an AI coding agent
  I want errors returned as JSON when in JSON mode
  So that I can handle failures programmatically

  Scenario: Command failure in JSON mode
    Given I am not in a git repository
    When I run "structured-cli git status --json"
    Then the output should be valid JSON
    And the JSON should contain "error" with value "fatal: not a git repository"
    And the JSON should contain "exitCode" equal to 128
    And the process exit code should be 128

  Scenario: Command failure in passthrough mode
    Given I am not in a git repository
    When I run "structured-cli git status"
    Then the output should match raw "git status" error output
    And the process exit code should be 128

  Scenario: Parser failure in JSON mode
    Given git status returns unexpected output format
    When I run "structured-cli git status --json"
    Then the output should be valid JSON
    And the JSON should contain "error" starting with "parser error:"
    And the JSON should contain "raw" with the original output
```

### Feature: Unsupported Commands

```gherkin
Feature: Unsupported Commands
  As a developer
  I want graceful handling of commands without parsers
  So that the tool doesn't break on obscure subcommands

  Scenario: Unsupported subcommand in JSON mode
    Given there is no parser for "git stash show"
    When I run "structured-cli git stash show --json"
    Then the output should be valid JSON
    And the JSON should contain "raw" with the command output
    And the JSON should contain "parsed" equal to false

  Scenario: Unsupported subcommand in passthrough mode
    Given there is no parser for "git stash show"
    When I run "structured-cli git stash show"
    Then the output should match raw "git stash show" output exactly
```

### Feature: Native JSON Passthrough

```gherkin
Feature: Native JSON Passthrough
  As an AI coding agent
  I want consistent JSON structure even from tools with native JSON
  So that I can rely on a single schema per command

  Scenario: GitHub CLI native JSON is normalized
    Given I have a GitHub repository with pull requests
    When I run "structured-cli gh pr list --json"
    Then the output should be valid JSON
    And the JSON should conform to the structured-cli gh-pr-list schema

  Scenario: kubectl native JSON is normalized
    Given I have a Kubernetes cluster with pods
    When I run "structured-cli kubectl get pods --json"
    Then the output should be valid JSON
    And the JSON should conform to the structured-cli kubectl-get-pods schema

  Scenario: npm native JSON is normalized
    Given I have a package.json with dependencies
    When I run "structured-cli npm ls --json"
    Then the output should be valid JSON
    And the JSON should conform to the structured-cli npm-list schema
```

### Feature: Streaming Command Buffering

```gherkin
Feature: Streaming Command Buffering
  As an AI coding agent
  I want streaming commands to return complete JSON
  So that I get a single parseable result

  Scenario: Docker build buffers output
    Given I have a Dockerfile in the current directory
    When I run "structured-cli docker build . --json"
    Then the command should wait for docker build to complete
    And the output should be valid JSON
    And the JSON should contain "success" as a boolean
    And the JSON should contain "imageId" if successful

  Scenario: npm install buffers output
    Given I have a package.json with dependencies
    When I run "structured-cli npm install --json"
    Then the command should wait for npm install to complete
    And the output should be valid JSON
    And the JSON should contain "installed" as an array

  Scenario: cargo build buffers output
    Given I have a Cargo.toml with dependencies
    When I run "structured-cli cargo build --json"
    Then the command should wait for cargo build to complete
    And the output should be valid JSON
    And the JSON should contain "success" as a boolean
```

### Feature: NPM Commands

```gherkin
Feature: NPM Commands
  As an AI coding agent
  I want structured output from npm commands
  So that I can manage dependencies programmatically

  Scenario: npm install - fresh install
    Given I have a package.json with 5 dependencies
    When I run "structured-cli npm install --json"
    Then the JSON should contain "installed" as an array
    And the JSON should contain "added" as an integer
    And the JSON should contain "duration" as a string

  Scenario: npm install - with warnings
    Given I have a package.json with peer dependency issues
    When I run "structured-cli npm install --json"
    Then the JSON should contain "warnings" as an array
    And each warning should have "type" and "message"

  Scenario: npm audit - vulnerabilities found
    Given I have dependencies with known vulnerabilities
    When I run "structured-cli npm audit --json"
    Then the JSON should contain "vulnerabilities" as an array
    And each vulnerability should have "severity", "package", "title", "url"
    And the JSON should contain "summary" with counts by severity

  Scenario: npm audit - clean
    Given I have no vulnerable dependencies
    When I run "structured-cli npm audit --json"
    Then the JSON should contain "vulnerabilities" as an empty array
    And the JSON "summary.total" should equal 0

  Scenario: npm outdated - packages need updates
    Given I have outdated dependencies
    When I run "structured-cli npm outdated --json"
    Then the JSON should contain "packages" as an array
    And each package should have "name", "current", "wanted", "latest"

  Scenario: npm list - dependency tree
    Given I have a project with nested dependencies
    When I run "structured-cli npm list --json"
    Then the JSON should contain "dependencies" as an object
    And each dependency should have "version" and optionally "dependencies"

  Scenario: npm list - flat with depth
    When I run "structured-cli npm list --depth=0 --json"
    Then the JSON should contain only top-level dependencies

  Scenario: npm run - script execution
    Given I have a script "build" in package.json
    When I run "structured-cli npm run build --json"
    Then the JSON should contain "script" equal to "build"
    And the JSON should contain "exitCode" as an integer
    And the JSON should contain "duration" as a string

  Scenario: npm test - test execution
    Given I have a test script in package.json
    When I run "structured-cli npm test --json"
    Then the JSON should contain "exitCode" as an integer
    And the JSON should contain "output" as a string

  Scenario: npm init - create package.json
    Given I am in an empty directory
    When I run "structured-cli npm init -y --json"
    Then the JSON should contain "created" equal to true
    And the JSON should contain "package" with name and version
```

### Feature: Docker Commands

```gherkin
Feature: Docker Commands
  As an AI coding agent
  I want structured output from docker commands
  So that I can manage containers programmatically

  Scenario: docker ps - list containers
    Given I have 3 running containers
    When I run "structured-cli docker ps --json"
    Then the JSON should contain "containers" as an array with 3 items
    And each container should have "id", "image", "status", "ports", "names"

  Scenario: docker ps - include stopped
    Given I have 2 running and 1 stopped container
    When I run "structured-cli docker ps -a --json"
    Then the JSON should contain "containers" as an array with 3 items

  Scenario: docker build - successful build
    Given I have a valid Dockerfile
    When I run "structured-cli docker build -t myapp . --json"
    Then the JSON should contain "success" equal to true
    And the JSON should contain "imageId" as a string
    And the JSON should contain "tags" containing "myapp"
    And the JSON should contain "steps" as an array

  Scenario: docker build - build failure
    Given I have a Dockerfile with an error
    When I run "structured-cli docker build . --json"
    Then the JSON should contain "success" equal to false
    And the JSON should contain "error" with failure details
    And the JSON should contain "failedStep" as an integer

  Scenario: docker logs - container logs
    Given I have a running container "myapp"
    When I run "structured-cli docker logs myapp --json"
    Then the JSON should contain "logs" as an array of log entries
    And each entry should have "timestamp", "stream", "message"

  Scenario: docker logs - with tail
    Given I have a container with 1000 log lines
    When I run "structured-cli docker logs --tail 10 myapp --json"
    Then the JSON "logs" should have 10 items

  Scenario: docker images - list images
    Given I have 5 docker images
    When I run "structured-cli docker images --json"
    Then the JSON should contain "images" as an array with 5 items
    And each image should have "id", "repository", "tag", "size", "created"

  Scenario: docker run - run container
    Given I have image "nginx:latest"
    When I run "structured-cli docker run -d nginx:latest --json"
    Then the JSON should contain "containerId" as a string
    And the JSON should contain "success" equal to true

  Scenario: docker exec - execute in container
    Given I have a running container "myapp"
    When I run "structured-cli docker exec myapp ls -la --json"
    Then the JSON should contain "exitCode" as an integer
    And the JSON should contain "stdout" as a string

  Scenario: docker pull - pull image
    When I run "structured-cli docker pull nginx:latest --json"
    Then the JSON should contain "image" equal to "nginx:latest"
    And the JSON should contain "digest" as a string
    And the JSON should contain "success" equal to true

  Scenario: docker compose up - start services
    Given I have a docker-compose.yml with 3 services
    When I run "structured-cli docker compose up -d --json"
    Then the JSON should contain "services" as an array with 3 items
    And each service should have "name", "status", "containerId"

  Scenario: docker compose down - stop services
    Given I have running compose services
    When I run "structured-cli docker compose down --json"
    Then the JSON should contain "stopped" as an array of service names
    And the JSON should contain "removed" as an array

  Scenario: docker compose ps - list compose services
    Given I have running compose services
    When I run "structured-cli docker compose ps --json"
    Then the JSON should contain "services" as an array
    And each service should have "name", "status", "ports"
```

### Feature: Python Commands

```gherkin
Feature: Python Commands
  As an AI coding agent
  I want structured output from Python tools
  So that I can manage Python projects programmatically

  Scenario: pip install - install packages
    Given I have a requirements.txt with 5 packages
    When I run "structured-cli pip install -r requirements.txt --json"
    Then the JSON should contain "installed" as an array
    And each package should have "name", "version"
    And the JSON should contain "success" equal to true

  Scenario: pip audit - security check
    Given I have installed packages with vulnerabilities
    When I run "structured-cli pip-audit --json"
    Then the JSON should contain "vulnerabilities" as an array
    And each vulnerability should have "package", "version", "id", "description"

  Scenario: uv install - fast install
    Given I have a pyproject.toml
    When I run "structured-cli uv pip install -r requirements.txt --json"
    Then the JSON should contain "installed" as an array
    And the JSON should contain "duration" as a number

  Scenario: uv run - run script
    Given I have a Python script
    When I run "structured-cli uv run python script.py --json"
    Then the JSON should contain "exitCode" as an integer
    And the JSON should contain "stdout" as a string

  Scenario: black - format check
    Given I have Python files
    When I run "structured-cli black --check . --json"
    Then the JSON should contain "wouldReformat" as an array
    And the JSON should contain "unchanged" as an array
```

### Feature: Cargo (Rust) Commands

```gherkin
Feature: Cargo Commands
  As an AI coding agent
  I want structured output from Cargo commands
  So that I can manage Rust projects programmatically

  Scenario: cargo build - successful build
    Given I have a valid Rust project
    When I run "structured-cli cargo build --json"
    Then the JSON should contain "success" equal to true
    And the JSON should contain "target" as a string
    And the JSON should contain "profile" as a string

  Scenario: cargo test - run tests
    Given I have a Rust project with tests
    When I run "structured-cli cargo test --json"
    Then the JSON should contain "passed", "failed", "ignored"
    And the JSON should contain "tests" as an array

  Scenario: cargo clippy - lint check
    Given I have a Rust project
    When I run "structured-cli cargo clippy --json"
    Then the JSON should contain "warnings" as an array
    And each warning should have "message", "file", "line", "code"

  Scenario: cargo run - execute binary
    Given I have a Rust project with a main function
    When I run "structured-cli cargo run --json"
    Then the JSON should contain "exitCode" as an integer
    And the JSON should contain "stdout" as a string

  Scenario: cargo add - add dependency
    Given I have a Rust project
    When I run "structured-cli cargo add serde --json"
    Then the JSON should contain "added" as an object
    And added should have "name" equal to "serde"
    And added should have "version" as a string

  Scenario: cargo remove - remove dependency
    Given I have a Rust project with "serde" dependency
    When I run "structured-cli cargo remove serde --json"
    Then the JSON should contain "removed" equal to "serde"
    And the JSON should contain "success" equal to true

  Scenario: cargo fmt - format check
    Given I have a Rust project
    When I run "structured-cli cargo fmt --check --json"
    Then the JSON should contain "formatted" equal to true or false
    And the JSON should contain "files" as an array

  Scenario: cargo doc - generate docs
    Given I have a Rust project
    When I run "structured-cli cargo doc --json"
    Then the JSON should contain "success" equal to true
    And the JSON should contain "outputDir" as a string

  Scenario: cargo check - type check
    Given I have a Rust project
    When I run "structured-cli cargo check --json"
    Then the JSON should contain "success" equal to true or false
    And the JSON should contain "errors" as an array
```

### Feature: Make/Just Commands

```gherkin
Feature: Make/Just Commands
  As an AI coding agent
  I want structured output from make and just commands
  So that I can run build tasks programmatically

  Scenario: make - successful build
    Given I have a Makefile with target "build"
    When I run "structured-cli make build --json"
    Then the JSON should contain "success" equal to true
    And the JSON should contain "target" equal to "build"
    And the JSON should contain "duration" as a number

  Scenario: make - build failure
    Given I have a Makefile with a failing target
    When I run "structured-cli make broken --json"
    Then the JSON should contain "success" equal to false
    And the JSON should contain "error" as a string
    And the JSON should contain "exitCode" as an integer

  Scenario: make - list targets
    Given I have a Makefile with multiple targets
    When I run "structured-cli make -pn --json"
    Then the JSON should contain "targets" as an array
    And each target should have "name" and "dependencies"

  Scenario: make - dry run
    Given I have a Makefile
    When I run "structured-cli make -n build --json"
    Then the JSON should contain "commands" as an array
    And the JSON should contain "wouldExecute" equal to true

  Scenario: just - successful recipe
    Given I have a justfile with recipe "build"
    When I run "structured-cli just build --json"
    Then the JSON should contain "success" equal to true
    And the JSON should contain "recipe" equal to "build"
    And the JSON should contain "duration" as a number

  Scenario: just - recipe failure
    Given I have a justfile with a failing recipe
    When I run "structured-cli just broken --json"
    Then the JSON should contain "success" equal to false
    And the JSON should contain "error" as a string

  Scenario: just - list recipes
    Given I have a justfile with multiple recipes
    When I run "structured-cli just --list --json"
    Then the JSON should contain "recipes" as an array
    And each recipe should have "name", "description", "parameters"

  Scenario: just - dry run
    Given I have a justfile
    When I run "structured-cli just --dry-run build --json"
    Then the JSON should contain "commands" as an array
    And the JSON should contain "wouldExecute" equal to true
```

### Feature: File Operations Commands

```gherkin
Feature: File Operations Commands
  As an AI coding agent
  I want structured output from file operations
  So that I can navigate and search filesystems programmatically

  Scenario: ls - list directory
    Given I have a directory with files and subdirectories
    When I run "structured-cli ls -la --json"
    Then the JSON should contain "entries" as an array
    And each entry should have "name", "type", "size", "permissions", "modified"
    And "type" should be "file", "directory", or "symlink"

  Scenario: ls - specific path
    Given I have a directory "/tmp/myproject"
    When I run "structured-cli ls /tmp/myproject --json"
    Then the JSON should contain "path" equal to "/tmp/myproject"
    And the JSON should contain "entries" as an array

  Scenario: find - search by name
    Given I have a directory tree with various files
    When I run "structured-cli find . -name '*.go' --json"
    Then the JSON should contain "matches" as an array
    And each match should have "path", "type", "size"

  Scenario: find - search by type
    Given I have a directory tree with files and directories
    When I run "structured-cli find . -type d --json"
    Then the JSON should contain "matches" as an array
    And each match should have "type" equal to "directory"

  Scenario: find - search with exec
    Given I have a directory tree
    When I run "structured-cli find . -name '*.txt' -exec wc -l {} \; --json"
    Then the JSON should contain "matches" as an array
    And each match should have "path" and "execResult"

  Scenario: grep - search in files
    Given I have files containing text
    When I run "structured-cli grep -rn 'TODO' src/ --json"
    Then the JSON should contain "matches" as an array
    And each match should have "file", "line", "column", "content"

  Scenario: grep - no matches
    Given I have files without the search term
    When I run "structured-cli grep 'nonexistent' src/ --json"
    Then the JSON should contain "matches" as an empty array
    And the JSON should contain "matchCount" equal to 0

  Scenario: grep - count only
    Given I have files containing text
    When I run "structured-cli grep -c 'function' src/*.js --json"
    Then the JSON should contain "counts" as an object
    And each key should be a filename with count as value

  Scenario: ripgrep - search in files
    Given I have a codebase
    When I run "structured-cli rg 'TODO' --json"
    Then the JSON should contain "matches" as an array
    And each match should have "file", "line", "column", "content"

  Scenario: ripgrep - search with context
    Given I have a codebase
    When I run "structured-cli rg -C 2 'error' --json"
    Then the JSON should contain "matches" as an array
    And each match should have "beforeContext", "content", "afterContext"

  Scenario: ripgrep - search with file type filter
    Given I have a codebase with multiple languages
    When I run "structured-cli rg -t go 'func' --json"
    Then the JSON should contain "matches" as an array
    And each match "file" should end with ".go"

  Scenario: ripgrep - stats only
    Given I have a codebase
    When I run "structured-cli rg --stats 'import' --json"
    Then the JSON should contain "stats" with "matches", "files", "lines"

  Scenario: fd - find files by name
    Given I have a directory tree
    When I run "structured-cli fd '\.go$' --json"
    Then the JSON should contain "matches" as an array of file paths

  Scenario: fd - find with type filter
    Given I have a directory tree
    When I run "structured-cli fd -t d 'test' --json"
    Then the JSON should contain "matches" as an array
    And each match should be a directory

  Scenario: fd - find with exclusions
    Given I have a directory tree with node_modules
    When I run "structured-cli fd -E node_modules '\.js$' --json"
    Then the JSON should contain "matches" as an array
    And no match should contain "node_modules"

  Scenario: cat - read file contents
    Given I have a file "README.md"
    When I run "structured-cli cat README.md --json"
    Then the JSON should contain "file" equal to "README.md"
    And the JSON should contain "content" as a string
    And the JSON should contain "lines" as an integer
    And the JSON should contain "bytes" as an integer

  Scenario: head - read first lines
    Given I have a file with 100 lines
    When I run "structured-cli head -n 10 file.txt --json"
    Then the JSON should contain "content" as a string
    And the JSON should contain "linesReturned" equal to 10

  Scenario: tail - read last lines
    Given I have a file with 100 lines
    When I run "structured-cli tail -n 10 file.txt --json"
    Then the JSON should contain "content" as a string
    And the JSON should contain "linesReturned" equal to 10

  Scenario: wc - word count
    Given I have a file "document.txt"
    When I run "structured-cli wc document.txt --json"
    Then the JSON should contain "file" equal to "document.txt"
    And the JSON should contain "lines", "words", "bytes"

  Scenario: wc - multiple files
    Given I have files "a.txt" and "b.txt"
    When I run "structured-cli wc a.txt b.txt --json"
    Then the JSON should contain "files" as an array
    And each file should have "name", "lines", "words", "bytes"
    And the JSON should contain "total" with aggregate counts

  Scenario: du - disk usage
    Given I have a directory with files
    When I run "structured-cli du -sh mydir --json"
    Then the JSON should contain "path" equal to "mydir"
    And the JSON should contain "size" as a string
    And the JSON should contain "bytes" as an integer

  Scenario: du - recursive
    Given I have a directory tree
    When I run "structured-cli du -h mydir --json"
    Then the JSON should contain "entries" as an array
    And each entry should have "path", "size", "bytes"

  Scenario: df - disk free
    When I run "structured-cli df -h --json"
    Then the JSON should contain "filesystems" as an array
    And each filesystem should have "device", "size", "used", "available", "usePercent", "mountpoint"
```

---

## Design Decisions

### Error Handling

When `--json` is used, errors are always JSON:

```bash
# Command failed (non-zero exit):
$ structured-cli git status --json
{"error": "fatal: not a git repository", "exitCode": 128}

# Parser failed (couldn't parse output):
$ structured-cli git status --json
{"error": "parser error: unexpected format", "exitCode": 0, "raw": "..."}
```

Exit code from `structured-cli` mirrors the underlying command's exit code.

### Unsupported Commands

When no parser exists for a command, return raw output in JSON wrapper:

```bash
$ structured-cli git some-obscure-subcommand --json
{"raw": "...", "parsed": false}
```

### Native JSON Passthrough

For tools with native JSON support (`gh --json`, `kubectl -o json`, `npm --json`), we validate and normalize against our schema for consistent structure across all tools.

### Streaming Output

Commands like `docker build`, `npm install`, `cargo build` that produce streaming progress output are **buffered**. JSON is emitted only when the command completes.

This matches Pare's approach and fits request/response semantics. Users who want live progress can run commands without `--json`.

### Schema Versioning

CLI output formats may change across tool versions. Strategy:
- Embed version detection where practical
- Design schemas to be forward-compatible (optional fields)
- Multiple parser implementations per tool if formats diverge significantly

---

## Success Criteria

1. `structured-cli git status --json` returns valid JSON matching schema
2. `structured-cli npm install lodash --json` returns structured install output
3. `alias git="structured-cli git"` works transparently in non-JSON mode
4. Token reduction comparable to Pare (~50-90% for verbose commands)
5. Binary size < 20MB (single static binary)
6. Startup overhead < 50ms
