package theme

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/curtbushko/flair/pkg/flair"
	"github.com/curtbushko/structured-cli/internal/domain"
)

// FlairThemeProvider provides theme colors using the flair theming library.
// It loads the currently selected flair theme and maps semantic status colors
// to SavingsCategory values.
type FlairThemeProvider struct {
	theme *flair.Theme
}

// NewFlairThemeProvider creates a new FlairThemeProvider using flair.MustLoad().
// It loads the currently selected flair theme or falls back to the built-in default.
func NewFlairThemeProvider() *FlairThemeProvider {
	return &FlairThemeProvider{
		theme: flair.MustLoad(),
	}
}

// ColorFor returns the lipgloss-rendered hex color string for the given savings category.
// It maps savings categories to flair status colors:
//   - Good -> status.success (green)
//   - Warning -> status.warning (yellow)
//   - Critical -> status.error (red)
func (p *FlairThemeProvider) ColorFor(category domain.SavingsCategory) string {
	status := p.theme.Status()

	var color flair.Color
	switch category {
	case domain.SavingsCategoryGood:
		color = status.Success
	case domain.SavingsCategoryWarning:
		color = status.Warning
	case domain.SavingsCategoryCritical:
		color = status.Error
	default:
		color = status.Warning // Default to warning for unknown categories
	}

	// Convert flair.Color to lipgloss.Color and return as string
	return string(lipgloss.Color(color.Hex()))
}

// EfficiencyColorFor returns the color for the given efficiency percentage.
// Thresholds: >80% = success, >50% = warning, <=50% = error.
func (p *FlairThemeProvider) EfficiencyColorFor(percent float64) string {
	switch {
	case percent > 80.0:
		return p.ColorFor(domain.SavingsCategoryGood)
	case percent > 50.0:
		return p.ColorFor(domain.SavingsCategoryWarning)
	default:
		return p.ColorFor(domain.SavingsCategoryCritical)
	}
}

// ImpactGradientColor returns the color for the given impact percentage.
// Gradient: >70% = success, >30% = warning, <=30% = error.
func (p *FlairThemeProvider) ImpactGradientColor(impact float64) string {
	switch {
	case impact > 70.0:
		return p.ColorFor(domain.SavingsCategoryGood)
	case impact > 30.0:
		return p.ColorFor(domain.SavingsCategoryWarning)
	default:
		return p.ColorFor(domain.SavingsCategoryCritical)
	}
}

// Name returns the name of the flair theme being used.
func (p *FlairThemeProvider) Name() string {
	return p.theme.Name()
}

// ListThemes returns all available built-in flair theme names.
func (p *FlairThemeProvider) ListThemes() []string {
	return flair.ListBuiltins()
}

// SetTheme installs and selects the given theme name via the flair store.
// Returns an error if the theme is not a valid built-in or cannot be persisted.
func (p *FlairThemeProvider) SetTheme(name string) error {
	if !flair.HasBuiltin(name) {
		return fmt.Errorf("theme %q is not a valid built-in theme", name)
	}

	store := flair.NewStore()
	if err := store.Install(name); err != nil {
		return fmt.Errorf("install theme %q: %w", name, err)
	}
	if err := store.Select(name); err != nil {
		return fmt.Errorf("select theme %q: %w", name, err)
	}
	return nil
}
