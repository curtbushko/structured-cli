package lint

import (
	"strings"
	"testing"
)

func TestESLintParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData ESLintResult
	}{
		{
			name:  "empty output indicates clean lint",
			input: "",
			wantData: ESLintResult{
				Success: true,
				Issues:  []ESLintIssue{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewESLintParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*ESLintResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *ESLintResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("ESLintResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if len(got.Issues) != len(tt.wantData.Issues) {
				t.Errorf("ESLintResult.Issues length = %d, want %d", len(got.Issues), len(tt.wantData.Issues))
			}
		})
	}
}

func TestESLintParser_SingleIssue(t *testing.T) {
	// ESLint default formatter output: /path/file.js:line:column: message (rule)
	input := `/home/user/project/src/index.js
  10:5  error  'foo' is not defined  no-undef`

	parser := NewESLintParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*ESLintResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ESLintResult", result.Data)
	}

	if got.Success {
		t.Error("ESLintResult.Success = true, want false when issues present")
	}

	if len(got.Issues) != 1 {
		t.Fatalf("ESLintResult.Issues length = %d, want 1", len(got.Issues))
	}

	wantIssue := ESLintIssue{
		File:     "/home/user/project/src/index.js",
		Line:     10,
		Column:   5,
		Severity: "error",
		Message:  "'foo' is not defined",
		Rule:     "no-undef",
	}

	if got.Issues[0] != wantIssue {
		t.Errorf("ESLintResult.Issues[0] = %+v, want %+v", got.Issues[0], wantIssue)
	}
}

func TestESLintParser_MultipleIssues(t *testing.T) {
	input := `/home/user/project/src/index.js
  10:5  error  'foo' is not defined  no-undef
  15:1  warning  Unexpected console statement  no-console

/home/user/project/src/utils.js
  5:10  error  'bar' is assigned a value but never used  no-unused-vars`

	parser := NewESLintParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*ESLintResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ESLintResult", result.Data)
	}

	if got.Success {
		t.Error("ESLintResult.Success = true, want false when issues present")
	}

	if len(got.Issues) != 3 {
		t.Fatalf("ESLintResult.Issues length = %d, want 3", len(got.Issues))
	}

	wantIssues := []ESLintIssue{
		{File: "/home/user/project/src/index.js", Line: 10, Column: 5, Severity: "error", Message: "'foo' is not defined", Rule: "no-undef"},
		{File: "/home/user/project/src/index.js", Line: 15, Column: 1, Severity: "warning", Message: "Unexpected console statement", Rule: "no-console"},
		{File: "/home/user/project/src/utils.js", Line: 5, Column: 10, Severity: "error", Message: "'bar' is assigned a value but never used", Rule: "no-unused-vars"},
	}

	for i, wantIssue := range wantIssues {
		if got.Issues[i] != wantIssue {
			t.Errorf("ESLintResult.Issues[%d] = %+v, want %+v", i, got.Issues[i], wantIssue)
		}
	}
}

func TestESLintParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches eslint with no subcommands",
			cmd:         "eslint",
			subcommands: []string{},
			want:        true,
		},
		{
			name:        "matches eslint with file path",
			cmd:         "eslint",
			subcommands: []string{"src/"},
			want:        true,
		},
		{
			name:        "matches eslint with flags",
			cmd:         "eslint",
			subcommands: []string{"--fix", "src/"},
			want:        true,
		},
		{
			name:        "does not match npx eslint (npx is the command)",
			cmd:         "npx",
			subcommands: []string{"eslint"},
			want:        false,
		},
		{
			name:        "does not match node",
			cmd:         "node",
			subcommands: []string{},
			want:        false,
		},
		{
			name:        "does not match empty command",
			cmd:         "",
			subcommands: []string{},
			want:        false,
		},
	}

	parser := NewESLintParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestESLintParser_Schema(t *testing.T) {
	parser := NewESLintParser()
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

func TestESLintParser_SummaryLine(t *testing.T) {
	// ESLint outputs a summary line at the end
	input := `/home/user/project/src/index.js
  10:5  error  'foo' is not defined  no-undef

✖ 1 problem (1 error, 0 warnings)`

	parser := NewESLintParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*ESLintResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ESLintResult", result.Data)
	}

	// Should parse exactly 1 issue, ignoring summary line
	if len(got.Issues) != 1 {
		t.Fatalf("ESLintResult.Issues length = %d, want 1", len(got.Issues))
	}
}

func TestESLintParser_WarningOnly(t *testing.T) {
	input := `/home/user/project/src/index.js
  10:5  warning  Unexpected console statement  no-console`

	parser := NewESLintParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*ESLintResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ESLintResult", result.Data)
	}

	if got.Success {
		t.Error("ESLintResult.Success = true, want false when issues present")
	}

	if len(got.Issues) != 1 {
		t.Fatalf("ESLintResult.Issues length = %d, want 1", len(got.Issues))
	}

	if got.Issues[0].Severity != "warning" {
		t.Errorf("Issue.Severity = %q, want %q", got.Issues[0].Severity, "warning")
	}
}
