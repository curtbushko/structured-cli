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
- [x] Implement `services/validator.go` (schema validation)

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
- [x] `git status` parser + schema
- [x] `git log` parser + schema
- [x] `git diff` parser + schema
- [x] `git branch` parser + schema
- [x] `git show` parser + schema
- [x] `git add` parser + schema
- [x] `git commit` parser + schema
- [x] `git push` parser + schema
- [x] `git pull` parser + schema
- [x] `git checkout` parser + schema
- [x] `git blame` parser + schema
- [x] `git reflog` parser + schema

### Phase 7: Go Parsers
- [x] `go build` parser + schema
- [x] `go test` parser + schema
- [x] `go test -cover` parser + schema
- [x] `go vet` parser + schema
- [x] `go run` parser + schema
- [x] `go mod tidy` parser + schema
- [x] `gofmt` parser + schema
- [x] `go generate` parser + schema

### Phase 8: Build Tool Parsers
- [x] `tsc` parser + schema
- [x] `esbuild` parser + schema
- [x] `vite build` parser + schema
- [x] `webpack` parser + schema
- [x] `cargo build` parser + schema

### Phase 9: Linter Parsers
- [x] `eslint` parser + schema
- [x] `prettier --check` parser + schema
- [x] `biome check` parser + schema
- [x] `golangci-lint` parser + schema
- [x] `ruff check` parser + schema
- [x] `mypy` parser + schema

### Phase 10: Test Runner Parsers
- [x] `pytest` parser + schema
- [x] `jest` parser + schema
- [x] `vitest` parser + schema
- [x] `mocha` parser + schema
- [x] `cargo test` parser + schema

### Phase 11: Schema Validation & Error Handling
- [x] Implement schema validation service
- [x] Implement error JSON output (`{"error": "...", "exitCode": N}`)
- [x] Implement unsupported command fallback (`{"raw": "...", "parsed": false}`)
- [x] Implement native JSON passthrough with validation
- [x] Implement streaming command buffering

### Phase 12: NPM Parsers
- [x] `npm install` parser + schema
- [x] `npm audit` parser + schema
- [x] `npm outdated` parser + schema
- [x] `npm list` parser + schema
- [x] `npm run` parser + schema
- [x] `npm test` parser + schema
- [x] `npm init` parser + schema

### Phase 13: Docker Parsers
- [x] `docker ps` parser + schema
- [x] `docker build` parser + schema
- [x] `docker logs` parser + schema
- [x] `docker images` parser + schema
- [x] `docker run` parser + schema
- [x] `docker exec` parser + schema
- [x] `docker pull` parser + schema
- [x] `docker compose up` parser + schema
- [x] `docker compose down` parser + schema
- [x] `docker compose ps` parser + schema

### Phase 14: Python Parsers
- [x] `pip install` parser + schema
- [x] `pip-audit` parser + schema
- [x] `uv pip install` parser + schema
- [x] `uv run` parser + schema
- [x] `black --check` parser + schema

### Phase 15: Cargo (Rust) Parsers
- [x] `cargo build` parser + schema
- [x] `cargo test` parser + schema
- [x] `cargo clippy` parser + schema
- [x] `cargo run` parser + schema
- [x] `cargo add` parser + schema
- [x] `cargo remove` parser + schema
- [x] `cargo fmt` parser + schema
- [x] `cargo doc` parser + schema
- [x] `cargo check` parser + schema

### Phase 16: Make/Just Parsers
- [x] `make` parser + schema
- [x] `just` parser + schema

### Phase 17: File Operations Parsers
- [x] `ls` parser + schema
- [x] `find` parser + schema
- [x] `grep` parser + schema
- [x] `ripgrep (rg)` parser + schema
- [x] `fd` parser + schema
- [x] `cat` parser + schema
- [x] `head` / `tail` parser + schema
- [x] `wc` parser + schema
- [x] `du` parser + schema
- [x] `df` parser + schema

### Phase 18: Polish & Release
- [x] Comprehensive test coverage (>80%)
- [x] Documentation (README, CLAUDE.md)

### Phase 19: End-to-End BDD Tests
Real-world e2e tests using godog with actual CLI tools and file systems.

