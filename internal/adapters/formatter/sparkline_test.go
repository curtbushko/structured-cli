package formatter

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateSparkline_TrendData(t *testing.T) {
	// given: an array of token savings values over time
	values := []int{1000, 1500, 2000, 1800, 2500}

	// when: generating sparkline
	result := GenerateSparkline(values)

	// then: returns visual representation with bars scaled to max value
	require.NotEmpty(t, result)
	// Each value maps to a sparkline character
	assert.Len(t, []rune(result), len(values))
	// The highest value (2500) should produce the tallest bar (█)
	runes := []rune(result)
	assert.Equal(t, '█', runes[4], "max value should produce tallest bar")
}

func TestGenerateSparkline_ScalingToMax(t *testing.T) {
	// given: values where one is clearly the max
	values := []int{0, 500, 1000}

	// when: generating sparkline
	result := GenerateSparkline(values)

	// then: bars are scaled relative to max
	runes := []rune(result)
	require.Len(t, runes, 3)
	// Zero should produce lowest bar
	assert.Equal(t, ' ', runes[0], "zero value should produce space")
	// Max should produce tallest bar
	assert.Equal(t, '█', runes[2], "max value should produce tallest bar")
}

func TestSparkline_EmptyData(t *testing.T) {
	// given: empty array of savings values
	values := []int{}

	// when: generating sparkline
	result := GenerateSparkline(values)

	// then: returns empty string
	assert.Empty(t, result)
}

func TestSparkline_NilData(t *testing.T) {
	// given: nil savings values
	// when: generating sparkline
	result := GenerateSparkline(nil)

	// then: returns empty string
	assert.Empty(t, result)
}

func TestSparkline_SingleValue(t *testing.T) {
	// given: array with single savings value
	values := []int{1000}

	// when: generating sparkline
	result := GenerateSparkline(values)

	// then: returns single bar at full height
	runes := []rune(result)
	require.Len(t, runes, 1)
	assert.Equal(t, '█', runes[0], "single value should be full bar")
}

func TestSparkline_AllSameValues(t *testing.T) {
	// given: all values are the same
	values := []int{500, 500, 500}

	// when: generating sparkline
	result := GenerateSparkline(values)

	// then: all bars should be at full height
	runes := []rune(result)
	require.Len(t, runes, 3)
	for i, r := range runes {
		assert.Equal(t, '█', r, "bar %d should be full for equal values", i)
	}
}

func TestSparkline_AllZeros(t *testing.T) {
	// given: all values are zero
	values := []int{0, 0, 0}

	// when: generating sparkline
	result := GenerateSparkline(values)

	// then: all bars should be spaces (no height)
	runes := []rune(result)
	require.Len(t, runes, 3)
	for i, r := range runes {
		assert.Equal(t, ' ', r, "bar %d should be space for zero values", i)
	}
}

func TestRenderSparklineWithColor_UsesFlairThemeColor(t *testing.T) {
	// given: sparkline values and a theme
	values := []int{1000, 1500, 2000, 1800, 2500}
	theme := newTestTheme()

	// when: rendering sparkline with flair color
	result := RenderSparklineWithColor(values, theme)

	// then: output is non-empty and contains sparkline characters
	require.NotEmpty(t, result)
	// The styled string should contain the raw sparkline chars somewhere
	rawSparkline := GenerateSparkline(values)
	assert.Contains(t, result, rawSparkline)
}

func TestRenderSparklineWithColor_EmptyValues(t *testing.T) {
	// given: empty values
	theme := newTestTheme()

	// when: rendering sparkline with color
	result := RenderSparklineWithColor(nil, theme)

	// then: returns empty string
	assert.Empty(t, result)
}

func TestGenerateSparkline_NegativeValues(t *testing.T) {
	// given: token savings with some negative values (JSON larger than raw)
	values := []int{-100, 0, 500, 1000}

	// when: generating sparkline
	result := GenerateSparkline(values)

	// then: handles negative values by shifting range
	runes := []rune(result)
	require.Len(t, runes, 4)
	// Min value (-100) should produce lowest bar
	assert.Equal(t, ' ', runes[0], "min value should produce lowest bar")
	// Max value (1000) should produce tallest bar
	assert.Equal(t, '█', runes[3], "max value should produce tallest bar")
}

func TestGenerateSparkline_AllNegativeValues(t *testing.T) {
	// given: all negative values
	values := []int{-1000, -500, -100}

	// when: generating sparkline
	result := GenerateSparkline(values)

	// then: scales correctly with -1000 as min and -100 as max
	runes := []rune(result)
	require.Len(t, runes, 3)
	// Most negative should produce lowest bar
	assert.Equal(t, ' ', runes[0], "most negative value should produce lowest bar")
	// Least negative should produce tallest bar
	assert.Equal(t, '█', runes[2], "least negative value should produce tallest bar")
}
