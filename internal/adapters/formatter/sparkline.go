package formatter

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

// sparklineChars are the Unicode braille/block characters used for sparkline bars,
// ordered from lowest to highest intensity.
var sparklineChars = []rune{' ', '▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

// GenerateSparkline creates a sparkline string from a series of integer values.
// Each value is mapped to a Unicode block character scaled relative to the range.
// Handles negative values by shifting the entire range to be positive.
// Returns an empty string for nil or empty input.
func GenerateSparkline(values []int) string {
	if len(values) == 0 {
		return ""
	}

	// Find min and max values
	minVal := values[0]
	maxVal := values[0]
	for _, v := range values {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	// Shift all values to be non-negative if there are negative values
	shift := 0
	if minVal < 0 {
		shift = -minVal
	}

	// Calculate range
	rangeVal := maxVal - minVal

	runes := make([]rune, len(values))
	for i, v := range values {
		if rangeVal == 0 {
			// All values are the same - use appropriate character
			if maxVal == 0 {
				runes[i] = sparklineChars[0] // all zeros -> space
			} else {
				runes[i] = sparklineChars[len(sparklineChars)-1] // all same non-zero -> full bar
			}
			continue
		}
		// Shift and scale value to sparkline character index (0-8)
		shiftedVal := v + shift
		idx := shiftedVal * (len(sparklineChars) - 1) / rangeVal
		if idx < 0 {
			idx = 0
		}
		if idx >= len(sparklineChars) {
			idx = len(sparklineChars) - 1
		}
		runes[i] = sparklineChars[idx]
	}

	return string(runes)
}

// RenderSparklineWithColor renders a sparkline with flair theme success color applied.
// Returns an empty string if values is nil or empty.
func RenderSparklineWithColor(values []int, theme ports.ThemeProvider) string {
	raw := GenerateSparkline(values)
	if raw == "" {
		return ""
	}

	color := theme.ColorFor(domain.SavingsCategoryGood)
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))

	return style.Render(raw)
}