#### Infrastructure
- [x] godog BDD framework setup
- [x] Test binary build automation
- [x] Temporary directory/repo management
- [x] Step definitions for common operations
- [x] Add required tools to flake.nix for CI

#### Git Commands E2E
- [x] `git status` - clean repo, untracked files, modified files
- [x] `git log` - basic output, commit fields, empty repo
- [x] `git diff` - unstaged changes, staged changes, hunks
- [x] `git branch` - list branches, current branch detection
- [x] `git show` - commit details
- [x] `git blame` - file attribution
- [x] `git reflog` - reference log

#### Output Modes E2E
- [x] Passthrough mode (default)
- [x] JSON output with `--json` flag
- [x] JSON output with environment variable
- [x] `--json` flag position (before/after command)

#### Unsupported Commands E2E
- [x] Fallback JSON for unsupported subcommands
- [x] Passthrough for unsupported commands

#### Error Handling E2E
- [x] Command failure in JSON mode
- [x] Command failure in passthrough mode
- [x] Parser failure with raw output fallback

#### File Operations E2E
- [x] `ls` - list directory, specific path
- [x] `cat` - read file contents
- [x] `head`/`tail` - read first/last lines
- [x] `wc` - word count
- [x] `find` - search by name, type
- [x] `grep` - search in files
- [x] `du` - disk usage
- [x] `df` - disk free

#### Go Commands E2E (if go available)
- [x] `go build` - successful build, build errors
- [x] `go test` - run tests, with coverage
- [x] `go vet` - static analysis
- [x] `go fmt` - format check

#### NPM Commands E2E (if npm available)
- [x] `npm list` - dependency tree
- [x] `npm outdated` - outdated packages

#### Docker Commands E2E (if docker available)
- [x] `docker ps` - list containers
- [x] `docker images` - list images

#### Make/Just Commands E2E
- [x] `make` - successful build, target listing
- [x] `just` - recipe execution, listing

### Phase 20: SQLite Usage Tracking System
Token usage tracking inspired by RTK (Rust Token Killer) for analytics and measuring value.

#### Domain Layer
- [x] Define `domain/tracking.go` (CommandRecord, ParseFailure types)
- [x] Define `domain/stats.go` (StatsSummary, CommandStats types)

#### Ports Layer
- [x] Define `ports/tracker.go` (Tracker interface with Record, GetStats, Cleanup methods)

#### Adapters - SQLite Implementation
- [x] Implement `adapters/tracking/sqlite.go` (SQLite tracker)
- [x] Implement `adapters/tracking/noop.go` (No-op tracker for testing/disabled)
- [x] XDG Base Directory support (`~/.local/share/structured-cli/tracking.db`)
- [x] Database schema creation (commands table, parse_failures table)
- [x] 90-day automatic retention cleanup on insert
- [x] Token estimation (chars/4 heuristic)

#### Application Layer Integration
- [x] Add TimedExecution pattern to executor
- [x] Track successful parses (command, tokens, savings, exec time)
- [x] Track parse failures (command, error, fallback success)
- [x] Wire tracker into composition root

#### Stats Subcommand
- [x] Implement `stats` subcommand in CLI handler
- [x] Summary report (total commands, tokens saved, avg savings %)
- [x] `--history` flag for recent commands
- [x] `--json` flag for JSON export
- [x] `--by-parser` flag for per-parser breakdown
- [x] `--project` flag for current directory only

#### E2E Tests
- [x] Tracking records commands after JSON output
- [x] Tracking calculates token savings correctly
- [x] Stats command shows summary
- [x] Stats --history shows recent commands
- [x] Stats --json outputs valid JSON
- [x] 90-day cleanup removes old records
- [x] Disabled tracking (STRUCTURED_CLI_NO_TRACKING=1)

### Phase 21: Output Deduplication System
Generic deduplication layer to reduce token usage by collapsing identical items.

#### Design Decision: Two-Stage Generic Deduplication

**Approach:** Pure generic deduplication at two stages:

1. **Raw text stage** - Before parsing, collapse identical lines in raw output
2. **JSON object stage** - After parsing, collapse identical objects within same array level

**Why generic instead of per-command?**
- Single implementation, zero per-parser maintenance
- Consistent behavior across all commands
- Objects must be truly identical (all fields match) to deduplicate
- Different objects naturally preserved (lint error on line 10 ≠ line 20)

