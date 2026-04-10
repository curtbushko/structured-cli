package formatter

import (
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

func TestStatsFormatter_RenderHeader_ContainsBorder(t *testing.T) {
	// given: a StatsFormatter with width 80
	f := NewStatsFormatter(80, newTestTheme())

	// when: rendering the header
	result := f.RenderHeader()

	// then: contains rounded border characters (lipgloss rounded border uses ╭╮╰╯)
	assert.Contains(t, result, "╭")
	assert.Contains(t, result, "╯")
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

func TestStatsFormatter_RenderSummary_ContainsCommands(t *testing.T) {
	// given: a StatsFormatter and a summary
	f := NewStatsFormatter(80, newTestTheme())
	summary := domain.NewStatsSummary(150, 27500000, 85.5, 12*time.Minute+20*time.Second)

	// when: rendering the summary
	result := f.RenderSummary(summary)

	// then: contains command count
	assert.Contains(t, result, "150")
}

func TestStatsFormatter_RenderSummary_ContainsDuration(t *testing.T) {
	// given: a StatsFormatter and a summary
	f := NewStatsFormatter(80, newTestTheme())
	summary := domain.NewStatsSummary(150, 27500000, 85.5, 12*time.Minute+20*time.Second)

	// when: rendering the summary
	result := f.RenderSummary(summary)

	// then: contains formatted duration
	assert.Contains(t, result, "12m20s")
}

func TestStatsFormatter_RenderSummary_ContainsPercentage(t *testing.T) {
	// given: a StatsFormatter and a summary
	f := NewStatsFormatter(80, newTestTheme())
	summary := domain.NewStatsSummary(150, 27500000, 85.5, 12*time.Minute+20*time.Second)

	// when: rendering the summary
	result := f.RenderSummary(summary)

	// then: contains percentage
	assert.Contains(t, result, "85.5%")
}

func TestStatsFormatter_RenderSummary_ContainsEfficiencyMeter(t *testing.T) {
	// given: a StatsFormatter and a summary with high efficiency
	f := NewStatsFormatter(80, newTestTheme())
	summary := domain.NewStatsSummary(150, 27500000, 85.5, 12*time.Minute+20*time.Second)

	// when: rendering the summary
	result := f.RenderSummary(summary)

	// then: contains progress bar characters (filled block)
	assert.Contains(t, result, "█")
}

func TestStatsFormatter_RenderSummary_ContainsBorder(t *testing.T) {
	// given: a StatsFormatter and a summary
	f := NewStatsFormatter(80, newTestTheme())
	summary := domain.NewStatsSummary(150, 27500000, 85.5, 12*time.Minute+20*time.Second)

	// when: rendering the summary
	result := f.RenderSummary(summary)

	// then: contains rounded border characters
	assert.Contains(t, result, "╭")
	assert.Contains(t, result, "╯")
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

func TestStatsFormatter_RenderCommandTable_ContainsBorder(t *testing.T) {
	// given: a StatsFormatter and command stats
	f := NewStatsFormatter(80, newTestTheme())
	commands := []domain.AggregatedCommandStats{
		domain.NewAggregatedCommandStats("git status", 42, 500000, 92.3, 150*time.Millisecond),
	}
	commands = domain.CalculateImpact(commands)

	// when: rendering the command table
	result := f.RenderCommandTable(commands)

	// then: contains rounded border characters
	assert.Contains(t, result, "╭")
	assert.Contains(t, result, "╯")
}

func TestStatsFormatter_RenderCommandTable_EmptyCommands(t *testing.T) {
	// given: a StatsFormatter with no commands
	f := NewStatsFormatter(80, newTestTheme())

	// when: rendering the command table
	result := f.RenderCommandTable(nil)

	// then: returns empty string
	assert.Empty(t, result)
}

func TestEfficiencyColor_HighEfficiency(t *testing.T) {
	// given: an efficiency percentage above 80%
	// when: getting the color category
	category := EfficiencyCategory(85.0)

	// then: returns good (success) category
	assert.Equal(t, domain.SavingsCategoryGood, category)
}

func TestEfficiencyColor_MediumEfficiency(t *testing.T) {
	// given: an efficiency percentage between 50% and 80%
	// when: getting the color category
	category := EfficiencyCategory(65.0)

	// then: returns warning category
	assert.Equal(t, domain.SavingsCategoryWarning, category)
}

func TestEfficiencyColor_LowEfficiency(t *testing.T) {
	// given: an efficiency percentage below 50%
	// when: getting the color category
	category := EfficiencyCategory(40.0)

	// then: returns critical (error) category
	assert.Equal(t, domain.SavingsCategoryCritical, category)
}

func TestEfficiencyColor_Boundary80(t *testing.T) {
	// given: an efficiency at the 80% boundary
	// when: getting the color category
	category := EfficiencyCategory(80.0)

	// then: returns warning (80 is not > 80)
	assert.Equal(t, domain.SavingsCategoryWarning, category)
}

func TestEfficiencyColor_Boundary50(t *testing.T) {
	// given: an efficiency at the 50% boundary
	// when: getting the color category
	category := EfficiencyCategory(50.0)

	// then: returns critical (50 is not > 50)
	assert.Equal(t, domain.SavingsCategoryCritical, category)
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
