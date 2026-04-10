package theme

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

func TestFlairThemeProvider_ImplementsInterface(t *testing.T) {
	// given: FlairThemeProvider
	var provider ports.ThemeProvider = NewFlairThemeProvider()

	// then: it implements ThemeProvider interface (compiles)
	require.NotNil(t, provider)
}

func TestFlairThemeProvider_Name(t *testing.T) {
	// given: FlairThemeProvider loaded with flair.MustLoad()
	provider := NewFlairThemeProvider()

	// when: call Name()
	name := provider.Name()

	// then: returns non-empty string (the flair theme name)
	assert.NotEmpty(t, name, "Name() should return a non-empty string")
}

func TestFlairThemeProvider_ColorFor_Good(t *testing.T) {
	// given: FlairThemeProvider
	provider := NewFlairThemeProvider()

	// when: ColorFor(domain.SavingsCategoryGood)
	color := provider.ColorFor(domain.SavingsCategoryGood)

	// then: returns non-empty string
	assert.NotEmpty(t, color, "ColorFor(Good) should return a non-empty color string")
}

func TestFlairThemeProvider_ColorFor_Warning(t *testing.T) {
	// given: FlairThemeProvider
	provider := NewFlairThemeProvider()

	// when: ColorFor(domain.SavingsCategoryWarning)
	color := provider.ColorFor(domain.SavingsCategoryWarning)

	// then: returns non-empty string
	assert.NotEmpty(t, color, "ColorFor(Warning) should return a non-empty color string")
}

func TestFlairThemeProvider_ColorFor_Critical(t *testing.T) {
	// given: FlairThemeProvider
	provider := NewFlairThemeProvider()

	// when: ColorFor(domain.SavingsCategoryCritical)
	color := provider.ColorFor(domain.SavingsCategoryCritical)

	// then: returns non-empty string
	assert.NotEmpty(t, color, "ColorFor(Critical) should return a non-empty color string")
}

func TestFlairThemeProvider_ColorFor_DistinctValues(t *testing.T) {
	// given: FlairThemeProvider
	provider := NewFlairThemeProvider()

	// when: get colors for all categories
	good := provider.ColorFor(domain.SavingsCategoryGood)
	warning := provider.ColorFor(domain.SavingsCategoryWarning)
	critical := provider.ColorFor(domain.SavingsCategoryCritical)

	// then: each category has a distinct color
	assert.NotEqual(t, good, warning, "Good and Warning should have distinct colors")
	assert.NotEqual(t, warning, critical, "Warning and Critical should have distinct colors")
	assert.NotEqual(t, good, critical, "Good and Critical should have distinct colors")
}