**Same-level rule:**
- Objects are only compared within their own array
- Nested arrays are processed independently
- Prevents false deduplication across different contexts

#### Domain Layer
- [x] Define `domain/dedup.go` (DedupConfig, DedupResult types)

#### Ports Layer
- [x] Define `ports/deduplicator.go` (Deduplicator interface)

#### Application Layer
- [x] Implement `application/dedup.go` (deduplication engine)
- [x] **Stage 1: Raw text dedup** - deduplicate identical lines before parsing
- [x] **Stage 2: JSON object dedup** - deduplicate identical objects at same array level
- [x] Objects must be at the same level in the JSON tree to be deduplicated
- [x] Add count field to grouped items (`"count": N`)
- [x] Keep first occurrence as sample
- [x] Integrate dedup step into executor pipeline

#### CLI Integration
- [x] Deduplication enabled by default in JSON mode
- [x] Add `--disable-filter=dedupe` flag to disable deduplication
- [x] Add `--disable-filter=all` to disable all filters
- [x] Add `STRUCTURED_CLI_DISABLE_FILTER=dedupe` environment variable
- [x] Support comma-separated filters: `--disable-filter=dedupe,compact`

#### Deduplication Rules
- [x] Raw text: identical lines are collapsed with count
- [x] JSON arrays: identical objects at the same level are collapsed
- [x] Object equality: deep comparison of all fields
- [x] Unique outputs (ls entries, container IDs) naturally don't dedupe
- [x] Repetitive outputs (lint errors, log lines) collapse automatically

#### Output Format
```json
// Default JSON output (dedup ON, 100 items → 3 groups):
// $ structured-cli eslint src/ --json
{"issues": [
  {"rule": "no-unused-vars", "file": "a.js", "line": 10, "count": 45, "sample": "first"},
  {"rule": "semi", "file": "x.js", "line": 5, "count": 30, "sample": "first"},
  {"rule": "indent", "file": "y.js", "line": 1, "count": 25, "sample": "first"}
],
"dedupStats": {"originalCount": 100, "dedupedCount": 3, "reduction": "97%"}}

// With --disable-filter=dedupe (full output, 100 items):
// $ structured-cli eslint src/ --json --disable-filter=dedupe
{"issues": [
  {"rule": "no-unused-vars", "file": "a.js", "line": 10},
  {"rule": "no-unused-vars", "file": "b.js", "line": 20},
  {"rule": "no-unused-vars", "file": "c.js", "line": 30},
  ...
]}
```

#### E2E Tests
- [x] Dedup enabled by default in JSON mode
- [x] Identical objects at same array level are collapsed with count
- [x] Different objects are preserved individually
- [x] Raw text dedup collapses identical lines
- [x] `--disable-filter=dedupe` disables deduplication
- [x] `--disable-filter=all` disables all filters
- [x] `STRUCTURED_CLI_DISABLE_FILTER=dedupe` env var works
- [x] Dedup preserves first occurrence as sample
- [x] Dedup adds `dedupStats` to output
- [x] Objects at different nesting levels are not deduplicated against each other
- [x] Dedup handles empty arrays gracefully
- [x] Passthrough mode (no --json) is unaffected by dedup

### Phase 22: Success Message Filter
Filter out success/passing results from test runners and linters to focus on actionable failures.

#### Design Decision: Failure-Focus Filter

**Problem:** Test and lint output often contains hundreds of passing items that consume tokens without providing actionable information. An AI agent typically only needs to see failures.

**Solution:** Filter that removes success/passing items from output arrays, keeping only failures and a summary count.

**Applies to:**
- Test runners: jest, pytest, vitest, go test, cargo test, mocha
- Linters: eslint, golangci-lint, ruff, mypy, tsc, prettier, biome
- Build tools: go build, cargo build, tsc (warnings vs errors)

**Behavior:**
- Enabled by default for test/lint commands in JSON mode
- Removes passing tests, successful checks from arrays
- Adds summary: `{"passed": 45, "failed": 2, "skipped": 1}`
- Preserves all failure details

