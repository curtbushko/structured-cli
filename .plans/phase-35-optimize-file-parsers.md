# Phase 35: Optimize File Operation Parsers - Ultra-Compact JSON Format

## Goal

Optimize file operation parsers (find, ls, ripgrep) to use ultra-compact array-based JSON format. These commands can output thousands of files and benefit significantly from array tuples and truncation.

## Context

**Current Problem:**
- `find` can return 10,000+ files with repeated field structure
- `ls -l` repeats field names for every file
- `ripgrep` has identical pattern to grep (already solved in Phase 33)
- Result: Massive token waste on repeated `{"path": "...", "type": "...", "size": ...}`

**Solution:**
- Use array tuples: `[path, type, size, permissions]`
- Add smart truncation (500 files max for find/ls, 200 for rg)
- Group by directory for ls
- Target: 30-35% token savings

## Affected Parsers

1. **find** - File search
2. **ls / ls -l** - Directory listings
3. **ripgrep (rg)** - Code search (same pattern as grep)

## Tasks

### 1. Update File Operation Types (internal/adapters/parsers/fileops/types.go)

- [ ] Create `FindOutputCompact` struct with array-based results
- [ ] Create `LsOutputCompact` struct with array-based results
- [ ] Create `RipgrepOutputCompact` struct (mirrors grep compact format)
- [ ] Add `FileTuple` type for find: `[path, type, size, modified_time]`
- [ ] Add `LsEntryTuple` type: `[name, size, permissions, owner, group, modified]`
- [ ] Add truncation metadata fields to all structs
- [ ] Keep old types temporarily for backward compatibility

### 2. Update Find Parser (internal/adapters/parsers/fileops/find.go)

- [ ] Modify `Parse()` to use array tuples instead of objects
- [ ] Implement truncation: 500 files maximum
- [ ] Track total files found vs displayed
- [ ] Extract file type (f=file, d=directory, l=symlink, etc.)
- [ ] Parse size from `-ls` output when available
- [ ] Parse modification time when available
- [ ] Filter out find warnings ("Permission denied", etc.)
- [ ] Update schema to reflect compact structure

### 3. Update Ls Parser (internal/adapters/parsers/fileops/ls.go)

- [ ] Group entries by directory if recursive `-R` used
- [ ] Use array tuples for file entries
- [ ] Handle both simple `ls` and detailed `ls -l` formats
- [ ] Implement truncation: 500 entries maximum per directory
- [ ] Parse permissions, owner, group, size, date from `-l` output
- [ ] Handle symlinks (show target)
- [ ] Filter directory total lines ("total 42")
- [ ] Update schema for compact format

### 4. Update Ripgrep Parser (internal/adapters/parsers/fileops/ripgrep.go)

- [ ] Apply same compact format as grep (Phase 33)
- [ ] Group matches by file
- [ ] Use tuple format: `[line, column, content]` (column if available)
- [ ] Implement truncation: 200 matches, 10 per file
- [ ] Handle ripgrep JSON mode (`--json`) if detected
- [ ] Filter ripgrep warnings and stats
- [ ] Update schema to match grep compact format

### 5. Update Tests

#### Find Tests (internal/adapters/parsers/fileops/find_test.go)
- [ ] Update `TestFindParser_BasicSearch` for tuple format
- [ ] Update `TestFindParser_WithType` for type extraction
- [ ] Update `TestFindParser_WithSize` for size parsing
- [ ] Add `TestFindParser_Truncation` for 500-file limit
- [ ] Add `TestFindParser_TruncationCount` to verify metadata
- [ ] Add `TestFindParser_PermissionDenied` for warning filtering
- [ ] Add `TestFindParser_EmptyResults` for no matches
- [ ] Update `TestFindParser_Schema` for new format

#### Ls Tests (internal/adapters/parsers/fileops/ls_test.go)
- [ ] Update `TestLsParser_SimpleList` for basic output
- [ ] Update `TestLsParser_LongFormat` for `-l` parsing
- [ ] Add `TestLsParser_DirectoryGrouping` for `-R` recursive
- [ ] Add `TestLsParser_Truncation` for 500-entry limit
- [ ] Add `TestLsParser_Symlinks` for symlink handling
- [ ] Add `TestLsParser_FilterTotals` to verify "total" lines filtered
- [ ] Update `TestLsParser_EmptyDirectory` for new format

#### Ripgrep Tests (internal/adapters/parsers/fileops/ripgrep_test.go)
- [ ] Mirror grep tests from Phase 33
- [ ] Add `TestRipgrepParser_WithColumn` for column numbers
- [ ] Add `TestRipgrepParser_JSONMode` for `--json` format
- [ ] Add truncation tests
- [ ] Add file grouping tests

### 6. Update JSON Schemas

- [ ] Update `schemas/find.json` for compact format
- [ ] Update `schemas/ls.json` for compact format
- [ ] Create `schemas/ripgrep.json` (mirror grep.json)
- [ ] Document tuple formats in each schema
- [ ] Add examples showing array structure
- [ ] Document truncation behavior

### 7. Update Feature Tests

- [ ] Update find scenarios in `features/fileops.feature`
- [ ] Update ls scenarios for new format
- [ ] Add ripgrep scenarios
- [ ] Add truncation scenarios for large directories
- [ ] Verify permission denied filtering

### 8. Token Savings Validation

