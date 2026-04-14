# Phase 36: Optimize Test Runner Parsers - Ultra-Compact JSON Format

## Goal

Optimize test runner parsers (pytest, jest, vitest, mocha, cargo test, go test) to use ultra-compact format grouped by test status. This achieves maximum token savings through status grouping and synergizes with the success filter to show only failures.

## Context

**Current Problem:**
- Test suites have 100-1000+ tests with repeated field structure
- Each test repeats: `{"name": "...", "status": "...", "duration": ..., "message": "..."}`
- Most tests pass (90-95%), wasting tokens on repeated "passed" status
- Result: Massive overhead for test results

**Solution:**
- Group tests by status: `passed`, `failed`, `skipped`
- Use array tuples: `[name, duration]` for passed, `[name, duration, error]` for failed
- Add smart truncation (1000 tests max, prioritize failures)
- Synergy with success filter: Only show failed tests by default
- Target: 40% token savings + better UX

## Affected Parsers

1. **pytest** (Python)
2. **jest** (JavaScript/TypeScript)
3. **vitest** (Vite test runner)
4. **mocha** (Node.js)
5. **cargo test** (Rust)
6. **go test** (Go)

## Tasks

### 1. Update Test Runner Types (internal/adapters/parsers/test/types.go)

- [ ] Create `TestOutputCompact` struct for all test runners
- [ ] Add `TestSummary` with counts by status
- [ ] Add `PassedTestTuple` type: `[name, duration]`
- [ ] Add `FailedTestTuple` type: `[name, duration, error_message, traceback]`
- [ ] Add `SkippedTestTuple` type: `[name, reason]`
- [ ] Add status groups: `passed`, `failed`, `skipped` arrays
- [ ] Keep old types temporarily for backward compatibility

### 2. Update Pytest Parser (internal/adapters/parsers/python/pytest.go)

- [ ] Modify `Parse()` to group tests by status
- [ ] Extract test name, duration, status from output
- [ ] For failures: extract assertion error and traceback
- [ ] For skips: extract skip reason
- [ ] Implement truncation: 1000 tests max, keep all failures
- [ ] Filter pytest internal output (collection stats, etc.)
- [ ] Update schema to reflect grouped structure

### 3. Update Jest Parser (internal/adapters/parsers/js/jest.go)

- [ ] Group tests by status (passed/failed/skipped/pending)
- [ ] Parse Jest JSON output if available (`--json`)
- [ ] Extract test suite and test name
- [ ] For failures: extract expected vs actual, stack trace
- [ ] Handle snapshot failures specially
- [ ] Implement truncation with failure priority
- [ ] Update schema

### 4. Update Vitest Parser (internal/adapters/parsers/js/vitest.go)

- [ ] Similar to Jest but handle Vitest-specific format
- [ ] Parse Vitest reporter output
- [ ] Group by status
- [ ] Extract duration, error messages
- [ ] Implement truncation
- [ ] Update schema

### 5. Update Mocha Parser (internal/adapters/parsers/js/mocha.go)

- [ ] Parse Mocha TAP or JSON output
- [ ] Group by status
- [ ] Extract test names from nested describe blocks
- [ ] For failures: extract assertion error
- [ ] Handle pending tests (not implemented)
- [ ] Implement truncation
- [ ] Update schema

### 6. Update Cargo Test Parser (internal/adapters/parsers/rust/test.go)

- [ ] Group tests by status (ok/FAILED/ignored)
- [ ] Parse test names (module::test_name format)
- [ ] For failures: extract panic message
- [ ] Handle doc tests separately
- [ ] Implement truncation
- [ ] Update schema

### 7. Update Go Test Parser (internal/adapters/parsers/golang/test.go)

- [ ] Parse `go test -json` output (preferred)
- [ ] Fallback to text output parsing
- [ ] Group by status (PASS/FAIL/SKIP)
- [ ] Extract package, test name, duration
- [ ] For failures: extract error message and location
- [ ] Handle table-driven test output
- [ ] Implement truncation
- [ ] Update schema

### 8. Update Tests for All Test Parsers

