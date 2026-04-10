# Phase 22: Success Message Filter

Filter out success/passing results from test runners and linters to focus on actionable failures.

## Design Decision: Failure-Focus Filter

**Problem:** Test and lint output often contains hundreds of passing items that consume tokens without providing actionable information.

**Solution:** Filter that removes success/passing items from output arrays, keeping only failures and a summary count.

## Domain Layer

- [x] Define `domain/filter.go` (FilterConfig, FilterResult types)
- [x] Define success/failure detection rules per output type

## Ports Layer

- [x] Define `ports/filter.go` (OutputFilter interface)

## Application Layer

- [x] Implement `application/success_filter.go`
- [x] Detect test results by status field (passed/failed/skipped/error)
- [x] Detect lint results by severity field (error/warning/info)
- [x] Remove passing items, preserve failures
- [x] Add summary stats to output
- [x] Integrate into executor pipeline (after parse, can chain with dedup)

## CLI Integration

- [x] Success filter enabled by default for test/lint commands
- [x] Add `--disable-filter=success` flag to show all results
- [x] Add `STRUCTURED_CLI_DISABLE_FILTER=success` environment variable
- [x] Combine with dedupe: `--disable-filter=success,dedupe`

## Parser Detection Rules

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

## E2E Tests

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
