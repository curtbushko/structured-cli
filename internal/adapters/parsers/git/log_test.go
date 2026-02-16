package git_test

import (
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/git"
)

func TestLogParser_EmptyLog(t *testing.T) {
	input := ""
	parser := git.NewLogParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	log, ok := result.Data.(*git.Log)
	if !ok {
		t.Fatalf("result.Data is not *git.Log, got %T", result.Data)
	}

	if len(log.Commits) != 0 {
		t.Errorf("Commits len = %d, want 0", len(log.Commits))
	}
}

func TestLogParser_SingleCommit(t *testing.T) {
	input := `COMMIT_START
abc123def456789012345678901234567890abcd
abc123d
John Doe
john@example.com
2024-01-15T10:30:00+00:00
Fix critical bug in parser

COMMIT_END
`
	parser := git.NewLogParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	log, ok := result.Data.(*git.Log)
	if !ok {
		t.Fatalf("result.Data is not *git.Log, got %T", result.Data)
	}

	if len(log.Commits) != 1 {
		t.Fatalf("Commits len = %d, want 1", len(log.Commits))
	}

	commit := log.Commits[0]
	if commit.Hash != "abc123def456789012345678901234567890abcd" {
		t.Errorf("Hash = %q, want %q", commit.Hash, "abc123def456789012345678901234567890abcd")
	}
	if commit.AbbrevHash != "abc123d" {
		t.Errorf("AbbrevHash = %q, want %q", commit.AbbrevHash, "abc123d")
	}
	if commit.Author != "John Doe" {
		t.Errorf("Author = %q, want %q", commit.Author, "John Doe")
	}
	if commit.Email != "john@example.com" {
		t.Errorf("Email = %q, want %q", commit.Email, "john@example.com")
	}
	if commit.Date != "2024-01-15T10:30:00+00:00" {
		t.Errorf("Date = %q, want %q", commit.Date, "2024-01-15T10:30:00+00:00")
	}
	if commit.Subject != "Fix critical bug in parser" {
		t.Errorf("Subject = %q, want %q", commit.Subject, "Fix critical bug in parser")
	}
}

func TestLogParser_MultipleCommits(t *testing.T) {
	input := `COMMIT_START
abc123def456
abc123
John
john@example.com
2024-01-15
First commit

COMMIT_END
COMMIT_START
def456abc789
def456
Jane
jane@example.com
2024-01-14
Second commit

COMMIT_END
`
	parser := git.NewLogParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	log, ok := result.Data.(*git.Log)
	if !ok {
		t.Fatalf("result.Data is not *git.Log, got %T", result.Data)
	}

	if len(log.Commits) != 2 {
		t.Fatalf("Commits len = %d, want 2", len(log.Commits))
	}
	if log.Commits[0].Subject != "First commit" {
		t.Errorf("Commits[0].Subject = %q, want %q", log.Commits[0].Subject, "First commit")
	}
	if log.Commits[1].Subject != "Second commit" {
		t.Errorf("Commits[1].Subject = %q, want %q", log.Commits[1].Subject, "Second commit")
	}
}

func TestLogParser_CommitWithBody(t *testing.T) {
	input := `COMMIT_START
abc123def456
abc
John
john@example.com
2024-01-15
Add feature X

This is a longer description
that spans multiple lines.

Fixes #123
COMMIT_END
`
	parser := git.NewLogParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	log, ok := result.Data.(*git.Log)
	if !ok {
		t.Fatalf("result.Data is not *git.Log, got %T", result.Data)
	}

	if len(log.Commits) != 1 {
		t.Fatalf("Commits len = %d, want 1", len(log.Commits))
	}

	commit := log.Commits[0]
	if commit.Subject != "Add feature X" {
		t.Errorf("Subject = %q, want %q", commit.Subject, "Add feature X")
	}
	if !strings.Contains(commit.Body, "longer description") {
		t.Errorf("Body = %q, want to contain %q", commit.Body, "longer description")
	}
	if !strings.Contains(commit.Body, "Fixes #123") {
		t.Errorf("Body = %q, want to contain %q", commit.Body, "Fixes #123")
	}
}

func TestLogParser_WithStats(t *testing.T) {
	input := `COMMIT_START
abc123def456
abc
John
john@example.com
2024-01-15
Update files

COMMIT_END
10	5	src/main.go
3	0	README.md
`
	parser := git.NewLogParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	log, ok := result.Data.(*git.Log)
	if !ok {
		t.Fatalf("result.Data is not *git.Log, got %T", result.Data)
	}

	if len(log.Commits) != 1 {
		t.Fatalf("Commits len = %d, want 1", len(log.Commits))
	}

	commit := log.Commits[0]
	if len(commit.Files) != 2 {
		t.Fatalf("Files len = %d, want 2", len(commit.Files))
	}
	if commit.Files[0].Path != "src/main.go" {
		t.Errorf("Files[0].Path = %q, want %q", commit.Files[0].Path, "src/main.go")
	}
	if commit.Files[0].Additions != 10 {
		t.Errorf("Files[0].Additions = %d, want 10", commit.Files[0].Additions)
	}
	if commit.Files[0].Deletions != 5 {
		t.Errorf("Files[0].Deletions = %d, want 5", commit.Files[0].Deletions)
	}
	if commit.Files[1].Path != "README.md" {
		t.Errorf("Files[1].Path = %q, want %q", commit.Files[1].Path, "README.md")
	}
	if commit.Files[1].Additions != 3 {
		t.Errorf("Files[1].Additions = %d, want 3", commit.Files[1].Additions)
	}
	if commit.Files[1].Deletions != 0 {
		t.Errorf("Files[1].Deletions = %d, want 0", commit.Files[1].Deletions)
	}
}

