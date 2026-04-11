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
// Each value is mapped to a Unicode block character scaled relative to the maximum value.
// Returns an empty string for nil or empty input.
func GenerateSparkline(values []int) string {
	if len(values) == 0 {
		return ""
	}

	maxVal := 0
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}

	runes := make([]rune, len(values))
	for i, v := range values {
		if maxVal == 0 {
			runes[i] = sparklineChars[0] // space for all-zero case
			continue
		}
		// Scale value to sparkline character index (0-8)
		idx := v * (len(sparklineChars) - 1) / maxVal
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
