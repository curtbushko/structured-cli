package git

import (
	"strings"
	"testing"
)

func TestDiffParser_SingleFile(t *testing.T) {
	input := `diff --git a/README.md b/README.md
index abc123..def456 100644
--- a/README.md
+++ b/README.md
@@ -1,3 +1,4 @@
 # Title
+New line
 Content
`
	parser := NewDiffParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	diff, ok := result.Data.(*Diff)
	if !ok {
		t.Fatalf("expected *Diff, got %T", result.Data)
	}

	if len(diff.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(diff.Files))
	}

	if diff.Files[0].Path != "README.md" {
		t.Errorf("expected path README.md, got %s", diff.Files[0].Path)
	}

	if diff.Files[0].Status != "modified" {
		t.Errorf("expected status modified, got %s", diff.Files[0].Status)
	}

	if diff.Files[0].Additions != 1 {
		t.Errorf("expected 1 addition, got %d", diff.Files[0].Additions)
	}

	if diff.Files[0].Deletions != 0 {
		t.Errorf("expected 0 deletions, got %d", diff.Files[0].Deletions)
	}
}

func TestDiffParser_Hunks(t *testing.T) {
	input := `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -10,5 +10,6 @@ func main() {
 context
-old
+new
+another
`
	parser := NewDiffParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	diff, ok := result.Data.(*Diff)
	if !ok {
		t.Fatalf("expected *Diff, got %T", result.Data)
	}

	if len(diff.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(diff.Files))
	}

	if len(diff.Files[0].Hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(diff.Files[0].Hunks))
	}

	hunk := diff.Files[0].Hunks[0]
	if hunk.OldStart != 10 {
		t.Errorf("expected OldStart 10, got %d", hunk.OldStart)
	}

	if hunk.OldLines != 5 {
		t.Errorf("expected OldLines 5, got %d", hunk.OldLines)
	}

	if hunk.NewStart != 10 {
		t.Errorf("expected NewStart 10, got %d", hunk.NewStart)
	}

	if hunk.NewLines != 6 {
		t.Errorf("expected NewLines 6, got %d", hunk.NewLines)
	}

	if len(hunk.Lines) != 4 {
		t.Errorf("expected 4 lines, got %d", len(hunk.Lines))
	}
}

func TestDiffParser_MultipleFiles(t *testing.T) {
	input := `diff --git a/file1.go b/file1.go
--- a/file1.go
+++ b/file1.go
@@ -1 +1 @@
-old
+new
diff --git a/file2.go b/file2.go
--- a/file2.go
+++ b/file2.go
@@ -1 +1 @@
-foo
+bar
`
	parser := NewDiffParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	diff, ok := result.Data.(*Diff)
	if !ok {
		t.Fatalf("expected *Diff, got %T", result.Data)
	}

	if len(diff.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(diff.Files))
	}
}

func TestDiffParser_NewFile(t *testing.T) {
	input := `diff --git a/newfile.go b/newfile.go
new file mode 100644
index 0000000..abc1234
--- /dev/null
+++ b/newfile.go
@@ -0,0 +1,3 @@
+package main
+
+func init() {}
`
	parser := NewDiffParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	diff, ok := result.Data.(*Diff)
	if !ok {
		t.Fatalf("expected *Diff, got %T", result.Data)
	}

	if len(diff.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(diff.Files))
	}

	if diff.Files[0].Status != "added" {
		t.Errorf("expected status added, got %s", diff.Files[0].Status)
	}
}

func TestDiffParser_DeletedFile(t *testing.T) {
	input := `diff --git a/oldfile.go b/oldfile.go
deleted file mode 100644
index abc1234..0000000
--- a/oldfile.go
+++ /dev/null
@@ -1,3 +0,0 @@
-package main
-
-func gone() {}
`
	parser := NewDiffParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	diff, ok := result.Data.(*Diff)
	if !ok {
		t.Fatalf("expected *Diff, got %T", result.Data)
	}

	if len(diff.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(diff.Files))
	}

	if diff.Files[0].Status != "deleted" {
		t.Errorf("expected status deleted, got %s", diff.Files[0].Status)
	}
}

func TestDiffParser_BinaryFile(t *testing.T) {
	input := `diff --git a/image.png b/image.png
Binary files a/image.png and b/image.png differ
`
	parser := NewDiffParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	diff, ok := result.Data.(*Diff)
	if !ok {
		t.Fatalf("expected *Diff, got %T", result.Data)
	}

	if len(diff.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(diff.Files))
	}

	if !diff.Files[0].Binary {
		t.Errorf("expected binary true, got false")
	}
}