#### Domain Layer
- [x] Define `domain/filter.go` (FilterConfig, FilterResult types)
- [x] Define success/failure detection rules per output type

#### Ports Layer
- [x] Define `ports/filter.go` (OutputFilter interface)

#### Application Layer
- [x] Implement `application/success_filter.go`
- [x] Detect test results by status field (passed/failed/skipped/error)
- [x] Detect lint results by severity field (error/warning/info)
- [x] Remove passing items, preserve failures
- [x] Add summary stats to output
- [x] Integrate into executor pipeline (after parse, can chain with dedup)

#### CLI Integration
- [x] Success filter enabled by default for test/lint commands
- [x] Add `--disable-filter=success` flag to show all results
- [x] Add `STRUCTURED_CLI_DISABLE_FILTER=success` environment variable
- [x] Combine with dedupe: `--disable-filter=success,dedupe`

#### Parser Detection Rules
Define which parsers use success filtering and how to detect success:

**Test Runners:**
- [x] `jest` - filter where `status === "passed"`
- [x] `pytest` - filter where `outcome === "passed"`
- [x] `vitest` - filter where `state === "pass"`
- [x] `go test` - filter where `action === "pass"`
- [x] `cargo test` - filter where `status === "ok"`
- [x] `mocha` - filter where `state === "passed"`

**Linters (filter non-errors):**
- [x] `eslint` - filter where `severity < 2` (keep only errors)
- [x] `golangci-lint` - filter where `severity === "warning"` (optional)
- [x] `tsc` - keep all (errors only by default)
- [x] `ruff` - keep all (no severity levels)

#### Output Format
```json
// Default JSON output (success filter ON):
// $ structured-cli pytest tests/ --json
{
  "tests": [
    {"name": "test_login_failure", "outcome": "failed", "message": "AssertionError..."},
    {"name": "test_invalid_input", "outcome": "failed", "message": "ValueError..."}
  ],
  "summary": {
    "total": 48,
    "passed": 45,
    "failed": 2,
    "skipped": 1
  },
  "filterStats": {"removed": 46, "kept": 2}
}

// With --disable-filter=success (full output):
// $ structured-cli pytest tests/ --json --disable-filter=success
{
  "tests": [
    {"name": "test_login", "outcome": "passed"},
    {"name": "test_logout", "outcome": "passed"},
    ... // all 48 tests
  ],
  "summary": {"total": 48, "passed": 45, "failed": 2, "skipped": 1}
}
```

#### E2E Tests
- [x] Success filter enabled by default for test commands
- [x] Success filter enabled by default for lint commands
- [x] Passing tests are removed, failures preserved
- [x] Summary stats added with pass/fail/skip counts
- [x] `--disable-filter=success` shows all results
- [x] `--disable-filter=all` disables success filter and dedupe
- [x] `STRUCTURED_CLI_DISABLE_FILTER=success` env var works
- [x] Filter works correctly with empty results (all pass)
- [x] Filter works correctly with all failures
- [x] Filter chains correctly with dedupe filter
- [x] Non-test/lint commands are unaffected
- [x] Passthrough mode is unaffected

### Phase 23: Small Output Filter
Filter that detects terse CLI outputs and returns minimal JSON status instead of verbose structured data.

#### Design Decision: Minimal Status for Terse Outputs

**Problem:** Some CLI commands produce very terse output (e.g., `git status` showing "nothing to commit"). When converted to JSON, the structured output with metadata becomes *larger* than the raw output, resulting in negative token savings. This defeats the purpose of structured output for simple cases.

**Solution:** A "small" filter that detects when raw output is minimal and returns a simplified status response instead of full structured data.

**Detection Criteria:**
- Raw output is below a threshold (e.g., < 100 characters or < 25 tokens)
- Output matches known "success/clean" patterns for the parser
- Parser indicates output represents a "minimal state"

**Applies to:**
- `git status` - "nothing to commit, working tree clean"
- `git stash list` - empty output (no stashes)
- `npm audit` - "found 0 vulnerabilities"
- `go build` - empty output (successful build)
- `cargo build` - "Finished" with no warnings
- Test runners with 0 failures
- Linters with 0 issues

