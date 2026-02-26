package git_test

import (
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/git"
)

const (
	branchMain        = "main"
	upstreamMain      = "origin/main"
	upstreamFeature   = "origin/feature"
	detachedHead      = "(HEAD detached at abc123)"
	branchFeatureAuth = "feature/auth"
	testCommitAbc1234 = "abc1234"
)

func TestBranchListType(t *testing.T) {
	// Test that BranchList type exists and can be constructed
	upstream := upstreamMain
	result := git.BranchList{
		Branches: []git.Branch{
			{Name: branchMain, Current: true, Upstream: &upstream, Ahead: 2, Behind: 1},
			{Name: "feature", Current: false},
		},
		Current: branchMain,
	}

	if len(result.Branches) != 2 {
		t.Errorf("expected 2 branches, got %d", len(result.Branches))
	}
	if result.Current != branchMain {
		t.Errorf("expected current branch %q, got %q", branchMain, result.Current)
	}
	if result.Branches[0].Name != branchMain {
		t.Errorf("expected first branch name %q, got %q", branchMain, result.Branches[0].Name)
	}
	if !result.Branches[0].Current {
		t.Error("expected first branch to be current")
	}
	if *result.Branches[0].Upstream != upstreamMain {
		t.Errorf("expected upstream %q, got %q", upstreamMain, *result.Branches[0].Upstream)
	}
}

func TestBranchParser_BasicList(t *testing.T) {
	input := `* main
  feature/auth
  bugfix/login
`
	parser := git.NewBranchParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	branches, ok := result.Data.(*git.BranchList)
	if !ok {
		t.Fatalf("result.Data is not *git.BranchList, got %T", result.Data)
	}

	if len(branches.Branches) != 3 {
		t.Errorf("expected 3 branches, got %d", len(branches.Branches))
	}
	if branches.Current != branchMain {
		t.Errorf("expected current branch %q, got %q", branchMain, branches.Current)
	}
	if !branches.Branches[0].Current {
		t.Error("expected first branch to be current")
	}
	if branches.Branches[0].Name != branchMain {
		t.Errorf("expected first branch name %q, got %q", branchMain, branches.Branches[0].Name)
	}
	if branches.Branches[1].Name != branchFeatureAuth {
		t.Errorf("expected second branch name %q, got %q", branchFeatureAuth, branches.Branches[1].Name)
	}
	if branches.Branches[1].Current {
		t.Error("expected second branch to not be current")
	}
}

func TestBranchParser_Verbose(t *testing.T) {
	input := `* main                  abc1234 [origin/main: ahead 2]
  feature              def5678 [origin/feature: behind 1]
`
	parser := git.NewBranchParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	branches, ok := result.Data.(*git.BranchList)
	if !ok {
		t.Fatalf("result.Data is not *git.BranchList, got %T", result.Data)
	}

	if len(branches.Branches) != 2 {
		t.Fatalf("expected 2 branches, got %d", len(branches.Branches))
	}

	// First branch - main with upstream tracking
	if branches.Branches[0].Upstream == nil {
		t.Fatal("expected first branch to have upstream")
	}
	if *branches.Branches[0].Upstream != upstreamMain {
		t.Errorf("expected upstream %q, got %q", upstreamMain, *branches.Branches[0].Upstream)
	}
	if branches.Branches[0].Ahead != 2 {
		t.Errorf("expected ahead 2, got %d", branches.Branches[0].Ahead)
	}
	if branches.Branches[0].Behind != 0 {
		t.Errorf("expected behind 0, got %d", branches.Branches[0].Behind)
	}

	// Second branch - feature with behind count
	if branches.Branches[1].Upstream == nil {
		t.Fatal("expected second branch to have upstream")
	}
	if *branches.Branches[1].Upstream != upstreamFeature {
		t.Errorf("expected upstream %q, got %q", upstreamFeature, *branches.Branches[1].Upstream)
	}
	if branches.Branches[1].Ahead != 0 {
		t.Errorf("expected ahead 0, got %d", branches.Branches[1].Ahead)
	}
	if branches.Branches[1].Behind != 1 {
		t.Errorf("expected behind 1, got %d", branches.Branches[1].Behind)
	}
}