#### Pytest Tests (internal/adapters/parsers/python/pytest_test.go)
- [ ] Update `TestPytestParser_AllPassed` for grouped format
- [ ] Update `TestPytestParser_WithFailures` to verify failure details
- [ ] Add `TestPytestParser_GroupByStatus` to verify grouping
- [ ] Add `TestPytestParser_Truncation` for 1000-test limit
- [ ] Add `TestPytestParser_FailurePriority` to verify failures kept
- [ ] Add `TestPytestParser_SkipReason` to verify skip extraction
- [ ] Update existing tests for new format

#### Jest Tests
- [ ] Update all existing tests for grouped format
- [ ] Add `TestJestParser_SnapshotFailures` for snapshot handling
- [ ] Add truncation tests
- [ ] Add grouping verification tests

#### Vitest Tests
- [ ] Mirror Jest test structure
- [ ] Add Vitest-specific format tests

#### Mocha Tests
- [ ] Update for grouped format
- [ ] Test nested describe block handling

#### Cargo Test Tests
- [ ] Update for grouped format
- [ ] Test doc test handling separately

#### Go Test Tests
- [ ] Update for grouped format
- [ ] Test JSON mode parsing
- [ ] Test text mode fallback
- [ ] Test table-driven test handling

### 9. Update JSON Schemas

- [ ] Update `schemas/pytest.json` for grouped format
- [ ] Update `schemas/jest.json` for grouped format
- [ ] Update `schemas/vitest.json` for grouped format
- [ ] Update `schemas/mocha.json` for grouped format
- [ ] Update `schemas/cargo-test.json` for grouped format
- [ ] Update `schemas/go-test.json` for grouped format
- [ ] Document tuple formats in each schema
- [ ] Add examples showing status grouping

### 10. Update Feature Tests

- [ ] Update test runner scenarios in `features/test_runners.feature`
- [ ] Add scenario for all tests passing
- [ ] Add scenario for mixed pass/fail/skip
- [ ] Add scenario for truncation with 500+ tests
- [ ] Add scenario showing success filter integration
- [ ] Verify grouped status structure

### 11. Success Filter Integration

- [ ] Verify success filter works with new grouped format
- [ ] Test that `--enable-filter=success` shows only failed tests
- [ ] Ensure default behavior filters passed tests for test commands
- [ ] Update success filter logic if needed for grouped structure
- [ ] Test filter with all test runners

### 12. Token Savings Validation

- [ ] Benchmark each test runner with 100 tests (95 pass, 5 fail)
- [ ] Benchmark with 1000 tests
- [ ] Measure impact of success filter + compact format
- [ ] Verify 40% token savings on base format
- [ ] Verify 90%+ savings with success filter enabled
- [ ] Test common repositories

### 13. Documentation

- [ ] Update CLAUDE.md with test runner examples
- [ ] Update README.md test section
- [ ] Document status grouping rationale
- [ ] Document success filter synergy
- [ ] Document truncation and failure priority
- [ ] Add migration guide

## Implementation Notes

### Compact Format Specification

```json
{
  "summary": {
    "total": 156,
    "passed": 148,
    "failed": 5,
    "skipped": 3,
    "duration": 12.5
  },
  "passed": [
    ["test_login", 0.1],
    ["test_logout", 0.05],
    ["test_signup", 0.2]
  ],
  "failed": [
    ["test_payment", 1.2, "AssertionError: expected 200, got 404", "  File test.py, line 45\n    assert response.status == 200"],
    ["test_refund", 0.8, "ValueError: invalid amount", "  File test.py, line 67"]
  ],
  "skipped": [
    ["test_slow_integration", "marked slow"],
    ["test_external_api", "requires network"]
  ],
  "truncated": 0
}
```

**Tuple Formats**:
- **Passed**: `[name, duration]`
- **Failed**: `[name, duration, error_message, traceback]`
- **Skipped**: `[name, skip_reason]`

### Status Grouping Benefits

1. **No repetition** of "passed"/"failed"/"skipped" strings
2. **Different tuple sizes** optimize each status
3. **Easy filtering** by status
4. **Synergy with success filter** - just omit `passed` array

### Success Filter Integration

