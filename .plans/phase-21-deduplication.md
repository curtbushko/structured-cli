# Phase 21: Output Deduplication System

Generic deduplication layer to reduce token usage by collapsing identical items.

## Design Decision: Two-Stage Generic Deduplication

**Approach:** Pure generic deduplication at two stages:
1. **Raw text stage** - Before parsing, collapse identical lines in raw output
2. **JSON object stage** - After parsing, collapse identical objects within same array level

## Domain Layer

- [x] Define `domain/dedup.go` (DedupConfig, DedupResult types)

## Ports Layer

- [x] Define `ports/deduplicator.go` (Deduplicator interface)

## Application Layer

- [x] Implement `application/dedup.go` (deduplication engine)
- [x] **Stage 1: Raw text dedup** - deduplicate identical lines before parsing
- [x] **Stage 2: JSON object dedup** - deduplicate identical objects at same array level
- [x] Objects must be at the same level in the JSON tree to be deduplicated
- [x] Add count field to grouped items (`"count": N`)
- [x] Keep first occurrence as sample
- [x] Integrate dedup step into executor pipeline

## CLI Integration

- [x] Deduplication enabled by default in JSON mode
- [x] Add `--disable-filter=dedupe` flag to disable deduplication
- [x] Add `--disable-filter=all` to disable all filters
- [x] Add `STRUCTURED_CLI_DISABLE_FILTER=dedupe` environment variable
- [x] Support comma-separated filters: `--disable-filter=dedupe,compact`

## Deduplication Rules

- [x] Raw text: identical lines are collapsed with count
- [x] JSON arrays: identical objects at the same level are collapsed
- [x] Object equality: deep comparison of all fields
- [x] Unique outputs (ls entries, container IDs) naturally don't dedupe
- [x] Repetitive outputs (lint errors, log lines) collapse automatically

## E2E Tests

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
