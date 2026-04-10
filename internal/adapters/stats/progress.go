package stats

import (
	"fmt"
	"io"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

const (
	// progressBarWidth is the total width of the progress bar in characters.
	progressBarWidth = 30
	// progressFillChar is the character used for the filled portion.
	progressFillChar = "\u2588" // Full block: █
	// progressEmptyChar is the character used for the empty portion.
	progressEmptyChar = "\u2591" // Light shade: ░
)

// ProgressRenderer renders a token savings percentage as a progress bar.
// It uses a ThemeProvider for color-coding based on savings category.
type ProgressRenderer struct {
	theme ports.ThemeProvider
}

// NewProgressRenderer creates a new ProgressRenderer with the given theme.
func NewProgressRenderer(theme ports.ThemeProvider) *ProgressRenderer {
	return &ProgressRenderer{theme: theme}
}

// Render writes a color-coded progress bar representing the savings percentage.
func (r *ProgressRenderer) Render(w io.Writer, savingsPercent float64) error {
	// Clamp to 0-100 range
	if savingsPercent < 0 {
		savingsPercent = 0
	}
	if savingsPercent > 100 {
		savingsPercent = 100
	}

	// Determine category and color
	category := categorize(savingsPercent)
	color := r.theme.ColorFor(category)
	reset := "\033[0m"

	// Calculate fill
	filled := int(savingsPercent / 100 * progressBarWidth)
	empty := progressBarWidth - filled

	bar := color +
		strings.Repeat(progressFillChar, filled) +
		strings.Repeat(progressEmptyChar, empty) +
		reset

	_, err := fmt.Fprintf(w, "%s %.1f%%\n", bar, savingsPercent)
	return err
}

// categorize returns the SavingsCategory for the given percentage.
func categorize(percent float64) domain.SavingsCategory {
	switch {
	case percent > 50:
		return domain.SavingsCategoryGood
	case percent >= 20:
		return domain.SavingsCategoryWarning
	default:
		return domain.SavingsCategoryCritical
	}
}
