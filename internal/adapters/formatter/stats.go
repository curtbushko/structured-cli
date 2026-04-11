package formatter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/lipgloss"

	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

const (
	// defaultWidth is the fallback terminal width when none is specified.
	defaultWidth = 80

	// commandColWidth is the maximum width for command names in the table.
	commandColWidth = 25

	// impactBarWidth is the width of impact visualization bars.
	impactBarWidth = 10

	// efficiencyBarWidth is the width of the efficiency meter bar.
	efficiencyBarWidth = 20

	// fillChar is the Unicode full block character for bars.
	fillChar = "\u2588" // █

	// efficiencyGoodThreshold is the percentage above which efficiency is "good".
	efficiencyGoodThreshold = 80.0
	// efficiencyWarningThreshold is the percentage above which efficiency is "warning".
	efficiencyWarningThreshold = 50.0

	// impactGoodThreshold is the percentage above which impact is "good".
	impactGoodThreshold = 70.0
	// impactWarningThreshold is the percentage above which impact is "warning".
	impactWarningThreshold = 30.0

	// viewportThreshold is the max commands shown before enabling viewport scrolling.
	viewportThreshold = 10

	// scrollIndicator is shown when the command table has more rows than viewportThreshold.
	scrollIndicator = "  ▼ ... and %d more commands"
)

// Adaptive colors for light/dark terminal support.
var (
	titleColor = lipgloss.AdaptiveColor{Light: "#333333", Dark: "#FFFFFF"}
	labelColor = lipgloss.AdaptiveColor{Light: "#555555", Dark: "#AAAAAA"}
	valueColor = lipgloss.AdaptiveColor{Light: "#222222", Dark: "#EEEEEE"}
)

// StatsFormatter renders statistics with styled header, summary, and command table sections.
// It uses lipgloss for styling with rounded borders and adaptive colors,
// and a ThemeProvider for category-based color tokens (success/warning/error).
type StatsFormatter struct {
	width        int
	theme        ports.ThemeProvider
	savingsTrend []int
}

// NewStatsFormatter creates a new StatsFormatter with the given terminal width and theme provider.
// If width is 0 or negative, a default width of 80 is used.
func NewStatsFormatter(width int, theme ports.ThemeProvider) *StatsFormatter {
	if width <= 0 {
		width = defaultWidth
	}
	return &StatsFormatter{width: width, theme: theme}
}

// SetSavingsTrend sets the token savings trend data used to render a sparkline in the summary.
func (f *StatsFormatter) SetSavingsTrend(values []int) {
	f.savingsTrend = values
}

// RenderHeader renders the header section with title and box-drawing separator.
func (f *StatsFormatter) RenderHeader() string {
	innerWidth := f.width - 4 // Account for border padding

	titleStyle := lipgloss.NewStyle().
		Foreground(titleColor).
		Bold(true)

	title := titleStyle.Render("Token Savings (Global Scope)")
	separator := strings.Repeat("═", innerWidth)

	content := title + "\n" + separator

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(labelColor).
		Width(innerWidth)

	return borderStyle.Render(content)
}

// RenderSummary renders the summary section with formatted stats and efficiency meter.
func (f *StatsFormatter) RenderSummary(summary domain.StatsSummary) string {
	innerWidth := f.width - 4

	labelStyle := lipgloss.NewStyle().
		Foreground(labelColor)

	valStyle := lipgloss.NewStyle().
		Foreground(valueColor).
		Bold(true)

	// Build key-value pairs
	lines := make([]string, 0, 5)
	lines = append(lines,
		fmt.Sprintf("  %s  %s",
			labelStyle.Render("Commands:"),
			valStyle.Render(strconv.Itoa(summary.TotalCommands))),
		fmt.Sprintf("  %s  %s",
			labelStyle.Render("Tokens Saved:"),
			valStyle.Render(FormatLargeNumber(summary.TotalTokensSaved))),
		fmt.Sprintf("  %s  %s",
			labelStyle.Render("Avg Savings:"),
			valStyle.Render(fmt.Sprintf("%.1f%%", summary.AvgSavingsPercent))),
		fmt.Sprintf("  %s  %s",
			labelStyle.Render("Total Time:"),
			valStyle.Render(FormatDuration(summary.TotalExecutionTime))),
	)

	// Add efficiency meter
	meter := renderEfficiencyMeter(summary.AvgSavingsPercent, f.theme)
	lines = append(lines, fmt.Sprintf("  %s  %s",
		labelStyle.Render("Efficiency:"),
		meter))

	// Add sparkline for token savings trend if data is available
	if sparkline := RenderSparklineWithColor(f.savingsTrend, f.theme); sparkline != "" {
		lines = append(lines, fmt.Sprintf("  %s  %s",
			labelStyle.Render("Trend:"),
			sparkline))
	}

	content := strings.Join(lines, "\n")

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(labelColor).
		Width(innerWidth)

	return borderStyle.Render(content)
}

