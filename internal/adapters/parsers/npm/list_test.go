package npm

import (
	"strings"
	"testing"
)

func TestListParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData ListResult
	}{
		{
			name:  "empty project",
			input: "myproject@1.0.0 /path/to/project",
			wantData: ListResult{
				Success:      true,
				Name:         "myproject",
				Version:      "1.0.0",
				Dependencies: []ListDependency{},
				Problems:     []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewListParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*ListResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *ListResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("ListResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if got.Name != tt.wantData.Name {
				t.Errorf("ListResult.Name = %v, want %v", got.Name, tt.wantData.Name)
			}

			if got.Version != tt.wantData.Version {
				t.Errorf("ListResult.Version = %v, want %v", got.Version, tt.wantData.Version)
			}
		})
	}
}

func TestListParser_WithDependencies(t *testing.T) {
	// npm list output format (tree view)
	input := `myproject@1.0.0 /path/to/project
+-- express@4.18.2
+-- lodash@4.17.21
`

	parser := NewListParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*ListResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ListResult", result.Data)
	}

	if !got.Success {
		t.Error("ListResult.Success = false, want true")
	}

	if got.Name != "myproject" {
		t.Errorf("ListResult.Name = %q, want %q", got.Name, "myproject")
	}

	if got.Version != "1.0.0" {
		t.Errorf("ListResult.Version = %q, want %q", got.Version, "1.0.0")
	}

	if len(got.Dependencies) != 2 {
		t.Fatalf("ListResult.Dependencies length = %d, want 2", len(got.Dependencies))
	}

	// Check first dependency
	if got.Dependencies[0].Name != "express" {
		t.Errorf("Dependencies[0].Name = %q, want %q", got.Dependencies[0].Name, "express")
	}

	if got.Dependencies[0].Version != "4.18.2" {
		t.Errorf("Dependencies[0].Version = %q, want %q", got.Dependencies[0].Version, "4.18.2")
	}
}

func TestListParser_WithProblems(t *testing.T) {
	input := `myproject@1.0.0 /path/to/project
+-- UNMET DEPENDENCY lodash@^4.0.0
npm ERR! missing: lodash@^4.0.0, required by myproject@1.0.0
`

	parser := NewListParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*ListResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ListResult", result.Data)
	}

	if got.Success {
		t.Error("ListResult.Success = true, want false when problems exist")
	}

	if len(got.Problems) == 0 {
		t.Error("ListResult.Problems should not be empty")
	}
}

func TestListParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches npm list",
			cmd:         "npm",
			subcommands: []string{"list"},
			want:        true,
		},
		{
			name:        "matches npm ls",
			cmd:         "npm",
			subcommands: []string{"ls"},
			want:        true,
		},
		{
			name:        "matches npm ll",
			cmd:         "npm",
			subcommands: []string{"ll"},
			want:        true,
		},
		{
			name:        "matches npm la",
			cmd:         "npm",
			subcommands: []string{"la"},
			want:        true,
		},
		{
			name:        "does not match npm install",
			cmd:         "npm",
			subcommands: []string{"install"},
			want:        false,
		},
		{
			name:        "does not match yarn list",
			cmd:         "yarn",
			subcommands: []string{"list"},
			want:        false,
		},
		{
			name:        "does not match empty",
			cmd:         "",
			subcommands: []string{},
			want:        false,
		},
	}

	parser := NewListParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestListParser_Schema(t *testing.T) {
	parser := NewListParser()
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

	requiredProps := []string{"success", "name", "version", "dependencies"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}
