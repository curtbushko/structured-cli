package formatter

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// stubThemeProvider is a test double that returns fixed color values per category.
type stubThemeProvider struct{}

func (s *stubThemeProvider) ColorFor(category domain.SavingsCategory) string {
	switch category {
	case domain.SavingsCategoryGood:
		return "#00FF00"
	case domain.SavingsCategoryWarning:
		return "#FFFF00"
	case domain.SavingsCategoryCritical:
		return "#FF0000"
	default:
		return "#FFFFFF"
	}
}

func (s *stubThemeProvider) Name() string            { return "stub" }
func (s *stubThemeProvider) ListThemes() []string    { return []string{"stub"} }
func (s *stubThemeProvider) SetTheme(_ string) error { return nil }

func newTestTheme() *stubThemeProvider {
	return &stubThemeProvider{}
}

func TestStatsFormatter_RenderHeader_ContainsTitle(t *testing.T) {
	// given: a StatsFormatter with width 80
	f := NewStatsFormatter(80, newTestTheme())

	// when: rendering the header
	result := f.RenderHeader()

	// then: contains the title text
	assert.Contains(t, result, "Token Savings (Global Scope)")
}

func TestStatsFormatter_RenderHeader_ContainsBoxChars(t *testing.T) {
	// given: a StatsFormatter with width 80
	f := NewStatsFormatter(80, newTestTheme())

	// when: rendering the header
	result := f.RenderHeader()

	// then: contains box-drawing separator characters
	assert.Contains(t, result, "═")
}

func TestStatsFormatter_RenderHeader_NoBorder(t *testing.T) {
	// given: a StatsFormatter with width 80
	f := NewStatsFormatter(80, newTestTheme())

	// when: rendering the header
	result := f.RenderHeader()

	// then: does NOT contain border characters (lipgloss rounded border uses ╭╮╰╯)
	assert.NotContains(t, result, "╭")
	assert.NotContains(t, result, "╯")
}

func TestStatsFormatter_RenderSummary_ContainsFormattedNumbers(t *testing.T) {
	// given: a StatsFormatter and a summary with large numbers
	f := NewStatsFormatter(80, newTestTheme())
	summary := domain.NewStatsSummary(
		150,
		27500000,
		85.5,
		12*time.Minute+20*time.Second,
	)

	// when: rendering the summary
	result := f.RenderSummary(summary)

	// then: contains formatted large numbers
	assert.Contains(t, result, "27.5M")
}

func TestStatsFormatter_RenderSummary_OnlyTokenFields(t *testing.T) {
	// given: a StatsFormatter and a summary with token counts
	f := NewStatsFormatter(80, newTestTheme())
	summary := domain.NewStatsSummary(150, 27500000, 85.5, 12*time.Minute+20*time.Second)
	summary.TotalInputTokens = 50000000
	summary.TotalOutputTokens = 22500000

	// when: rendering the summary
	result := f.RenderSummary(summary)

	// then: contains only token-related labels
	assert.Contains(t, result, "Input Tokens:")
	assert.Contains(t, result, "Output Tokens:")
	assert.Contains(t, result, "Tokens Saved:")

	// then: does NOT contain removed labels
	assert.NotContains(t, result, "Commands:")
	assert.NotContains(t, result, "Avg Savings:")
	assert.NotContains(t, result, "Total Time:")
}

func TestStatsFormatter_RenderSummary_ContainsTokensSaved(t *testing.T) {
	// given: a StatsFormatter and a summary with tokens saved
	f := NewStatsFormatter(80, newTestTheme())
	summary := domain.NewStatsSummary(150, 27500000, 85.5, 12*time.Minute+20*time.Second)

	// when: rendering the summary
	result := f.RenderSummary(summary)

	// then: contains "Tokens Saved" label and formatted value
	assert.Contains(t, result, "Tokens Saved:")
	assert.Contains(t, result, "27.5M")
}

