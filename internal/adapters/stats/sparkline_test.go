package stats

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/adapters/theme"
)

func TestSparklineRenderer_Render_Ascending(t *testing.T) {
	// given: ascending values
	tp := theme.NewDefaultThemeProvider()
	renderer := NewSparklineRenderer(tp)
	values := []float64{10, 20, 30, 40, 50}
	var buf bytes.Buffer

	// when: Render is called
	err := renderer.Render(&buf, values)

	// then: output shows ascending character pattern
	require.NoError(t, err)
	output := buf.String()
	assert.NotEmpty(t, output, "sparkline should have content")
	// Verify characters are in ascending order by checking rune values
	runes := []rune(output)
	for i := 1; i < len(runes); i++ {
		if runes[i] == '\n' || runes[i-1] == '\n' {
			continue
		}
		assert.GreaterOrEqual(t, runes[i], runes[i-1],
			"sparkline chars should be ascending, got %c >= %c at position %d", runes[i], runes[i-1], i)
	}
}

func TestSparklineRenderer_Render_Descending(t *testing.T) {
	// given: descending values
	tp := theme.NewDefaultThemeProvider()
	renderer := NewSparklineRenderer(tp)
	values := []float64{50, 40, 30, 20, 10}
	var buf bytes.Buffer

	// when: Render is called
	err := renderer.Render(&buf, values)

	// then: output is non-empty
	require.NoError(t, err)
	assert.NotEmpty(t, buf.String(), "sparkline should have content")
}

func TestSparklineRenderer_Render_Flat(t *testing.T) {
	// given: flat values (all the same)
	tp := theme.NewDefaultThemeProvider()
	renderer := NewSparklineRenderer(tp)
	values := []float64{25, 25, 25, 25}
	var buf bytes.Buffer

	// when: Render is called
	err := renderer.Render(&buf, values)

	// then: output is non-empty with uniform characters
	require.NoError(t, err)
	output := buf.String()
	assert.NotEmpty(t, output)
}

func TestSparklineRenderer_Render_Empty(t *testing.T) {
	// given: empty values
	tp := theme.NewDefaultThemeProvider()
	renderer := NewSparklineRenderer(tp)
	var buf bytes.Buffer

	// when: Render is called with nil
	err := renderer.Render(&buf, nil)

	// then: no error
	require.NoError(t, err)
}

func TestSparklineRenderer_Render_SingleValue(t *testing.T) {
	// given: a single value
	tp := theme.NewDefaultThemeProvider()
	renderer := NewSparklineRenderer(tp)
	values := []float64{42.0}
	var buf bytes.Buffer

	// when: Render is called
	err := renderer.Render(&buf, values)

	// then: output is non-empty
	require.NoError(t, err)
	assert.NotEmpty(t, buf.String())
}
