# Phase 29: Stats Output - Table Format

## Tasks

- [x] Add box-drawing header with scope info
- [x] Add aggregated stats (commands, tokens, time, efficiency meter)
- [x] Create 'By Command' table with proper columns
- [x] Add visual impact bars (block characters)
- [x] Implement efficiency meter visualization
- [x] Add column alignment and formatting
- [x] Update tests for new format

## Notes

Target format example:
```
Token Savings (Global Scope)
════════════════════════════════════════════════════════════

Total commands:    1148
Input tokens:      27.9M
Output tokens:     1.1M
Tokens saved:      26.8M (96.0%)
Total exec time:   13m4s (avg 683ms)
Efficiency meter: ███████████████████████░ 96.0%

By Command
────────────────────────────────────────────────────────────────────────
  #  Command                   Count   Saved    Avg%    Time  Impact
────────────────────────────────────────────────────────────────────────
 1.  go test -race ./...          14    6.8M  100.0%    5.3s  ██████████
 2.  go test ./...                26    5.8M  100.0%   762ms  █████████░
 3.  go test ./feature...         26    1.5M   99.8%    4.6s  ██░░░░░░░░
```

- Commands displayed without 'rtk' prefix
- Box-drawing characters for headers (═ and ─)
- Block characters for visual bars (█ and ░)
- Proper column alignment
- Human-readable number formatting (M, K suffixes)
