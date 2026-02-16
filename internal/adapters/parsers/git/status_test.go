package git_test

import (
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/git"
)

func TestStatusParser_CleanRepo(t *testing.T) {
	input := `# branch.oid abc123def456
# branch.head main
`
	parser := git.NewStatusParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	status, ok := result.Data.(*git.Status)
	if !ok {
		t.Fatalf("result.Data is not *git.Status, got %T", result.Data)
	}

	if status.Branch != "main" {
		t.Errorf("Branch = %q, want %q", status.Branch, "main")
	}
	if !status.Clean {
		t.Error("Clean = false, want true")
	}
	if len(status.Staged) != 0 {
		t.Errorf("Staged = %v, want empty", status.Staged)
	}
	if len(status.Modified) != 0 {
		t.Errorf("Modified = %v, want empty", status.Modified)
	}
	if len(status.Untracked) != 0 {
		t.Errorf("Untracked = %v, want empty", status.Untracked)
	}
}

func TestStatusParser_StagedFiles(t *testing.T) {
	input := `# branch.oid abc123
# branch.head feature
1 A. N... 000000 100644 100644 0000000 abc1234 src/new.go
1 M. N... 100644 100644 100644 abc1234 def5678 src/modified.go
`
	parser := git.NewStatusParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	status, ok := result.Data.(*git.Status)
	if !ok {
		t.Fatalf("result.Data is not *git.Status, got %T", result.Data)
	}

	if len(status.Staged) != 2 {
		t.Fatalf("Staged len = %d, want 2", len(status.Staged))
	}
	if status.Staged[0].File != "src/new.go" {
		t.Errorf("Staged[0].File = %q, want %q", status.Staged[0].File, "src/new.go")
	}
	if status.Staged[0].Status != "added" {
		t.Errorf("Staged[0].Status = %q, want %q", status.Staged[0].Status, "added")
	}
	if status.Staged[1].File != "src/modified.go" {
		t.Errorf("Staged[1].File = %q, want %q", status.Staged[1].File, "src/modified.go")
	}
	if status.Staged[1].Status != "modified" {
		t.Errorf("Staged[1].Status = %q, want %q", status.Staged[1].Status, "modified")
	}
}

func TestStatusParser_WorktreeChanges(t *testing.T) {
	input := `# branch.oid abc123
# branch.head main
1 .M N... 100644 100644 100644 abc1234 abc1234 README.md
? temp.log
`
	parser := git.NewStatusParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	status, ok := result.Data.(*git.Status)
	if !ok {
		t.Fatalf("result.Data is not *git.Status, got %T", result.Data)
	}

	if !contains(status.Modified, "README.md") {
		t.Errorf("Modified = %v, want to contain %q", status.Modified, "README.md")
	}
	if !contains(status.Untracked, "temp.log") {
		t.Errorf("Untracked = %v, want to contain %q", status.Untracked, "temp.log")
	}
	if status.Clean {
		t.Error("Clean = true, want false")
	}
}

func TestStatusParser_UpstreamTracking(t *testing.T) {
	input := `# branch.oid abc123
# branch.head main
# branch.upstream origin/main
# branch.ab +3 -1
`
	parser := git.NewStatusParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	status, ok := result.Data.(*git.Status)
	if !ok {
		t.Fatalf("result.Data is not *git.Status, got %T", result.Data)
	}

	if status.Upstream == nil {
		t.Fatal("Upstream = nil, want non-nil")
	}
	if *status.Upstream != "origin/main" {
		t.Errorf("Upstream = %q, want %q", *status.Upstream, "origin/main")
	}
	if status.Ahead != 3 {
		t.Errorf("Ahead = %d, want 3", status.Ahead)
	}
	if status.Behind != 1 {
		t.Errorf("Behind = %d, want 1", status.Behind)
	}
}

func TestStatusParser_DeletedFiles(t *testing.T) {
	input := `# branch.oid abc123
# branch.head main
1 D. N... 100644 000000 000000 abc1234 0000000 old-file.txt
1 .D N... 100644 100644 000000 abc1234 abc1234 deleted-unstaged.txt
`
	parser := git.NewStatusParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	status, ok := result.Data.(*git.Status)
	if !ok {
		t.Fatalf("result.Data is not *git.Status, got %T", result.Data)
	}

	if len(status.Staged) != 1 {
		t.Fatalf("Staged len = %d, want 1", len(status.Staged))
	}
	if status.Staged[0].File != "old-file.txt" {
		t.Errorf("Staged[0].File = %q, want %q", status.Staged[0].File, "old-file.txt")
	}
	if status.Staged[0].Status != "deleted" {
		t.Errorf("Staged[0].Status = %q, want %q", status.Staged[0].Status, "deleted")
	}
	if !contains(status.Deleted, "deleted-unstaged.txt") {
		t.Errorf("Deleted = %v, want to contain %q", status.Deleted, "deleted-unstaged.txt")
	}
}

func TestStatusParser_RenamedFiles(t *testing.T) {
	input := `# branch.oid abc123
# branch.head main
2 R. N... 100644 100644 100644 abc1234 abc1234 R100 new-name.txt	old-name.txt
`
	parser := git.NewStatusParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	status, ok := result.Data.(*git.Status)
	if !ok {
		t.Fatalf("result.Data is not *git.Status, got %T", result.Data)
	}

	if len(status.Staged) != 1 {
		t.Fatalf("Staged len = %d, want 1", len(status.Staged))
	}
	if status.Staged[0].File != "new-name.txt" {
		t.Errorf("Staged[0].File = %q, want %q", status.Staged[0].File, "new-name.txt")
	}
	if status.Staged[0].Status != "renamed" {
		t.Errorf("Staged[0].Status = %q, want %q", status.Staged[0].Status, "renamed")
	}
}

func TestStatusParser_MergeConflicts(t *testing.T) {
	input := `# branch.oid abc123
# branch.head main
u UU N... 100644 100644 100644 100644 abc1234 def5678 ghi9012 conflict.txt
`
	parser := git.NewStatusParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	status, ok := result.Data.(*git.Status)
	if !ok {
		t.Fatalf("result.Data is not *git.Status, got %T", result.Data)
	}

	if !contains(status.Conflicts, "conflict.txt") {
		t.Errorf("Conflicts = %v, want to contain %q", status.Conflicts, "conflict.txt")
	}
	if status.Clean {
		t.Error("Clean = true, want false")
	}
}

func TestStatusParser_Schema(t *testing.T) {
	parser := git.NewStatusParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema().ID is empty")
	}
	if schema.Title == "" {
		t.Error("Schema().Title is empty")
	}
}

func TestStatusParser_Matches(t *testing.T) {
	parser := git.NewStatusParser()

	tests := []struct {
		cmd         string
		subcommands []string
		want        bool
	}{
		{"git", []string{"status"}, true},
		{"git", []string{"status", "--porcelain=v2"}, true},
		{"git", []string{"commit"}, false},
		{"git", []string{}, false},
		{"docker", []string{"status"}, false},
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

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
