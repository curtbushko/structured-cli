// Package ports defines the interfaces (contracts) for the structured-cli.
// This layer only imports from domain - never from adapters or application.
// Adapters implement these interfaces; application depends on them.
package ports

import (
	"github.com/curtbushko/structured-cli/internal/domain"
)

// ThemeProvider provides color tokens for rendering statistics.
// Implementations return terminal escape codes, hex colors, or other format-specific tokens.
type ThemeProvider interface {
	// ColorFor returns the color token for the given savings category.
	// The returned string format depends on the implementation (ANSI codes, hex, etc.).
	ColorFor(category domain.SavingsCategory) string

	// Name returns the name of this theme (e.g., "default", "dark", "light").
	Name() string

	// ListThemes returns the names of all available themes provided by this implementation.
	ListThemes() []string

	// SetTheme persists the given theme name as the active selection.
	// Returns an error if the theme cannot be persisted.
	SetTheme(name string) error
}
