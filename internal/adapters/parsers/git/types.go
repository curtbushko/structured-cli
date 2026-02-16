// Package git provides parsers for git command output.
// This package is in the adapters layer and implements parsers for
// converting raw git command output into structured domain types.
package git

// Status represents the structured output of 'git status'.
// It captures branch information, tracking status, and file changes.
type Status struct {
	// Branch is the current branch name.
	Branch string `json:"branch"`

	// Upstream is the upstream branch being tracked, or nil if not tracking.
	Upstream *string `json:"upstream"`

	// Ahead is the number of commits ahead of upstream.
	Ahead int `json:"ahead"`

	// Behind is the number of commits behind upstream.
	Behind int `json:"behind"`

	// Staged contains files that are staged for commit.
	Staged []StagedFile `json:"staged"`

	// Modified contains paths of files modified in the working tree.
	Modified []string `json:"modified"`

	// Deleted contains paths of files deleted in the working tree.
	Deleted []string `json:"deleted"`

	// Untracked contains paths of untracked files.
	Untracked []string `json:"untracked"`

	// Conflicts contains paths of files with merge conflicts.
	Conflicts []string `json:"conflicts"`

	// Clean is true if the working tree has no changes.
	Clean bool `json:"clean"`
}

// StagedFile represents a file staged for commit.
type StagedFile struct {
	// File is the path of the staged file.
	File string `json:"file"`

	// Status indicates the type of staged change.
	// Valid values: added, modified, deleted, renamed, copied
	Status string `json:"status"`
}

// Log represents the structured output of 'git log'.
// It contains an array of commits with their metadata and optional stats.
type Log struct {
	// Commits is the list of commits in the log.
	Commits []Commit `json:"commits"`
}

// Commit represents a single git commit.
type Commit struct {
	// Hash is the full commit hash (SHA-1).
	Hash string `json:"hash"`

	// AbbrevHash is the abbreviated commit hash.
	AbbrevHash string `json:"abbrevHash"`

	// Author is the name of the commit author.
	Author string `json:"author"`

	// Email is the email address of the commit author.
	Email string `json:"email"`

	// Date is the commit date in ISO 8601 format.
	Date string `json:"date"`

	// Message is the full commit message (subject + body).
	Message string `json:"message"`

	// Subject is the first line of the commit message.
	Subject string `json:"subject"`

	// Body is the remainder of the commit message after the subject.
	Body string `json:"body,omitempty"`

	// Files contains per-file change statistics.
	Files []FileChange `json:"files,omitempty"`

	// Insertions is the total number of lines added in this commit.
	Insertions int `json:"insertions,omitempty"`

	// Deletions is the total number of lines removed in this commit.
	Deletions int `json:"deletions,omitempty"`
}

// FileChange represents changes to a file in a commit.
type FileChange struct {
	// Path is the file path relative to the repository root.
	Path string `json:"path"`

	// Additions is the number of lines added to this file.
	Additions int `json:"additions"`

	// Deletions is the number of lines removed from this file.
	Deletions int `json:"deletions"`
}

// Diff represents the structured output of 'git diff'.
// It contains an array of files with their changes including hunks and line-level detail.
type Diff struct {
	// Files is the list of files changed in the diff.
	Files []DiffFile `json:"files"`
}

// DiffFile represents changes to a single file in a diff.
type DiffFile struct {
	// Path is the file path relative to the repository root.
	Path string `json:"path"`

	// OldPath is the previous path for renamed files.
	OldPath string `json:"oldPath,omitempty"`

	// Status indicates the type of change.
	// Valid values: added, modified, deleted, renamed
	Status string `json:"status"`

	// Binary is true if the file is binary.
	Binary bool `json:"binary"`

	// Additions is the number of lines added.
	Additions int `json:"additions"`

	// Deletions is the number of lines removed.
	Deletions int `json:"deletions"`

	// Hunks contains the contiguous blocks of changes.
	Hunks []DiffHunk `json:"hunks,omitempty"`
}

// DiffHunk represents a contiguous block of changes in a diff.
type DiffHunk struct {
	// OldStart is the starting line number in the old file.
	OldStart int `json:"oldStart"`

	// OldLines is the number of lines from the old file in this hunk.
	OldLines int `json:"oldLines"`

	// NewStart is the starting line number in the new file.
	NewStart int `json:"newStart"`

	// NewLines is the number of lines in the new file in this hunk.
	NewLines int `json:"newLines"`

	// Header is the hunk header line (e.g., "@@ -10,5 +10,7 @@ func main()").
	Header string `json:"header"`

	// Lines contains the individual line changes.
	Lines []DiffLine `json:"lines"`
}

// DiffLine represents a single line in a diff hunk.
type DiffLine struct {
	// Type indicates the line type.
	// Valid values: context, add, delete
	Type string `json:"type"`

	// Content is the line content including the prefix (+, -, or space).
	Content string `json:"content"`

	// OldLineNo is the line number in the old file, or nil for added lines.
	OldLineNo *int `json:"oldLineNo,omitempty"`

	// NewLineNo is the line number in the new file, or nil for deleted lines.
	NewLineNo *int `json:"newLineNo,omitempty"`
}

// BranchList represents the structured output of 'git branch'.
// It contains information about all branches and which one is current.
type BranchList struct {
	// Branches is the list of all branches.
	Branches []Branch `json:"branches"`

	// Current is the name of the currently checked out branch.
	Current string `json:"current"`
}

// Branch represents a single git branch.
type Branch struct {
	// Name is the branch name.
	Name string `json:"name"`

	// Current is true if this is the currently checked out branch.
	Current bool `json:"current"`

	// Upstream is the upstream tracking branch, or nil if not tracking.
	Upstream *string `json:"upstream,omitempty"`

	// Ahead is the number of commits ahead of the upstream branch.
	Ahead int `json:"ahead,omitempty"`

	// Behind is the number of commits behind the upstream branch.
	Behind int `json:"behind,omitempty"`

	// LastCommit is the abbreviated hash of the last commit on this branch.
	LastCommit string `json:"lastCommit,omitempty"`
}

// Show represents the structured output of 'git show'.
// It combines commit details with the diff for that commit.
type Show struct {
	// Commit contains the commit metadata.
	Commit Commit `json:"commit"`

	// Diff contains the changes introduced by the commit.
	Diff Diff `json:"diff"`
}
