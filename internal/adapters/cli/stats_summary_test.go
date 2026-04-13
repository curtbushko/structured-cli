package cli

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

// mockStatsFormatter implements ports.StatsFormatter for testing.
type mockStatsFormatter struct{}

var _ ports.StatsFormatter = (*mockStatsFormatter)(nil)

func (m *mockStatsFormatter) RenderHeader() string {
	return "== Token Savings (Global Scope) =="
}

func (m *mockStatsFormatter) RenderSummary(summary domain.StatsSummary) string {
	return fmt.Sprintf("Commands: %d | Tokens Saved: %d | Avg: %.1f%% | Time: %s",
		summary.TotalCommands,
		summary.TotalTokensSaved,
		summary.AvgSavingsPercent,
		summary.TotalExecutionTime)
}

func (m *mockStatsFormatter) RenderCommandTable(commands []domain.AggregatedCommandStats) string {
	if len(commands) == 0 {
		return ""
	}
	var b strings.Builder
	for i, cmd := range commands {
		fmt.Fprintf(&b, "%d. %s count=%d saved=%d\n", i+1, cmd.CommandName, cmd.Count, cmd.TotalTokensSaved)
	}
	return b.String()
}

func TestExecuteSummaryStats_UsesNewFormatter(t *testing.T) {
	// given: stats command with summary data and a stats formatter
	reader := &mockTrackingReader{
		summary: domain.NewStatsSummary(
			150,
			27500000,
			85.5,
			12*time.Minute+20*time.Second,
		),
	}
	sf := &mockStatsFormatter{}

	// when: execute stats without flags (default summary)
	var buf bytes.Buffer
	err := executeStatsCommand(context.Background(), reader, statsFlags{}, &buf, sf)

	// then: output uses new formatter with header and summary
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Token Savings")
	assert.Contains(t, output, "150")
	assert.Contains(t, output, "85.5%")
}

func TestExecuteHistoryStats_UnchangedBehavior(t *testing.T) {
	// given: stats command with --history flag
	now := time.Now()
	reader := &mockTrackingReader{
		history: []domain.CommandRecord{
			{
				ID:             1,
				Command:        "git",
				Subcommands:    []string{"status"},
				TokensSaved:    500,
				SavingsPercent: 80.0,
				Timestamp:      now,
			},
		},
	}
	sf := &mockStatsFormatter{}

	// when: execute stats --history
	var buf bytes.Buffer
	err := executeStatsCommand(context.Background(), reader, statsFlags{history: 10}, &buf, sf)

	// then: output uses existing history formatter (unchanged)
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "git")
	assert.Contains(t, output, "status")
	assert.Contains(t, output, "Recent Command History")
}

func TestStatsFormatter_Integration_EmptyCommands(t *testing.T) {
	// given: stats data with no commands (zero summary)
	reader := &mockTrackingReader{
		summary: domain.NewStatsSummary(0, 0, 0, 0),
	}
	sf := &mockStatsFormatter{}

	// when: render stats output
	var buf bytes.Buffer
	err := executeStatsCommand(context.Background(), reader, statsFlags{}, &buf, sf)

	// then: shows header and summary, no error
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Token Savings")
	assert.Contains(t, output, "0")
}

func TestCommandAggregation_GroupsByNormalizedName(t *testing.T) {
	// given: history with multiple executions of same command with different paths
	records := []domain.CommandRecord{
		{Command: "git", Subcommands: []string{"status"}, TokensSaved: 100, SavingsPercent: 80.0, ExecutionTime: 100 * time.Millisecond},
		{Command: "git", Subcommands: []string{"status"}, TokensSaved: 200, SavingsPercent: 90.0, ExecutionTime: 150 * time.Millisecond},
		{Command: "npm", Subcommands: []string{"test"}, TokensSaved: 300, SavingsPercent: 70.0, ExecutionTime: 200 * time.Millisecond},
	}

	// when: aggregate command stats
	result := aggregateCommands(records)

	// then: commands grouped by normalized name, counts and savings summed
	require.Len(t, result, 2)

	// Find git status aggregate (sorted by tokens saved desc: npm=300, git=300, but npm appears second due to insertion order tie)
	var gitStats, npmStats domain.AggregatedCommandStats
	for _, s := range result {
		switch s.CommandName {
		case "git status":
			gitStats = s
		case "npm test":
			npmStats = s
		}
	}

	assert.Equal(t, 2, gitStats.Count)
	assert.Equal(t, 300, gitStats.TotalTokensSaved) // 100 + 200
	assert.InDelta(t, 85.0, gitStats.AvgSavingsPercent, 0.1)

	assert.Equal(t, 1, npmStats.Count)
	assert.Equal(t, 300, npmStats.TotalTokensSaved)
}

func TestCommandAggregation_EmptyRecords(t *testing.T) {
	// given: empty records
	// when: aggregate
	result := aggregateCommands(nil)

	// then: returns nil
	assert.Nil(t, result)
}

