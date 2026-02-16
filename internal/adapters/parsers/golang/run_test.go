package golang

import (
	"strings"
	"testing"
)

func TestRunParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		exitCode int
		wantData RunResult
	}{
		{
			name:     "capture stdout from successful run",
			input:    "Hello, World!",
			exitCode: 0,
			wantData: RunResult{
				ExitCode: 0,
				Stdout:   "Hello, World!",
				Stderr:   "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewRunParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*RunResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *RunResult", result.Data)
			}

			if got.ExitCode != tt.wantData.ExitCode {
				t.Errorf("RunResult.ExitCode = %d, want %d", got.ExitCode, tt.wantData.ExitCode)
			}

			if got.Stdout != tt.wantData.Stdout {
				t.Errorf("RunResult.Stdout = %q, want %q", got.Stdout, tt.wantData.Stdout)
			}

			if got.Stderr != tt.wantData.Stderr {
				t.Errorf("RunResult.Stderr = %q, want %q", got.Stderr, tt.wantData.Stderr)
			}
		})
	}
}

func TestRunParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches go run",
			cmd:         "go",
			subcommands: []string{"run"},
			want:        true,
		},
		{
			name:        "matches go run with file",
			cmd:         "go",
			subcommands: []string{"run", "main.go"},
			want:        true,
		},
		{
			name:        "does not match go build",
			cmd:         "go",
			subcommands: []string{"build"},
			want:        false,
		},
		{
			name:        "does not match go test",
			cmd:         "go",
			subcommands: []string{"test"},
			want:        false,
		},
		{
			name:        "does not match git",
			cmd:         "git",
			subcommands: []string{"run"},
			want:        false,
		},
		{
			name:        "does not match go without subcommand",
			cmd:         "go",
			subcommands: []string{},
			want:        false,
		},
		{
			name:        "does not match empty command",
			cmd:         "",
			subcommands: []string{"run"},
			want:        false,
		},
	}

	parser := NewRunParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestRunParser_Schema(t *testing.T) {
	parser := NewRunParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema.ID should not be empty")
	}

	if schema.Title == "" {
		t.Error("Schema.Title should not be empty")
	}

	if schema.Type != "object" {
		t.Errorf("Schema.Type = %q, want %q", schema.Type, "object")
	}

	// Verify required properties exist
	requiredProps := []string{"exitCode", "stdout", "stderr"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}

func TestRunParser_MultilineOutput(t *testing.T) {
	input := `Line 1
Line 2
Line 3`

	parser := NewRunParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*RunResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *RunResult", result.Data)
	}

	if got.Stdout != input {
		t.Errorf("RunResult.Stdout = %q, want %q", got.Stdout, input)
	}
}

func TestRunParser_EmptyOutput(t *testing.T) {
	parser := NewRunParser()
	result, err := parser.Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*RunResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *RunResult", result.Data)
	}

	if got.Stdout != "" {
		t.Errorf("RunResult.Stdout = %q, want empty string", got.Stdout)
	}

	if got.Stderr != "" {
		t.Errorf("RunResult.Stderr = %q, want empty string", got.Stderr)
	}
}
