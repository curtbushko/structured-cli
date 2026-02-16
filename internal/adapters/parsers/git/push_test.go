package git_test

import (
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/git"
)

func TestPushParser_ExistingBranch(t *testing.T) {
	input := `To github.com:user/repo.git
   abc123..def456  main -> main
`
	parser := git.NewPushParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	push, ok := result.Data.(*git.Push)
	if !ok {
		t.Fatalf("result.Data is not *git.Push, got %T", result.Data)
	}

	if !push.Success {
		t.Error("Success = false, want true")
	}
	if push.Remote != "github.com:user/repo.git" {
		t.Errorf("Remote = %q, want %q", push.Remote, "github.com:user/repo.git")
	}
	if push.Branch != "main" {
		t.Errorf("Branch = %q, want %q", push.Branch, "main")
	}
	if push.NewBranch {
		t.Error("NewBranch = true, want false")
	}
}

func TestPushParser_NewBranch(t *testing.T) {
	input := `To github.com:user/repo.git
 * [new branch]      feature -> feature
`
	parser := git.NewPushParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	push, ok := result.Data.(*git.Push)
	if !ok {
		t.Fatalf("result.Data is not *git.Push, got %T", result.Data)
	}

	if !push.Success {
		t.Error("Success = false, want true")
	}
	if push.Branch != "feature" {
		t.Errorf("Branch = %q, want %q", push.Branch, "feature")
	}
	if !push.NewBranch {
		t.Error("NewBranch = false, want true")
	}
}

func TestPushParser_Rejected(t *testing.T) {
	input := `To github.com:user/repo.git
 ! [rejected]        main -> main (fetch first)
error: failed to push some refs to 'github.com:user/repo.git'
hint: Updates were rejected because the remote contains work that you do not
hint: have locally. This is usually caused by another repository pushing to
hint: the same ref. If you want to integrate the remote changes, use
hint: 'git pull' before pushing again.
`
	parser := git.NewPushParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	push, ok := result.Data.(*git.Push)
	if !ok {
		t.Fatalf("result.Data is not *git.Push, got %T", result.Data)
	}

	if push.Success {
		t.Error("Success = true, want false")
	}
	if len(push.Errors) == 0 {
		t.Error("Errors is empty, want non-empty")
	}
}

func TestPushParser_AlreadyUpToDate(t *testing.T) {
	input := `Everything up-to-date
`
	parser := git.NewPushParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	push, ok := result.Data.(*git.Push)
	if !ok {
		t.Fatalf("result.Data is not *git.Push, got %T", result.Data)
	}

	if !push.Success {
		t.Error("Success = false, want true")
	}
}

func TestPushParser_ForcePush(t *testing.T) {
	input := `To github.com:user/repo.git
 + abc123...def456 main -> main (forced update)
`
	parser := git.NewPushParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	push, ok := result.Data.(*git.Push)
	if !ok {
		t.Fatalf("result.Data is not *git.Push, got %T", result.Data)
	}

	if !push.Success {
		t.Error("Success = false, want true")
	}
	if push.Branch != "main" {
		t.Errorf("Branch = %q, want %q", push.Branch, "main")
	}
}

func TestPushParser_Schema(t *testing.T) {
	parser := git.NewPushParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema().ID is empty")
	}
	if schema.Title == "" {
		t.Error("Schema().Title is empty")
	}
}

func TestPushParser_Matches(t *testing.T) {
	parser := git.NewPushParser()

	tests := []struct {
		cmd         string
		subcommands []string
		want        bool
	}{
		{"git", []string{"push"}, true},
		{"git", []string{"push", "origin", "main"}, true},
		{"git", []string{"pull"}, false},
		{"git", []string{}, false},
		{"docker", []string{"push"}, false},
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
