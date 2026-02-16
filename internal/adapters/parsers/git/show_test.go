package git_test

import (
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/git"
)

func TestGitShowType(t *testing.T) {
	show := git.Show{
		Commit: git.Commit{
			Hash:    "abc123",
			Author:  "John",
			Date:    "2024-01-15",
			Message: "Fix bug",
			Subject: "Fix bug",
		},
		Diff: git.Diff{Files: []git.DiffFile{}},
	}
	if show.Commit.Hash != "abc123" {
		t.Errorf("Commit.Hash = %q, want %q", show.Commit.Hash, "abc123")
	}
}

func TestShowParser_Basic(t *testing.T) {
	input := `commit abc123def456789
Author: John Doe <john@example.com>
Date:   Mon Jan 15 10:30:00 2024 +0000

    Fix critical bug

diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1 +1 @@
-old
+new
`
	parser := git.NewShowParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	show, ok := result.Data.(*git.Show)
	if !ok {
		t.Fatalf("result.Data is not *git.Show, got %T", result.Data)
	}

	if show.Commit.Hash != "abc123def456789" {
		t.Errorf("Commit.Hash = %q, want %q", show.Commit.Hash, "abc123def456789")
	}
	if show.Commit.Author != "John Doe" {
		t.Errorf("Commit.Author = %q, want %q", show.Commit.Author, "John Doe")
	}
	if show.Commit.Subject != "Fix critical bug" {
		t.Errorf("Commit.Subject = %q, want %q", show.Commit.Subject, "Fix critical bug")
	}
	if len(show.Diff.Files) != 1 {
		t.Errorf("Diff.Files len = %d, want 1", len(show.Diff.Files))
	}
}

func TestShowParser_MultilineBody(t *testing.T) {
	input := `commit abc123
Author: John <john@example.com>
Date:   Mon Jan 15 10:30:00 2024

    Subject line

    Body paragraph one.

    Body paragraph two.

diff --git a/file.go b/file.go
--- a/file.go
+++ b/file.go
@@ -1 +1 @@
-old
+new
`
	parser := git.NewShowParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	show, ok := result.Data.(*git.Show)
	if !ok {
		t.Fatalf("result.Data is not *git.Show, got %T", result.Data)
	}

	if show.Commit.Subject != "Subject line" {
		t.Errorf("Commit.Subject = %q, want %q", show.Commit.Subject, "Subject line")
	}
	if !strings.Contains(show.Commit.Body, "paragraph one") {
		t.Errorf("Commit.Body = %q, want to contain %q", show.Commit.Body, "paragraph one")
	}
	if !strings.Contains(show.Commit.Body, "paragraph two") {
		t.Errorf("Commit.Body = %q, want to contain %q", show.Commit.Body, "paragraph two")
	}
}

func TestShowParser_NoDiff(t *testing.T) {
	// Merge commits may have no diff
	input := `commit abc123
Author: John <john@example.com>
Date:   Mon Jan 15 10:30:00 2024

    Merge branch 'feature'

`
	parser := git.NewShowParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	show, ok := result.Data.(*git.Show)
	if !ok {
		t.Fatalf("result.Data is not *git.Show, got %T", result.Data)
	}

	if show.Commit.Subject != "Merge branch 'feature'" {
		t.Errorf("Commit.Subject = %q, want %q", show.Commit.Subject, "Merge branch 'feature'")
	}
	if len(show.Diff.Files) != 0 {
		t.Errorf("Diff.Files len = %d, want 0", len(show.Diff.Files))
	}
}

func TestShowParser_BinaryFile(t *testing.T) {
	input := `commit abc123
Author: John <john@example.com>
Date:   Mon Jan 15 10:30:00 2024

    Add image

diff --git a/image.png b/image.png
new file mode 100644
Binary files /dev/null and b/image.png differ
`
	parser := git.NewShowParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	show, ok := result.Data.(*git.Show)
	if !ok {
		t.Fatalf("result.Data is not *git.Show, got %T", result.Data)
	}

	if len(show.Diff.Files) != 1 {
		t.Fatalf("Diff.Files len = %d, want 1", len(show.Diff.Files))
	}
	if !show.Diff.Files[0].Binary {
		t.Error("Diff.Files[0].Binary = false, want true")
	}
}

func TestShowParser_WithEmail(t *testing.T) {
	input := `commit abc123def
Author: Jane Smith <jane.smith@example.org>
Date:   Tue Feb 20 14:45:30 2024 -0500

    Update README

diff --git a/README.md b/README.md
--- a/README.md
+++ b/README.md
@@ -1 +1,2 @@
 # Project
+Description
`
	parser := git.NewShowParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	show, ok := result.Data.(*git.Show)
	if !ok {
		t.Fatalf("result.Data is not *git.Show, got %T", result.Data)
	}

	if show.Commit.Author != "Jane Smith" {
		t.Errorf("Commit.Author = %q, want %q", show.Commit.Author, "Jane Smith")
	}
	if show.Commit.Email != "jane.smith@example.org" {
		t.Errorf("Commit.Email = %q, want %q", show.Commit.Email, "jane.smith@example.org")
	}
}

func TestShowParser_Schema(t *testing.T) {
	parser := git.NewShowParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema().ID is empty")
	}
	if schema.Title == "" {
		t.Error("Schema().Title is empty")
	}
}

func TestShowParser_Matches(t *testing.T) {
	parser := git.NewShowParser()

	tests := []struct {
		cmd         string
		subcommands []string
		want        bool
	}{
		{"git", []string{"show"}, true},
		{"git", []string{"show", "HEAD"}, true},
		{"git", []string{"show", "abc123"}, true},
		{"git", []string{"status"}, false},
		{"git", []string{"log"}, false},
		{"git", []string{}, false},
		{"docker", []string{"show"}, false},
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