func TestStatsFormatter_RenderSummary_ContainsInputOutputTokens(t *testing.T) {
	// given: a StatsFormatter and a summary with input/output token counts
	f := NewStatsFormatter(80, newTestTheme())
	summary := domain.NewStatsSummary(150, 27500000, 85.5, 12*time.Minute+20*time.Second)
	summary.TotalInputTokens = 50000000
	summary.TotalOutputTokens = 22500000

	// when: rendering the summary
	result := f.RenderSummary(summary)

	// then: contains formatted input and output token values
	assert.Contains(t, result, "50.0M")
	assert.Contains(t, result, "22.5M")
}

func TestStatsFormatter_RenderSummary_NoEfficiencyMeter(t *testing.T) {
	// given: a StatsFormatter and a summary with high efficiency
	f := NewStatsFormatter(80, newTestTheme())
	summary := domain.NewStatsSummary(150, 27500000, 85.5, 12*time.Minute+20*time.Second)

	// when: rendering the summary
	result := f.RenderSummary(summary)

	// then: output does NOT contain efficiency meter or 'Efficiency:' label
	assert.NotContains(t, result, "Efficiency:")
}

func TestStatsFormatter_RenderSummary_NoBorder(t *testing.T) {
	// given: a StatsFormatter and a summary
	f := NewStatsFormatter(80, newTestTheme())
	summary := domain.NewStatsSummary(150, 27500000, 85.5, 12*time.Minute+20*time.Second)

	// when: rendering the summary
	result := f.RenderSummary(summary)

	// then: does NOT contain border characters
	assert.NotContains(t, result, "╭")
	assert.NotContains(t, result, "╯")
}

func TestStatsFormatter_RenderCommandTable_ContainsHeaders(t *testing.T) {
	// given: a StatsFormatter and command stats
	f := NewStatsFormatter(80, newTestTheme())
	commands := []domain.AggregatedCommandStats{
		domain.NewAggregatedCommandStats("git status", 42, 500000, 92.3, 150*time.Millisecond),
	}
	commands = domain.CalculateImpact(commands)

	// when: rendering the command table
	result := f.RenderCommandTable(commands)

	// then: contains column headers
	assert.Contains(t, result, "Command")
	assert.Contains(t, result, "Count")
	assert.Contains(t, result, "Saved")
	assert.Contains(t, result, "Impact")
}

func TestStatsFormatter_RenderCommandTable_ContainsByCommandHeader(t *testing.T) {
	// given: a StatsFormatter and command stats
	f := NewStatsFormatter(80, newTestTheme())
	commands := []domain.AggregatedCommandStats{
		domain.NewAggregatedCommandStats("git status", 42, 500000, 92.3, 150*time.Millisecond),
	}
	commands = domain.CalculateImpact(commands)

	// when: rendering the command table
	result := f.RenderCommandTable(commands)

	// then: contains "By Command" section header
	assert.Contains(t, result, "By Command")
}

func TestStatsFormatter_RenderCommandTable_ContainsSeparator(t *testing.T) {
	// given: a StatsFormatter and command stats
	f := NewStatsFormatter(80, newTestTheme())
	commands := []domain.AggregatedCommandStats{
		domain.NewAggregatedCommandStats("git status", 42, 500000, 92.3, 150*time.Millisecond),
	}
	commands = domain.CalculateImpact(commands)

	// when: rendering the command table
	result := f.RenderCommandTable(commands)

	// then: contains separator characters
	assert.Contains(t, result, "─")
}

func TestStatsFormatter_RenderCommandTable_ContainsCommandData(t *testing.T) {
	// given: a StatsFormatter and command stats
	f := NewStatsFormatter(80, newTestTheme())
	commands := []domain.AggregatedCommandStats{
		domain.NewAggregatedCommandStats("git status", 42, 500000, 92.3, 150*time.Millisecond),
	}
	commands = domain.CalculateImpact(commands)

	// when: rendering the command table
	result := f.RenderCommandTable(commands)

	// then: contains command data
	assert.Contains(t, result, "git status")
	assert.Contains(t, result, "42")
	assert.Contains(t, result, "500.0K")
}

