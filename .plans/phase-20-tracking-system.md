# Phase 20: SQLite Usage Tracking System

Token usage tracking inspired by RTK (Rust Token Killer) for analytics and measuring value.

## Domain Layer

- [x] Define `domain/tracking.go` (CommandRecord, ParseFailure types)
- [x] Define `domain/stats.go` (StatsSummary, CommandStats types)

## Ports Layer

- [x] Define `ports/tracker.go` (Tracker interface with Record, GetStats, Cleanup methods)

## Adapters - SQLite Implementation

- [x] Implement `adapters/tracking/sqlite.go` (SQLite tracker)
- [x] Implement `adapters/tracking/noop.go` (No-op tracker for testing/disabled)
- [x] XDG Base Directory support (`~/.local/share/structured-cli/tracking.db`)
- [x] Database schema creation (commands table, parse_failures table)
- [x] 90-day automatic retention cleanup on insert
- [x] Token estimation (chars/4 heuristic)

## Application Layer Integration

- [x] Add TimedExecution pattern to executor
- [x] Track successful parses (command, tokens, savings, exec time)
- [x] Track parse failures (command, error, fallback success)
- [x] Wire tracker into composition root

## Stats Subcommand

- [x] Implement `stats` subcommand in CLI handler
- [x] Summary report (total commands, tokens saved, avg savings %)
- [x] `--history` flag for recent commands
- [x] `--json` flag for JSON export
- [x] `--by-parser` flag for per-parser breakdown
- [x] `--project` flag for current directory only

## E2E Tests

- [x] Tracking records commands after JSON output
- [x] Tracking calculates token savings correctly
- [x] Stats command shows summary
- [x] Stats --history shows recent commands
- [x] Stats --json outputs valid JSON
- [x] 90-day cleanup removes old records
- [x] Disabled tracking (STRUCTURED_CLI_NO_TRACKING=1)
