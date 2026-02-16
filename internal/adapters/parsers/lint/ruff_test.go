package lint

import (
	"strings"
	"testing"
)

func TestRuffParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData RuffResult
	}{
		{
			name:  "empty output indicates clean lint",
			input: "",
			wantData: RuffResult{
				Success: true,
				Issues:  []RuffIssue{},
			},
		},
		{
			name:  "all checks passed message",
			input: "All checks passed!",
			wantData: RuffResult{
				Success: true,
				Issues:  []RuffIssue{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewRuffParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*RuffResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *RuffResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("RuffResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if len(got.Issues) != len(tt.wantData.Issues) {
				t.Errorf("RuffResult.Issues length = %d, want %d", len(got.Issues), len(tt.wantData.Issues))
			}
		})
	}
}

func TestRuffParser_SingleIssue(t *testing.T) {
	// Ruff outputs issues in format: file:line:column: CODE message
	input := `main.py:10:1: F401 ` + "`" + `os` + "`" + ` imported but unused`

	parser := NewRuffParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*RuffResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *RuffResult", result.Data)
	}

	if got.Success {
		t.Error("RuffResult.Success = true, want false when issues present")
	}

	if len(got.Issues) != 1 {
		t.Fatalf("RuffResult.Issues length = %d, want 1", len(got.Issues))
	}

	wantIssue := RuffIssue{
		File:    "main.py",
		Line:    10,
		Column:  1,
		Code:    "F401",
		Message: "`os` imported but unused",
	}

	if got.Issues[0] != wantIssue {
		t.Errorf("RuffResult.Issues[0] = %+v, want %+v", got.Issues[0], wantIssue)
	}
}

func TestRuffParser_MultipleIssues(t *testing.T) {
	input := `main.py:10:1: F401 ` + "`" + `os` + "`" + ` imported but unused
utils.py:25:5: E501 Line too long (120 > 88 characters)
app.py:42:1: I001 Import block is un-sorted or un-formatted`

	parser := NewRuffParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*RuffResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *RuffResult", result.Data)
	}

	if got.Success {
		t.Error("RuffResult.Success = true, want false when issues present")
	}

	if len(got.Issues) != 3 {
		t.Fatalf("RuffResult.Issues length = %d, want 3", len(got.Issues))
	}

	wantIssues := []RuffIssue{
		{File: "main.py", Line: 10, Column: 1, Code: "F401", Message: "`os` imported but unused"},
		{File: "utils.py", Line: 25, Column: 5, Code: "E501", Message: "Line too long (120 > 88 characters)"},
		{File: "app.py", Line: 42, Column: 1, Code: "I001", Message: "Import block is un-sorted or un-formatted"},
	}

	for i, wantIssue := range wantIssues {
		if got.Issues[i] != wantIssue {
			t.Errorf("RuffResult.Issues[%d] = %+v, want %+v", i, got.Issues[i], wantIssue)
		}
	}
}

func TestRuffParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches ruff check",
			cmd:         "ruff",
			subcommands: []string{"check"},
			want:        true,
		},
		{
			name:        "matches ruff check with path",
			cmd:         "ruff",
			subcommands: []string{"check", "."},
			want:        true,
		},
		{
			name:        "matches ruff with no subcommands",
			cmd:         "ruff",
			subcommands: []string{},
			want:        true,
		},
		{
			name:        "does not match python",
			cmd:         "python",
			subcommands: []string{"-m", "ruff"},
			want:        false,
		},
		{
			name:        "does not match empty command",
			cmd:         "",
			subcommands: []string{},
			want:        false,
		},
	}

	parser := NewRuffParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestRuffParser_Schema(t *testing.T) {
	parser := NewRuffParser()
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

func TestRuffParser_DifferentFormats(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantIssues []RuffIssue
	}{
		{
			name:  "full path file",
			input: "/home/user/project/src/main.py:10:1: F401 unused import",
			wantIssues: []RuffIssue{
				{File: "/home/user/project/src/main.py", Line: 10, Column: 1, Code: "F401", Message: "unused import"},
			},
		},
		{
			name:  "relative path file",
			input: "./src/main.py:15:2: E999 SyntaxError",
			wantIssues: []RuffIssue{
				{File: "./src/main.py", Line: 15, Column: 2, Code: "E999", Message: "SyntaxError"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewRuffParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*RuffResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *RuffResult", result.Data)
			}

			if len(got.Issues) != len(tt.wantIssues) {
				t.Fatalf("RuffResult.Issues length = %d, want %d", len(got.Issues), len(tt.wantIssues))
			}

			for i, wantIssue := range tt.wantIssues {
				if got.Issues[i] != wantIssue {
					t.Errorf("RuffResult.Issues[%d] = %+v, want %+v", i, got.Issues[i], wantIssue)
				}
			}
		})
	}
}

func TestRuffParser_FoundIssuesSummary(t *testing.T) {
	// Ruff may output a summary line at the end
	input := `main.py:10:1: F401 ` + "`" + `os` + "`" + ` imported but unused
Found 1 error.`

	parser := NewRuffParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*RuffResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *RuffResult", result.Data)
	}

	// Should parse exactly 1 issue, ignoring summary line
	if len(got.Issues) != 1 {
		t.Fatalf("RuffResult.Issues length = %d, want 1", len(got.Issues))
	}
}
