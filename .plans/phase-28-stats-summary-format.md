# Phase 28: Stats Output - Aggregated Summary Format

## Overview

Redesign the `stats` subcommand output with three main sections: header with box-drawing characters, summary with efficiency meter bar, and command table with impact visualization. Keep the `--history` flag for detailed command history view. Uses advanced Charm ecosystem features for polished, adaptive output.

## Dependencies

- `github.com/charmbracelet/lipgloss` - Styling, borders, adaptive colors
- `github.com/charmbracelet/bubbles/progress` - Efficiency meter progress bar
- `github.com/charmbracelet/bubbles/viewport` - Scrollable command list
- `github.com/charmbracelet/bubbles/table` - Command table (from Phase 27)
- `github.com/curtbushko/flair` - Theme integration (from Phase 27)

## Tasks

### Domain Updates

- [x] Add command aggregation types to `domain/stats.go` (CommandStats with count, saved tokens, avg%, time)
- [x] Add grouping logic for aggregating commands by name (strip variable parts like paths)
- [x] Add impact calculation (sort commands by tokens saved)

### Formatter - Header Section

- [x] Implement header formatter with title "Token Savings (Global Scope)"
- [x] Add box-drawing character separator line (═══)
- [x] Support dynamic width based on terminal size
- [x] Add rounded border around header using lipgloss.RoundedBorder()
- [x] Use adaptive colors for title (light/dark mode support)

### Formatter - Summary Section

- [x] Implement summary formatter with aligned key-value pairs
- [x] Add large number formatting (27.5M, 1.1M instead of raw numbers)
- [x] Add duration formatting (12m20s, avg 696ms)
- [x] Add percentage calculation for tokens saved
- [x] Implement efficiency meter using bubbles/progress (instead of raw █░)
- [x] Add color-coded efficiency using flair theme colors: success (>80%), warning (50-80%), error (<50%)
- [x] Add sparkline showing token savings trend over time using flair theme colors
- [x] Add rounded border around summary section
- [x] Use adaptive colors for all text elements

### Formatter - Command Table

- [x] Implement "By Command" section header with separator (────)
- [x] Add table with columns: #, Command, Count, Saved, Avg%, Time, Impact
- [x] Truncate long commands with ellipsis (e.g., "go test ./feature...")
- [x] Add gradient impact bars using flair theme colors (success → warning → error)
- [x] Scale impact bars based on relative token savings (highest = full bar)
- [x] Align columns properly with padding
- [x] Add viewport for scrollable command list if >10 commands
- [x] Add rounded border around command table section
- [x] Use adaptive colors for table headers and data

### Theme Integration

- [x] Load flair theme for all color references (success, warning, error)
- [x] Map efficiency thresholds to flair semantic colors
- [x] Map impact bar gradients to flair color scales

### CLI Integration

- [x] Update `stats` subcommand to use new aggregated summary formatter by default
- [x] Keep `--history` flag to show detailed per-execution history (existing behavior)

### Testing

- [x] Unit tests for command aggregation logic
- [x] Unit tests for large number formatting (K, M, B suffixes)
- [x] Unit tests for bubbles/progress efficiency meter
- [x] Unit tests for color-coded efficiency thresholds using flair theme (80%, 50%)
- [x] Unit tests for gradient impact bar color calculation using flair theme
- [x] Unit tests for impact bar scaling
- [x] Unit tests for command truncation
- [x] Unit tests for sparkline data generation
- [x] Unit tests for viewport scrolling with >10 commands
- [x] Unit tests for adaptive color selection
- [x] Integration tests for stats command with summary format
- [x] Integration tests for --history flag compatibility

## Notes

### Charm Packages
- Use `github.com/charmbracelet/bubbles/progress` for efficiency meter
- Use `github.com/charmbracelet/bubbles/viewport` for scrollable command list
- Use `lipgloss.AdaptiveColor{Light: "#333", Dark: "#fff"}` for light/dark support
- Use `lipgloss.RoundedBorder()` for section borders

### Color Coding (from Flair Theme)
- Efficiency thresholds: >80% success color, 50-80% warning color, <50% error color
- Impact bar gradients: success → warning → error (flair semantic colors)
- Use flair's adaptive colors for light/dark terminal support
- All colors sourced from flair theme (no hardcoded hex values)

### Visual Elements
- Efficiency meter: bubbles/progress component with color-coded percentage
- Impact bars: gradient colors based on relative token savings
- Sparkline: mini trend graph showing token savings over last N executions
- Borders: rounded borders around header, summary, and table sections
- Command truncation: keep first part + "..." for readability

### Viewport
- Show top 10 commands by default
- If >10 commands, enable viewport scrolling
- Display scroll indicator at bottom

### Flair Theme Integration
- All colors come from flair theme semantic colors (success, warning, error, etc.)
- Use flair's adaptive color support for light/dark terminals
- No hardcoded color values (#00ff00, etc.) - only flair theme references
- Flair theme from Phase 27 provides consistent styling across all output
- Box drawing chars: ═ (double horizontal), ─ (single horizontal)
