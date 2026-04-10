# Phase 27: Enhanced Stats Output

## Overview

Improve the stats output with rich tables, progress bars, and visualizations using Charm libraries (lipgloss, bubbles). Use [flair](https://github.com/curtbushko/flair) for consistent theming across all output components.

## Dependencies

- `github.com/charmbracelet/lipgloss` - Styling
- `github.com/charmbracelet/bubbles/table` - Table rendering
- `github.com/charmbracelet/bubbles/progress` - Progress bars
- `github.com/curtbushko/flair` - Theme management

## Tasks

### Domain & Ports

- [x] Define `domain/stats.go` - Stats types (TokenStats, CompressionStats, CommandStats)
- [x] Define `ports/stats.go` - StatsRenderer interface
- [x] Define `ports/theme.go` - ThemeProvider interface

### Adapters - Rendering

- [x] Implement `adapters/stats/table.go` - Table renderer using bubbles/table
- [x] Implement `adapters/stats/progress.go` - Progress bar for token savings percentage
- [x] Implement `adapters/stats/sparkline.go` - Sparkline for historical trends
- [x] Implement `adapters/stats/gauge.go` - Gauge component for compression ratio

### Adapters - Theming

- [x] Implement `adapters/theme/flair.go` - Flair theme adapter
- [x] Create theme presets for stats output (colors for good/warning/critical savings)

### CLI Integration

- [x] Add `--stats` flag to show enhanced stats after command execution
- [x] Add `--theme` flag to select theme (uses flair's built-in themes)
- [x] Add `structured-cli theme list` subcommand to show available themes
- [x] Add `structured-cli theme set <name>` subcommand to set default theme

### Testing

- [x] Unit tests for stats domain types
- [x] Unit tests for table renderer
- [x] Unit tests for progress bar renderer
- [x] Unit tests for sparkline/gauge components
- [x] Unit tests for theme adapter
- [x] Integration tests for CLI flags

## Notes

- Flair provides zero-setup theme loading: `theme := flair.MustLoad()`
- Use flair's bubbles adapters for pre-configured table/list styles
- Token savings visualization: show raw bytes vs structured bytes with percentage bar
- Consider color-coding: green (>50% savings), yellow (20-50%), red (<20%)
