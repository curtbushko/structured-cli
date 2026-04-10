package stats

import (
	"fmt"
	"io"
	"math"

	"github.com/curtbushko/structured-cli/internal/ports"
)

// sparklineChars maps values to Unicode block characters for sparkline rendering.
// Characters are ordered from lowest to highest: ▁▂▃▄▅▆▇█
var sparklineChars = []rune{
	'\u2581', // ▁
	'\u2582', // ▂
	'\u2583', // ▃
	'\u2584', // ▄
	'\u2585', // ▅
	'\u2586', // ▆
	'\u2587', // ▇
	'\u2588', // █
}

// SparklineRenderer renders historical token savings trends as ASCII sparklines.
// It maps float64 values to Unicode block characters for a compact visual display.
type SparklineRenderer struct {
	theme ports.ThemeProvider
}

// NewSparklineRenderer creates a new SparklineRenderer with the given theme.
func NewSparklineRenderer(theme ports.ThemeProvider) *SparklineRenderer {
	return &SparklineRenderer{theme: theme}
}

// Render writes a sparkline representing the given values.
// Each value maps to a block character proportional to its magnitude within the range.
func (r *SparklineRenderer) Render(w io.Writer, values []float64) error {
	if len(values) == 0 {
		return nil
	}

	// Find min and max
	minVal := values[0]
	maxVal := values[0]
	for _, v := range values[1:] {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	rangeVal := maxVal - minVal
	maxIdx := len(sparklineChars) - 1

	result := make([]rune, len(values))
	for i, v := range values {
		var idx int
		if rangeVal == 0 {
			// All values equal - use middle character
			idx = maxIdx / 2
		} else {
			normalized := (v - minVal) / rangeVal
			idx = int(math.Round(normalized * float64(maxIdx)))
			if idx > maxIdx {
				idx = maxIdx
			}
		}
		result[i] = sparklineChars[idx]
	}

	_, err := fmt.Fprintf(w, "%s\n", string(result))
	return err
}