func TestCommandAggregation_SortsByTokensSavedDescending(t *testing.T) {
	// given: records with varying savings
	records := []domain.CommandRecord{
		{Command: "low", Subcommands: nil, TokensSaved: 100, SavingsPercent: 50.0, ExecutionTime: 100 * time.Millisecond},
		{Command: "high", Subcommands: nil, TokensSaved: 1000, SavingsPercent: 90.0, ExecutionTime: 100 * time.Millisecond},
		{Command: "mid", Subcommands: nil, TokensSaved: 500, SavingsPercent: 70.0, ExecutionTime: 100 * time.Millisecond},
	}

	// when: aggregate
	result := aggregateCommands(records)

	// then: sorted by total tokens saved descending
	require.Len(t, result, 3)
	assert.Equal(t, "high", result[0].CommandName)
	assert.Equal(t, "mid", result[1].CommandName)
	assert.Equal(t, "low", result[2].CommandName)
}

func TestStatsCommand_JSON_Unchanged(t *testing.T) {
	// given: stats command with --json flag
	reader := &mockTrackingReader{
		summary: domain.NewStatsSummary(100, 50000, 75.5, 10*time.Minute),
	}
	sf := &mockStatsFormatter{}

	// when: execute stats --json
	var buf bytes.Buffer
	err := executeStatsCommand(context.Background(), reader, statsFlags{json: true}, &buf, sf)

	// then: JSON output unchanged, new formatter only affects text output
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, `"total_commands":100`)
	assert.Contains(t, output, `"total_tokens_saved":50000`)
	assert.Contains(t, output, `"avg_savings_percent":75.5`)
}

func TestStatsFormatter_Integration_ViewportEnabled(t *testing.T) {
	// given: history that would generate 15 different commands
	records := make([]domain.CommandRecord, 15)
	for i := range records {
		records[i] = domain.CommandRecord{
			Command:        fmt.Sprintf("cmd%d", i),
			Subcommands:    []string{"run"},
			TokensSaved:    (i + 1) * 1000,
			SavingsPercent: 50.0,
			ExecutionTime:  100 * time.Millisecond,
		}
	}

	// when: aggregate and check count
	result := aggregateCommands(records)

	// then: should have 15 distinct commands
	assert.Len(t, result, 15)
}

func TestExecuteSummaryStats_WithHistory_IncludesCommandTable(t *testing.T) {
	// given: stats with history records available
	reader := &mockTrackingReader{
		summary: domain.NewStatsSummary(5, 5000, 70.0, 5*time.Minute),
		history: []domain.CommandRecord{
			{Command: "git", Subcommands: []string{"status"}, TokensSaved: 3000, SavingsPercent: 80.0, ExecutionTime: 100 * time.Millisecond},
			{Command: "npm", Subcommands: []string{"test"}, TokensSaved: 2000, SavingsPercent: 60.0, ExecutionTime: 200 * time.Millisecond},
		},
	}
	sf := &mockStatsFormatter{}

	// when: execute summary stats
	var buf bytes.Buffer
	err := executeStatsCommand(context.Background(), reader, statsFlags{}, &buf, sf)

	// then: output includes command table from aggregated data
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "git status")
	assert.Contains(t, output, "npm test")
}

func TestExecuteSummaryStats_RendersWithoutTrend(t *testing.T) {
	// given: stats with history records (trend no longer used)
	reader := &mockTrackingReader{
		summary: domain.NewStatsSummary(3, 3000, 60.0, 3*time.Minute),
		history: []domain.CommandRecord{
			{Command: "git", Subcommands: []string{"status"}, TokensSaved: 1000, SavingsPercent: 60.0, ExecutionTime: 100 * time.Millisecond},
			{Command: "git", Subcommands: []string{"log"}, TokensSaved: 1500, SavingsPercent: 70.0, ExecutionTime: 150 * time.Millisecond},
			{Command: "npm", Subcommands: []string{"test"}, TokensSaved: 500, SavingsPercent: 50.0, ExecutionTime: 200 * time.Millisecond},
		},
	}
	sf := &mockStatsFormatter{}

	// when: execute summary stats
	var buf bytes.Buffer
	err := executeStatsCommand(context.Background(), reader, statsFlags{}, &buf, sf)

	// then: renders successfully without trend data
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "Token Savings")
}

func TestExecuteSummaryStats_NilFormatterFallsBack(t *testing.T) {
	// given: stats command with nil formatter
	reader := &mockTrackingReader{
		summary: domain.NewStatsSummary(100, 50000, 75.5, 10*time.Minute),
	}

	// when: execute stats without formatter
	var buf bytes.Buffer
	err := executeStatsCommand(context.Background(), reader, statsFlags{}, &buf, nil)

	// then: still produces output (falls back to plain text)
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "100")
	assert.Contains(t, output, "50000")
	assert.Contains(t, output, "75.5")
}
