package domain

import (
	"errors"
	"reflect"
	"testing"
)

func TestParseResultWithData(t *testing.T) {
	tests := []struct {
		name     string
		data     any
		raw      string
		exitCode int
	}{
		{
			name: "result with map data",
			data: map[string]any{
				"branch": "main",
				"clean":  true,
			},
			raw:      "On branch main\nnothing to commit",
			exitCode: 0,
		},
		{
			name:     "result with slice data",
			data:     []string{"file1.go", "file2.go"},
			raw:      "file1.go\nfile2.go",
			exitCode: 0,
		},
		{
			name:     "result with nil data",
			data:     nil,
			raw:      "",
			exitCode: 0,
		},
		{
			name:     "result with non-zero exit code",
			data:     map[string]any{"error": "not found"},
			raw:      "fatal: not a git repository",
			exitCode: 128,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewParseResult(tt.data, tt.raw, tt.exitCode)

			if !reflect.DeepEqual(result.Data, tt.data) {
				t.Errorf("ParseResult.Data = %v, want %v", result.Data, tt.data)
			}
			if result.Raw != tt.raw {
				t.Errorf("ParseResult.Raw = %v, want %v", result.Raw, tt.raw)
			}
			if result.ExitCode != tt.exitCode {
				t.Errorf("ParseResult.ExitCode = %v, want %v", result.ExitCode, tt.exitCode)
			}
			if result.Error != nil {
				t.Errorf("ParseResult.Error = %v, want nil", result.Error)
			}
		})
	}
}

func TestParseResultWithError(t *testing.T) {
	testErr := errors.New("parse failed")

	tests := []struct {
		name     string
		err      error
		raw      string
		exitCode int
	}{
		{
			name:     "result with error",
			err:      testErr,
			raw:      "invalid output",
			exitCode: 1,
		},
		{
			name:     "result with error and empty raw",
			err:      testErr,
			raw:      "",
			exitCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewParseResultWithError(tt.err, tt.raw, tt.exitCode)

			if result.Data != nil {
				t.Errorf("ParseResult.Data = %v, want nil", result.Data)
			}
			if result.Raw != tt.raw {
				t.Errorf("ParseResult.Raw = %v, want %v", result.Raw, tt.raw)
			}
			if result.ExitCode != tt.exitCode {
				t.Errorf("ParseResult.ExitCode = %v, want %v", result.ExitCode, tt.exitCode)
			}
			if !errors.Is(result.Error, tt.err) {
				t.Errorf("ParseResult.Error = %v, want %v", result.Error, tt.err)
			}
		})
	}
}

func TestParseResult_Success(t *testing.T) {
	tests := []struct {
		name   string
		result ParseResult
		want   bool
	}{
		{
			name: "successful result",
			result: ParseResult{
				Data:     map[string]any{"key": "value"},
				ExitCode: 0,
				Error:    nil,
			},
			want: true,
		},
		{
			name: "result with error",
			result: ParseResult{
				Data:     nil,
				ExitCode: 0,
				Error:    errors.New("failed"),
			},
			want: false,
		},
		{
			name: "result with non-zero exit code",
			result: ParseResult{
				Data:     nil,
				ExitCode: 1,
				Error:    nil,
			},
			want: false,
		},
		{
			name: "result with both error and non-zero exit",
			result: ParseResult{
				Data:     nil,
				ExitCode: 1,
				Error:    errors.New("failed"),
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.result.Success()
			if got != tt.want {
				t.Errorf("ParseResult.Success() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseResult_ZeroValue(t *testing.T) {
	// Test that zero value is usable
	var result ParseResult

	if result.Data != nil {
		t.Errorf("Zero value Data should be nil, got %v", result.Data)
	}
	if result.Raw != "" {
		t.Errorf("Zero value Raw should be empty, got %v", result.Raw)
	}
	if result.ExitCode != 0 {
		t.Errorf("Zero value ExitCode should be 0, got %v", result.ExitCode)
	}
	if result.Error != nil {
		t.Errorf("Zero value Error should be nil, got %v", result.Error)
	}
	// Zero value should be considered successful (no error, exit code 0)
	if !result.Success() {
		t.Error("Zero value ParseResult.Success() should return true")
	}
}
