package cli

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

// mockTrackingReader implements ports.TrackingReader for testing.
type mockTrackingReader struct {
	summary       domain.StatsSummary
	history       []domain.CommandRecord
	byParser      []domain.CommandStats
	byFilter      []domain.FilterStats
	statsErr      error
	historyErr    error
	byParserErr   error
	byFilterErr   error
	lastStatsOpts ports.StatsOptions
	lastLimit     int
}

func (m *mockTrackingReader) Stats(_ context.Context, opts ports.StatsOptions) (domain.StatsSummary, error) {
	m.lastStatsOpts = opts
	return m.summary, m.statsErr
}

func (m *mockTrackingReader) History(_ context.Context, limit int) ([]domain.CommandRecord, error) {
	m.lastLimit = limit
	return m.history, m.historyErr
}

func (m *mockTrackingReader) StatsByParser(_ context.Context) ([]domain.CommandStats, error) {
	return m.byParser, m.byParserErr
}

func (m *mockTrackingReader) StatsByFilter(_ context.Context) ([]domain.FilterStats, error) {
	return m.byFilter, m.byFilterErr
}

func TestStatsCommand_Summary(t *testing.T) {
	// Arrange
	reader := &mockTrackingReader{
		summary: domain.NewStatsSummary(
			100,            // total commands
			50000,          // total tokens saved
			75.5,           // avg savings percent
			10*time.Minute, // total execution time
		),
	}

	// Act - nil theme falls back to plain text
	var buf bytes.Buffer
	err := executeStatsCommand(context.Background(), reader, statsFlags{}, &buf, nil)

	// Assert
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "100")   // total commands
	assert.Contains(t, output, "50000") // tokens saved (or formatted version)
	assert.Contains(t, output, "75.5")  // avg savings
}

func TestStatsCommand_JSON(t *testing.T) {
	// Arrange
	reader := &mockTrackingReader{
		summary: domain.NewStatsSummary(
			100,
			50000,
			75.5,
			10*time.Minute,
		),
	}

	// Act
	var buf bytes.Buffer
	err := executeStatsCommand(context.Background(), reader, statsFlags{json: true}, &buf, nil)

	// Assert
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, `"total_commands":100`)
	assert.Contains(t, output, `"total_tokens_saved":50000`)
	assert.Contains(t, output, `"avg_savings_percent":75.5`)
}

func TestStatsCommand_History(t *testing.T) {
	// Arrange
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
			{
				ID:             2,
				Command:        "npm",
				Subcommands:    []string{"list"},
				TokensSaved:    300,
				SavingsPercent: 70.0,
				Timestamp:      now.Add(-time.Hour),
			},
		},
	}

	// Act
	var buf bytes.Buffer
	// history=10 means show 10 entries (default when --history is used without explicit value)
	err := executeStatsCommand(context.Background(), reader, statsFlags{history: 10}, &buf, nil)

	// Assert
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "git")
	assert.Contains(t, output, "status")
	assert.Contains(t, output, "npm")
	assert.Contains(t, output, "list")
	assert.Equal(t, 10, reader.lastLimit, "default history limit should be 10")
}

func TestStatsCommand_HistoryLimit(t *testing.T) {
	// Arrange
	reader := &mockTrackingReader{
		history: make([]domain.CommandRecord, 20),
	}

	// Act
	var buf bytes.Buffer
	// history=20 means show 20 entries
	err := executeStatsCommand(context.Background(), reader, statsFlags{history: 20}, &buf, nil)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 20, reader.lastLimit, "custom history limit should be 20")
}

func TestStatsCommand_ByParser(t *testing.T) {
	// Arrange
	reader := &mockTrackingReader{
		byParser: []domain.CommandStats{
			domain.NewCommandStats("git-status", 50, 25000, 100*time.Millisecond),
			domain.NewCommandStats("npm-list", 30, 15000, 200*time.Millisecond),
			domain.NewCommandStats("docker-ps", 20, 10000, 150*time.Millisecond),
		},
	}

	// Act
	var buf bytes.Buffer
	err := executeStatsCommand(context.Background(), reader, statsFlags{byParser: true}, &buf, nil)

	// Assert
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "git-status")
	assert.Contains(t, output, "npm-list")
	assert.Contains(t, output, "docker-ps")
	assert.Contains(t, output, "50") // invocation count for git-status
}

func TestStatsCommand_Project(t *testing.T) {
	// Arrange
	reader := &mockTrackingReader{
		summary: domain.NewStatsSummary(25, 12500, 80.0, 2*time.Minute),
	}

	// Act
	var buf bytes.Buffer
	// project=true means use current working directory
	err := executeStatsCommand(context.Background(), reader, statsFlags{project: true}, &buf, nil)

	// Assert
	require.NoError(t, err)
	// The project should be set to the actual current working directory
	assert.NotEmpty(t, reader.lastStatsOpts.Project, "project should be set to current working directory")
}

