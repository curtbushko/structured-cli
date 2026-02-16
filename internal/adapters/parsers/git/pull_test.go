package git_test

import (
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/git"
)

func TestPullParser_FastForward(t *testing.T) {
	input := `Updating abc123..def456
Fast-forward
 file.go | 10 +++++-----
 1 file changed, 5 insertions(+), 5 deletions(-)
`
	parser := git.NewPullParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	pull, ok := result.Data.(*git.Pull)
	if !ok {
		t.Fatalf("result.Data is not *git.Pull, got %T", result.Data)
	}

	if !pull.Success {
		t.Error("Success = false, want true")
	}
	if !pull.FastForward {
		t.Error("FastForward = false, want true")
	}
	if pull.FilesChanged != 1 {
		t.Errorf("FilesChanged = %d, want 1", pull.FilesChanged)
	}
	if pull.Insertions != 5 {
		t.Errorf("Insertions = %d, want 5", pull.Insertions)
	}
	if pull.Deletions != 5 {
		t.Errorf("Deletions = %d, want 5", pull.Deletions)
	}
}

func TestPullParser_AlreadyUpToDate(t *testing.T) {
	input := `Already up to date.
`
	parser := git.NewPullParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	pull, ok := result.Data.(*git.Pull)
	if !ok {
		t.Fatalf("result.Data is not *git.Pull, got %T", result.Data)
	}

	if !pull.Success {
		t.Error("Success = false, want true")
	}
	if pull.FilesChanged != 0 {
		t.Errorf("FilesChanged = %d, want 0", pull.FilesChanged)
	}
}

func TestPullParser_MergeCommit(t *testing.T) {
	input := `Updating abc123..def456
Merge made by the 'ort' strategy.
 file1.go | 5 +++++
 file2.go | 3 ---
 2 files changed, 5 insertions(+), 3 deletions(-)
`
	parser := git.NewPullParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	pull, ok := result.Data.(*git.Pull)
	if !ok {
		t.Fatalf("result.Data is not *git.Pull, got %T", result.Data)
	}

	if !pull.Success {
		t.Error("Success = false, want true")
	}
	if pull.FastForward {
		t.Error("FastForward = true, want false")
	}
	if pull.FilesChanged != 2 {
		t.Errorf("FilesChanged = %d, want 2", pull.FilesChanged)
	}
	if pull.Insertions != 5 {
		t.Errorf("Insertions = %d, want 5", pull.Insertions)
	}
	if pull.Deletions != 3 {
		t.Errorf("Deletions = %d, want 3", pull.Deletions)
	}
}

func TestPullParser_Conflicts(t *testing.T) {
	input := `Auto-merging file.go
CONFLICT (content): Merge conflict in file.go
Automatic merge failed; fix conflicts and then commit the result.
`
	parser := git.NewPullParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	pull, ok := result.Data.(*git.Pull)
	if !ok {
		t.Fatalf("result.Data is not *git.Pull, got %T", result.Data)
	}

	if pull.Success {
		t.Error("Success = true, want false")
	}
	if len(pull.Conflicts) != 1 {
		t.Fatalf("Conflicts len = %d, want 1", len(pull.Conflicts))
	}
	if pull.Conflicts[0] != "file.go" {
		t.Errorf("Conflicts[0] = %q, want %q", pull.Conflicts[0], "file.go")
	}
}

func TestPullParser_MultipleConflicts(t *testing.T) {
	input := `Auto-merging a.go
CONFLICT (content): Merge conflict in a.go
Auto-merging b.go
CONFLICT (content): Merge conflict in b.go
Automatic merge failed; fix conflicts and then commit the result.
`
	parser := git.NewPullParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	pull, ok := result.Data.(*git.Pull)
	if !ok {
		t.Fatalf("result.Data is not *git.Pull, got %T", result.Data)
	}

	if pull.Success {
		t.Error("Success = true, want false")
	}
	if len(pull.Conflicts) != 2 {
		t.Fatalf("Conflicts len = %d, want 2", len(pull.Conflicts))
	}
}

func TestPullParser_Schema(t *testing.T) {
	parser := git.NewPullParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema().ID is empty")
	}
	if schema.Title == "" {
		t.Error("Schema().Title is empty")
	}
}

func TestPullParser_Matches(t *testing.T) {
	parser := git.NewPullParser()

	tests := []struct {
		cmd         string
		subcommands []string
		want        bool
	}{
		{"git", []string{"pull"}, true},
		{"git", []string{"pull", "origin", "main"}, true},
		{"git", []string{"push"}, false},
		{"git", []string{}, false},
		{"docker", []string{"pull"}, false},
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