**Without success filter** (show all):
```json
{
  "summary": {"total": 100, "passed": 95, "failed": 5},
  "passed": [[...], ...],  // 95 tests
  "failed": [[...], ...],  // 5 tests
  "skipped": []
}
```

**With success filter** (default for tests):
```json
{
  "summary": {"total": 100, "passed": 95, "failed": 5},
  "failed": [[...], ...],  // Only failures shown
  "skipped": []
}
```

**Token savings with filter**: ~95% reduction (only show failures!)

### Truncation Strategy

1. **Global limit**: 1000 tests maximum
2. **Failure priority**: Always keep ALL failures, truncate passed tests first
3. **Skip handling**: Truncate skipped before passed
4. **Order**: Keep first N passed tests (likely most important)
5. **Metadata**: `summary` reflects total counts even when truncated

**Truncation order**:
1. Keep all failed tests (highest priority)
2. Keep all skipped tests
3. Truncate passed tests to fit limit
4. If still over limit, truncate skipped tests

### Test Runner Specific Notes

**Pytest**:
- Parse pytest markers (xfail, skip, parametrize)
- Handle fixture failures separately
- Extract assertion introspection details

**Jest**:
- Parse JSON mode for best results
- Handle snapshot failures with diff
- Extract expected vs actual for assertions
- Group by test suite if available

**Vitest**:
- Similar to Jest
- Handle Vite-specific reporters
- Parse inline snapshots

**Mocha**:
- Parse TAP or JSON reporter
- Handle nested describe blocks (flatten to test name)
- Extract "pending" tests

**Cargo Test**:
- Parse "test result: ok" summary
- Handle ignored tests
- Extract panic messages for failures
- Separate doc tests from unit tests

**Go Test**:
- Prefer `-json` output (structured)
- Fallback to text parsing
- Handle subtests (TestFoo/subtest_name)
- Extract file:line for failures

### Token Savings Math

**100 tests (95 passed, 5 failed)**:

Old format (~3,800 tokens):
```json
{"tests": [
  {"name": "test_login", "status": "passed", "duration": 0.1, "message": ""},
  // ... repeated 95 times
  {"name": "test_payment", "status": "failed", "duration": 1.2, "message": "AssertionError: ..."},
  // ... 5 failures
]}
```

New format without success filter (~2,300 tokens):
```json
{"summary": {"total": 100, "passed": 95, "failed": 5},
 "passed": [["test_login", 0.1], ...],  // 95 tests
 "failed": [["test_payment", 1.2, "AssertionError: ..."]]  // 5 tests
}
```

**Savings**: ~40% reduction

New format WITH success filter (~300 tokens):
```json
{"summary": {"total": 100, "passed": 95, "failed": 5},
 "failed": [["test_payment", 1.2, "AssertionError: ..."]]  // Only failures
}
```

**Savings with filter**: ~92% reduction!

### Edge Cases to Handle

- All tests pass: `{"passed": [...], "failed": [], "skipped": []}`
- All tests fail: Still group by status
- No tests found: `{"summary": {"total": 0}, "passed": [], "failed": [], "skipped": []}`
- Parametrized tests: Include parameter in name or separate field
- Retried tests: Show final status only
- Timeout failures: Include timeout in error message
- Setup/teardown failures: Include in failed tests

## Acceptance Criteria

- [ ] All 6 test parsers use grouped status format
- [ ] Tests are grouped by passed/failed/skipped
- [ ] Token savings are 40% for base format
- [ ] Token savings are 90%+ with success filter
- [ ] Truncation works correctly (1000 test limit)
- [ ] Failures are never truncated
- [ ] All existing tests pass with new format
- [ ] Feature tests pass
- [ ] Success filter integration works
- [ ] Schemas are updated and validated
- [ ] No regressions in test runner functionality

## Success Metrics

- 40% token savings on base format for 100-test suite
- 90%+ token savings with success filter enabled
- Truncation at 1000 tests works correctly
- All failures preserved even with truncation
- Success filter shows only failures by default
- Token cost for 100 tests (5 failures) with filter: ~300 tokens (vs ~3,800 currently)
