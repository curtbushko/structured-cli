package git_test

import (
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/git"
)

func TestCommitParser_SuccessfulCommit(t *testing.T) {
	input := `[main abc1234] Add new feature
 2 files changed, 10 insertions(+), 5 deletions(-)
`
	parser := git.NewCommitParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	commit, ok := result.Data.(*git.CommitOutput)
	if !ok {
		t.Fatalf("result.Data is not *git.CommitOutput, got %T", result.Data)
	}

	if commit.Branch != "main" {
		t.Errorf("Branch = %q, want %q", commit.Branch, "main")
	}
	if commit.Hash != "abc1234" {
		t.Errorf("Hash = %q, want %q", commit.Hash, "abc1234")
	}
	if commit.Message != "Add new feature" {
		t.Errorf("Message = %q, want %q", commit.Message, "Add new feature")
	}
	if commit.FilesChanged != 2 {
		t.Errorf("FilesChanged = %d, want 2", commit.FilesChanged)
	}
	if commit.Insertions != 10 {
		t.Errorf("Insertions = %d, want 10", commit.Insertions)
	}
	if commit.Deletions != 5 {
		t.Errorf("Deletions = %d, want 5", commit.Deletions)
	}
}

func TestCommitParser_CommitNoStats(t *testing.T) {
	// Commit with no file changes shown (e.g., --allow-empty)
	input := `[feature/test 123abcd] Empty commit message
`
	parser := git.NewCommitParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	commit, ok := result.Data.(*git.CommitOutput)
	if !ok {
		t.Fatalf("result.Data is not *git.CommitOutput, got %T", result.Data)
	}

	if commit.Branch != "feature/test" {
		t.Errorf("Branch = %q, want %q", commit.Branch, "feature/test")
	}
	if commit.Hash != "123abcd" {
		t.Errorf("Hash = %q, want %q", commit.Hash, "123abcd")
	}
	if commit.Message != "Empty commit message" {
		t.Errorf("Message = %q, want %q", commit.Message, "Empty commit message")
	}
}

func TestCommitParser_NothingToCommit(t *testing.T) {
	input := `On branch main
nothing to commit, working tree clean
`
	parser := git.NewCommitParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	commit, ok := result.Data.(*git.CommitOutput)
	if !ok {
		t.Fatalf("result.Data is not *git.CommitOutput, got %T", result.Data)
	}

	// Empty commit when nothing to commit
	if commit.Hash != "" {
		t.Errorf("Hash = %q, want empty", commit.Hash)
	}
}

func TestCommitParser_RootCommit(t *testing.T) {
	// Initial commit (root commit)
	input := `[main (root-commit) a1b2c3d] Initial commit
 1 file changed, 1 insertion(+)
 create mode 100644 README.md
`
	parser := git.NewCommitParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	commit, ok := result.Data.(*git.CommitOutput)
	if !ok {
		t.Fatalf("result.Data is not *git.CommitOutput, got %T", result.Data)
	}

	if commit.Branch != "main" {
		t.Errorf("Branch = %q, want %q", commit.Branch, "main")
	}
	if commit.Hash != "a1b2c3d" {
		t.Errorf("Hash = %q, want %q", commit.Hash, "a1b2c3d")
	}
	if commit.FilesChanged != 1 {
		t.Errorf("FilesChanged = %d, want 1", commit.FilesChanged)
	}
	if commit.Insertions != 1 {
		t.Errorf("Insertions = %d, want 1", commit.Insertions)
	}
}

func TestCommitParser_InsertionsOnly(t *testing.T) {
	input := `[main def4567] Add new file
 1 file changed, 50 insertions(+)
`
	parser := git.NewCommitParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	commit, ok := result.Data.(*git.CommitOutput)
	if !ok {
		t.Fatalf("result.Data is not *git.CommitOutput, got %T", result.Data)
	}

	if commit.Insertions != 50 {
		t.Errorf("Insertions = %d, want 50", commit.Insertions)
	}
	if commit.Deletions != 0 {
		t.Errorf("Deletions = %d, want 0", commit.Deletions)
	}
}

func TestCommitParser_DeletionsOnly(t *testing.T) {
	input := `[main ghi8901] Remove old file
 1 file changed, 30 deletions(-)
`
	parser := git.NewCommitParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	commit, ok := result.Data.(*git.CommitOutput)
	if !ok {
		t.Fatalf("result.Data is not *git.CommitOutput, got %T", result.Data)
	}

	if commit.Insertions != 0 {
		t.Errorf("Insertions = %d, want 0", commit.Insertions)
	}
	if commit.Deletions != 30 {
		t.Errorf("Deletions = %d, want 30", commit.Deletions)
	}
}

func TestCommitParser_Schema(t *testing.T) {
	parser := git.NewCommitParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema().ID is empty")
	}
	if schema.Title == "" {
		t.Error("Schema().Title is empty")
	}
}

func TestCommitParser_Matches(t *testing.T) {
	parser := git.NewCommitParser()

	tests := []struct {
		cmd         string
		subcommands []string
		want        bool
	}{
		{"git", []string{"commit"}, true},
		{"git", []string{"commit", "-m", "message"}, true},
		{"git", []string{"status"}, false},
		{"git", []string{}, false},
		{"docker", []string{"commit"}, false},
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
