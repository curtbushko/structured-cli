// Package theme provides theme implementations for colorized output.
// This adapter package implements ports.ThemeProvider using various theming approaches.
package theme

import (
	"github.com/curtbushko/structured-cli/internal/domain"
)

// ANSI color codes for terminal output.
const (
	// Green ANSI escape code for good savings (>50%).
	colorGood = "\033[32m"
	// Yellow ANSI escape code for warning savings (20-50%).
	colorWarning = "\033[33m"
	// Red ANSI escape code for critical savings (<20%).
	colorCritical = "\033[31m"
)

// DefaultThemeProvider is a fallback theme using ANSI color codes.
// It provides hardcoded colors: green for good, yellow for warning, red for critical.
type DefaultThemeProvider struct{}

// NewDefaultThemeProvider creates a new DefaultThemeProvider instance.
func NewDefaultThemeProvider() *DefaultThemeProvider {
	return &DefaultThemeProvider{}
}

// ColorFor returns the ANSI color code for the given savings category.
func (p *DefaultThemeProvider) ColorFor(category domain.SavingsCategory) string {
	switch category {
	case domain.SavingsCategoryGood:
		return colorGood
	case domain.SavingsCategoryWarning:
		return colorWarning
	case domain.SavingsCategoryCritical:
		return colorCritical
	default:
		return colorWarning // Default to warning for unknown categories
	}
}

// EfficiencyColorFor returns the ANSI color for the given efficiency percentage.
// Thresholds: >80% = good, >50% = warning, <=50% = critical.
func (p *DefaultThemeProvider) EfficiencyColorFor(percent float64) string {
	switch {
	case percent > 80.0:
		return p.ColorFor(domain.SavingsCategoryGood)
	case percent > 50.0:
		return p.ColorFor(domain.SavingsCategoryWarning)
	default:
		return p.ColorFor(domain.SavingsCategoryCritical)
	}
}

// ImpactGradientColor returns the ANSI color for the given impact percentage.
// Gradient: >70% = good, >30% = warning, <=30% = critical.
func (p *DefaultThemeProvider) ImpactGradientColor(impact float64) string {
	switch {
	case impact > 70.0:
		return p.ColorFor(domain.SavingsCategoryGood)
	case impact > 30.0:
		return p.ColorFor(domain.SavingsCategoryWarning)
	default:
		return p.ColorFor(domain.SavingsCategoryCritical)
	}
}

// Name returns the name of this theme.
func (p *DefaultThemeProvider) Name() string {
	return "default"
}

// ListThemes returns the single "default" theme name.
func (p *DefaultThemeProvider) ListThemes() []string {
	return []string{"default"}
}

// SetTheme is a no-op for the default theme provider since it only supports one theme.
func (p *DefaultThemeProvider) SetTheme(_ string) error {
	return nil
}
