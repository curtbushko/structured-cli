package stats

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/adapters/theme"
)

func TestProgressRenderer_Render_AboveHalf(t *testing.T) {
	// given: a progress renderer with default theme and 60% savings
	tp := theme.NewDefaultThemeProvider()
	renderer := NewProgressRenderer(tp)
	var buf bytes.Buffer

	// when: Render is called with 60.0
	err := renderer.Render(&buf, 60.0)

	// then: output is non-empty and represents ~60% fill
	require.NoError(t, err)
	output := buf.String()
	assert.NotEmpty(t, output, "progress bar should have content")
}

func TestProgressRenderer_Render_Zero(t *testing.T) {
	// given: a progress renderer with 0% savings
	tp := theme.NewDefaultThemeProvider()
	renderer := NewProgressRenderer(tp)
	var buf bytes.Buffer

	// when: Render is called with 0.0
	err := renderer.Render(&buf, 0.0)

	// then: output is non-empty (empty bar still rendered)
	require.NoError(t, err)
	assert.NotEmpty(t, buf.String(), "empty bar should still be rendered")
}

func TestProgressRenderer_Render_Full(t *testing.T) {
	// given: a progress renderer with 100% savings
	tp := theme.NewDefaultThemeProvider()
	renderer := NewProgressRenderer(tp)
	var buf bytes.Buffer

	// when: Render is called with 100.0
	err := renderer.Render(&buf, 100.0)

	// then: output is non-empty
	require.NoError(t, err)
	assert.NotEmpty(t, buf.String(), "full bar should be rendered")
}

func TestProgressRenderer_Render_LowSavings(t *testing.T) {
	// given: a progress renderer with low savings (critical category)
	tp := theme.NewDefaultThemeProvider()
	renderer := NewProgressRenderer(tp)
	var buf bytes.Buffer

	// when: Render is called with 10.0
	err := renderer.Render(&buf, 10.0)

	// then: output is non-empty
	require.NoError(t, err)
	assert.NotEmpty(t, buf.String(), "low savings bar should be rendered")
}

func TestProgressRenderer_Render_ContainsPercentage(t *testing.T) {
	// given: a progress renderer
	tp := theme.NewDefaultThemeProvider()
	renderer := NewProgressRenderer(tp)
	var buf bytes.Buffer

	// when: Render is called with 45.5
	err := renderer.Render(&buf, 45.5)

	// then: output contains the percentage value
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "45.5%", "output should contain the percentage")
}
