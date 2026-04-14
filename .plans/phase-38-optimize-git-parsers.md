# Phase 38: Optimize Git Parsers - Ultra-Compact JSON Format

## Goal

Optimize git parsers (git log, git diff --stat, git blame, git reflog) to use ultra-compact array-based JSON format with smart truncation. Git log can produce thousands of commits and benefits significantly from compact representation.

## Context

**Current Problem:**
- `git log` can return thousands of commits with repeated field structure
- Each commit repeats: `{"hash": "...", "author": "...", "date": "...", "message": "..."}`
- `git diff --stat` repeats file info for every changed file
- Result: Massive token waste, especially for long histories

**Solution:**
- Use array tuples for commits: `[hash_short, author, date, subject]`
- Add smart truncation (100 commits default, 500 max for git log)
- Group file changes by type (added/modified/deleted) for diff
- Target: 25-30% token savings

## Affected Parsers

1. **git log** - Commit history
2. **git diff --stat** - File change statistics
3. **git blame** - Line-by-line attribution
4. **git reflog** - Reference logs

## Tasks

### 1. Update Git Types (internal/adapters/parsers/git/types.go)

- [ ] Create `GitLogCompact` struct with array-based commits
- [ ] Create `GitDiffStatCompact` struct grouped by change type
- [ ] Create `GitBlameCompact` struct with array-based lines
- [ ] Create `GitReflogCompact` struct with array-based entries
- [ ] Add `CommitTuple` type: `[hash_short, author_name, date, subject, files_changed, insertions, deletions]`
- [ ] Add `FileChangeTuple` type: `[filepath, insertions, deletions]`
- [ ] Add `BlameLine` type: `[line_number, hash_short, author, date, content]`
- [ ] Add `ReflogEntry` type: `[hash_short, ref_name, action, message]`
- [ ] Keep old types temporarily for backward compatibility

### 2. Update Git Log Parser (internal/adapters/parsers/git/log.go)

- [ ] Modify `Parse()` to use array tuples for commits
- [ ] Use short hash (7-8 chars) instead of full 40-char hash
- [ ] Extract author name only (not email) for compactness
- [ ] Parse date as relative ("2 days ago") or ISO8601
- [ ] Extract first line of commit message (subject only)
- [ ] Include file stats if `--stat` flag present
- [ ] Implement truncation: 100 commits default, configurable to 500
- [ ] Add `--max-count` detection to respect user limits
- [ ] Update schema to reflect compact structure

### 3. Update Git Diff Stat Parser (internal/adapters/parsers/git/diff.go)

- [ ] Group files by change type (added/modified/deleted)
- [ ] Use array tuples for file changes
- [ ] Parse insertions/deletions from stat line
- [ ] Include summary (total files, insertions, deletions)
- [ ] Handle binary file changes
- [ ] Implement truncation: 200 files max
- [ ] Update schema for grouped structure

### 4. Update Git Blame Parser (internal/adapters/parsers/git/blame.go)

- [ ] Use array tuples for blame lines
- [ ] Use short hash for commits
- [ ] Extract author name and date
- [ ] Include line number and content
- [ ] Group consecutive lines from same commit
- [ ] Implement truncation: 1000 lines max
- [ ] Update schema

### 5. Update Git Reflog Parser (internal/adapters/parsers/git/reflog.go)

- [ ] Use array tuples for reflog entries
- [ ] Use short hash
- [ ] Parse ref name (HEAD@{0}, HEAD@{1}, etc.)
- [ ] Extract action (commit, checkout, rebase, etc.)
- [ ] Include message/description
- [ ] Implement truncation: 100 entries max
- [ ] Update schema

### 6. Update Tests for All Git Parsers

#### Git Log Tests (internal/adapters/parsers/git/log_test.go)
- [ ] Update `TestGitLogParser_BasicLog` for tuple format
- [ ] Update `TestGitLogParser_WithStats` for file stats inclusion
- [ ] Add `TestGitLogParser_Truncation` for 100-commit limit
- [ ] Add `TestGitLogParser_MaxCount` to respect `--max-count`
- [ ] Add `TestGitLogParser_ShortHash` to verify hash truncation
- [ ] Add `TestGitLogParser_RelativeDate` for date parsing
- [ ] Update `TestGitLogParser_EmptyLog` for new format