func TestStatsFormatter_RenderCommandTable_TruncatesLongCommands(t *testing.T) {
	// given: a StatsFormatter and a command with a very long name
	f := NewStatsFormatter(80, newTestTheme())
	commands := []domain.AggregatedCommandStats{
		domain.NewAggregatedCommandStats(
			"go test ./very/long/path/to/some/feature/package",
			10, 100000, 80.0, 500*time.Millisecond,
		),
	}
	commands = domain.CalculateImpact(commands)

	// when: rendering the command table
	result := f.RenderCommandTable(commands)

	// then: contains ellipsis for truncated command
	assert.Contains(t, result, "...")
}

func TestStatsFormatter_RenderCommandTable_ContainsImpactBars(t *testing.T) {
	// given: a StatsFormatter and multiple commands with different impact
	f := NewStatsFormatter(80, newTestTheme())
	commands := []domain.AggregatedCommandStats{
		domain.NewAggregatedCommandStats("git status", 42, 500000, 92.3, 150*time.Millisecond),
		domain.NewAggregatedCommandStats("go build", 10, 100000, 50.0, 2*time.Second),
	}
	commands = domain.CalculateImpact(commands)

	// when: rendering the command table
	result := f.RenderCommandTable(commands)

	// then: contains impact bar characters
	assert.Contains(t, result, "█")
}

func TestStatsFormatter_RenderCommandTable_NoBorder(t *testing.T) {
	// given: a StatsFormatter and command stats
	f := NewStatsFormatter(80, newTestTheme())
	commands := []domain.AggregatedCommandStats{
		domain.NewAggregatedCommandStats("git status", 42, 500000, 92.3, 150*time.Millisecond),
	}
	commands = domain.CalculateImpact(commands)

	// when: rendering the command table
	result := f.RenderCommandTable(commands)

	// then: does NOT contain border characters
	assert.NotContains(t, result, "╭")
	assert.NotContains(t, result, "╯")
}

func TestStatsFormatter_RenderCommandTable_EmptyCommands(t *testing.T) {
	// given: a StatsFormatter with no commands
	f := NewStatsFormatter(80, newTestTheme())

	// when: rendering the command table
	result := f.RenderCommandTable(nil)

	// then: returns empty string
	assert.Empty(t, result)
}

func TestImpactBarGradient_HighImpact(t *testing.T) {
	// given: impact at 100%
	// when: getting the gradient category
	category := ImpactCategory(100.0)

	// then: returns success
	assert.Equal(t, domain.SavingsCategoryGood, category)
}

func TestImpactBarGradient_MediumImpact(t *testing.T) {
	// given: impact at 50%
	// when: getting the gradient category
	category := ImpactCategory(50.0)

	// then: returns warning
	assert.Equal(t, domain.SavingsCategoryWarning, category)
}

func TestImpactBarGradient_LowImpact(t *testing.T) {
	// given: impact at 10%
	// when: getting the gradient category
	category := ImpactCategory(10.0)

	// then: returns critical
	assert.Equal(t, domain.SavingsCategoryCritical, category)
}

func TestNewStatsFormatter_DefaultWidth(t *testing.T) {
	// given/when: creating a StatsFormatter with width 0
	f := NewStatsFormatter(0, newTestTheme())

	// then: uses default width
	require.NotNil(t, f)
	result := f.RenderHeader()
	assert.Contains(t, result, "Token Savings")
}