func TestBranchParser_AheadAndBehind(t *testing.T) {
	input := `* main [origin/main: ahead 2, behind 1]
`
	parser := git.NewBranchParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	branches, ok := result.Data.(*git.BranchList)
	if !ok {
		t.Fatalf("result.Data is not *git.BranchList, got %T", result.Data)
	}

	if len(branches.Branches) != 1 {
		t.Fatalf("expected 1 branch, got %d", len(branches.Branches))
	}

	if branches.Branches[0].Ahead != 2 {
		t.Errorf("expected ahead 2, got %d", branches.Branches[0].Ahead)
	}
	if branches.Branches[0].Behind != 1 {
		t.Errorf("expected behind 1, got %d", branches.Branches[0].Behind)
	}
}

func TestBranchParser_DetachedHead(t *testing.T) {
	input := `* (HEAD detached at abc123)
  main
`
	parser := git.NewBranchParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	branches, ok := result.Data.(*git.BranchList)
	if !ok {
		t.Fatalf("result.Data is not *git.BranchList, got %T", result.Data)
	}

	if branches.Current != detachedHead {
		t.Errorf("expected current %q, got %q", detachedHead, branches.Current)
	}
	if len(branches.Branches) != 2 {
		t.Errorf("expected 2 branches, got %d", len(branches.Branches))
	}
}

func TestBranchParser_UpstreamOnly(t *testing.T) {
	input := `* main [origin/main]
`
	parser := git.NewBranchParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	branches, ok := result.Data.(*git.BranchList)
	if !ok {
		t.Fatalf("result.Data is not *git.BranchList, got %T", result.Data)
	}

	if branches.Branches[0].Upstream == nil {
		t.Fatal("expected branch to have upstream")
	}
	if *branches.Branches[0].Upstream != upstreamMain {
		t.Errorf("expected upstream %q, got %q", upstreamMain, *branches.Branches[0].Upstream)
	}
}

func TestBranchParser_Schema(t *testing.T) {
	parser := git.NewBranchParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema().ID is empty")
	}
	if schema.Title == "" {
		t.Error("Schema().Title is empty")
	}
}

func TestBranchParser_Matches(t *testing.T) {
	parser := git.NewBranchParser()

	tests := []struct {
		cmd         string
		subcommands []string
		want        bool
	}{
		{"git", []string{"branch"}, true},
		{"git", []string{"branch", "-v"}, true},
		{"git", []string{"branch", "-vv"}, true},
		{"git", []string{"status"}, false},
		{"git", []string{}, false},
		{"docker", []string{"branch"}, false},
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

func TestBranchParser_EmptyInput(t *testing.T) {
	input := ``
	parser := git.NewBranchParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	branches, ok := result.Data.(*git.BranchList)
	if !ok {
		t.Fatalf("result.Data is not *git.BranchList, got %T", result.Data)
	}

	if len(branches.Branches) != 0 {
		t.Errorf("expected 0 branches, got %d", len(branches.Branches))
	}
	if branches.Current != "" {
		t.Errorf("expected current to be empty, got %q", branches.Current)
	}
}

func TestBranchParser_VerboseWithHash(t *testing.T) {
	input := `* main                  abc1234 Some commit message
  feature/auth          def5678 Another commit
`
	parser := git.NewBranchParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	branches, ok := result.Data.(*git.BranchList)
	if !ok {
		t.Fatalf("result.Data is not *git.BranchList, got %T", result.Data)
	}

	if len(branches.Branches) != 2 {
		t.Errorf("expected 2 branches, got %d", len(branches.Branches))
	}
	if branches.Branches[0].Name != branchMain {
		t.Errorf("expected first branch %q, got %q", branchMain, branches.Branches[0].Name)
	}
	if branches.Branches[0].LastCommit != testCommitAbc1234 {
		t.Errorf("expected lastCommit '%s', got %q", testCommitAbc1234, branches.Branches[0].LastCommit)
	}
}

func TestBranchParser_GoneUpstream(t *testing.T) {
	// When remote branch is deleted, git shows [origin/branch: gone]
	input := `* main [origin/main: gone]
`
	parser := git.NewBranchParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	branches, ok := result.Data.(*git.BranchList)
	if !ok {
		t.Fatalf("result.Data is not *git.BranchList, got %T", result.Data)
	}

	// Upstream should still be set
	if branches.Branches[0].Upstream == nil {
		t.Fatal("expected branch to have upstream")
	}
	if *branches.Branches[0].Upstream != upstreamMain {
		t.Errorf("expected upstream %q, got %q", upstreamMain, *branches.Branches[0].Upstream)
	}
}
