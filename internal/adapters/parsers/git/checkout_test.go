package git_test

import (
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/git"
)

func TestCheckoutParser_ExistingBranch(t *testing.T) {
	input := `Switched to branch 'main'
`
	parser := git.NewCheckoutParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	checkout, ok := result.Data.(*git.Checkout)
	if !ok {
		t.Fatalf("result.Data is not *git.Checkout, got %T", result.Data)
	}

	if !checkout.Success {
		t.Error("Success = false, want true")
	}
	if checkout.Branch != "main" {
		t.Errorf("Branch = %q, want %q", checkout.Branch, "main")
	}
	if checkout.NewBranch {
		t.Error("NewBranch = true, want false")
	}
}

func TestCheckoutParser_NewBranch(t *testing.T) {
	input := `Switched to a new branch 'feature/test'
`
	parser := git.NewCheckoutParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	checkout, ok := result.Data.(*git.Checkout)
	if !ok {
		t.Fatalf("result.Data is not *git.Checkout, got %T", result.Data)
	}

	if !checkout.Success {
		t.Error("Success = false, want true")
	}
	if checkout.Branch != "feature/test" {
		t.Errorf("Branch = %q, want %q", checkout.Branch, "feature/test")
	}
	if !checkout.NewBranch {
		t.Error("NewBranch = false, want true")
	}
}

func TestCheckoutParser_DetachedHead(t *testing.T) {
	input := `Note: switching to 'abc123'.

You are in 'detached HEAD' state. You can look around, make experimental
changes and commit them, and you can discard any commits you make in this
state without impacting any branches by switching back to a branch.

HEAD is now at abc123 Some commit message
`
	parser := git.NewCheckoutParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	checkout, ok := result.Data.(*git.Checkout)
	if !ok {
		t.Fatalf("result.Data is not *git.Checkout, got %T", result.Data)
	}

	if !checkout.Success {
		t.Error("Success = false, want true")
	}
	if checkout.Commit != "abc123" {
		t.Errorf("Commit = %q, want %q", checkout.Commit, "abc123")
	}
}

func TestCheckoutParser_AlreadyOnBranch(t *testing.T) {
	input := `Already on 'main'
Your branch is up to date with 'origin/main'.
`
	parser := git.NewCheckoutParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	checkout, ok := result.Data.(*git.Checkout)
	if !ok {
		t.Fatalf("result.Data is not *git.Checkout, got %T", result.Data)
	}

	if !checkout.Success {
		t.Error("Success = false, want true")
	}
	if checkout.Branch != "main" {
		t.Errorf("Branch = %q, want %q", checkout.Branch, "main")
	}
}

func TestCheckoutParser_HeadIsNow(t *testing.T) {
	input := `Previous HEAD position was abc123 Old commit
HEAD is now at def456 New commit
`
	parser := git.NewCheckoutParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	checkout, ok := result.Data.(*git.Checkout)
	if !ok {
		t.Fatalf("result.Data is not *git.Checkout, got %T", result.Data)
	}

	if checkout.Commit != "def456" {
		t.Errorf("Commit = %q, want %q", checkout.Commit, "def456")
	}
}

func TestCheckoutParser_Schema(t *testing.T) {
	parser := git.NewCheckoutParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema().ID is empty")
	}
	if schema.Title == "" {
		t.Error("Schema().Title is empty")
	}
}

func TestCheckoutParser_Matches(t *testing.T) {
	parser := git.NewCheckoutParser()

	tests := []struct {
		cmd         string
		subcommands []string
		want        bool
	}{
		{"git", []string{"checkout"}, true},
		{"git", []string{"checkout", "main"}, true},
		{"git", []string{"checkout", "-b", "feature"}, true},
		{"git", []string{"branch"}, false},
		{"git", []string{}, false},
		{"docker", []string{"checkout"}, false},
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
