package cargo

import (
	"strings"
	"testing"
)

func TestAddParser_EmptyOutput(t *testing.T) {
	parser := NewAddParser()
	result, err := parser.Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*AddResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *AddResult", result.Data)
	}

	if !got.Success {
		t.Error("AddResult.Success = false, want true for empty output")
	}

	if len(got.Dependencies) != 0 {
		t.Errorf("AddResult.Dependencies length = %d, want 0", len(got.Dependencies))
	}
}

func TestAddParser_SingleDependency(t *testing.T) {
	input := `    Updating crates.io index
      Adding serde v1.0.193 to dependencies
             Features:
             + derive
             + std
`

	parser := NewAddParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*AddResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *AddResult", result.Data)
	}

	if !got.Success {
		t.Error("AddResult.Success = false, want true")
	}

	if len(got.Dependencies) != 1 {
		t.Fatalf("AddResult.Dependencies length = %d, want 1", len(got.Dependencies))
	}

	dep := got.Dependencies[0]
	if dep.Name != "serde" {
		t.Errorf("Dependency.Name = %q, want %q", dep.Name, "serde")
	}

	if dep.Version != "1.0.193" {
		t.Errorf("Dependency.Version = %q, want %q", dep.Version, "1.0.193")
	}
}

func TestAddParser_DevDependency(t *testing.T) {
	input := `    Updating crates.io index
      Adding mockall v0.12.1 to dev-dependencies
`

	parser := NewAddParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*AddResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *AddResult", result.Data)
	}

	if len(got.Dependencies) != 1 {
		t.Fatalf("AddResult.Dependencies length = %d, want 1", len(got.Dependencies))
	}

	if !got.Dependencies[0].Dev {
		t.Error("Dependency.Dev = false, want true")
	}
}

func TestAddParser_BuildDependency(t *testing.T) {
	input := `    Updating crates.io index
      Adding cc v1.0.83 to build-dependencies
`

	parser := NewAddParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*AddResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *AddResult", result.Data)
	}

	if len(got.Dependencies) != 1 {
		t.Fatalf("AddResult.Dependencies length = %d, want 1", len(got.Dependencies))
	}

	if !got.Dependencies[0].Build {
		t.Error("Dependency.Build = false, want true")
	}
}

func TestAddParser_MultipleDependencies(t *testing.T) {
	input := `    Updating crates.io index
      Adding serde v1.0.193 to dependencies
      Adding tokio v1.35.0 to dependencies
`

	parser := NewAddParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*AddResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *AddResult", result.Data)
	}

	if len(got.Dependencies) != 2 {
		t.Errorf("AddResult.Dependencies length = %d, want 2", len(got.Dependencies))
	}
}

func TestAddParser_Error(t *testing.T) {
	input := `error: the crate 'nonexistent_crate_xyz' could not be found in registry index.
`

	parser := NewAddParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*AddResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *AddResult", result.Data)
	}

	if got.Success {
		t.Error("AddResult.Success = true, want false")
	}

	if len(got.Errors) != 1 {
		t.Fatalf("AddResult.Errors length = %d, want 1", len(got.Errors))
	}
}

func TestAddParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches cargo add",
			cmd:         "cargo",
			subcommands: []string{"add"},
			want:        true,
		},
		{
			name:        "matches cargo add with package",
			cmd:         "cargo",
			subcommands: []string{"add", "serde"},
			want:        true,
		},
		{
			name:        "matches cargo add with flags",
			cmd:         "cargo",
			subcommands: []string{"add", "--dev", "mockall"},
			want:        true,
		},
		{
			name:        "does not match cargo build",
			cmd:         "cargo",
			subcommands: []string{"build"},
			want:        false,
		},
		{
			name:        "does not match cargo without subcommand",
			cmd:         "cargo",
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

	parser := NewAddParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestAddParser_Schema(t *testing.T) {
	parser := NewAddParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema.ID should not be empty")
	}

	if schema.Title == "" {
		t.Error("Schema.Title should not be empty")
	}

	if schema.Type != schemaTypeObject {
		t.Errorf("Schema.Type = %q, want %q", schema.Type, schemaTypeObject)
	}

	requiredProps := []string{"success", "dependencies", "errors"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}