func TestDiffParser_RenamedFile(t *testing.T) {
	input := `diff --git a/old.go b/new.go
similarity index 95%
rename from old.go
rename to new.go
index abc123..def456 100644
--- a/old.go
+++ b/new.go
@@ -1 +1 @@
-old content
+new content
`
	parser := NewDiffParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	diff, ok := result.Data.(*Diff)
	if !ok {
		t.Fatalf("expected *Diff, got %T", result.Data)
	}

	if len(diff.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(diff.Files))
	}

	if diff.Files[0].Status != "renamed" {
		t.Errorf("expected status renamed, got %s", diff.Files[0].Status)
	}

	if diff.Files[0].Path != "new.go" {
		t.Errorf("expected path new.go, got %s", diff.Files[0].Path)
	}

	if diff.Files[0].OldPath != "old.go" {
		t.Errorf("expected oldPath old.go, got %s", diff.Files[0].OldPath)
	}
}

func TestDiffParser_EmptyDiff(t *testing.T) {
	input := ""
	parser := NewDiffParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	diff, ok := result.Data.(*Diff)
	if !ok {
		t.Fatalf("expected *Diff, got %T", result.Data)
	}

	if len(diff.Files) != 0 {
		t.Errorf("expected 0 files, got %d", len(diff.Files))
	}
}

func TestDiffParser_MultipleHunks(t *testing.T) {
	input := `diff --git a/main.go b/main.go
--- a/main.go
+++ b/main.go
@@ -1,3 +1,4 @@
 first
+added
 second
@@ -10,3 +11,3 @@
 tenth
-removed
+replaced
`
	parser := NewDiffParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	diff, ok := result.Data.(*Diff)
	if !ok {
		t.Fatalf("expected *Diff, got %T", result.Data)
	}

	if len(diff.Files) != 1 {
		t.Fatalf("expected 1 file, got %d", len(diff.Files))
	}

	if len(diff.Files[0].Hunks) != 2 {
		t.Errorf("expected 2 hunks, got %d", len(diff.Files[0].Hunks))
	}
}

func TestDiffParser_LineTypes(t *testing.T) {
	input := `diff --git a/test.go b/test.go
--- a/test.go
+++ b/test.go
@@ -1,3 +1,3 @@
 context
-deleted
+added
`
	parser := NewDiffParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	diff, ok := result.Data.(*Diff)
	if !ok {
		t.Fatalf("expected *Diff, got %T", result.Data)
	}

	if len(diff.Files[0].Hunks) != 1 {
		t.Fatalf("expected 1 hunk, got %d", len(diff.Files[0].Hunks))
	}

	lines := diff.Files[0].Hunks[0].Lines
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}

	if lines[0].Type != "context" {
		t.Errorf("expected context line, got %s", lines[0].Type)
	}

	if lines[1].Type != "delete" {
		t.Errorf("expected delete line, got %s", lines[1].Type)
	}

	if lines[2].Type != "add" {
		t.Errorf("expected add line, got %s", lines[2].Type)
	}
}

func TestDiffParser_Matches(t *testing.T) {
	parser := NewDiffParser()

	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		expected    bool
	}{
		{"git diff", "git", []string{"diff"}, true},
		{"git diff --staged", "git", []string{"diff", "--staged"}, true},
		{"git diff --cached", "git", []string{"diff", "--cached"}, true},
		{"git status", "git", []string{"status"}, false},
		{"git log", "git", []string{"log"}, false},
		{"not git", "svn", []string{"diff"}, false},
		{"empty subcommands", "git", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parser.Matches(tt.cmd, tt.subcommands); got != tt.expected {
				t.Errorf("Matches(%q, %v) = %v, expected %v", tt.cmd, tt.subcommands, got, tt.expected)
			}
		})
	}
}

func TestDiffParser_Schema(t *testing.T) {
	parser := NewDiffParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("expected schema ID to be set")
	}

	if schema.Title != "Git Diff Output" {
		t.Errorf("expected schema Title 'Git Diff Output', got %s", schema.Title)
	}

	if schema.Type != "object" {
		t.Errorf("expected schema Type object, got %s", schema.Type)
	}

	// Verify required fields include "files"
	found := false
	for _, req := range schema.Required {
		if req == "files" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected schema Required to contain 'files', got %v", schema.Required)
	}
}
