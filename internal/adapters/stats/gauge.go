package stats

import (
	"fmt"
	"io"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

const (
	// gaugeWidth is the total width of the gauge bar in characters.
	gaugeWidth = 30
	// gaugeMaxRatio is the maximum ratio for full-scale gauge display.
	gaugeMaxRatio = 10.0
	// gaugeFillChar is the character used for the filled portion.
	gaugeFillChar = "\u2588" // Full block: █
	// gaugeEmptyChar is the character used for the empty portion.
	gaugeEmptyChar = "\u2591" // Light shade: ░
)

// GaugeRenderer renders a compression ratio as a horizontal gauge.
// It uses a ThemeProvider for color-coding based on the ratio level.
type GaugeRenderer struct {
	theme ports.ThemeProvider
}

// NewGaugeRenderer creates a new GaugeRenderer with the given theme.
func NewGaugeRenderer(theme ports.ThemeProvider) *GaugeRenderer {
	return &GaugeRenderer{theme: theme}
}

// Render writes a color-coded gauge representing the compression ratio.
// A ratio of 1.0 means no compression; higher values indicate more compression.
func (r *GaugeRenderer) Render(w io.Writer, ratio float64) error {
	// Clamp to gauge range
	if ratio < 0 {
		ratio = 0
	}

	// Determine category based on ratio
	category := ratioCategory(ratio)
	color := r.theme.ColorFor(category)
	reset := "\033[0m"

	// Calculate fill proportional to ratio (capped at gaugeMaxRatio)
	displayRatio := ratio
	if displayRatio > gaugeMaxRatio {
		displayRatio = gaugeMaxRatio
	}

	filled := int(displayRatio / gaugeMaxRatio * gaugeWidth)
	if filled > gaugeWidth {
		filled = gaugeWidth
	}
	empty := gaugeWidth - filled

	bar := color +
		strings.Repeat(gaugeFillChar, filled) +
		strings.Repeat(gaugeEmptyChar, empty) +
		reset

	_, err := fmt.Fprintf(w, "%s %.1fx\n", bar, ratio)
	return err
}

// ratioCategory maps a compression ratio to a savings category for theming.
func ratioCategory(ratio float64) domain.SavingsCategory {
	switch {
	case ratio >= 3.0:
		return domain.SavingsCategoryGood
	case ratio >= 1.5:
		return domain.SavingsCategoryWarning
	default:
		return domain.SavingsCategoryCritical
	}
}
