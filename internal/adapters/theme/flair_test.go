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

func TestFlairThemeProvider_EfficiencyColorFor_Success(t *testing.T) {
	// given: FlairThemeProvider with flair theme
	provider := NewFlairThemeProvider()

	// when: get efficiency color for 85%
	color := provider.EfficiencyColorFor(85.0)

	// then: returns flair success color (same as ColorFor Good)
	expected := provider.ColorFor(domain.SavingsCategoryGood)
	assert.Equal(t, expected, color)
}

func TestFlairThemeProvider_EfficiencyColorFor_Warning(t *testing.T) {
	// given: FlairThemeProvider with flair theme
	provider := NewFlairThemeProvider()

	// when: get efficiency color for 65%
	color := provider.EfficiencyColorFor(65.0)

	// then: returns flair warning color
	expected := provider.ColorFor(domain.SavingsCategoryWarning)
	assert.Equal(t, expected, color)
}

func TestFlairThemeProvider_EfficiencyColorFor_Error(t *testing.T) {
	// given: FlairThemeProvider with flair theme
	provider := NewFlairThemeProvider()

	// when: get efficiency color for 40%
	color := provider.EfficiencyColorFor(40.0)

	// then: returns flair error color
	expected := provider.ColorFor(domain.SavingsCategoryCritical)
	assert.Equal(t, expected, color)
}

func TestFlairThemeProvider_EfficiencyColorFor_Boundary80(t *testing.T) {
	// given: FlairThemeProvider with flair theme
	provider := NewFlairThemeProvider()

	// when: get efficiency color for exactly 80%
	color := provider.EfficiencyColorFor(80.0)

	// then: returns warning (80 is not > 80)
	expected := provider.ColorFor(domain.SavingsCategoryWarning)
	assert.Equal(t, expected, color)
}

func TestFlairThemeProvider_EfficiencyColorFor_Boundary50(t *testing.T) {
	// given: FlairThemeProvider with flair theme
	provider := NewFlairThemeProvider()

	// when: get efficiency color for exactly 50%
	color := provider.EfficiencyColorFor(50.0)

	// then: returns critical (50 is not > 50)
	expected := provider.ColorFor(domain.SavingsCategoryCritical)
	assert.Equal(t, expected, color)
}

func TestFlairThemeProvider_ImpactGradientColor_FullBar(t *testing.T) {
	// given: FlairThemeProvider with flair theme
	provider := NewFlairThemeProvider()

	// when: get gradient color for impact=100%
	color := provider.ImpactGradientColor(100.0)

	// then: returns flair success color
	expected := provider.ColorFor(domain.SavingsCategoryGood)
	assert.Equal(t, expected, color)
}

func TestFlairThemeProvider_ImpactGradientColor_MidBar(t *testing.T) {
	// given: FlairThemeProvider with flair theme
	provider := NewFlairThemeProvider()

	// when: get gradient color for impact=50%
	color := provider.ImpactGradientColor(50.0)

	// then: returns flair warning color
	expected := provider.ColorFor(domain.SavingsCategoryWarning)
	assert.Equal(t, expected, color)
}

func TestFlairThemeProvider_ImpactGradientColor_LowBar(t *testing.T) {
	// given: FlairThemeProvider with flair theme
	provider := NewFlairThemeProvider()

	// when: get gradient color for impact=10%
	color := provider.ImpactGradientColor(10.0)

	// then: returns flair error color
	expected := provider.ColorFor(domain.SavingsCategoryCritical)
	assert.Equal(t, expected, color)
}
