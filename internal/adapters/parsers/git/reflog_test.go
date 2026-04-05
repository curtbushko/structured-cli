package git_test

import (
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/git"
)

func TestReflogParser_BasicReflog(t *testing.T) {
	input := `abc1234 HEAD@{0}: commit: Add new feature
def5678 HEAD@{1}: checkout: moving from main to feature
a1b2c34 HEAD@{2}: commit: Initial commit
`
	parser := git.NewReflogParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	reflog, ok := result.Data.(*git.Reflog)
	if !ok {
		t.Fatalf("result.Data is not *git.Reflog, got %T", result.Data)
	}

	if len(reflog.Entries) != 3 {
		t.Fatalf("Entries len = %d, want 3", len(reflog.Entries))
	}

	// Check first entry
	entry := reflog.Entries[0]
	if entry.Hash != "abc1234" {
		t.Errorf("Hash = %q, want %q", entry.Hash, "abc1234")
	}
	if entry.Index != 0 {
		t.Errorf("Index = %d, want 0", entry.Index)
	}
	if entry.Action != "commit" {
		t.Errorf("Action = %q, want %q", entry.Action, "commit")
	}
	if entry.Message != "Add new feature" {
		t.Errorf("Message = %q, want %q", entry.Message, "Add new feature")
	}

	// Check second entry
	entry2 := reflog.Entries[1]
	if entry2.Action != "checkout" {
		t.Errorf("Action = %q, want %q", entry2.Action, "checkout")
	}
	if entry2.Message != "moving from main to feature" {
		t.Errorf("Message = %q, want %q", entry2.Message, "moving from main to feature")
	}
}

func TestReflogParser_VariousActions(t *testing.T) {
	input := `abc1234 HEAD@{0}: commit: Latest commit
def5678 HEAD@{1}: merge feature: Merge branch 'feature'
a1b2c34 HEAD@{2}: rebase -i (finish): rebasing main
c5d6e78 HEAD@{3}: reset: moving to HEAD~1
f9e8d70 HEAD@{4}: pull: Fast-forward
`
	parser := git.NewReflogParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	reflog, ok := result.Data.(*git.Reflog)
	if !ok {
		t.Fatalf("result.Data is not *git.Reflog, got %T", result.Data)
	}

	if len(reflog.Entries) != 5 {
		t.Fatalf("Entries len = %d, want 5", len(reflog.Entries))
	}

	// Check merge action
	if reflog.Entries[1].Action != "merge feature" {
		t.Errorf("Entry[1].Action = %q, want %q", reflog.Entries[1].Action, "merge feature")
	}

	// Check rebase action
	if reflog.Entries[2].Action != "rebase -i (finish)" {
		t.Errorf("Entry[2].Action = %q, want %q", reflog.Entries[2].Action, "rebase -i (finish)")
	}

	// Check reset action
	if reflog.Entries[3].Action != "reset" {
		t.Errorf("Entry[3].Action = %q, want %q", reflog.Entries[3].Action, "reset")
	}
}

func TestReflogParser_EmptyReflog(t *testing.T) {
	input := ``
	parser := git.NewReflogParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	reflog, ok := result.Data.(*git.Reflog)
	if !ok {
		t.Fatalf("result.Data is not *git.Reflog, got %T", result.Data)
	}

	if len(reflog.Entries) != 0 {
		t.Errorf("Entries = %v, want empty", reflog.Entries)
	}
}

func TestReflogParser_LongHash(t *testing.T) {
	input := `abc123def456789012345678901234567890abcd HEAD@{0}: commit: Test
`
	parser := git.NewReflogParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	reflog, ok := result.Data.(*git.Reflog)
	if !ok {
		t.Fatalf("result.Data is not *git.Reflog, got %T", result.Data)
	}

	if len(reflog.Entries) != 1 {
		t.Fatalf("Entries len = %d, want 1", len(reflog.Entries))
	}

	if reflog.Entries[0].Hash != "abc123def456789012345678901234567890abcd" {
		t.Errorf("Hash = %q, want full hash", reflog.Entries[0].Hash)
	}
}

func TestReflogParser_Schema(t *testing.T) {
	parser := git.NewReflogParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema().ID is empty")
	}
	if schema.Title == "" {
		t.Error("Schema().Title is empty")
	}
}

func TestReflogParser_Matches(t *testing.T) {
	parser := git.NewReflogParser()

	tests := []struct {
		cmd         string
		subcommands []string
		want        bool
	}{
		{"git", []string{"reflog"}, true},
		{"git", []string{"reflog", "show"}, true},
		{"git", []string{"log"}, false},
		{"git", []string{}, false},
		{"docker", []string{"reflog"}, false},
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
