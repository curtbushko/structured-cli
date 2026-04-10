package stats

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

func TestTableStatsRenderer_ImplementsInterface(t *testing.T) {
	// given: a TableStatsRenderer
	var renderer ports.StatsRenderer = NewTableStatsRenderer()

	// then: it implements StatsRenderer interface
	require.NotNil(t, renderer)
}

func TestTableStatsRenderer_RenderSummary(t *testing.T) {
	// given: a summary with known values
	renderer := NewTableStatsRenderer()
	summary := domain.NewStatsSummary(10, 5000, 45.0, 2*time.Second)
	var buf bytes.Buffer

	// when: RenderSummary is called
	err := renderer.RenderSummary(&buf, summary)

	// then: output contains key values
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "10", "output should contain total commands")
	assert.Contains(t, output, "5000", "output should contain total tokens saved")
	assert.Contains(t, output, "45.0%", "output should contain avg savings percent")
}

func TestTableStatsRenderer_RenderByParser(t *testing.T) {
	// given: command stats for two parsers
	renderer := NewTableStatsRenderer()
	stats := []domain.CommandStats{
		domain.NewCommandStats("git-status", 15, 3000, 100*time.Millisecond),
		domain.NewCommandStats("kubectl-get", 8, 1500, 200*time.Millisecond),
	}
	var buf bytes.Buffer

	// when: RenderByParser is called
	err := renderer.RenderByParser(&buf, stats)

	// then: output contains parser names and invocation counts
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "git-status", "output should contain parser name")
	assert.Contains(t, output, "kubectl-get", "output should contain parser name")
	assert.Contains(t, output, "15", "output should contain invocation count")
	assert.Contains(t, output, "8", "output should contain invocation count")
}

func TestTableStatsRenderer_RenderByFilter(t *testing.T) {
	// given: filter stats for two filters
	renderer := NewTableStatsRenderer()
	stats := []domain.FilterStats{
		domain.NewFilterStats("small", 20, 4000),
		domain.NewFilterStats("dedupe", 5, 800),
	}
	var buf bytes.Buffer

	// when: RenderByFilter is called
	err := renderer.RenderByFilter(&buf, stats)

	// then: output contains filter names and activation counts
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "small", "output should contain filter name")
	assert.Contains(t, output, "dedupe", "output should contain filter name")
	assert.Contains(t, output, "20", "output should contain activation count")
	assert.Contains(t, output, "5", "output should contain activation count")
}

func TestTableStatsRenderer_RenderHistory(t *testing.T) {
	// given: command records with known values
	renderer := NewTableStatsRenderer()
	now := time.Now()
	records := []domain.CommandRecord{
		{
			Command:        "git",
			Subcommands:    []string{"status"},
			RawTokens:      1000,
			ParsedTokens:   500,
			TokensSaved:    500,
			SavingsPercent: 50.0,
			ExecutionTime:  100 * time.Millisecond,
			Timestamp:      now,
		},
		{
			Command:        "kubectl",
			Subcommands:    []string{"get", "pods"},
			RawTokens:      2000,
			ParsedTokens:   800,
			TokensSaved:    1200,
			SavingsPercent: 60.0,
			ExecutionTime:  200 * time.Millisecond,
			Timestamp:      now.Add(-1 * time.Hour),
		},
		{
			Command:        "docker",
			Subcommands:    []string{"ps"},
			RawTokens:      500,
			ParsedTokens:   300,
			TokensSaved:    200,
			SavingsPercent: 40.0,
			ExecutionTime:  50 * time.Millisecond,
			Timestamp:      now.Add(-2 * time.Hour),
		},
	}
	var buf bytes.Buffer

	// when: RenderHistory is called
	err := renderer.RenderHistory(&buf, records)

	// then: output contains command names and timestamps
	require.NoError(t, err)
	output := buf.String()
	assert.Contains(t, output, "git", "output should contain command name")
	assert.Contains(t, output, "kubectl", "output should contain command name")
	assert.Contains(t, output, "docker", "output should contain command name")
}

func TestTableStatsRenderer_RenderSummary_Empty(t *testing.T) {
	// given: an empty summary
	renderer := NewTableStatsRenderer()
	summary := domain.StatsSummary{}
	var buf bytes.Buffer

	// when: RenderSummary is called
	err := renderer.RenderSummary(&buf, summary)

	// then: no error and output is non-empty (header still rendered)
	require.NoError(t, err)
	assert.NotEmpty(t, buf.String())
}

func TestTableStatsRenderer_RenderByParser_Empty(t *testing.T) {
	// given: empty parser stats
	renderer := NewTableStatsRenderer()
	var buf bytes.Buffer

	// when: RenderByParser is called with nil
	err := renderer.RenderByParser(&buf, nil)

	// then: no error
	require.NoError(t, err)
}

func TestTableStatsRenderer_RenderHistory_SubcommandJoin(t *testing.T) {
	// given: a record with multiple subcommands
	renderer := NewTableStatsRenderer()
	records := []domain.CommandRecord{
		{
			Command:     "kubectl",
			Subcommands: []string{"get", "pods"},
			Timestamp:   time.Now(),
		},
	}
	var buf bytes.Buffer

	// when: RenderHistory is called
	err := renderer.RenderHistory(&buf, records)

	// then: subcommands are joined in output
	require.NoError(t, err)
	output := buf.String()
	assert.True(t, strings.Contains(output, "get pods") || strings.Contains(output, "get, pods") || strings.Contains(output, "get-pods"),
		"output should contain joined subcommands, got: %s", output)
}