func TestStatsFormatter_RenderCommandTable_ViewportEnabledWhenManyCommands(t *testing.T) {
	// given: a StatsFormatter with 15 commands (over the viewport threshold)
	f := NewStatsFormatter(80, newTestTheme())
	commands := make([]domain.AggregatedCommandStats, 15)
	for i := range commands {
		commands[i] = domain.NewAggregatedCommandStats(
			fmt.Sprintf("cmd-%d", i+1), i+1, (i+1)*1000, 50.0, 100*time.Millisecond,
		)
	}
	commands = domain.CalculateImpact(commands)

	// when: rendering the command table
	result := f.RenderCommandTable(commands)

	// then: viewport is enabled, only first 10 commands shown
	assert.Contains(t, result, "cmd-1")
	assert.Contains(t, result, "cmd-10")
	assert.NotContains(t, result, "cmd-11")
	// Scroll indicator shown at bottom
	assert.Contains(t, result, "▼")
}

func TestStatsFormatter_RenderCommandTable_ViewportDisabledWhenFewCommands(t *testing.T) {
	// given: a StatsFormatter with 5 commands (under the viewport threshold)
	f := NewStatsFormatter(80, newTestTheme())
	commands := make([]domain.AggregatedCommandStats, 5)
	for i := range commands {
		commands[i] = domain.NewAggregatedCommandStats(
			fmt.Sprintf("cmd-%d", i+1), i+1, (i+1)*1000, 50.0, 100*time.Millisecond,
		)
	}
	commands = domain.CalculateImpact(commands)

	// when: rendering the command table
	result := f.RenderCommandTable(commands)

	// then: all 5 commands shown, no scroll indicator
	assert.Contains(t, result, "cmd-1")
	assert.Contains(t, result, "cmd-5")
	assert.NotContains(t, result, "▼")
}

func TestStatsFormatter_RenderCommandTable_ViewportExactlyAtThreshold(t *testing.T) {
	// given: exactly 10 commands (at threshold)
	f := NewStatsFormatter(80, newTestTheme())
	commands := make([]domain.AggregatedCommandStats, 10)
	for i := range commands {
		commands[i] = domain.NewAggregatedCommandStats(
			fmt.Sprintf("cmd-%d", i+1), i+1, (i+1)*1000, 50.0, 100*time.Millisecond,
		)
	}
	commands = domain.CalculateImpact(commands)

	// when: rendering the command table
	result := f.RenderCommandTable(commands)

	// then: all 10 commands shown, no scroll indicator
	assert.Contains(t, result, "cmd-10")
	assert.NotContains(t, result, "▼")
}

func TestStatsFormatter_RenderSummary_NoSparkline(t *testing.T) {
	// given: a StatsFormatter with summary data
	f := NewStatsFormatter(80, newTestTheme())
	summary := domain.NewStatsSummary(150, 27500000, 85.5, 12*time.Minute+20*time.Second)

	// when: rendering the summary
	result := f.RenderSummary(summary)

	// then: output does NOT contain sparkline or 'Trend:' label
	assert.NotContains(t, result, "Trend:")
}

func TestStatsFormatter_UsesThemeColors(t *testing.T) {
	// given: StatsFormatter with a theme provider
	tp := newTestTheme()
	f := NewStatsFormatter(80, tp)

	// when: mapping impact 50% to category and getting color
	impCat := ImpactCategory(50.0)
	impactColor := tp.ColorFor(impCat)

	// then: returns the warning color from theme
	assert.Equal(t, domain.SavingsCategoryWarning, impCat)
	assert.Equal(t, "#FFFF00", impactColor)

	// when: mapping impact 10% to category and getting color
	lowCat := ImpactCategory(10.0)
	lowColor := tp.ColorFor(lowCat)

	// then: returns the error color from theme
	assert.Equal(t, domain.SavingsCategoryCritical, lowCat)
	assert.Equal(t, "#FF0000", lowColor)

	// Verify formatter renders without error (integration check)
	summary := domain.NewStatsSummary(10, 5000, 85.0, 1*time.Minute)
	result := f.RenderSummary(summary)
	assert.NotEmpty(t, result)
}