func TestLogParser_BinaryFileInStats(t *testing.T) {
	input := `COMMIT_START
abc123def456
abc
John
john@example.com
2024-01-15
Add image

COMMIT_END
-	-	image.png
10	5	src/main.go
`
	parser := git.NewLogParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	log, ok := result.Data.(*git.Log)
	if !ok {
		t.Fatalf("result.Data is not *git.Log, got %T", result.Data)
	}

	if len(log.Commits) != 1 {
		t.Fatalf("Commits len = %d, want 1", len(log.Commits))
	}

	commit := log.Commits[0]
	if len(commit.Files) != 2 {
		t.Fatalf("Files len = %d, want 2", len(commit.Files))
	}

	// Binary file should have -1 for additions/deletions to indicate binary
	if commit.Files[0].Path != "image.png" {
		t.Errorf("Files[0].Path = %q, want %q", commit.Files[0].Path, "image.png")
	}
	if commit.Files[0].Additions != -1 {
		t.Errorf("Files[0].Additions = %d, want -1 (binary)", commit.Files[0].Additions)
	}
	if commit.Files[0].Deletions != -1 {
		t.Errorf("Files[0].Deletions = %d, want -1 (binary)", commit.Files[0].Deletions)
	}
}

func TestLogParser_MultipleCommitsWithStats(t *testing.T) {
	input := `COMMIT_START
abc123def456
abc
John
john@example.com
2024-01-15
First commit

COMMIT_END
5	2	file1.go
COMMIT_START
def456abc789
def
Jane
jane@example.com
2024-01-14
Second commit

COMMIT_END
3	1	file2.go
`
	parser := git.NewLogParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	log, ok := result.Data.(*git.Log)
	if !ok {
		t.Fatalf("result.Data is not *git.Log, got %T", result.Data)
	}

	if len(log.Commits) != 2 {
		t.Fatalf("Commits len = %d, want 2", len(log.Commits))
	}

	if len(log.Commits[0].Files) != 1 {
		t.Fatalf("Commits[0].Files len = %d, want 1", len(log.Commits[0].Files))
	}
	if log.Commits[0].Files[0].Path != "file1.go" {
		t.Errorf("Commits[0].Files[0].Path = %q, want %q", log.Commits[0].Files[0].Path, "file1.go")
	}

	if len(log.Commits[1].Files) != 1 {
		t.Fatalf("Commits[1].Files len = %d, want 1", len(log.Commits[1].Files))
	}
	if log.Commits[1].Files[0].Path != "file2.go" {
		t.Errorf("Commits[1].Files[0].Path = %q, want %q", log.Commits[1].Files[0].Path, "file2.go")
	}
}

func TestLogParser_MessageField(t *testing.T) {
	input := `COMMIT_START
abc123def456
abc
John
john@example.com
2024-01-15
Add feature

Body text here
COMMIT_END
`
	parser := git.NewLogParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	log, ok := result.Data.(*git.Log)
	if !ok {
		t.Fatalf("result.Data is not *git.Log, got %T", result.Data)
	}

	commit := log.Commits[0]
	// Message should be subject + body
	if !strings.Contains(commit.Message, "Add feature") {
		t.Errorf("Message = %q, want to contain subject", commit.Message)
	}
	if !strings.Contains(commit.Message, "Body text here") {
		t.Errorf("Message = %q, want to contain body", commit.Message)
	}
}

func TestLogParser_InsertionsDeletionsTotals(t *testing.T) {
	input := `COMMIT_START
abc123def456
abc
John
john@example.com
2024-01-15
Update files

COMMIT_END
10	5	src/main.go
3	2	README.md
`
	parser := git.NewLogParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	log, ok := result.Data.(*git.Log)
	if !ok {
		t.Fatalf("result.Data is not *git.Log, got %T", result.Data)
	}

	commit := log.Commits[0]
	// Total insertions: 10 + 3 = 13
	if commit.Insertions != 13 {
		t.Errorf("Insertions = %d, want 13", commit.Insertions)
	}
	// Total deletions: 5 + 2 = 7
	if commit.Deletions != 7 {
		t.Errorf("Deletions = %d, want 7", commit.Deletions)
	}
}

func TestLogParser_Schema(t *testing.T) {
	parser := git.NewLogParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema().ID is empty")
	}
	if schema.Title == "" {
		t.Error("Schema().Title is empty")
	}
}

func TestLogParser_Matches(t *testing.T) {
	parser := git.NewLogParser()

	tests := []struct {
		cmd         string
		subcommands []string
		want        bool
	}{
		{"git", []string{"log"}, true},
		{"git", []string{"log", "--oneline"}, true},
		{"git", []string{"log", "-n", "10"}, true},
		{"git", []string{"status"}, false},
		{"git", []string{}, false},
		{"docker", []string{"log"}, false},
	}

	for _, tt := range tests {
		name := tt.cmd + "_" + strings.Join(tt.subcommands, "_")
		if name == tt.cmd+"_" {
			name = tt.cmd + "_empty"
		}
		t.Run(name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}
