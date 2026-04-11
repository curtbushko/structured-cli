package formatter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

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

	// fillChar is the Unicode full block character for bars.
	fillChar = "\u2588" // █

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
	width int
	theme ports.ThemeProvider
}

// NewStatsFormatter creates a new StatsFormatter with the given terminal width and theme provider.
// If width is 0 or negative, a default width of 80 is used.
func NewStatsFormatter(width int, theme ports.ThemeProvider) *StatsFormatter {
	if width <= 0 {
		width = defaultWidth
	}
	return &StatsFormatter{width: width, theme: theme}
}

// RenderHeader renders the header section with title and box-drawing separator.
func (f *StatsFormatter) RenderHeader() string {
	titleStyle := lipgloss.NewStyle().
		Foreground(titleColor).
		Bold(true)

	title := titleStyle.Render("Token Savings (Global Scope)")
	separator := strings.Repeat("═", f.width)

	return title + "\n" + separator
}

// RenderSummary renders the summary section with formatted stats.
func (f *StatsFormatter) RenderSummary(summary domain.StatsSummary) string {
	labelStyle := lipgloss.NewStyle().
		Foreground(labelColor)

	valStyle := lipgloss.NewStyle().
		Foreground(valueColor).
		Bold(true)

	// Build key-value pairs (token counts only)
	lines := make([]string, 0, 3)
	lines = append(lines,
		fmt.Sprintf("%s  %s",
			labelStyle.Render("Input Tokens:"),
			valStyle.Render(FormatLargeNumber(summary.TotalInputTokens))),
		fmt.Sprintf("%s  %s",
			labelStyle.Render("Output Tokens:"),
			valStyle.Render(FormatLargeNumber(summary.TotalOutputTokens))),
		fmt.Sprintf("%s  %s",
			labelStyle.Render("Tokens Saved:"),
			valStyle.Render(FormatLargeNumber(summary.TotalTokensSaved))),
	)

	return strings.Join(lines, "\n")
}

// RenderCommandTable renders the command table with all columns and impact bars.
func (f *StatsFormatter) RenderCommandTable(commands []domain.AggregatedCommandStats) string {
	if len(commands) == 0 {
		return ""
	}

	headerStyle := lipgloss.NewStyle().
		Foreground(titleColor).
		Bold(true)

	sectionHeader := headerStyle.Render("By Command")
	separator := strings.Repeat("─", f.width)

	// Apply viewport: limit visible rows when command count exceeds threshold
	visibleCommands := commands
	hasMore := false
	remaining := 0
	if len(commands) > viewportThreshold {
		visibleCommands = commands[:viewportThreshold]
		hasMore = true
		remaining = len(commands) - viewportThreshold
	}

	// Build table using lipgloss/table for proper column alignment
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(labelColor)).
		BorderTop(false).
		BorderBottom(false).
		BorderLeft(false).
		BorderRight(false).
		BorderColumn(true).
		BorderHeader(true).
		Headers("#", "Command", "Count", "Saved", "Avg%", "Time", "Impact")

	// Add rows
	for i, cmd := range visibleCommands {
		truncated := TruncateCommand(cmd.CommandName, commandColWidth)
		bar := renderImpactBar(cmd.ImpactPercent, f.theme)
		t.Row(
			strconv.Itoa(i+1),
			truncated,
			strconv.Itoa(cmd.Count),
			FormatLargeNumber(cmd.TotalTokensSaved),
			fmt.Sprintf("%.1f%%", cmd.AvgSavingsPercent),
			FormatDuration(cmd.AvgExecutionTime),
			bar,
		)
	}

	parts := []string{
		sectionHeader,
		separator,
		t.Render(),
	}

	// Add scroll indicator when viewport is active
	if hasMore {
		parts = append(parts, fmt.Sprintf(scrollIndicator, remaining))
	}

	return strings.Join(parts, "\n")
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
