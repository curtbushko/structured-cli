package formatter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatLargeNumber_Billions(t *testing.T) {
	// given: a value in the billions
	// when: formatting 2,500,000,000
	result := FormatLargeNumber(2500000000)

	// then: returns "2.5B"
	assert.Equal(t, "2.5B", result)
}

func TestFormatLargeNumber_Millions(t *testing.T) {
	// given: various values in the millions
	tests := []struct {
		input    int
		expected string
	}{
		{27500000, "27.5M"},
		{1100000, "1.1M"},
		{1000000, "1.0M"},
	}

	for _, tc := range tests {
		// when: formatting the value
		result := FormatLargeNumber(tc.input)

		// then: returns formatted string with M suffix
		assert.Equal(t, tc.expected, result)
	}
}

func TestFormatLargeNumber_Thousands(t *testing.T) {
	// given: a value in the thousands
	// when: formatting 45,000
	result := FormatLargeNumber(45000)

	// then: returns "45.0K"
	assert.Equal(t, "45.0K", result)
}

func TestFormatLargeNumber_Small(t *testing.T) {
	// given: a value below 1000
	// when: formatting 850
	result := FormatLargeNumber(850)

	// then: returns the raw number as string
	assert.Equal(t, "850", result)
}

func TestFormatLargeNumber_Zero(t *testing.T) {
	// given: zero value
	// when: formatting 0
	result := FormatLargeNumber(0)

	// then: returns "0"
	assert.Equal(t, "0", result)
}

func TestFormatDuration_MinutesAndSeconds(t *testing.T) {
	// given: a duration of 12m20s
	d := 12*time.Minute + 20*time.Second

	// when: formatting the duration
	result := FormatDuration(d)

	// then: returns human-readable string
	assert.Equal(t, "12m20s", result)
}

func TestFormatDuration_Milliseconds(t *testing.T) {
	// given: a duration of 696ms
	d := 696 * time.Millisecond

	// when: formatting the duration
	result := FormatDuration(d)

	// then: returns millisecond format
	assert.Equal(t, "696ms", result)
}

func TestFormatDuration_HoursAndMinutes(t *testing.T) {
	// given: a duration of 1h5m
	d := 1*time.Hour + 5*time.Minute

	// when: formatting the duration
	result := FormatDuration(d)

	// then: returns hours and minutes
	assert.Equal(t, "1h5m", result)
}

func TestFormatDuration_Zero(t *testing.T) {
	// given: a zero duration
	d := time.Duration(0)

	// when: formatting the duration
	result := FormatDuration(d)

	// then: returns "0ms"
	assert.Equal(t, "0ms", result)
}

func TestFormatDuration_SubMillisecond(t *testing.T) {
	// given: a sub-millisecond duration
	d := 500 * time.Microsecond

	// when: formatting the duration
	result := FormatDuration(d)

	// then: returns "<1ms"
	assert.Equal(t, "<1ms", result)
}

func TestTruncateCommand_ShortCommand(t *testing.T) {
	// given: a command shorter than maxLen
	cmd := "git status"

	// when: truncating to 30 chars
	result := TruncateCommand(cmd, 30)

	// then: returns the command unchanged
	assert.Equal(t, "git status", result)
}

func TestTruncateCommand_LongCommand(t *testing.T) {
	// given: a long command string
	cmd := "go test ./very/long/path/to/feature"

	// when: truncating to 20 chars
	result := TruncateCommand(cmd, 20)

	// then: returns truncated string with ellipsis
	assert.Equal(t, "go test ./very/lo...", result)
	assert.LessOrEqual(t, len(result), 20)
}

func TestTruncateCommand_ExactLength(t *testing.T) {
	// given: a command exactly at maxLen
	cmd := "git status"

	// when: truncating to exactly its length
	result := TruncateCommand(cmd, len(cmd))

	// then: returns unchanged
	assert.Equal(t, "git status", result)
}

func TestTruncateCommand_VeryShortMax(t *testing.T) {
	// given: a command with very short maxLen
	cmd := "go test ./path"

	// when: truncating to 5 chars
	result := TruncateCommand(cmd, 5)

	// then: returns truncated with ellipsis
	assert.Equal(t, "go...", result)
	assert.LessOrEqual(t, len(result), 5)
}
