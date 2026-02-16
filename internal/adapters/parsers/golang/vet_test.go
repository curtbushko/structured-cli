package golang

import (
	"strings"
	"testing"
)

func TestVetParser_NoIssues(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData VetResult
	}{
		{
			name:  "empty output indicates clean vet",
			input: "",
			wantData: VetResult{
				Issues: []VetIssue{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewVetParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*VetResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *VetResult", result.Data)
			}

			if len(got.Issues) != len(tt.wantData.Issues) {
				t.Errorf("VetResult.Issues length = %d, want %d", len(got.Issues), len(tt.wantData.Issues))
			}
		})
	}
}

func TestVetParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches go vet",
			cmd:         "go",
			subcommands: []string{"vet"},
			want:        true,
		},
		{
			name:        "matches go vet with path",
			cmd:         "go",
			subcommands: []string{"vet", "./..."},
			want:        true,
		},
		{
			name:        "does not match go build",
			cmd:         "go",
			subcommands: []string{"build"},
			want:        false,
		},
		{
			name:        "does not match git",
			cmd:         "git",
			subcommands: []string{"vet"},
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
			subcommands: []string{"vet"},
			want:        false,
		},
	}

	parser := NewVetParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestVetParser_Schema(t *testing.T) {
	parser := NewVetParser()
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
	requiredProps := []string{"issues"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}

func TestVetParser_SingleIssue(t *testing.T) {
	input := "main.go:10:5: printf call has arguments but no formatting directives"

	parser := NewVetParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*VetResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *VetResult", result.Data)
	}

	if len(got.Issues) != 1 {
		t.Fatalf("VetResult.Issues length = %d, want 1", len(got.Issues))
	}

	wantIssue := VetIssue{
		File:    "main.go",
		Line:    10,
		Column:  5,
		Message: "printf call has arguments but no formatting directives",
	}

	if got.Issues[0] != wantIssue {
		t.Errorf("VetResult.Issues[0] = %+v, want %+v", got.Issues[0], wantIssue)
	}
}

func TestVetParser_MultipleIssues(t *testing.T) {
	input := `main.go:10:5: printf call has arguments but no formatting directives
utils.go:25:10: unreachable code
handler.go:50:3: loop variable i captured by func literal`

	parser := NewVetParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*VetResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *VetResult", result.Data)
	}

	if len(got.Issues) != 3 {
		t.Fatalf("VetResult.Issues length = %d, want 3", len(got.Issues))
	}

	wantIssues := []VetIssue{
		{File: "main.go", Line: 10, Column: 5, Message: "printf call has arguments but no formatting directives"},
		{File: "utils.go", Line: 25, Column: 10, Message: "unreachable code"},
		{File: "handler.go", Line: 50, Column: 3, Message: "loop variable i captured by func literal"},
	}

	for i, wantIssue := range wantIssues {
		if got.Issues[i] != wantIssue {
			t.Errorf("VetResult.Issues[%d] = %+v, want %+v", i, got.Issues[i], wantIssue)
		}
	}
}

func TestVetParser_DifferentFormats(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantIssues []VetIssue
	}{
		{
			name:  "issue with no column (file:line: message)",
			input: "main.go:10: suspicious check",
			wantIssues: []VetIssue{
				{File: "main.go", Line: 10, Column: 0, Message: "suspicious check"},
			},
		},
		{
			name:  "full path file",
			input: "/home/user/project/pkg/handler.go:42:8: unused result of call",
			wantIssues: []VetIssue{
				{File: "/home/user/project/pkg/handler.go", Line: 42, Column: 8, Message: "unused result of call"},
			},
		},
		{
			name:  "relative path file",
			input: "./internal/app/main.go:15:2: composite literal uses unkeyed fields",
			wantIssues: []VetIssue{
				{File: "./internal/app/main.go", Line: 15, Column: 2, Message: "composite literal uses unkeyed fields"},
			},
		},
		{
			name:  "mixed output with empty lines",
			input: "\nmain.go:10:5: issue one\n\nutils.go:20:3: issue two\n",
			wantIssues: []VetIssue{
				{File: "main.go", Line: 10, Column: 5, Message: "issue one"},
				{File: "utils.go", Line: 20, Column: 3, Message: "issue two"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewVetParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*VetResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *VetResult", result.Data)
			}

			if len(got.Issues) != len(tt.wantIssues) {
				t.Fatalf("VetResult.Issues length = %d, want %d", len(got.Issues), len(tt.wantIssues))
			}

			for i, wantIssue := range tt.wantIssues {
				if got.Issues[i] != wantIssue {
					t.Errorf("VetResult.Issues[%d] = %+v, want %+v", i, got.Issues[i], wantIssue)
				}
			}
		})
	}
}