#### Git Diff Stat Tests (internal/adapters/parsers/git/diff_test.go)
- [ ] Update for grouped file changes
- [ ] Test added/modified/deleted grouping
- [ ] Test binary file handling
- [ ] Add truncation tests
- [ ] Test summary calculation

#### Git Blame Tests (internal/adapters/parsers/git/blame_test.go)
- [ ] Update for tuple format
- [ ] Test consecutive line grouping
- [ ] Add truncation tests
- [ ] Test line number parsing

#### Git Reflog Tests (internal/adapters/parsers/git/reflog_test.go)
- [ ] Update for tuple format
- [ ] Test action parsing
- [ ] Add truncation tests

### 7. Update JSON Schemas

- [ ] Update `schemas/git-log.json` for compact format
- [ ] Update `schemas/git-diff-stat.json` for grouped format
- [ ] Update `schemas/git-blame.json` for compact format
- [ ] Update `schemas/git-reflog.json` for compact format
- [ ] Document tuple formats in each schema
- [ ] Add examples showing array structure

### 8. Update Feature Tests

- [ ] Update git scenarios in `features/git_commands.feature`
- [ ] Add scenario for git log truncation
- [ ] Add scenario for git diff --stat grouping
- [ ] Test various git log options
- [ ] Verify hash truncation

### 9. Token Savings Validation

- [ ] Benchmark git log with 20 commits
- [ ] Benchmark git log with 100 commits (truncation point)
- [ ] Benchmark git log with 500 commits (max truncation)
- [ ] Benchmark git diff --stat with 50 files
- [ ] Verify 25-30% token savings target
- [ ] Test on real repositories

### 10. Documentation

- [ ] Update CLAUDE.md with git parser examples
- [ ] Update README.md git section
- [ ] Document truncation limits and rationale
- [ ] Document short hash usage
- [ ] Document date format choices
- [ ] Add migration guide

## Implementation Notes

### Git Log Compact Format

```json
{
  "total_commits": 523,
  "commits": [
    ["a1b2c3d", "John Doe", "2024-01-15", "Fix authentication bug", 3, 45, 12],
    ["e4f5g6h", "Jane Smith", "2024-01-14", "Add user profile feature", 8, 234, 56],
    ["i7j8k9l", "Bob Johnson", "2024-01-13", "Update dependencies", 1, 5, 2]
  ],
  "truncated": 423
}
```

**Tuple Format**: `[hash_short, author_name, date, subject, files_changed, insertions, deletions]`
- `hash_short` (string) - First 7-8 chars of commit hash
- `author_name` (string) - Author name only (no email)
- `date` (string) - Relative ("2 days ago") or ISO8601
- `subject` (string) - First line of commit message
- `files_changed` (int, optional) - Number of files changed (if --stat)
- `insertions` (int, optional) - Lines added (if --stat)
- `deletions` (int, optional) - Lines removed (if --stat)

### Git Diff Stat Compact Format

```json
{
  "summary": {
    "files_changed": 12,
    "insertions": 234,
    "deletions": 56
  },
  "added": [
    ["src/new_feature.go", 145, 0],
    ["tests/new_feature_test.go", 89, 0]
  ],
  "modified": [
    ["src/main.go", 12, 8],
    ["README.md", 5, 3]
  ],
  "deleted": [
    ["src/old_code.go", 0, 45]
  ],
  "binary": [
    "images/logo.png"
  ],
  "truncated": 0
}
```

**Tuple Format**: `[filepath, insertions, deletions]`

**Grouping**:
- `added` - New files (deletions = 0)
- `modified` - Changed files (both insertions and deletions > 0)
- `deleted` - Removed files (insertions = 0)
- `binary` - Binary files (no line count)

### Git Blame Compact Format

```json
{
  "file": "src/main.go",
  "total_lines": 156,
  "lines": [
    [1, "a1b2c3d", "John Doe", "2024-01-10", "package main"],
    [2, "a1b2c3d", "John Doe", "2024-01-10", ""],
    [3, "a1b2c3d", "John Doe", "2024-01-10", "import ("],
    [4, "e4f5g6h", "Jane Smith", "2024-01-12", "    \"fmt\""]
  ],
  "truncated": 0
}
```

**Tuple Format**: `[line_number, hash_short, author, date, content]`