**Behavior:**
- Enabled by default in JSON mode
- Detects terse/minimal outputs using raw token count + pattern matching
- Returns compact status JSON: `{"status": "clean", "summary": "nothing to commit"}`
- Preserves full output when content exceeds threshold or contains actionable data

#### Domain Layer
- [x] Define `domain/small_filter.go` (SmallOutputConfig, SmallOutputResult types)
- [x] Add `IsMinimal() bool` method signature to parser interface (optional)
- [x] Define threshold constants (MIN_TOKEN_THRESHOLD = 25)

#### Ports Layer
- [x] Extend `ports/filter.go` with SmallOutputFilter interface
- [x] Define MinimalPattern type for parser-specific detection

#### Application Layer
- [x] Implement `application/small_filter.go`
- [x] Token count check: if rawTokens < threshold, check for minimal pattern
- [x] Pattern matching for known "clean/empty" outputs per command
- [x] Return compact status JSON when filter triggers
- [x] Integrate into executor pipeline (before other filters)

#### Parser Integration
Add minimal output patterns to existing parsers:

**Git:**
- [x] `git status` - matches "nothing to commit" or "working tree clean"
- [x] `git stash list` - empty output
- [x] `git diff` - empty output (no changes)

**Package Managers:**
- [x] `npm audit` - matches "found 0 vulnerabilities"
- [x] `npm outdated` - empty output (all up to date)
- [x] `pip-audit` - matches "No known vulnerabilities"

**Build Tools:**
- [x] `go build` - empty output
- [x] `cargo build` - matches "Finished" without warnings/errors
- [x] `tsc` - empty output (no errors)

**Linters:**
- [x] `eslint` - empty output or 0 problems
- [x] `golangci-lint` - empty output
- [x] `ruff` - empty output

#### CLI Integration
- [x] Small filter enabled by default in JSON mode
- [x] Add `--disable-filter=small` flag to get full structured output
- [x] Add `STRUCTURED_CLI_DISABLE_FILTER=small` environment variable
- [x] Combine with other filters: `--disable-filter=small,success,dedupe`

#### Output Format
```json
// Default JSON output (small filter ON) for clean git status:
// $ structured-cli git status --json
{
  "status": "clean",
  "summary": "nothing to commit, working tree clean"
}

// With --disable-filter=small (full structured output):
// $ structured-cli git status --json --disable-filter=small
{
  "clean": true,
  "branch": "main",
  "upstream": "origin/main",
  "ahead": 0,
  "behind": 0,
  "staged": [],
  "modified": [],
  "untracked": [],
  "renamed": [],
  "deleted": []
}
```

#### Tracking Integration
- [x] Small filter outputs should show positive token savings in stats
- [x] Track filter activation count in stats (`--by-filter` breakdown)
- [x] Exclude filtered commands from "negative savings" calculations

#### E2E Tests
- [x] Small filter enabled by default for JSON mode
- [x] Clean `git status` returns compact status JSON
- [x] Empty `git stash list` returns compact status JSON
- [x] `npm audit` with 0 vulnerabilities returns compact status
- [x] `--disable-filter=small` returns full structured output
- [x] `--disable-filter=all` disables small filter
- [x] `STRUCTURED_CLI_DISABLE_FILTER=small` env var works
- [x] Filter does not trigger when output exceeds threshold
- [x] Filter does not trigger when output contains actionable data
- [x] Filter chains correctly with success and dedupe filters
- [x] Passthrough mode is unaffected
- [x] Stats tracking shows improved savings with filter enabled

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

