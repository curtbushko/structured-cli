# Phase 34: Optimize Linter Parsers - Ultra-Compact JSON Format

## Goal

Optimize all linter parsers (eslint, golangci-lint, ruff, mypy, prettier, cargo clippy) to use ultra-compact array-based JSON format grouped by file. This addresses the massive token overhead from repeated file paths and field names in linter output.

## Context

**Current Problem:**
- Linters output 100+ issues with repeated field structure
- Each issue repeats: `{"file": "...", "line": ..., "severity": "...", "message": "..."}`
- Result: Massive token waste on field names and file path repetition
- Example: 100 issues = ~250 tokens just for repeated "file", "line", "severity", "message" keys

**Solution:**
- Group issues by file (like grep optimization)
- Use array tuples: `[line, severity, message]`
- Add smart truncation (200 issues max, 20 per file)
- Target: 30-40% token savings

## Affected Parsers

1. **eslint** (JavaScript/TypeScript)
2. **golangci-lint** (Go)
3. **ruff** (Python)
4. **mypy** (Python type checker)
5. **prettier** (formatter output)
6. **cargo clippy** (Rust)

## Tasks

### 1. Update Linter Types (internal/adapters/parsers/linters/types.go)

- [ ] Create `LinterOutputCompact` struct for all linters
- [ ] Add `FileIssueGroup` type: `[filename, issue_count, [[line, severity, message], ...]]`
- [ ] Add `IssueTuple` type: `[line int, severity string, message string, rule_id string]`
- [ ] Add severity enum: "error", "warning", "info", "style"
- [ ] Add optional fields for rule_id, column, end_line (when available)
- [ ] Keep old types temporarily for backward compatibility

### 2. Update ESLint Parser (internal/adapters/parsers/linters/eslint.go)

- [ ] Modify `Parse()` to group issues by file
- [ ] Implement truncation: 200 total issues, 20 per file
- [ ] Extract severity mapping (error/warning/info)
- [ ] Include rule_id in tuple when available
- [ ] Filter ESLint internal warnings/metadata
- [ ] Update schema to reflect compact structure

### 3. Update golangci-lint Parser (internal/adapters/parsers/linters/golangci.go)

- [ ] Group issues by file path
- [ ] Map severity levels to standard enum
- [ ] Include linter name in tuple (which linter triggered)
- [ ] Handle multi-line error messages (truncate to first line)
- [ ] Filter out summary statistics
- [ ] Implement same truncation limits

### 4. Update Ruff Parser (internal/adapters/parsers/linters/ruff.go)

- [ ] Group by file path
- [ ] Extract rule codes (e.g., "F401", "E501")
- [ ] Map severity from ruff format
- [ ] Handle fix suggestions (store in separate field)
- [ ] Implement truncation

### 5. Update Mypy Parser (internal/adapters/parsers/linters/mypy.go)

- [ ] Group type errors by file
- [ ] Extract error codes (e.g., "arg-type", "return-value")
- [ ] Handle note/hint messages (lower severity)
- [ ] Truncate verbose type signatures
- [ ] Implement truncation

### 6. Update Prettier Parser (internal/adapters/parsers/linters/prettier.go)

- [ ] Group formatting issues by file
- [ ] Since prettier is binary (format/no-format), use simpler structure
- [ ] Just list files needing formatting
- [ ] Track total files checked vs files needing formatting

### 7. Update Cargo Clippy Parser (internal/adapters/parsers/rust/clippy.go)

- [ ] Group warnings by file
- [ ] Extract lint names (e.g., "needless_borrow")
- [ ] Map severity (deny/warn/allow)
- [ ] Include help text in separate field (not in main message)
- [ ] Implement truncation

### 8. Update Tests for All Linter Parsers

#### ESLint Tests (internal/adapters/parsers/linters/eslint_test.go)
- [ ] Update `TestESLintParser_BasicIssues` for compact format
- [ ] Update `TestESLintParser_MultipleFiles` to verify file grouping
- [ ] Add `TestESLintParser_Truncation` for 200-issue limit
- [ ] Add `TestESLintParser_PerFileTruncation` for 20-per-file limit
- [ ] Add `TestESLintParser_SeverityMapping` to verify error/warning/info
- [ ] Update `TestESLintParser_NoIssues` for new format

#### golangci-lint Tests
- [ ] Update all existing tests for compact format
- [ ] Add `TestGolangciLint_MultiLinter` to verify linter name in tuple
- [ ] Add truncation tests
- [ ] Add file grouping tests

#### Ruff Tests
- [ ] Update for compact format
- [ ] Add rule code extraction tests
- [ ] Add truncation tests

#### Mypy Tests
- [ ] Update for compact format
- [ ] Add error code extraction tests
- [ ] Add truncation tests

#### Prettier Tests
- [ ] Update for simpler file list format
- [ ] Test files needing formatting vs total checked

#### Clippy Tests
- [ ] Update for compact format
- [ ] Add lint name extraction tests
- [ ] Add help text separation tests
- [ ] Add truncation tests

### 9. Update JSON Schemas

