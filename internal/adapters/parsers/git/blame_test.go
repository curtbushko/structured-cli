package git_test

import (
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/git"
)

const (
	testHash   = "abc123def456789012345678901234567890abcd"
	testAuthor = "John Doe"
)

func TestBlameParser_SingleAuthor(t *testing.T) {
	input := `abc123def456789012345678901234567890abcd 1 1 1
author John Doe
author-mail <john@example.com>
author-time 1609459200
author-tz +0000
committer John Doe
committer-mail <john@example.com>
committer-time 1609459200
committer-tz +0000
summary Initial commit
filename main.go
	package main
`
	parser := git.NewBlameParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	blame, ok := result.Data.(*git.Blame)
	if !ok {
		t.Fatalf("result.Data is not *git.Blame, got %T", result.Data)
	}

	if len(blame.Lines) != 1 {
		t.Fatalf("Lines len = %d, want 1", len(blame.Lines))
	}

	line := blame.Lines[0]
	if line.LineNumber != 1 {
		t.Errorf("LineNumber = %d, want 1", line.LineNumber)
	}
	if line.Hash != testHash {
		t.Errorf("Hash = %q, want %q", line.Hash, testHash)
	}
	if line.Author != testAuthor {
		t.Errorf("Author = %q, want %q", line.Author, testAuthor)
	}
	if line.Content != "package main" {
		t.Errorf("Content = %q, want %q", line.Content, "package main")
	}
}

func TestBlameParser_MultipleAuthors(t *testing.T) {
	input := `abc123def456789012345678901234567890abcd 1 1 1
author Alice
author-mail <alice@example.com>
author-time 1609459200
author-tz +0000
committer Alice
committer-mail <alice@example.com>
committer-time 1609459200
committer-tz +0000
summary First commit
filename main.go
	package main
def456789012345678901234567890abcdef5678 2 2 1
author Bob
author-mail <bob@example.com>
author-time 1609545600
author-tz +0000
committer Bob
committer-mail <bob@example.com>
committer-time 1609545600
committer-tz +0000
summary Second commit
filename main.go
	import "fmt"
`
	parser := git.NewBlameParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	blame, ok := result.Data.(*git.Blame)
	if !ok {
		t.Fatalf("result.Data is not *git.Blame, got %T", result.Data)
	}

	if len(blame.Lines) != 2 {
		t.Fatalf("Lines len = %d, want 2", len(blame.Lines))
	}

	if blame.Lines[0].Author != "Alice" {
		t.Errorf("Lines[0].Author = %q, want %q", blame.Lines[0].Author, "Alice")
	}
	if blame.Lines[1].Author != "Bob" {
		t.Errorf("Lines[1].Author = %q, want %q", blame.Lines[1].Author, "Bob")
	}
}

func TestBlameParser_WithFilename(t *testing.T) {
	input := `abc123def456789012345678901234567890abcd 1 1 1
author Test User
author-mail <test@example.com>
author-time 1609459200
author-tz +0000
committer Test User
committer-mail <test@example.com>
committer-time 1609459200
committer-tz +0000
summary Test
filename src/app.go
	func main() {}
`
	parser := git.NewBlameParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	blame, ok := result.Data.(*git.Blame)
	if !ok {
		t.Fatalf("result.Data is not *git.Blame, got %T", result.Data)
	}

	if blame.File != "src/app.go" {
		t.Errorf("File = %q, want %q", blame.File, "src/app.go")
	}
}

func TestBlameParser_UncommittedChanges(t *testing.T) {
	// Uncommitted changes show as 0000... hash
	input := `0000000000000000000000000000000000000000 1 1 1
author Not Committed Yet
author-mail <not.committed.yet>
author-time 0
author-tz +0000
committer Not Committed Yet
committer-mail <not.committed.yet>
committer-time 0
committer-tz +0000
summary Not yet committed
filename test.go
	// new line
`
	parser := git.NewBlameParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	blame, ok := result.Data.(*git.Blame)
	if !ok {
		t.Fatalf("result.Data is not *git.Blame, got %T", result.Data)
	}

	if len(blame.Lines) != 1 {
		t.Fatalf("Lines len = %d, want 1", len(blame.Lines))
	}
	if blame.Lines[0].Author != "Not Committed Yet" {
		t.Errorf("Author = %q, want %q", blame.Lines[0].Author, "Not Committed Yet")
	}
}

func TestBlameParser_Schema(t *testing.T) {
	parser := git.NewBlameParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema().ID is empty")
	}
	if schema.Title == "" {
		t.Error("Schema().Title is empty")
	}
}

func TestBlameParser_Matches(t *testing.T) {
	parser := git.NewBlameParser()

	tests := []struct {
		cmd         string
		subcommands []string
		want        bool
	}{
		{"git", []string{"blame"}, true},
		{"git", []string{"blame", "file.go"}, true},
		{"git", []string{"blame", "--porcelain", "file.go"}, true},
		{"git", []string{"log"}, false},
		{"git", []string{}, false},
		{"docker", []string{"blame"}, false},
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