Feature: Usage Tracking System
  As a user of structured-cli
  I want to track command usage and token savings
  So that I can measure the value and optimize my workflow

  Background:
    Given structured-cli is configured with tracking enabled
    And the tracking database is at "~/.local/share/structured-cli/tracking.db"

  # Recording Commands

  Scenario: Track successful JSON output command
    Given I have a git repository
    When I run "structured-cli git status --json"
    Then the command should be recorded in the tracking database
    And the record should contain "command" equal to "git"
    And the record should contain "subcommands" containing "status"
    And the record should contain "input_tokens" as a positive integer
    And the record should contain "output_tokens" as a positive integer
    And the record should contain "saved_tokens" calculated correctly
    And the record should contain "savings_pct" between 0 and 100
    And the record should contain "exec_time_ms" as a positive integer
    And the record should contain "parser_used" equal to "git-status"

  Scenario: Track passthrough mode (no savings)
    Given I have a git repository
    When I run "structured-cli git status" without --json flag
    Then the command should NOT be recorded in the tracking database

  Scenario: Track unsupported command fallback
    Given I have a git repository
    When I run "structured-cli git some-obscure-subcommand --json"
    Then the command should be recorded in the tracking database
    And the record should contain "parser_used" equal to "fallback"
    And the record should contain "savings_pct" near 0

  Scenario: Track parse failure
    Given a command produces output that cannot be parsed
    When I run the command with --json flag
    Then a parse failure should be recorded
    And the failure record should contain "error_message"
    And the failure record should contain "fallback_succeeded" as true or false

  Scenario: Token estimation uses chars/4 heuristic
    Given I have a command that produces 400 characters of raw output
    And the JSON output is 100 characters
    When the command is tracked
    Then "input_tokens" should be approximately 100
    And "output_tokens" should be approximately 25
    And "saved_tokens" should be approximately 75
    And "savings_pct" should be approximately 75

  # Stats Subcommand

  Scenario: Stats summary report
    Given I have tracked 100 commands over the past week
    When I run "structured-cli stats"
    Then the output should show total commands executed
    And the output should show total tokens saved
    And the output should show average savings percentage
    And the output should show top commands by usage

  Scenario: Stats with JSON output
    Given I have tracked commands
    When I run "structured-cli stats --json"
    Then the JSON should contain "total_commands" as an integer
    And the JSON should contain "total_input_tokens" as an integer
    And the JSON should contain "total_output_tokens" as an integer
    And the JSON should contain "total_saved_tokens" as an integer
    And the JSON should contain "avg_savings_pct" as a float
    And the JSON should contain "by_command" as an array
    And the JSON should contain "by_day" as an array

  Scenario: Stats history shows recent commands
    Given I have tracked 50 commands
    When I run "structured-cli stats --history"
    Then the output should show the 20 most recent commands
    And each entry should show timestamp, command, and savings

  Scenario: Stats history with JSON output
    Given I have tracked commands
    When I run "structured-cli stats --history --json"
    Then the JSON should contain "commands" as an array
    And each command should have "timestamp", "command", "subcommands", "saved_tokens", "savings_pct"

  Scenario: Stats by parser breakdown
    Given I have tracked commands using multiple parsers
    When I run "structured-cli stats --by-parser"
    Then the output should group stats by parser name
    And each parser should show count, total savings, average savings

  Scenario: Stats by parser with JSON output
    Given I have tracked commands using multiple parsers
    When I run "structured-cli stats --by-parser --json"
    Then the JSON should contain "by_parser" as an array
    And each entry should have "parser", "count", "total_saved", "avg_savings_pct"

  Scenario: Stats scoped to current project
    Given I have tracked commands from multiple projects
    When I run "structured-cli stats --project" from "/home/user/myproject"
    Then the output should only show stats for commands run in "/home/user/myproject"

  # Database Management

  Scenario: 90-day automatic cleanup
    Given I have tracking records older than 90 days
    When a new command is tracked
    Then records older than 90 days should be deleted
    And records within 90 days should be preserved

  Scenario: Database is created on first use
    Given the tracking database does not exist
    When I run "structured-cli git status --json"
    Then the tracking database should be created
    And the database should have "commands" table
    And the database should have "parse_failures" table

  Scenario: XDG Base Directory compliance
    Given XDG_DATA_HOME is set to "/custom/data"
    When I run "structured-cli git status --json"
    Then the tracking database should be at "/custom/data/structured-cli/tracking.db"

  Scenario: XDG fallback when not set
    Given XDG_DATA_HOME is not set
    When I run "structured-cli git status --json"
    Then the tracking database should be at "~/.local/share/structured-cli/tracking.db"

  # Disable Tracking

  Scenario: Disable tracking via environment variable
    Given STRUCTURED_CLI_NO_TRACKING is set to "1"
    When I run "structured-cli git status --json"
    Then the command should NOT be recorded in the tracking database
    And the JSON output should still be produced normally

  Scenario: Disable tracking via config file
    Given tracking is disabled in config file
    When I run "structured-cli git status --json"
    Then the command should NOT be recorded in the tracking database