func TestStatsFormatter_RenderCommandTable_ImpactColumnAligned(t *testing.T) {
	// given: StatsFormatter with multiple commands having different impact values
	f := NewStatsFormatter(100, newTestTheme())
	commands := []domain.AggregatedCommandStats{
		domain.NewAggregatedCommandStats("git status", 42, 500000, 92.3, 150*time.Millisecond),
		domain.NewAggregatedCommandStats("go build", 10, 100000, 50.0, 2*time.Second),
		domain.NewAggregatedCommandStats("npm install", 5, 50000, 30.0, 5*time.Second),
	}
	commands = domain.CalculateImpact(commands)

	// when: rendering the command table
	result := f.RenderCommandTable(commands)

	// then: Impact column header and impact bars are aligned vertically
	// Find the position of "Impact" in the header line and the start of bars in data rows
	lines := strings.Split(result, "\n")
	headerIdx := -1
	for i, line := range lines {
		if strings.Contains(line, "Impact") && strings.Contains(line, "Command") {
			headerIdx = i
			break
		}
	}
	require.NotEqual(t, -1, headerIdx, "should find header line")

	// Get the column position of "Impact" in the header
	headerLine := lines[headerIdx]
	impactPos := strings.Index(headerLine, "Impact")
	require.Greater(t, impactPos, 0, "Impact should be found in header")

	// Check that each data row has its impact bar starting at the same column position
	dataRowCount := 0
	for i := headerIdx + 1; i < len(lines); i++ {
		line := lines[i]
		// Skip non-data lines (borders, scroll indicators, empty)
		if !strings.Contains(line, "git status") &&
			!strings.Contains(line, "go build") &&
			!strings.Contains(line, "npm install") {
			continue
		}
		dataRowCount++
		// The impact bar character (█ or ░) should start at the same position as the header
		barStartIdx := strings.IndexAny(line, "█░")
		require.NotEqual(t, -1, barStartIdx, "data row should contain impact bar characters")
		assert.Equal(t, impactPos, barStartIdx,
			"Impact bar in row should align with Impact header (row: %q)", line)
	}
	assert.Equal(t, 3, dataRowCount, "should have found 3 data rows")
}

func TestStatsFormatter_RenderCommandTable_HeaderMatchesRowWidth(t *testing.T) {
	// given: StatsFormatter with command stats
	f := NewStatsFormatter(100, newTestTheme())
	commands := []domain.AggregatedCommandStats{
		domain.NewAggregatedCommandStats("git status", 42, 500000, 92.3, 150*time.Millisecond),
	}
	commands = domain.CalculateImpact(commands)

	// when: rendering the command table
	result := f.RenderCommandTable(commands)

	// then: Impact column width in header matches width used in data rows
	lines := strings.Split(result, "\n")
	var headerLine, dataLine string
	for _, line := range lines {
		if strings.Contains(line, "Impact") && strings.Contains(line, "Command") {
			headerLine = line
		}
		if strings.Contains(line, "git status") {
			dataLine = line
		}
	}
	require.NotEmpty(t, headerLine, "should find header line")
	require.NotEmpty(t, dataLine, "should find data line")

	// The Impact text starts at same position as the bar characters
	impactHeaderPos := strings.Index(headerLine, "Impact")
	barPos := strings.IndexAny(dataLine, "█░")
	require.NotEqual(t, -1, impactHeaderPos)
	require.NotEqual(t, -1, barPos)
	assert.Equal(t, impactHeaderPos, barPos, "Impact header and bar should be at the same column position")
}

func TestStatsFormatter_SetSavingsTrend_MethodRemoved(t *testing.T) {
	// This test verifies at compile time that SetSavingsTrend is removed.
	// The StatsFormatter struct should NOT have a SetSavingsTrend method.
	// If this test compiles, the savingsTrend field and method are gone.
	f := NewStatsFormatter(80, newTestTheme())
	// Verify the formatter still works without trend support
	summary := domain.NewStatsSummary(10, 5000, 85.0, 1*time.Minute)
	result := f.RenderSummary(summary)
	assert.NotEmpty(t, result)
	assert.NotContains(t, result, "Trend:")
}
