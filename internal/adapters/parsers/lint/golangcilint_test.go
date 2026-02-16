package lint

import (
	"strings"
	"testing"
)

func TestGolangCILintParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData GolangCILintResult
	}{
		{
			name:  "empty output indicates clean lint",
			input: "",
			wantData: GolangCILintResult{
				Success: true,
				Issues:  []GolangCILintIssue{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewGolangCILintParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*GolangCILintResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *GolangCILintResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("GolangCILintResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if len(got.Issues) != len(tt.wantData.Issues) {
				t.Errorf("GolangCILintResult.Issues length = %d, want %d", len(got.Issues), len(tt.wantData.Issues))
			}
		})
	}
}

func TestGolangCILintParser_SingleIssue(t *testing.T) {
	// golangci-lint outputs issues in format: file:line:column: message (linter)
	input := `main.go:10:5: Error return value of 'foo' is not checked (errcheck)`

	parser := NewGolangCILintParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*GolangCILintResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *GolangCILintResult", result.Data)
	}

	if got.Success {
		t.Error("GolangCILintResult.Success = true, want false when issues present")
	}

	if len(got.Issues) != 1 {
		t.Fatalf("GolangCILintResult.Issues length = %d, want 1", len(got.Issues))
	}

	wantIssue := GolangCILintIssue{
		File:     "main.go",
		Line:     10,
		Column:   5,
		Message:  "Error return value of 'foo' is not checked",
		Linter:   "errcheck",
		Severity: "error",
	}

	if got.Issues[0] != wantIssue {
		t.Errorf("GolangCILintResult.Issues[0] = %+v, want %+v", got.Issues[0], wantIssue)
	}
}

func TestGolangCILintParser_MultipleIssues(t *testing.T) {
	input := `main.go:10:5: Error return value of 'foo' is not checked (errcheck)
utils.go:25:10: S1000: should use a simple channel send/receive instead of select (gosimple)
handler.go:50:3: printf: fmt.Printf format %s has arg x of wrong type int (govet)`

	parser := NewGolangCILintParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*GolangCILintResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *GolangCILintResult", result.Data)
	}

	if got.Success {
		t.Error("GolangCILintResult.Success = true, want false when issues present")
	}

	if len(got.Issues) != 3 {
		t.Fatalf("GolangCILintResult.Issues length = %d, want 3", len(got.Issues))
	}

	wantIssues := []GolangCILintIssue{
		{File: "main.go", Line: 10, Column: 5, Message: "Error return value of 'foo' is not checked", Linter: "errcheck", Severity: "error"},
		{File: "utils.go", Line: 25, Column: 10, Message: "S1000: should use a simple channel send/receive instead of select", Linter: "gosimple", Severity: "error"},
		{File: "handler.go", Line: 50, Column: 3, Message: "printf: fmt.Printf format %s has arg x of wrong type int", Linter: "govet", Severity: "error"},
	}

	for i, wantIssue := range wantIssues {
		if got.Issues[i] != wantIssue {
			t.Errorf("GolangCILintResult.Issues[%d] = %+v, want %+v", i, got.Issues[i], wantIssue)
		}
	}
}

func TestGolangCILintParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches golangci-lint run",
			cmd:         "golangci-lint",
			subcommands: []string{"run"},
			want:        true,
		},
		{
			name:        "matches golangci-lint with path",
			cmd:         "golangci-lint",
			subcommands: []string{"run", "./..."},
			want:        true,
		},
		{
			name:        "matches golangci-lint with no subcommands",
			cmd:         "golangci-lint",
			subcommands: []string{},
			want:        true,
		},
		{
			name:        "does not match go tool",
			cmd:         "go",
			subcommands: []string{"vet"},
			want:        false,
		},
		{
			name:        "does not match empty command",
			cmd:         "",
			subcommands: []string{},
			want:        false,
		},
	}

	parser := NewGolangCILintParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestGolangCILintParser_Schema(t *testing.T) {
	parser := NewGolangCILintParser()
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
	requiredProps := []string{"success", "issues"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}

func TestGolangCILintParser_DifferentFormats(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantIssues []GolangCILintIssue
	}{
		{
			name:  "issue without column (file:line: message)",
			input: "main.go:10: unused variable (deadcode)",
			wantIssues: []GolangCILintIssue{
				{File: "main.go", Line: 10, Column: 0, Message: "unused variable", Linter: "deadcode", Severity: "error"},
			},
		},
		{
			name:  "full path file",
			input: "/home/user/project/pkg/handler.go:42:8: ineffectual assignment (ineffassign)",
			wantIssues: []GolangCILintIssue{
				{File: "/home/user/project/pkg/handler.go", Line: 42, Column: 8, Message: "ineffectual assignment", Linter: "ineffassign", Severity: "error"},
			},
		},
		{
			name:  "relative path file",
			input: "./internal/app/main.go:15:2: exported function without comment (golint)",
			wantIssues: []GolangCILintIssue{
				{File: "./internal/app/main.go", Line: 15, Column: 2, Message: "exported function without comment", Linter: "golint", Severity: "error"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewGolangCILintParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*GolangCILintResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *GolangCILintResult", result.Data)
			}

			if len(got.Issues) != len(tt.wantIssues) {
				t.Fatalf("GolangCILintResult.Issues length = %d, want %d", len(got.Issues), len(tt.wantIssues))
			}

			for i, wantIssue := range tt.wantIssues {
				if got.Issues[i] != wantIssue {
					t.Errorf("GolangCILintResult.Issues[%d] = %+v, want %+v", i, got.Issues[i], wantIssue)
				}
			}
		})
	}
}