Feature: Output Deduplication System
  As an AI coding agent
  I want to deduplicate repetitive output items
  So that I consume fewer tokens while retaining essential information

  Background:
    Given structured-cli is installed
    And the command produces JSON output with --json flag

  # Default Deduplication (ON by default)

  Scenario: Dedupe is enabled by default in JSON mode
    Given a linter produces 50 identical "no-unused-vars" errors
    When I run "structured-cli eslint src/ --json"
    Then the JSON should contain a single grouped item
    And the item should have "count" equal to 50
    And the item should have "sample" equal to "first"

  Scenario: Dedupe collapses identical objects at same level
    Given eslint produces 10 identical error objects
      | rule           | file   | line |
      | no-unused-vars | a.js   | 10   |
    When I run "structured-cli eslint src/ --json"
    Then the JSON should contain 1 grouped item
    And the item should have "count" equal to 10

  Scenario: Dedupe preserves different objects
    Given eslint produces errors with different content
      | rule           | file   | line |
      | no-unused-vars | a.js   | 10   |
      | no-unused-vars | a.js   | 20   |
      | semi           | b.js   | 5    |
    When I run "structured-cli eslint src/ --json"
    Then the JSON should contain 3 items (all different)
    And no item should have a "count" field greater than 1

  Scenario: Disable dedupe with --disable-filter flag
    Given eslint produces 50 errors
    When I run "structured-cli eslint src/ --json --disable-filter=dedupe"
    Then the JSON should contain all 50 individual items
    And no item should have a "count" field

  Scenario: Disable all filters with --disable-filter=all
    Given eslint produces 50 errors
    When I run "structured-cli eslint src/ --json --disable-filter=all"
    Then the JSON should contain all 50 individual items
    And no filtering or transformation should be applied

  # Dedup Stats

  Scenario: Dedupe adds stats to output by default
    Given a command produces 100 items that dedupe to 5
    When I run the command with --json
    Then the JSON should contain "dedupStats" object
    And "dedupStats.originalCount" should equal 100
    And "dedupStats.dedupedCount" should equal 5
    And "dedupStats.reduction" should equal "95%"

  # Environment Variable

  Scenario: Disable dedupe via environment variable
    Given STRUCTURED_CLI_DISABLE_FILTER is set to "dedupe"
    When I run "structured-cli eslint src/ --json"
    Then deduplication should NOT be applied
    And all individual items should be present

  # Edge Cases

  Scenario: Dedupe handles empty arrays
    Given a linter produces no errors
    When I run "structured-cli eslint src/ --json --dedupe"
    Then the JSON should contain an empty array
    And "dedupStats.originalCount" should equal 0

  Scenario: Dedupe handles single item
    Given a linter produces exactly 1 error
    When I run "structured-cli eslint src/ --json --dedupe"
    Then the JSON should contain 1 item
    And the item should have "count" equal to 1

  Scenario: Dedupe preserves non-array fields
    Given command output has both arrays and scalar fields
    When I run the command with --dedupe
    Then only array fields should be deduplicated
    And scalar fields should be unchanged

  Scenario: Dedupe handles nested arrays
    Given output has nested structures with arrays
    When I run the command with --dedupe
    Then top-level arrays should be deduplicated
    And nested arrays within items should be preserved

  # Two-Stage Deduplication

  Scenario: Raw text deduplication before parsing
    Given a command produces repeated identical lines
      """
      ERROR: connection refused
      ERROR: connection refused
      ERROR: connection refused
      """
    When I run the command with --json
    Then raw text should be deduplicated before parsing
    And the output should reflect 1 unique error with count 3

  Scenario: JSON object deduplication at same array level
    Given parsed output has an array with identical objects
    When deduplication is applied
    Then identical objects should be collapsed
    And "count" field should reflect the number of occurrences

  Scenario: Objects at different levels are not deduplicated
    Given parsed output has nested arrays
    And identical objects exist at different nesting levels
    When deduplication is applied
    Then objects are only compared within their own array level
    And cross-level duplicates are preserved