- [ ] Update `schemas/eslint.json` for compact format
- [ ] Update `schemas/golangci-lint.json` for compact format
- [ ] Update `schemas/ruff.json` for compact format
- [ ] Update `schemas/mypy.json` for compact format
- [ ] Update `schemas/prettier.json` for simpler format
- [ ] Update `schemas/clippy.json` for compact format
- [ ] Add examples showing grouped structure
- [ ] Document tuple format in each schema

### 10. Update Feature Tests

- [ ] Update linter scenarios in `features/linters.feature`
- [ ] Add scenario for file grouping behavior
- [ ] Add scenario for truncation with 100+ issues
- [ ] Add scenario for mixed severity levels
- [ ] Verify backward compatibility if needed

### 11. Token Savings Validation

- [ ] Create benchmark comparing old vs new format for each linter
- [ ] Test with 10 issues (minimal overhead)
- [ ] Test with 100 issues (target case)
- [ ] Test with 500+ issues (truncation case)
- [ ] Verify 30-40% token savings target
- [ ] Measure impact on common repositories

### 12. Documentation

- [ ] Update CLAUDE.md with linter compact format examples
- [ ] Update README.md linter section
- [ ] Add comment in each parser explaining grouping rationale
- [ ] Document truncation behavior and limits
- [ ] Add migration guide for users relying on old format

## Implementation Notes

### Compact Format Specification

```json
{
  "total_issues": 156,
  "files_with_issues": 23,
  "severity_counts": {
    "error": 12,
    "warning": 140,
    "info": 4
  },
  "results": [
    ["src/main.go", 15, [
      [23, "error", "undefined variable 'foo'", "undefined-var"],
      [45, "warning", "unused import 'fmt'", "unused-import"],
      [67, "info", "consider using strings.Builder", "string-concat"]
    ]],
    ["src/util.go", 8, [
      [10, "error", "missing return statement", "missing-return"]
    ]]
  ],
  "truncated": 0
}
```

**Tuple Format**: `[line, severity, message, rule_id]`
- `line` (int) - Line number where issue occurs
- `severity` (string) - "error" | "warning" | "info" | "style"
- `message` (string) - Human-readable issue description
- `rule_id` (string, optional) - Linter rule identifier

**File Group Format**: `[filename, issue_count, issues_array]`

### Severity Standardization

Map all linters to consistent severity levels:

| Linter | Error | Warning | Info |
|--------|-------|---------|------|
| eslint | 2 | 1 | - |
| golangci-lint | error | warning | - |
| ruff | E* | W*, F* | I* |
| mypy | error | note | note |
| clippy | deny | warn | allow |

### Truncation Strategy

1. **Global limit**: 200 total issues maximum
2. **Per-file limit**: 20 issues per file maximum
3. **Severity priority**: Show errors first, then warnings, then info
4. **File priority**: Files with most errors shown first
5. **Truncation metadata**: Report counts by severity even when truncated

### Token Savings Math

For 100 linter issues across 20 files:

**Old format** (~3,500 tokens):
```json
{"issues": [
  {"file": "/src/components/Button.tsx", "line": 23, "severity": "warning", "message": "...", "rule": "..."},
  // ... repeated 100 times
]}
```

**New format** (~2,100 tokens):
```json
{"total_issues": 100, "files_with_issues": 20, "results": [
  ["/src/components/Button.tsx", 8, [[23, "warning", "...", "..."], ...]],
  // ... grouped by file
]}
```

**Savings**: 1,400 tokens (40%)

### Edge Cases to Handle

- Empty results: `{"total_issues": 0, "files_with_issues": 0, "results": [], "truncated": 0}`
- Single file: Still use array format for consistency
- No rule_id: Omit from tuple or use empty string
- Multi-line messages: Truncate to first line, add "..." if truncated
- Column numbers: Optional 5th element in tuple when available
- Fix suggestions: Store in separate `fixes` array if linter provides them
- Summary stats: Filter out, don't include in results

### Linter-Specific Notes

**ESLint**:
- Has rich metadata (fix suggestions, rule docs)
- Consider separate `fixable` field for auto-fixable issues

**golangci-lint**:
- Multiple linters in one output
- Include linter name as additional tuple element

**Ruff**:
- Very fast, can produce 1000+ issues
- Truncation is critical

**Mypy**:
- Type errors can be verbose
- Truncate long type signatures

**Prettier**:
- Binary output (formatted vs needs formatting)
- Simpler structure: just list files needing formatting

**Clippy**:
- Includes help text (separate from message)
- Store help in additional field, not main message

## Acceptance Criteria

- [ ] All 6 linter parsers use compact format
- [ ] Issues are grouped by file
- [ ] Token savings are 30-40% for typical linter output
- [ ] Truncation works correctly for large outputs (500+ issues)
- [ ] Severity levels are standardized across linters
- [ ] All existing tests pass with new format
- [ ] Feature tests pass
- [ ] Schemas are updated and validated
- [ ] No regressions in linter functionality
- [ ] Documentation is complete

## Success Metrics

- 30-40% token savings on linter output with 100+ issues
- Linters with 500+ issues automatically truncated to 200
- Consistent severity levels across all linters
- File grouping eliminates filename repetition
- Token cost for 100 issues: ~2,100 tokens (vs ~3,500 currently)
