package stats

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/adapters/theme"
)

func TestGaugeRenderer_Render_HighRatio(t *testing.T) {
	// given: a gauge renderer with high compression ratio
	tp := theme.NewDefaultThemeProvider()
	renderer := NewGaugeRenderer(tp)
	var buf bytes.Buffer

	// when: Render is called with ratio 3.0 (300%)
	err := renderer.Render(&buf, 3.0)

	// then: output is non-empty and shows ratio value
	require.NoError(t, err)
	output := buf.String()
	assert.NotEmpty(t, output, "gauge should have content")
	assert.Contains(t, output, "3.0", "gauge should show ratio value")
}

func TestGaugeRenderer_Render_NoCompression(t *testing.T) {
	// given: a gauge renderer with no compression (ratio 1.0)
	tp := theme.NewDefaultThemeProvider()
	renderer := NewGaugeRenderer(tp)
	var buf bytes.Buffer

	// when: Render is called with ratio 1.0
	err := renderer.Render(&buf, 1.0)

	// then: output is non-empty
	require.NoError(t, err)
	assert.NotEmpty(t, buf.String(), "gauge should render even with no compression")
}

func TestGaugeRenderer_Render_ZeroRatio(t *testing.T) {
	// given: a gauge renderer with zero ratio
	tp := theme.NewDefaultThemeProvider()
	renderer := NewGaugeRenderer(tp)
	var buf bytes.Buffer

	// when: Render is called with ratio 0.0
	err := renderer.Render(&buf, 0.0)

	// then: output is non-empty
	require.NoError(t, err)
	assert.NotEmpty(t, buf.String(), "gauge should render even with zero ratio")
}

func TestGaugeRenderer_Render_VeryHighRatio(t *testing.T) {
	// given: a gauge renderer with very high ratio
	tp := theme.NewDefaultThemeProvider()
	renderer := NewGaugeRenderer(tp)
	var buf bytes.Buffer

	// when: Render is called with ratio 10.0
	err := renderer.Render(&buf, 10.0)

	// then: output is non-empty and contains ratio
	require.NoError(t, err)
	output := buf.String()
	assert.NotEmpty(t, output)
	assert.Contains(t, output, "10.0", "gauge should show ratio value")
}

func TestGaugeRenderer_Render_ContainsLabel(t *testing.T) {
	// given: a gauge renderer
	tp := theme.NewDefaultThemeProvider()
	renderer := NewGaugeRenderer(tp)
	var buf bytes.Buffer

	// when: Render is called
	err := renderer.Render(&buf, 2.5)

	// then: output contains a ratio indicator
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "2.5x", "gauge should show ratio with 'x' suffix")
}
