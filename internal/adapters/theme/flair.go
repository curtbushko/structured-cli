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