- [ ] Benchmark find with 100 files
- [ ] Benchmark find with 1,000 files (truncation)
- [ ] Benchmark ls with 50 entries
- [ ] Benchmark ripgrep with 100 matches
- [ ] Verify 30-35% token savings target
- [ ] Measure before/after for common use cases

### 9. Documentation

- [ ] Update CLAUDE.md with find/ls/rg examples
- [ ] Update README.md file operations section
- [ ] Document truncation limits
- [ ] Add comments explaining tuple formats
- [ ] Document migration for users relying on old format

## Implementation Notes

### Find Compact Format

```json
{
  "total_files": 1523,
  "results": [
    ["./src/main.go", "f", 1024, "2024-01-15T10:30:00Z"],
    ["./src/util.go", "f", 512, "2024-01-14T15:20:00Z"],
    ["./build", "d", 4096, "2024-01-15T09:00:00Z"]
  ],
  "truncated": 1023
}
```

**Tuple Format**: `[path, type, size, modified_time]`
- `path` (string) - File path
- `type` (string) - "f" (file), "d" (directory), "l" (symlink), "b" (block), "c" (char), "p" (pipe), "s" (socket)
- `size` (int) - File size in bytes
- `modified_time` (string, optional) - ISO 8601 timestamp

### Ls Compact Format

**Simple ls**:
```json
{
  "entries": [
    "main.go",
    "util.go",
    "README.md"
  ]
}
```

**ls -l (detailed)**:
```json
{
  "total_entries": 45,
  "entries": [
    ["main.go", 1024, "-rw-r--r--", "user", "group", "2024-01-15T10:30:00Z"],
    ["util.go", 512, "-rw-r--r--", "user", "group", "2024-01-14T15:20:00Z"]
  ],
  "truncated": 0
}
```

**Tuple Format**: `[name, size, permissions, owner, group, modified]`

**ls -lR (recursive with directory grouping)**:
```json
{
  "total_entries": 234,
  "directories": [
    ["./src", 15, [
      ["main.go", 1024, "-rw-r--r--", "user", "group", "2024-01-15"],
      ["util.go", 512, "-rw-r--r--", "user", "group", "2024-01-14"]
    ]],
    ["./tests", 8, [
      ["main_test.go", 256, "-rw-r--r--", "user", "group", "2024-01-15"]
    ]]
  ],
  "truncated": 0
}
```

### Ripgrep Compact Format

**Same as grep (Phase 33)**:
```json
{
  "total": 156,
  "files": 23,
  "results": [
    ["src/main.go", 5, [
      [10, 15, "func main() {"],
      [45, 20, "    fmt.Println(\"test\")"]
    ]]
  ],
  "truncated": 0
}
```

**Tuple Format**: `[line, column, content]`
- Column is optional (0 if not available)

### Truncation Strategy

**Find**:
- **Limit**: 500 files maximum
- **Priority**: Directories first, then files (when using `-type`)
- **Metadata**: Track total found vs displayed

**Ls**:
- **Limit**: 500 entries per directory
- **Priority**: Directories first, then by name
- **Recursive**: Limit applies per directory

**Ripgrep**:
- **Same as grep**: 200 matches total, 10 per file

### Token Savings Math

**Find - 1000 files**:

Old format (~8,500 tokens):
```json
{"files": [
  {"path": "./src/components/Button.tsx", "type": "file", "size": 1024},
  // ... repeated 1000 times
]}
```

New format (~5,800 tokens after truncation to 500):
```json
{"total_files": 1000, "results": [
  ["./src/components/Button.tsx", "f", 1024],
  // ... 500 entries
], "truncated": 500}
```

**Savings**: ~32% reduction + truncation benefit

**Ls -l - 100 files**:

Old format (~3,200 tokens):
```json
{"files": [
  {"name": "main.go", "size": 1024, "permissions": "-rw-r--r--", "owner": "user", "group": "group"},
  // ... repeated 100 times
]}
```

New format (~2,100 tokens):
```json
{"entries": [
  ["main.go", 1024, "-rw-r--r--", "user", "group"],
  // ... 100 entries
]}
```

**Savings**: ~34% reduction

### Edge Cases to Handle

**Find**:
- Permission denied: Filter out warning lines
- No results: `{"total_files": 0, "results": [], "truncated": 0}`
- Symlinks: Show target in additional field or use "l" type
- Special files: Use appropriate type codes

**Ls**:
- Empty directory: `{"entries": []}`
- Total lines: Filter out "total 42" lines
- Symlinks: Show as `"name -> target"` in name field
- Hidden files: Include when `-a` used
- Color codes: Strip ANSI escape sequences

**Ripgrep**:
- Same edge cases as grep (Phase 33)
- JSON mode: Parse `--json` output if detected
- Binary files: Skip like grep

## Acceptance Criteria

- [ ] Find uses compact array format
- [ ] Ls uses compact array format (both simple and -l)
- [ ] Ripgrep mirrors grep compact format
- [ ] Token savings are 30-35% for typical outputs
- [ ] Truncation works for large outputs (500+ files)
- [ ] All existing tests pass with new format
- [ ] Feature tests pass
- [ ] Schemas are updated and validated
- [ ] Warning/error lines are filtered
- [ ] No regressions in functionality

## Success Metrics

- 30-35% token savings on file operations with 100+ results
- Find with 1000+ files truncated to 500
- Ls with large directories truncated appropriately
- Ripgrep achieves same savings as grep (25%)
- Token cost for 100 find results: ~1,400 tokens (vs ~2,100 currently)