// RenderCommandTable renders the command table with all columns and impact bars.
func (f *StatsFormatter) RenderCommandTable(commands []domain.AggregatedCommandStats) string {
	if len(commands) == 0 {
		return ""
	}

	innerWidth := f.width - 4

	headerStyle := lipgloss.NewStyle().
		Foreground(titleColor).
		Bold(true)

	sectionHeader := headerStyle.Render("By Command")
	separator := strings.Repeat("─", innerWidth)

	// Table header
	tableHeader := fmt.Sprintf("  %-3s %-*s %6s %8s %6s %7s  %-*s",
		"#", commandColWidth, "Command", "Count", "Saved", "Avg%", "Time", impactBarWidth, "Impact")

	// Apply viewport: limit visible rows when command count exceeds threshold
	visibleCommands := commands
	hasMore := false
	remaining := 0
	if len(commands) > viewportThreshold {
		visibleCommands = commands[:viewportThreshold]
		hasMore = true
		remaining = len(commands) - viewportThreshold
	}

	// Table rows
	rows := make([]string, 0, len(visibleCommands))
	for i, cmd := range visibleCommands {
		truncated := TruncateCommand(cmd.CommandName, commandColWidth)
		bar := renderImpactBar(cmd.ImpactPercent, f.theme)
		row := fmt.Sprintf("  %-3d %-*s %6d %8s %5.1f%% %7s  %s",
			i+1,
			commandColWidth, truncated,
			cmd.Count,
			FormatLargeNumber(cmd.TotalTokensSaved),
			cmd.AvgSavingsPercent,
			FormatDuration(cmd.AvgExecutionTime),
			bar)
		rows = append(rows, row)
	}

	parts := []string{
		sectionHeader,
		separator,
		tableHeader,
		strings.Join(rows, "\n"),
	}

	// Add scroll indicator when viewport is active
	if hasMore {
		parts = append(parts, fmt.Sprintf(scrollIndicator, remaining))
	}

	content := strings.Join(parts, "\n")

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(labelColor).
		Width(innerWidth)

	return borderStyle.Render(content)
}

// EfficiencyCategory returns the SavingsCategory for the given efficiency percentage.
// Thresholds: >80% = Good, >50% = Warning, <=50% = Critical.
func EfficiencyCategory(percent float64) domain.SavingsCategory {
	switch {
	case percent > efficiencyGoodThreshold:
		return domain.SavingsCategoryGood
	case percent > efficiencyWarningThreshold:
		return domain.SavingsCategoryWarning
	default:
		return domain.SavingsCategoryCritical
	}
}

// ImpactCategory returns the SavingsCategory for an impact percentage.
// Uses gradient: >70% = Good, >30% = Warning, <=30% = Critical.
func ImpactCategory(percent float64) domain.SavingsCategory {
	switch {
	case percent > impactGoodThreshold:
		return domain.SavingsCategoryGood
	case percent > impactWarningThreshold:
		return domain.SavingsCategoryWarning
	default:
		return domain.SavingsCategoryCritical
	}
}

// renderEfficiencyMeter creates a progress bar for efficiency percentage
// using the bubbles/progress component with flair theme colors.
func renderEfficiencyMeter(percent float64, theme ports.ThemeProvider) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	category := EfficiencyCategory(percent)
	color := theme.ColorFor(category)

	bar := progress.New(
		progress.WithSolidFill(color),
		progress.WithWidth(efficiencyBarWidth),
		progress.WithoutPercentage(),
	)

	return bar.ViewAs(percent/100) + fmt.Sprintf(" %.1f%%", percent)
}

// renderImpactBar creates a scaled impact visualization bar with flair theme gradient colors.
func renderImpactBar(impactPercent float64, theme ports.ThemeProvider) string {
	if impactPercent < 0 {
		impactPercent = 0
	}
	if impactPercent > 100 {
		impactPercent = 100
	}

	filled := int(impactPercent / 100 * impactBarWidth)
	empty := impactBarWidth - filled

	category := ImpactCategory(impactPercent)
	color := theme.ColorFor(category)

	barStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color))

	return barStyle.Render(strings.Repeat(fillChar, filled)) + strings.Repeat("░", empty)
}