**Optimization**: Group consecutive lines from same commit:
```json
{
  "lines": [
    {"commit": "a1b2c3d", "author": "John Doe", "date": "2024-01-10", "lines": [
      [1, "package main"],
      [2, ""],
      [3, "import ("]
    ]},
    {"commit": "e4f5g6h", "author": "Jane Smith", "date": "2024-01-12", "lines": [
      [4, "    \"fmt\""]
    ]}
  ]
}
```

### Git Reflog Compact Format

```json
{
  "total_entries": 45,
  "entries": [
    ["a1b2c3d", "HEAD@{0}", "commit", "fix: authentication bug"],
    ["e4f5g6h", "HEAD@{1}", "commit", "feat: user profile"],
    ["i7j8k9l", "HEAD@{2}", "checkout", "moving from feature to main"]
  ],
  "truncated": 0
}
```

**Tuple Format**: `[hash_short, ref_name, action, message]`

### Truncation Strategy

**Git Log**:
- **Default limit**: 100 commits
- **Max limit**: 500 commits (with flag or env var)
- **Priority**: Most recent commits first (chronological)
- **User limits**: Respect `--max-count=N` if present

**Git Diff Stat**:
- **Limit**: 200 files max
- **Priority**: No special priority (order preserved)

**Git Blame**:
- **Limit**: 1000 lines max
- **Priority**: First N lines of file

**Git Reflog**:
- **Limit**: 100 entries max
- **Priority**: Most recent first

### Token Savings Math

**Git log - 100 commits**:

Old format (~12,000 tokens):
```json
{"commits": [
  {
    "hash": "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0",
    "author": "John Doe <john@example.com>",
    "date": "2024-01-15T10:30:00-05:00",
    "message": "Fix authentication bug\n\nThis commit fixes the authentication bug by updating the JWT token validation logic.",
    "files_changed": 3,
    "insertions": 45,
    "deletions": 12
  },
  // ... repeated 100 times
]}
```

New format (~8,400 tokens):
```json
{"total_commits": 100, "commits": [
  ["a1b2c3d", "John Doe", "2024-01-15", "Fix authentication bug", 3, 45, 12],
  // ... 100 entries
]}
```

**Savings**: ~30% reduction

**Git log - 500 commits** (with truncation):

Old format without truncation: ~60,000 tokens
New format with truncation to 100: ~8,400 tokens

**Savings with truncation**: ~86% reduction

### Edge Cases to Handle

**Git Log**:
- Empty repository: `{"total_commits": 0, "commits": []}`
- Merge commits: Include in regular commits
- No stats: Omit files_changed/insertions/deletions from tuple
- Initial commit: Handle as regular commit
- Multi-line messages: Use only first line (subject)

**Git Diff Stat**:
- No changes: `{"summary": {"files_changed": 0}, "added": [], "modified": [], "deleted": []}`
- Binary files: List in separate array
- Renamed files: Show as deleted + added (or special "renamed" group)

**Git Blame**:
- Not in git repo: Return error
- Binary file: Return error or empty
- New file (no history): All lines from one commit

**Git Reflog**:
- Fresh repository: Few entries
- Actions: commit, checkout, rebase, merge, cherry-pick, etc.

### Hash Truncation

Use **7-8 character** short hashes:
- Git default is 7 chars
- 8 chars for larger repos (more unique)
- Include full hash in metadata field if needed for tools

### Date Format Options

1. **Relative** (preferred for compactness): "2 days ago", "3 weeks ago"
2. **ISO8601**: "2024-01-15T10:30:00Z"
3. **Unix timestamp**: 1705320600

Choose relative for best readability and compactness.

## Acceptance Criteria

- [ ] All 4 git parsers use compact array format
- [ ] Commits use short hashes (7-8 chars)
- [ ] Token savings are 25-30% for typical outputs
- [ ] Truncation works correctly for large histories
- [ ] Git log respects `--max-count` when present
- [ ] Diff stat groups files by change type
- [ ] All existing tests pass with new format
- [ ] Feature tests pass
- [ ] Schemas are updated and validated
- [ ] No regressions in git functionality

## Success Metrics

- 25-30% token savings on git log with 100 commits
- 86% token savings with truncation on large histories (500+ commits → 100)
- Consistent short hash usage across all git parsers
- Token cost for 100 git log commits: ~8,400 tokens (vs ~12,000 currently)
- Git log with 500+ commits automatically truncated to 100
