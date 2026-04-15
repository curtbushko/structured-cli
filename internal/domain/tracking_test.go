package domain

import (
	"testing"
	"time"
)

func TestNewCommandRecord(t *testing.T) {
	tests := []struct {
		name         string
		command      string
		subcommands  []string
		rawTokens    int
		parsedTokens int
		execTime     time.Duration
		project      string
	}{
		{
			name:         "git status command",
			command:      "git",
			subcommands:  []string{"status"},
			rawTokens:    1000,
			parsedTokens: 200,
			execTime:     50 * time.Millisecond,
			project:      "my-project",
		},
		{
			name:         "command with no subcommands",
			command:      "ls",
			subcommands:  nil,
			rawTokens:    500,
			parsedTokens: 100,
			execTime:     10 * time.Millisecond,
			project:      "",
		},
		{
			name:         "command with multiple subcommands",
			command:      "kubectl",
			subcommands:  []string{"config", "view"},
			rawTokens:    2000,
			parsedTokens: 400,
			execTime:     100 * time.Millisecond,
			project:      "k8s-cluster",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			record := NewCommandRecord(tt.command, tt.subcommands, tt.rawTokens, tt.parsedTokens, tt.execTime, tt.project)
			after := time.Now()

			if record.Command != tt.command {
				t.Errorf("CommandRecord.Command = %v, want %v", record.Command, tt.command)
			}
			if len(record.Subcommands) != len(tt.subcommands) {
				t.Errorf("CommandRecord.Subcommands length = %v, want %v", len(record.Subcommands), len(tt.subcommands))
			}
			for i, sub := range tt.subcommands {
				if record.Subcommands[i] != sub {
					t.Errorf("CommandRecord.Subcommands[%d] = %v, want %v", i, record.Subcommands[i], sub)
				}
			}
			if record.RawTokens != tt.rawTokens {
				t.Errorf("CommandRecord.RawTokens = %v, want %v", record.RawTokens, tt.rawTokens)
			}
			if record.ParsedTokens != tt.parsedTokens {
				t.Errorf("CommandRecord.ParsedTokens = %v, want %v", record.ParsedTokens, tt.parsedTokens)
			}

			expectedSaved := tt.rawTokens - tt.parsedTokens
			if record.TokensSaved != expectedSaved {
				t.Errorf("CommandRecord.TokensSaved = %v, want %v", record.TokensSaved, expectedSaved)
			}

			var expectedPercent float64
			if tt.rawTokens > 0 {
				expectedPercent = float64(expectedSaved) / float64(tt.rawTokens) * 100
			}
			if record.SavingsPercent != expectedPercent {
				t.Errorf("CommandRecord.SavingsPercent = %v, want %v", record.SavingsPercent, expectedPercent)
			}

			if record.ExecutionTime != tt.execTime {
				t.Errorf("CommandRecord.ExecutionTime = %v, want %v", record.ExecutionTime, tt.execTime)
			}
			if record.Project != tt.project {
				t.Errorf("CommandRecord.Project = %v, want %v", record.Project, tt.project)
			}

			// Timestamp should be between before and after
			if record.Timestamp.Before(before) || record.Timestamp.After(after) {
				t.Errorf("CommandRecord.Timestamp = %v, expected between %v and %v", record.Timestamp, before, after)
			}
		})
	}
}

func TestEstimateTokens(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{
			name:     "empty string",
			input:    "",
			expected: 0,
		},
		{
			name:     "short string",
			input:    "hello",
			expected: 1, // 5 chars / 4 = 1
		},
		{
			name:     "longer string",
			input:    "this is a longer string for testing",
			expected: 8, // 35 chars / 4 = 8
		},
		{
			name:     "exact multiple of 4",
			input:    "12345678",
			expected: 2, // 8 chars / 4 = 2
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EstimateTokens(tt.input)
			if got != tt.expected {
				t.Errorf("EstimateTokens(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestEstimateTokens_MultiLineText(t *testing.T) {
	// Multi-line text should count newlines as contributing to tokens.
	// LLM tokenizers typically treat newlines as separate tokens.
	// The estimation should reflect this by counting newlines.

	tests := []struct {
		name        string
		input       string
		minExpected int // minimum expected tokens (newlines should add to count)
	}{
		{
			name: "multi-line text with 5 newlines",
			// "line1\nline2\nline3\nline4\nline5\n" = 30 bytes total (25 chars + 5 newlines)
			// Base tokens: 30/4 = 7 tokens
			// Newline tokens: 5
			// Total: 7 + 5 = 12 tokens
			input:       "line1\nline2\nline3\nline4\nline5\n",
			minExpected: 12, // base tokens + newlines
		},
		{
			name: "multi-line text with 10 newlines",
			// 10 lines of "lineN\n" where N is 1-9 (5 chars each) + "line10\n" (7 chars)
			// Total: 9*6 + 7 = 61 chars = 61 bytes with 10 newlines
			// Without newline counting: 61/4 = 15 tokens
			// With newline counting: 15 + 10 = 25 tokens minimum
			input:       "line1\nline2\nline3\nline4\nline5\nline6\nline7\nline8\nline9\nline10\n",
			minExpected: 25, // base tokens + newlines
		},
		{
			name: "single line no newline",
			// No newlines, so just regular char/4 estimation
			input:       "single line text here",
			minExpected: 5, // 21 chars / 4 = 5
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := EstimateTokens(tt.input)
			if got < tt.minExpected {
				t.Errorf("EstimateTokens() = %v, want at least %v (input has %d newlines)",
					got, tt.minExpected, testCountNewlines(tt.input))
			}
		})
	}
}

// testCountNewlines counts the number of newline characters in a string (test helper).
func testCountNewlines(s string) int {
	count := 0
	for _, c := range s {
		if c == '\n' {
			count++
		}
	}
	return count
}

func TestNewParseFailure(t *testing.T) {
	tests := []struct {
		name            string
		command         string
		errorMsg        string
		fallbackSuccess bool
	}{
		{
			name:            "parse failure with fallback success",
			command:         "git status",
			errorMsg:        "unknown output format",
			fallbackSuccess: true,
		},
		{
			name:            "parse failure without fallback",
			command:         "docker ps",
			errorMsg:        "failed to parse container list",
			fallbackSuccess: false,
		},
		{
			name:            "parse failure with empty error message",
			command:         "npm list",
			errorMsg:        "",
			fallbackSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			failure := NewParseFailure(tt.command, tt.errorMsg, tt.fallbackSuccess)
			after := time.Now()

			if failure.Command != tt.command {
				t.Errorf("ParseFailure.Command = %v, want %v", failure.Command, tt.command)
			}
			if failure.ErrorMessage != tt.errorMsg {
				t.Errorf("ParseFailure.ErrorMessage = %v, want %v", failure.ErrorMessage, tt.errorMsg)
			}
			if failure.FallbackSuccess != tt.fallbackSuccess {
				t.Errorf("ParseFailure.FallbackSuccess = %v, want %v", failure.FallbackSuccess, tt.fallbackSuccess)
			}

			// Timestamp should be between before and after
			if failure.Timestamp.Before(before) || failure.Timestamp.After(after) {
				t.Errorf("ParseFailure.Timestamp = %v, expected between %v and %v", failure.Timestamp, before, after)
			}
		})
	}
}
