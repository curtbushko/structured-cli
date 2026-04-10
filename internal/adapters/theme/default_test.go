package theme

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

func TestDefaultThemeProvider_ImplementsInterface(t *testing.T) {
	// given: a DefaultThemeProvider
	var provider ports.ThemeProvider = NewDefaultThemeProvider()

	// then: it implements ThemeProvider interface
	require.NotNil(t, provider)
}

func TestDefaultThemeProvider_ColorFor_Good(t *testing.T) {
	// given: DefaultThemeProvider instance
	provider := NewDefaultThemeProvider()

	// when: ColorFor(domain.SavingsCategoryGood)
	color := provider.ColorFor(domain.SavingsCategoryGood)

	// then: returns non-empty string (green ANSI or lipgloss color)
	assert.NotEmpty(t, color, "ColorFor(Good) should return a non-empty color string")
}

func TestDefaultThemeProvider_ColorFor_Warning(t *testing.T) {
	// given: DefaultThemeProvider instance
	provider := NewDefaultThemeProvider()

	// when: ColorFor(domain.SavingsCategoryWarning)
	color := provider.ColorFor(domain.SavingsCategoryWarning)

	// then: returns non-empty string (yellow ANSI or lipgloss color)
	assert.NotEmpty(t, color, "ColorFor(Warning) should return a non-empty color string")
}

func TestDefaultThemeProvider_ColorFor_Critical(t *testing.T) {
	// given: DefaultThemeProvider instance
	provider := NewDefaultThemeProvider()

	// when: ColorFor(domain.SavingsCategoryCritical)
	color := provider.ColorFor(domain.SavingsCategoryCritical)

	// then: returns non-empty string (red ANSI or lipgloss color)
	assert.NotEmpty(t, color, "ColorFor(Critical) should return a non-empty color string")
}

func TestDefaultThemeProvider_Name(t *testing.T) {
	// given: DefaultThemeProvider instance
	provider := NewDefaultThemeProvider()

	// when: call Name()
	name := provider.Name()

	// then: returns 'default'
	assert.Equal(t, "default", name)
}

func TestDefaultThemeProvider_ColorFor_DistinctValues(t *testing.T) {
	// given: DefaultThemeProvider instance
	provider := NewDefaultThemeProvider()

	// when: get colors for all categories
	good := provider.ColorFor(domain.SavingsCategoryGood)
	warning := provider.ColorFor(domain.SavingsCategoryWarning)
	critical := provider.ColorFor(domain.SavingsCategoryCritical)

	// then: each category has a distinct color
	assert.NotEqual(t, good, warning, "Good and Warning should have distinct colors")
	assert.NotEqual(t, warning, critical, "Warning and Critical should have distinct colors")
	assert.NotEqual(t, good, critical, "Good and Critical should have distinct colors")
}