func TestStatsCommand_ProjectJSON(t *testing.T) {
	// Arrange
	reader := &mockTrackingReader{
		summary: domain.NewStatsSummary(25, 12500, 80.0, 2*time.Minute),
	}

	// Act
	var buf bytes.Buffer
	// project=true means use current working directory, json=true for JSON output
	err := executeStatsCommand(context.Background(), reader, statsFlags{project: true, json: true}, &buf, nil)

	// Assert
	require.NoError(t, err)
	// The project should be set to the actual current working directory
	assert.NotEmpty(t, reader.lastStatsOpts.Project, "project should be set to current working directory")
	output := buf.String()
	assert.Contains(t, output, `"total_commands":25`)
}

func TestStatsCommand_EmptyDatabase(t *testing.T) {
	// Arrange
	reader := &mockTrackingReader{
		summary: domain.NewStatsSummary(0, 0, 0, 0),
	}

	// Act
	var buf bytes.Buffer
	err := executeStatsCommand(context.Background(), reader, statsFlags{}, &buf, nil)

	// Assert
	require.NoError(t, err)
	output := buf.String()
	// Should show friendly message or zeros
	assert.Contains(t, output, "0")
}

func TestStatsCommand_HistoryJSON(t *testing.T) {
	// Arrange
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

	// Act
	var buf bytes.Buffer
	// history=10 with json=true
	err := executeStatsCommand(context.Background(), reader, statsFlags{history: 10, json: true}, &buf, nil)

	// Assert
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, `"command":"git"`)
	assert.Contains(t, output, `"subcommands":["status"]`)
	assert.Contains(t, output, `"tokens_saved":500`)
}

func TestStatsCommand_ByParserJSON(t *testing.T) {
	// Arrange
	reader := &mockTrackingReader{
		byParser: []domain.CommandStats{
			domain.NewCommandStats("git-status", 50, 25000, 100*time.Millisecond),
		},
	}

	// Act
	var buf bytes.Buffer
	err := executeStatsCommand(context.Background(), reader, statsFlags{byParser: true, json: true}, &buf, nil)

	// Assert
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, `"parser_name":"git-status"`)
	assert.Contains(t, output, `"invocation_count":50`)
}

func TestStatsCommand_ProjectUsesCurrentDir(t *testing.T) {
	// Arrange
	reader := &mockTrackingReader{
		summary: domain.NewStatsSummary(25, 12500, 80.0, 2*time.Minute),
	}

	// Act
	var buf bytes.Buffer
	// When project=true, it should use the current working directory
	err := executeStatsCommand(context.Background(), reader, statsFlags{project: true}, &buf, nil)

	// Assert
	require.NoError(t, err)
	// Should be set to the actual current working directory (non-empty)
	assert.NotEmpty(t, reader.lastStatsOpts.Project, "project should be set to current working directory")
}

func TestStatsCommand_HistoryWithLimit(t *testing.T) {
	// Arrange - test that history limit is respected when set
	reader := &mockTrackingReader{
		history: make([]domain.CommandRecord, 20),
	}

	// Act
	var buf bytes.Buffer
	// history=15 means show 15 entries
	err := executeStatsCommand(context.Background(), reader, statsFlags{history: 15}, &buf, nil)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 15, reader.lastLimit, "history limit should be 15")
}

func TestBuildStatsCommand_ProjectFlag(t *testing.T) {
	// Arrange - test that --project flag uses current directory
	reader := &mockTrackingReader{
		summary: domain.NewStatsSummary(25, 12500, 80.0, 2*time.Minute),
	}

	cmd := buildStatsCommand(reader, nil)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	// Use --project - should filter to current directory
	cmd.SetArgs([]string{"--project"})

	// Act
	err := cmd.Execute()

	// Assert
	require.NoError(t, err)
	// The project should be set to the actual current working directory
	assert.NotEmpty(t, reader.lastStatsOpts.Project, "project should be set to current directory")
}

func TestBuildStatsCommand_HistoryFlagNoValue(t *testing.T) {
	// Arrange - test that --history flag with no value uses default limit of 10
	reader := &mockTrackingReader{
		history: make([]domain.CommandRecord, 5),
	}

	cmd := buildStatsCommand(reader, nil)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	// Use --history without a value - should default to 10
	cmd.SetArgs([]string{"--history"})

	// Act
	err := cmd.Execute()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 10, reader.lastLimit, "history limit should default to 10 when --history used without value")
}

func TestBuildStatsCommand_HistoryFlagWithValue(t *testing.T) {
	// Arrange - test that --history=N uses the specified value
	reader := &mockTrackingReader{
		history: make([]domain.CommandRecord, 5),
	}

	cmd := buildStatsCommand(reader, nil)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	// Use --history=25
	cmd.SetArgs([]string{"--history=25"})

	// Act
	err := cmd.Execute()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 25, reader.lastLimit, "history limit should be 25")
}
