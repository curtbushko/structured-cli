package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStatsSummary(t *testing.T) {
	tests := []struct {
		name              string
		totalCommands     int
		totalTokensSaved  int
		avgSavingsPercent float64
		totalExecTime     time.Duration
	}{
		{
			name:              "normal usage stats",
			totalCommands:     100,
			totalTokensSaved:  50000,
			avgSavingsPercent: 75.5,
			totalExecTime:     5 * time.Second,
		},
		{
			name:              "zero values",
			totalCommands:     0,
			totalTokensSaved:  0,
			avgSavingsPercent: 0.0,
			totalExecTime:     0,
		},
		{
			name:              "single command",
			totalCommands:     1,
			totalTokensSaved:  250,
			avgSavingsPercent: 80.0,
			totalExecTime:     100 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := NewStatsSummary(tt.totalCommands, tt.totalTokensSaved, tt.avgSavingsPercent, tt.totalExecTime)

			if summary.TotalCommands != tt.totalCommands {
				t.Errorf("StatsSummary.TotalCommands = %v, want %v", summary.TotalCommands, tt.totalCommands)
			}
			if summary.TotalTokensSaved != tt.totalTokensSaved {
				t.Errorf("StatsSummary.TotalTokensSaved = %v, want %v", summary.TotalTokensSaved, tt.totalTokensSaved)
			}
			if summary.AvgSavingsPercent != tt.avgSavingsPercent {
				t.Errorf("StatsSummary.AvgSavingsPercent = %v, want %v", summary.AvgSavingsPercent, tt.avgSavingsPercent)
			}
			if summary.TotalExecutionTime != tt.totalExecTime {
				t.Errorf("StatsSummary.TotalExecutionTime = %v, want %v", summary.TotalExecutionTime, tt.totalExecTime)
			}
		})
	}
}

func TestNewCommandStats(t *testing.T) {
	tests := []struct {
		name            string
		parserName      string
		invocationCount int
		totalTokens     int
		avgExecTime     time.Duration
	}{
		{
			name:            "git status parser stats",
			parserName:      "git-status",
			invocationCount: 50,
			totalTokens:     25000,
			avgExecTime:     45 * time.Millisecond,
		},
		{
			name:            "kubectl parser stats",
			parserName:      "kubectl-get",
			invocationCount: 100,
			totalTokens:     80000,
			avgExecTime:     200 * time.Millisecond,
		},
		{
			name:            "zero invocations",
			parserName:      "docker-ps",
			invocationCount: 0,
			totalTokens:     0,
			avgExecTime:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stats := NewCommandStats(tt.parserName, tt.invocationCount, tt.totalTokens, tt.avgExecTime)

			if stats.ParserName != tt.parserName {
				t.Errorf("CommandStats.ParserName = %v, want %v", stats.ParserName, tt.parserName)
			}
			if stats.InvocationCount != tt.invocationCount {
				t.Errorf("CommandStats.InvocationCount = %v, want %v", stats.InvocationCount, tt.invocationCount)
			}
			if stats.TotalTokensSaved != tt.totalTokens {
				t.Errorf("CommandStats.TotalTokensSaved = %v, want %v", stats.TotalTokensSaved, tt.totalTokens)
			}
			if stats.AvgExecutionTime != tt.avgExecTime {
				t.Errorf("CommandStats.AvgExecutionTime = %v, want %v", stats.AvgExecutionTime, tt.avgExecTime)
			}
		})
	}
}

func TestAggregatedCommandStats_Fields(t *testing.T) {
	stats := NewAggregatedCommandStats(
		"git status",
		5,
		10000,
		75.5,
		500*time.Millisecond,
	)

	assert.Equal(t, "git status", stats.CommandName)
	assert.Equal(t, 5, stats.Count)
	assert.Equal(t, 10000, stats.TotalTokensSaved)
	assert.InDelta(t, 75.5, stats.AvgSavingsPercent, 0.001)
	assert.Equal(t, 500*time.Millisecond, stats.AvgExecutionTime)
}

func TestAggregatedCommandStats_AvgSavingsPercent(t *testing.T) {
	stats := NewAggregatedCommandStats(
		"npm test",
		3,
		5000,
		82.5,
		200*time.Millisecond,
	)

	assert.InDelta(t, 82.5, stats.AvgSavingsPercent, 0.001)
}

func TestNormalizeCommandName_StripsPaths(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "git status with path",
			input:    "git status /path/to/repo",
			expected: "git status",
		},
		{
			name:     "npm test with file",
			input:    "npm test file.js",
			expected: "npm test",
		},
		{
			name:     "go build with pattern",
			input:    "go build ./...",
			expected: "go build",
		},
		{
			name:     "simple command no args",
			input:    "git status",
			expected: "git status",
		},
		{
			name:     "docker compose up",
			input:    "docker compose up",
			expected: "docker compose up",
		},
		{
			name:     "command with absolute path arg",
			input:    "cat /etc/hosts",
			expected: "cat",
		},
		{
			name:     "command with relative path arg",
			input:    "go test ./internal/domain/...",
			expected: "go test",
		},
		{
			name:     "single command",
			input:    "ls",
			expected: "ls",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeCommandName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateImpact_RelativeScaling(t *testing.T) {
	t.Run("scales relative to highest savings", func(t *testing.T) {
		stats := []AggregatedCommandStats{
			NewAggregatedCommandStats("git status", 10, 10000, 80.0, 100*time.Millisecond),
			NewAggregatedCommandStats("npm test", 5, 5000, 60.0, 200*time.Millisecond),
			NewAggregatedCommandStats("go build", 2, 2000, 50.0, 300*time.Millisecond),
		}

		impacts := CalculateImpact(stats)

		require.Len(t, impacts, 3)
		// Highest saving = 100%
		assert.InDelta(t, 100.0, impacts[0].ImpactPercent, 0.001)
		// 5000/10000 = 50%
		assert.InDelta(t, 50.0, impacts[1].ImpactPercent, 0.001)
		// 2000/10000 = 20%
		assert.InDelta(t, 20.0, impacts[2].ImpactPercent, 0.001)
	})

	t.Run("empty input returns empty", func(t *testing.T) {
		impacts := CalculateImpact(nil)
		assert.Empty(t, impacts)
	})

	t.Run("single command gets 100%", func(t *testing.T) {
		stats := []AggregatedCommandStats{
			NewAggregatedCommandStats("git status", 10, 5000, 80.0, 100*time.Millisecond),
		}

		impacts := CalculateImpact(stats)

		require.Len(t, impacts, 1)
		assert.InDelta(t, 100.0, impacts[0].ImpactPercent, 0.001)
	})

	t.Run("all zero savings returns zero impact", func(t *testing.T) {
		stats := []AggregatedCommandStats{
			NewAggregatedCommandStats("git status", 10, 0, 0.0, 100*time.Millisecond),
			NewAggregatedCommandStats("npm test", 5, 0, 0.0, 200*time.Millisecond),
		}

		impacts := CalculateImpact(stats)

		require.Len(t, impacts, 2)
		assert.InDelta(t, 0.0, impacts[0].ImpactPercent, 0.001)
		assert.InDelta(t, 0.0, impacts[1].ImpactPercent, 0.001)
	})
}
