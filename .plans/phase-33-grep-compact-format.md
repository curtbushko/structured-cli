# Phase 33: Optimize Grep Parser - Ultra-Compact JSON Format

## Goal

Optimize the grep parser to use an ultra-compact array-based JSON format that achieves 25% token savings compared to the current verbose object format. This addresses the issue where grep currently produces negative token savings due to JSON overhead.

## Context

**Current Problem:**
- Grep output is naturally compact: `file:line:content`
- Current JSON format adds massive overhead with repeated field names
- Result: grep produces -30 tokens (costs tokens instead of saving them)
- These small negative savings are filtered out by the [-100, 100] threshold

**Solution:**
- Use array-based format: `[filename, count, [[line, content], ...]]`
- Group matches by file (like rtk does)
- Add smart truncation to prevent token explosion
- Target: 25% token savings (measured: 147 → 110 tokens for 6 matches)

## Tasks

### 1. Update Types (internal/adapters/parsers/fileops/types.go)

- [ ] Create new `GrepOutputCompact` struct with array-friendly structure
- [ ] Add `FileMatchGroup` type: contains filename, count, and match tuples
- [ ] Add `MatchTuple` type: `[line int, content string]` representation
- [ ] Keep old `GrepOutput` temporarily for backward compatibility testing

### 2. Update Parser Logic (internal/adapters/parsers/fileops/grep.go)

- [ ] Modify `Parse()` to group matches by file instead of flat list
- [ ] Implement smart truncation: default 200 total matches, 10 per file
- [ ] Track truncation count (how many matches were omitted)
- [ ] Preserve binary file detection logic (skip "Binary file X matches")
- [ ] Add directory warning filter (skip "X is a directory" messages)
- [ ] Add permission error filter (skip "Permission denied" messages)
- [ ] Update schema to reflect new compact structure

### 3. Update Tests (internal/adapters/parsers/fileops/grep_test.go)

- [ ] Update `TestGrepParser_WithLineNumbers` for new format
- [ ] Update `TestGrepParser_WithoutLineNumbers` for new format
- [ ] Update `TestGrepParser_SingleFile` for grouped output
- [ ] Update `TestGrepParser_NoResults` for new structure
- [ ] Add `TestGrepParser_Truncation` to verify 200-match limit
- [ ] Add `TestGrepParser_PerFileTruncation` to verify 10-per-file limit
- [ ] Add `TestGrepParser_FileGrouping` to verify grouping by filename
- [ ] Add `TestGrepParser_TruncationCount` to verify truncated field
- [ ] Update `TestGrepParser_BinaryFile` for new format
- [ ] Update `TestGrepParser_ColonInContent` for new format
- [ ] Add `TestGrepParser_DirectoryWarnings` to verify directory warnings are filtered
- [ ] Add `TestGrepParser_PermissionErrors` to verify permission errors are filtered

### 4. Update JSON Schema (schemas/grep.json)

- [ ] Update schema to reflect array-based structure
- [ ] Document format: `{"total": int, "files": int, "results": [...], "truncated": int}`
- [ ] Document result format: `[filename, count, [[line, content], ...]]`
- [ ] Add examples showing the compact format
- [ ] Add description explaining token efficiency benefits

### 5. Update Feature Tests (features/fileops.feature)

- [ ] Update grep scenarios to expect grouped file structure
- [ ] Update grep with line numbers scenario for tuple format
- [ ] Add scenario testing truncation behavior
- [ ] Verify backward compatibility if needed

### 6. Token Savings Validation

- [ ] Create benchmark test comparing old vs new format token counts
- [ ] Verify 25% token savings on representative grep outputs
- [ ] Test with various sizes: 5 matches, 50 matches, 200+ matches
- [ ] Confirm grep now shows positive token savings in stats

### 7. Documentation

- [ ] Update CLAUDE.md with grep compact format example
- [ ] Update README.md grep section with new format
- [ ] Add comment in grep.go explaining the compact format rationale
- [ ] Document truncation behavior and limits

## Implementation Notes

### Compact Format Specification

```json
{
  "total": 3812,
  "files": 206,
  "results": [
    ["CLAUDE.md", 9, [
      [127, "### 3. Write tests"],
      [129, "Location: `internal/adapters/parsers/{category}/{command}_test.go`"],
      [152, "# Run all tests with race detection"]
    ]],
    ["README.md", 15, [
      [20, "go install github.com/curtbushko/structured-cli/cmd/structured-cli@latest"],
      [104, "| `go test` | Test results with pass/fail/skip counts |"]
    ]]
  ],
  "truncated": 3612
}
```

**Format Rules:**
- `total` - total matches found before truncation
- `files` - total files matched
- `results` - array of `[filename, match_count, match_array]`
- `match_array` - array of `[line_number, content]` tuples
- `truncated` - number of matches omitted (0 if no truncation)

### Truncation Strategy

1. **Global limit**: 200 total matches maximum
2. **Per-file limit**: 10 matches per file maximum
3. **Order preservation**: Show first N matches per file, first M files
4. **Truncation metadata**: Report how many matches were omitted

### Token Savings Math

For 100 matches across 20 files:
- **Old format**: ~2,450 tokens (repeated `"file":`, `"line":`, `"content":`)
- **New format**: ~1,830 tokens (array tuples, grouped files)
- **Savings**: 620 tokens (25.3%)

### Edge Cases to Handle

- Empty results: `{"total": 0, "files": 0, "results": [], "truncated": 0}`
- Single file: Still use array format for consistency
- Binary files: Skip "Binary file X matches" from results, don't count in total
- Directory warnings: Skip "X is a directory" messages
- Permission errors: Skip "grep: X: Permission denied" messages
- Unreadable files: Skip "grep: X: No such file or directory" messages
- No line numbers: Use 0 for line number in tuple
- Truncation: Ensure `total` reflects actual match count, not displayed count

## Acceptance Criteria

- [ ] All existing grep tests pass with new format
- [ ] Token savings are positive (> 100 tokens for typical grep usage)
- [ ] Truncation works correctly for large outputs
- [ ] File grouping reduces redundancy
- [ ] Schema validation passes
- [ ] Feature tests pass
- [ ] No regressions in grep functionality
- [ ] grep commands now appear in stats (positive token savings)

## Success Metrics

- Grep achieves 20-30% token savings on typical outputs
- Large grep results (500+ matches) are automatically truncated
- Token cost for 6-match grep: 110 tokens (vs 147 currently)
- Grep commands with >100 token savings show up in stats
