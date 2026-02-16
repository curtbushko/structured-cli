package domain

import (
	"testing"
)

func TestFallbackResult_Fields(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		exitCode int
	}{
		{
			name:     "basic fallback result",
			raw:      "some raw output from unsupported command",
			exitCode: 0,
		},
		{
			name:     "fallback with non-zero exit code",
			raw:      "error: command failed",
			exitCode: 1,
		},
		{
			name:     "empty raw output",
			raw:      "",
			exitCode: 0,
		},
		{
			name:     "multiline raw output",
			raw:      "line1\nline2\nline3",
			exitCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewFallbackResult(tt.raw, tt.exitCode)

			// Verify Raw field
			if result.Raw != tt.raw {
				t.Errorf("FallbackResult.Raw = %q, want %q", result.Raw, tt.raw)
			}

			// Verify Parsed is always false for fallback results
			if result.Parsed != false {
				t.Errorf("FallbackResult.Parsed = %v, want false", result.Parsed)
			}

			// Verify ExitCode field
			if result.ExitCode != tt.exitCode {
				t.Errorf("FallbackResult.ExitCode = %d, want %d", result.ExitCode, tt.exitCode)
			}
		})
	}
}

func TestFallbackResult_ZeroValue(t *testing.T) {
	// Test that zero value is usable
	var result FallbackResult

	if result.Raw != "" {
		t.Errorf("Zero value Raw should be empty, got %q", result.Raw)
	}
	if result.Parsed != false {
		t.Errorf("Zero value Parsed should be false, got %v", result.Parsed)
	}
	if result.ExitCode != 0 {
		t.Errorf("Zero value ExitCode should be 0, got %d", result.ExitCode)
	}
}

func TestNewFallbackResult(t *testing.T) {
	// Arrange
	raw := "git stash show output"
	exitCode := 0

	// Act
	result := NewFallbackResult(raw, exitCode)

	// Assert
	if result.Raw != raw {
		t.Errorf("NewFallbackResult() Raw = %q, want %q", result.Raw, raw)
	}
	if result.Parsed != false {
		t.Errorf("NewFallbackResult() Parsed = %v, want false", result.Parsed)
	}
	if result.ExitCode != exitCode {
		t.Errorf("NewFallbackResult() ExitCode = %d, want %d", result.ExitCode, exitCode)
	}
}
