package golang

import (
	"strings"
	"testing"
)

func TestModTidyParser_NoChanges(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData ModTidyResult
	}{
		{
			name:  "empty output indicates no changes needed",
			input: "",
			wantData: ModTidyResult{
				Added:   []string{},
				Removed: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewModTidyParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*ModTidyResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *ModTidyResult", result.Data)
			}

			if len(got.Added) != len(tt.wantData.Added) {
				t.Errorf("ModTidyResult.Added length = %d, want %d", len(got.Added), len(tt.wantData.Added))
			}

			if len(got.Removed) != len(tt.wantData.Removed) {
				t.Errorf("ModTidyResult.Removed length = %d, want %d", len(got.Removed), len(tt.wantData.Removed))
			}
		})
	}
}

func TestModTidyParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches go mod tidy",
			cmd:         "go",
			subcommands: []string{"mod", "tidy"},
			want:        true,
		},
		{
			name:        "matches go mod tidy with flags",
			cmd:         "go",
			subcommands: []string{"mod", "tidy", "-v"},
			want:        true,
		},
		{
			name:        "does not match go mod without tidy",
			cmd:         "go",
			subcommands: []string{"mod"},
			want:        false,
		},
		{
			name:        "does not match go mod download",
			cmd:         "go",
			subcommands: []string{"mod", "download"},
			want:        false,
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
			subcommands: []string{"mod", "tidy"},
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
			subcommands: []string{"mod", "tidy"},
			want:        false,
		},
	}

	parser := NewModTidyParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestModTidyParser_Schema(t *testing.T) {
	parser := NewModTidyParser()
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

	// Verify required properties exist
	requiredProps := []string{"added", "removed"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}

func TestModTidyParser_WithAdditions(t *testing.T) {
	// Go 1.17+ uses "go: added module@version" format
	input := `go: added github.com/example/pkg1 v1.0.0
go: added github.com/example/pkg2 v2.3.4`

	parser := NewModTidyParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*ModTidyResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ModTidyResult", result.Data)
	}

	wantAdded := []string{
		"github.com/example/pkg1 v1.0.0",
		"github.com/example/pkg2 v2.3.4",
	}

	if len(got.Added) != len(wantAdded) {
		t.Fatalf("ModTidyResult.Added length = %d, want %d", len(got.Added), len(wantAdded))
	}

	for i, want := range wantAdded {
		if got.Added[i] != want {
			t.Errorf("ModTidyResult.Added[%d] = %q, want %q", i, got.Added[i], want)
		}
	}

	if len(got.Removed) != 0 {
		t.Errorf("ModTidyResult.Removed length = %d, want 0", len(got.Removed))
	}
}

func TestModTidyParser_WithRemovals(t *testing.T) {
	// Go mod tidy can also show removed modules in verbose mode
	input := `go: removed github.com/unused/pkg1 v1.0.0
go: removed github.com/unused/pkg2 v0.5.0`

	parser := NewModTidyParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*ModTidyResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ModTidyResult", result.Data)
	}

	wantRemoved := []string{
		"github.com/unused/pkg1 v1.0.0",
		"github.com/unused/pkg2 v0.5.0",
	}

	if len(got.Added) != 0 {
		t.Errorf("ModTidyResult.Added length = %d, want 0", len(got.Added))
	}

	if len(got.Removed) != len(wantRemoved) {
		t.Fatalf("ModTidyResult.Removed length = %d, want %d", len(got.Removed), len(wantRemoved))
	}

	for i, want := range wantRemoved {
		if got.Removed[i] != want {
			t.Errorf("ModTidyResult.Removed[%d] = %q, want %q", i, got.Removed[i], want)
		}
	}
}

func TestModTidyParser_Mixed(t *testing.T) {
	input := `go: added github.com/new/pkg v1.0.0
go: removed github.com/old/pkg v0.9.0
go: added github.com/another/pkg v2.0.0`

	parser := NewModTidyParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*ModTidyResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ModTidyResult", result.Data)
	}

	wantAdded := []string{
		"github.com/new/pkg v1.0.0",
		"github.com/another/pkg v2.0.0",
	}
	wantRemoved := []string{
		"github.com/old/pkg v0.9.0",
	}

	if len(got.Added) != len(wantAdded) {
		t.Fatalf("ModTidyResult.Added length = %d, want %d", len(got.Added), len(wantAdded))
	}

	for i, want := range wantAdded {
		if got.Added[i] != want {
			t.Errorf("ModTidyResult.Added[%d] = %q, want %q", i, got.Added[i], want)
		}
	}

	if len(got.Removed) != len(wantRemoved) {
		t.Fatalf("ModTidyResult.Removed length = %d, want %d", len(got.Removed), len(wantRemoved))
	}

	for i, want := range wantRemoved {
		if got.Removed[i] != want {
			t.Errorf("ModTidyResult.Removed[%d] = %q, want %q", i, got.Removed[i], want)
		}
	}
}