Feature: Success Message Filter
  As an AI coding agent
  I want to filter out passing tests and successful lint checks
  So that I only see actionable failures and save tokens

  Background:
    Given structured-cli is installed
    And the command produces JSON output with --json flag

  # Default Behavior (ON by default for test/lint)

  Scenario: Test runner filters passing tests by default
    Given pytest runs 50 tests with 48 passing and 2 failing
    When I run "structured-cli pytest tests/ --json"
    Then the JSON "tests" array should contain only 2 items (failures)
    And the JSON should contain "summary" with pass/fail counts
    And "summary.passed" should equal 48
    And "summary.failed" should equal 2

  Scenario: Linter filters successful checks by default
    Given eslint checks 100 files with 95 passing and 5 with errors
    When I run "structured-cli eslint src/ --json"
    Then the JSON "issues" array should contain only errors
    And passing files should not appear in output

  Scenario: All tests pass shows empty array with summary
    Given pytest runs 50 tests with all passing
    When I run "structured-cli pytest tests/ --json"
    Then the JSON "tests" array should be empty
    And "summary.passed" should equal 50
    And "summary.failed" should equal 0

  # Disable Filter

  Scenario: Disable success filter shows all results
    Given pytest runs 50 tests with 48 passing and 2 failing
    When I run "structured-cli pytest tests/ --json --disable-filter=success"
    Then the JSON "tests" array should contain all 50 items
    And passing tests should be included

  Scenario: Disable all filters shows unfiltered output
    Given pytest runs tests with duplicates and passing results
    When I run "structured-cli pytest tests/ --json --disable-filter=all"
    Then no filtering should be applied
    And all results should be present

  Scenario: Disable via environment variable
    Given STRUCTURED_CLI_DISABLE_FILTER is set to "success"
    When I run "structured-cli pytest tests/ --json"
    Then all test results should be included
    And passing tests should not be filtered

  # Filter Stats

  Scenario: Filter adds stats to output
    Given pytest runs 50 tests with 48 passing and 2 failing
    When I run "structured-cli pytest tests/ --json"
    Then the JSON should contain "filterStats" object
    And "filterStats.removed" should equal 48
    And "filterStats.kept" should equal 2

  # Test Runner Specific

  Scenario: Jest filters passing tests
    Given jest runs tests with mixed results
    When I run "structured-cli jest --json"
    Then only tests with "status" not equal to "passed" should appear

  Scenario: Go test filters passing tests
    Given go test runs with mixed results
    When I run "structured-cli go test ./... --json"
    Then only tests with "action" equal to "fail" should appear
    And package pass events should be summarized

  Scenario: Vitest filters passing tests
    Given vitest runs tests with mixed results
    When I run "structured-cli vitest --json"
    Then only tests with "state" not equal to "pass" should appear

  Scenario: Cargo test filters passing tests
    Given cargo test runs with mixed results
    When I run "structured-cli cargo test --json"
    Then only tests with "status" not equal to "ok" should appear

  # Linter Specific

  Scenario: ESLint keeps only errors by default
    Given eslint produces warnings and errors
    When I run "structured-cli eslint src/ --json"
    Then only items with "severity" equal to 2 (error) should appear
    And warnings should be counted in summary but not shown

  Scenario: TypeScript compiler keeps all errors
    Given tsc produces type errors
    When I run "structured-cli tsc --json"
    Then all errors should appear (tsc only outputs errors)

  # Filter Chaining

  Scenario: Success filter chains with dedupe filter
    Given pytest produces duplicate failure messages
    When I run "structured-cli pytest tests/ --json"
    Then passing tests should be filtered (success filter)
    And duplicate failures should be collapsed (dedupe filter)
    And both filterStats and dedupStats should be present

  Scenario: Disable success but keep dedupe
    Given pytest produces results with duplicates
    When I run "structured-cli pytest tests/ --json --disable-filter=success"
    Then all tests should be present (success filter disabled)
    And duplicates should still be collapsed (dedupe active)

  # Non-applicable Commands

  Scenario: Non-test commands are unaffected
    Given git status shows modified files
    When I run "structured-cli git status --json"
    Then success filter should not apply
    And all status information should be present

  Scenario: File operations are unaffected
    Given ls lists directory contents
    When I run "structured-cli ls -la --json"
    Then success filter should not apply
    And all entries should be present
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
