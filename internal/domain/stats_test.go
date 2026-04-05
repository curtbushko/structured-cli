package domain

import (
	"testing"
	"time"
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
