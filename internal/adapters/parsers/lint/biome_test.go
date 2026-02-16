package lint

import (
	"strings"
	"testing"
)

func TestBiomeParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData BiomeResult
	}{
		{
			name:  "empty output indicates clean check",
			input: "",
			wantData: BiomeResult{
				Success: true,
				Issues:  []BiomeIssue{},
			},
		},
		{
			name:  "checked files message with no issues",
			input: "Checked 10 files in 50ms. No fixes applied.",
			wantData: BiomeResult{
				Success: true,
				Issues:  []BiomeIssue{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewBiomeParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*BiomeResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *BiomeResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("BiomeResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if len(got.Issues) != len(tt.wantData.Issues) {
				t.Errorf("BiomeResult.Issues length = %d, want %d", len(got.Issues), len(tt.wantData.Issues))
			}
		})
	}
}

func TestBiomeParser_SingleIssue(t *testing.T) {
	// Biome check outputs issues in format:
	// path/file.js:line:column lint/category message
	input := `src/index.js:10:5 lint/suspicious/noExplicitAny Unexpected any. Specify a different type.`

	parser := NewBiomeParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*BiomeResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *BiomeResult", result.Data)
	}

	if got.Success {
		t.Error("BiomeResult.Success = true, want false when issues present")
	}

	if len(got.Issues) != 1 {
		t.Fatalf("BiomeResult.Issues length = %d, want 1", len(got.Issues))
	}

	wantIssue := BiomeIssue{
		File:     "src/index.js",
		Line:     10,
		Column:   5,
		Severity: "error",
		Category: "lint",
		Message:  "Unexpected any. Specify a different type.",
		Rule:     "suspicious/noExplicitAny",
	}

	if got.Issues[0] != wantIssue {
		t.Errorf("BiomeResult.Issues[0] = %+v, want %+v", got.Issues[0], wantIssue)
	}
}

func TestBiomeParser_MultipleIssues(t *testing.T) {
	input := `src/index.js:10:5 lint/suspicious/noExplicitAny Unexpected any. Specify a different type.
src/utils.ts:25:1 lint/complexity/noForEach Prefer for...of instead of forEach.
src/app.tsx:42:10 format The file contains formatting issues.`

	parser := NewBiomeParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*BiomeResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *BiomeResult", result.Data)
	}

	if got.Success {
		t.Error("BiomeResult.Success = true, want false when issues present")
	}

	if len(got.Issues) != 3 {
		t.Fatalf("BiomeResult.Issues length = %d, want 3", len(got.Issues))
	}

	wantIssues := []BiomeIssue{
		{File: "src/index.js", Line: 10, Column: 5, Severity: "error", Category: "lint", Message: "Unexpected any. Specify a different type.", Rule: "suspicious/noExplicitAny"},
		{File: "src/utils.ts", Line: 25, Column: 1, Severity: "error", Category: "lint", Message: "Prefer for...of instead of forEach.", Rule: "complexity/noForEach"},
		{File: "src/app.tsx", Line: 42, Column: 10, Severity: "error", Category: "format", Message: "The file contains formatting issues.", Rule: ""},
	}

	for i, wantIssue := range wantIssues {
		if got.Issues[i] != wantIssue {
			t.Errorf("BiomeResult.Issues[%d] = %+v, want %+v", i, got.Issues[i], wantIssue)
		}
	}
}

func TestBiomeParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches biome check",
			cmd:         "biome",
			subcommands: []string{"check"},
			want:        true,
		},
		{
			name:        "matches biome lint",
			cmd:         "biome",
			subcommands: []string{"lint"},
			want:        true,
		},
		{
			name:        "matches biome format",
			cmd:         "biome",
			subcommands: []string{"format"},
			want:        true,
		},
		{
			name:        "matches biome with no subcommands",
			cmd:         "biome",
			subcommands: []string{},
			want:        true,
		},
		{
			name:        "does not match npx biome",
			cmd:         "npx",
			subcommands: []string{"biome"},
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

	parser := NewBiomeParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestBiomeParser_Schema(t *testing.T) {
	parser := NewBiomeParser()
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

func TestBiomeParser_FormattingIssue(t *testing.T) {
	input := `src/index.js:10:5 format The file is not formatted correctly.`

	parser := NewBiomeParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*BiomeResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *BiomeResult", result.Data)
	}

	if len(got.Issues) != 1 {
		t.Fatalf("BiomeResult.Issues length = %d, want 1", len(got.Issues))
	}

	if got.Issues[0].Category != "format" {
		t.Errorf("Issue.Category = %q, want %q", got.Issues[0].Category, "format")
	}
}
