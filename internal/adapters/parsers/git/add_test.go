package git_test

import (
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/git"
)

func TestAddParser_SuccessfulAdd(t *testing.T) {
	// git add -v outputs "add 'filename'" for each file added
	input := `add 'src/main.go'
add 'README.md'
`
	parser := git.NewAddParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	add, ok := result.Data.(*git.Add)
	if !ok {
		t.Fatalf("result.Data is not *git.Add, got %T", result.Data)
	}

	if len(add.Staged) != 2 {
		t.Fatalf("Staged len = %d, want 2", len(add.Staged))
	}
	if add.Staged[0] != "src/main.go" {
		t.Errorf("Staged[0] = %q, want %q", add.Staged[0], "src/main.go")
	}
	if add.Staged[1] != "README.md" {
		t.Errorf("Staged[1] = %q, want %q", add.Staged[1], "README.md")
	}
	if len(add.Errors) != 0 {
		t.Errorf("Errors = %v, want empty", add.Errors)
	}
}

func TestAddParser_EmptyOutput(t *testing.T) {
	// Silent success when files are already staged or git add without -v
	input := ``
	parser := git.NewAddParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	add, ok := result.Data.(*git.Add)
	if !ok {
		t.Fatalf("result.Data is not *git.Add, got %T", result.Data)
	}

	if len(add.Staged) != 0 {
		t.Errorf("Staged = %v, want empty", add.Staged)
	}
}

func TestAddParser_FileNotFound(t *testing.T) {
	input := `fatal: pathspec 'nonexistent.txt' did not match any files
`
	parser := git.NewAddParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	add, ok := result.Data.(*git.Add)
	if !ok {
		t.Fatalf("result.Data is not *git.Add, got %T", result.Data)
	}

	if len(add.Errors) != 1 {
		t.Fatalf("Errors len = %d, want 1", len(add.Errors))
	}
	if !strings.Contains(add.Errors[0], "nonexistent.txt") {
		t.Errorf("Errors[0] = %q, want to contain 'nonexistent.txt'", add.Errors[0])
	}
}

func TestAddParser_DryRun(t *testing.T) {
	// git add --dry-run output
	input := `add 'file1.go'
add 'file2.go'
`
	parser := git.NewAddParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	add, ok := result.Data.(*git.Add)
	if !ok {
		t.Fatalf("result.Data is not *git.Add, got %T", result.Data)
	}

	if len(add.Staged) != 2 {
		t.Errorf("Staged len = %d, want 2", len(add.Staged))
	}
}

func TestAddParser_Schema(t *testing.T) {
	parser := git.NewAddParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema().ID is empty")
	}
	if schema.Title == "" {
		t.Error("Schema().Title is empty")
	}
}

func TestAddParser_Matches(t *testing.T) {
	parser := git.NewAddParser()

	tests := []struct {
		cmd         string
		subcommands []string
		want        bool
	}{
		{"git", []string{"add"}, true},
		{"git", []string{"add", "."}, true},
		{"git", []string{"add", "-v", "file.go"}, true},
		{"git", []string{"status"}, false},
		{"git", []string{}, false},
		{"docker", []string{"add"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.cmd+"_"+strings.Join(tt.subcommands, "_"), func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}
