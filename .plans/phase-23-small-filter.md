# Phase 23: Small Output Filter

Filter that detects terse CLI outputs and returns minimal JSON status instead of verbose structured data.

## Design Decision: Minimal Status for Terse Outputs

**Problem:** Some CLI commands produce very terse output. When converted to JSON, the structured output with metadata becomes *larger* than the raw output, resulting in negative token savings.

**Solution:** A "small" filter that detects when raw output is minimal and returns a simplified status response.

## Domain Layer

- [x] Define `domain/small_filter.go` (SmallOutputConfig, SmallOutputResult types)
- [x] Add `IsMinimal() bool` method signature to parser interface (optional)
- [x] Define threshold constants (MIN_TOKEN_THRESHOLD = 25)

## Ports Layer

- [x] Extend `ports/filter.go` with SmallOutputFilter interface
- [x] Define MinimalPattern type for parser-specific detection

## Application Layer

- [x] Implement `application/small_filter.go`
- [x] Token count check: if rawTokens < threshold, check for minimal pattern
- [x] Pattern matching for known "clean/empty" outputs per command
- [x] Return compact status JSON when filter triggers
- [x] Integrate into executor pipeline (before other filters)

## Parser Integration

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

## CLI Integration

- [x] Small filter enabled by default in JSON mode
- [x] Add `--disable-filter=small` flag to get full structured output
- [x] Add `STRUCTURED_CLI_DISABLE_FILTER=small` environment variable
- [x] Combine with other filters: `--disable-filter=small,success,dedupe`

## Tracking Integration

- [x] Small filter outputs should show positive token savings in stats
- [x] Track filter activation count in stats (`--by-filter` breakdown)
- [x] Exclude filtered commands from "negative savings" calculations

## E2E Tests

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
