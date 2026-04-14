# Phase 32: Fix Tracking Test Failures

## Tasks

- [ ] Fix tracking filter - always record commands to database, only filter during stats display
- [ ] Update tracking service to record all commands regardless of token savings
- [ ] Ensure token savings filter only applies to stats output, not command recording
- [ ] Verify all 5 failing tracking tests pass after fix

## Notes

- Problem: Task 5 implemented a filter that prevents commands with small token savings from being recorded in the database
- Solution: Commands should always be recorded for history; filter should only apply to stats display/calculation
- Affected code: `internal/adapters/tracking/` and `internal/application/executor.go`
- Tests failing: 5 scenarios in `features/tracking.feature`
- Follow TDD: write/update tests first, then fix implementation
