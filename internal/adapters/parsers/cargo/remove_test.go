package cargo

import (
	"strings"
	"testing"
)

func TestRemoveParser_EmptyOutput(t *testing.T) {
	parser := NewRemoveParser()
	result, err := parser.Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*RemoveResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *RemoveResult", result.Data)
	}

	if !got.Success {
		t.Error("RemoveResult.Success = false, want true for empty output")
	}

	if len(got.Removed) != 0 {
		t.Errorf("RemoveResult.Removed length = %d, want 0", len(got.Removed))
	}
}

func TestRemoveParser_SingleDependency(t *testing.T) {
	input := `    Removing serde from dependencies
`

	parser := NewRemoveParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*RemoveResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *RemoveResult", result.Data)
	}

	if !got.Success {
		t.Error("RemoveResult.Success = false, want true")
	}

	if len(got.Removed) != 1 {
		t.Fatalf("RemoveResult.Removed length = %d, want 1", len(got.Removed))
	}

	if got.Removed[0] != "serde" {
		t.Errorf("Removed[0] = %q, want %q", got.Removed[0], "serde")
	}
}

func TestRemoveParser_MultipleDependencies(t *testing.T) {
	input := `    Removing serde from dependencies
    Removing tokio from dependencies
`

	parser := NewRemoveParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*RemoveResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *RemoveResult", result.Data)
	}

	if len(got.Removed) != 2 {
		t.Errorf("RemoveResult.Removed length = %d, want 2", len(got.Removed))
	}
}

func TestRemoveParser_DevDependency(t *testing.T) {
	input := `    Removing mockall from dev-dependencies
`

	parser := NewRemoveParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*RemoveResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *RemoveResult", result.Data)
	}

	if len(got.Removed) != 1 {
		t.Fatalf("RemoveResult.Removed length = %d, want 1", len(got.Removed))
	}

	if got.Removed[0] != "mockall" {
		t.Errorf("Removed[0] = %q, want %q", got.Removed[0], "mockall")
	}
}

func TestRemoveParser_Error(t *testing.T) {
	input := `error: the dependency 'nonexistent' could not be found in 'Cargo.toml'
`

	parser := NewRemoveParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*RemoveResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *RemoveResult", result.Data)
	}

	if got.Success {
		t.Error("RemoveResult.Success = true, want false")
	}

	if len(got.Errors) != 1 {
		t.Fatalf("RemoveResult.Errors length = %d, want 1", len(got.Errors))
	}
}

func TestRemoveParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches cargo remove",
			cmd:         "cargo",
			subcommands: []string{"remove"},
			want:        true,
		},
		{
			name:        "matches cargo remove with package",
			cmd:         "cargo",
			subcommands: []string{"remove", "serde"},
			want:        true,
		},
		{
			name:        "matches cargo rm (alias)",
			cmd:         "cargo",
			subcommands: []string{"rm"},
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

	parser := NewRemoveParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestRemoveParser_Schema(t *testing.T) {
	parser := NewRemoveParser()
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

	requiredProps := []string{"success", "removed", "errors"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}
