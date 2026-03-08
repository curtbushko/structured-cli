package git

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// File status constants.
const (
	statusAdded    = "added"
	statusDeleted  = "deleted"
	statusModified = "modified"
	statusRenamed  = "renamed"
	statusCopied   = "copied"
	statusUnknown  = "unknown"
)

// Line type constants.
const (
	lineTypeContext = "context"
	lineTypeAdd     = "add"
	lineTypeDelete  = "delete"
)

// Command constant.
const gitCommand = "git"

// DiffParser parses the output of 'git diff'.
type DiffParser struct {
	schema domain.Schema
}

// hunkHeaderRegex matches the @@ -old,count +new,count @@ pattern.
var hunkHeaderRegex = regexp.MustCompile(`^@@\s+-(\d+)(?:,(\d+))?\s+\+(\d+)(?:,(\d+))?\s+@@`)

// NewDiffParser creates a new DiffParser with the git-diff schema.
func NewDiffParser() *DiffParser {
	return &DiffParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/git-diff.json",
			"Git Diff Output",
			"object",
			map[string]domain.PropertySchema{
				"files": {Type: "array", Description: "Files changed in the diff"},
			},
			[]string{"files"},
		),
	}
}

// Parse reads git diff output and returns structured data.
func (p *DiffParser) Parse(r io.Reader) (domain.ParseResult, error) {
	diff := &Diff{
		Files: []DiffFile{},
	}

	scanner := bufio.NewScanner(r)
	var rawBuilder strings.Builder
	var lines []string

	// Read all lines first
	for scanner.Scan() {
		line := scanner.Text()
		rawBuilder.WriteString(line)
		rawBuilder.WriteString("\n")
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return domain.NewParseResultWithError(
			fmt.Errorf("read input: %w", err),
			rawBuilder.String(),
			0,
		), nil
	}

	// Parse all lines
	p.parseLines(lines, diff)

	return domain.NewParseResult(diff, rawBuilder.String(), 0), nil
}

// diffParseState holds the current parsing state.
type diffParseState struct {
	currentFile *DiffFile
	currentHunk *DiffHunk
	oldLineNo   int
	newLineNo   int
}

// parseLines parses all lines and populates the diff struct.
func (p *DiffParser) parseLines(lines []string, diff *Diff) {
	state := &diffParseState{}

	for _, line := range lines {
		p.parseLine(line, diff, state)
	}

	// Finalize the last file and hunk
	p.finalizeCurrentFile(diff, state)
}

// parseLine processes a single line based on its prefix.
func (p *DiffParser) parseLine(line string, diff *Diff, state *diffParseState) {
	switch {
	case strings.HasPrefix(line, "diff --git "):
		p.handleNewFileDiff(line, diff, state)
	case strings.HasPrefix(line, "new file mode"):
		p.setFileStatus(state, statusAdded)
	case strings.HasPrefix(line, "deleted file mode"):
		p.setFileStatus(state, statusDeleted)
	case strings.HasPrefix(line, "rename from "):
		p.handleRenameFrom(line, state)
	case strings.HasPrefix(line, "rename to "):
		p.handleRenameTo(line, state)
	case strings.HasPrefix(line, "Binary files"):
		p.handleBinaryFile(state)
	case strings.HasPrefix(line, "@@"):
		p.handleHunkHeader(line, state)
	case p.isSkippableLine(line):
		// Skip metadata lines
	case state.currentHunk != nil && p.isContentLine(line):
		p.handleContentLine(line, state)
	}
}

// handleNewFileDiff starts a new file diff section.
func (p *DiffParser) handleNewFileDiff(line string, diff *Diff, state *diffParseState) {
	p.finalizeCurrentFile(diff, state)
	state.currentFile = p.parseFileDiffHeader(line)
	state.currentHunk = nil
}

// setFileStatus sets the status on the current file if present.
func (p *DiffParser) setFileStatus(state *diffParseState, status string) {
	if state.currentFile != nil {
		state.currentFile.Status = status
	}
}

// handleRenameFrom processes "rename from" line.
func (p *DiffParser) handleRenameFrom(line string, state *diffParseState) {
	if state.currentFile != nil {
		state.currentFile.OldPath = strings.TrimPrefix(line, "rename from ")
		state.currentFile.Status = statusRenamed
	}
}

// handleRenameTo processes "rename to" line.
func (p *DiffParser) handleRenameTo(line string, state *diffParseState) {
	if state.currentFile != nil {
		state.currentFile.Path = strings.TrimPrefix(line, "rename to ")
	}
}

// handleBinaryFile marks the current file as binary.
func (p *DiffParser) handleBinaryFile(state *diffParseState) {
	if state.currentFile != nil {
		state.currentFile.Binary = true
	}
}

// handleHunkHeader processes a hunk header line.
func (p *DiffParser) handleHunkHeader(line string, state *diffParseState) {
	if state.currentFile == nil {
		return
	}
	if state.currentHunk != nil {
		state.currentFile.Hunks = append(state.currentFile.Hunks, *state.currentHunk)
	}
	state.currentHunk = p.parseHunkHeader(line)
	if state.currentHunk != nil {
		state.oldLineNo = state.currentHunk.OldStart
		state.newLineNo = state.currentHunk.NewStart
	}
}

// isSkippableLine returns true for lines that should be skipped.
func (p *DiffParser) isSkippableLine(line string) bool {
	return strings.HasPrefix(line, "--- ") ||
		strings.HasPrefix(line, "+++ ") ||
		strings.HasPrefix(line, "index ") ||
		strings.HasPrefix(line, "similarity index")
}

// isContentLine returns true if the line is content within a hunk.
func (p *DiffParser) isContentLine(line string) bool {
	return len(line) == 0 ||
		strings.HasPrefix(line, " ") ||
		strings.HasPrefix(line, "+") ||
		strings.HasPrefix(line, "-")
}

// handleContentLine processes a content line within a hunk.
func (p *DiffParser) handleContentLine(line string, state *diffParseState) {
	diffLine := p.parseDiffLine(line, &state.oldLineNo, &state.newLineNo)
	if diffLine == nil {
		return
	}
	state.currentHunk.Lines = append(state.currentHunk.Lines, *diffLine)
	if state.currentFile != nil {
		switch diffLine.Type {
		case lineTypeAdd:
			state.currentFile.Additions++
		case lineTypeDelete:
			state.currentFile.Deletions++
		}
	}
}

// finalizeCurrentFile adds the current file to diff if present.
func (p *DiffParser) finalizeCurrentFile(diff *Diff, state *diffParseState) {
	if state.currentFile == nil {
		return
	}
	if state.currentHunk != nil {
		state.currentFile.Hunks = append(state.currentFile.Hunks, *state.currentHunk)
		state.currentHunk = nil
	}
	diff.Files = append(diff.Files, *state.currentFile)
	state.currentFile = nil
}

// parseFileDiffHeader parses "diff --git a/X b/Y" line.
func (p *DiffParser) parseFileDiffHeader(line string) *DiffFile {
	// Format: diff --git a/path b/path
	// For paths with spaces, we need to find the "b/" prefix that starts the second path
	// The line looks like: "diff --git a/path with spaces b/path with spaces"

	// Remove "diff --git " prefix
	const prefix = "diff --git "
	if !strings.HasPrefix(line, prefix) {
		return nil
	}
	remainder := strings.TrimPrefix(line, prefix)

	// The remainder is "a/path b/path" - we need to find where "b/" starts
	// Since paths can contain spaces, we look for " b/" pattern
	bIndex := strings.Index(remainder, " b/")
	if bIndex == -1 {
		// Fallback for simple paths without spaces
		parts := strings.SplitN(remainder, " ", 2)
		if len(parts) < 2 {
			return nil
		}
		path := strings.TrimPrefix(parts[1], "b/")
		return &DiffFile{
			Path:   path,
			Status: statusModified,
			Hunks:  []DiffHunk{},
		}
	}

	// Extract path from b/path (after the " b/" we found)
	bPath := remainder[bIndex+1:] // Skip the space before "b/"
	path := strings.TrimPrefix(bPath, "b/")

	return &DiffFile{
		Path:   path,
		Status: statusModified, // Default, may be overwritten
		Hunks:  []DiffHunk{},
	}
}

// parseHunkHeader parses "@@ -old,count +new,count @@" line.
func (p *DiffParser) parseHunkHeader(line string) *DiffHunk {
	matches := hunkHeaderRegex.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	oldStart, _ := strconv.Atoi(matches[1])
	oldLines := 1
	if matches[2] != "" {
		oldLines, _ = strconv.Atoi(matches[2])
	}
	newStart, _ := strconv.Atoi(matches[3])
	newLines := 1
	if matches[4] != "" {
		newLines, _ = strconv.Atoi(matches[4])
	}

	return &DiffHunk{
		OldStart: oldStart,
		OldLines: oldLines,
		NewStart: newStart,
		NewLines: newLines,
		Header:   line,
		Lines:    []DiffLine{},
	}
}

// parseDiffLine parses a content line (context, add, or delete).
func (p *DiffParser) parseDiffLine(line string, oldLineNo, newLineNo *int) *DiffLine {
	if len(line) == 0 {
		return p.createContextLine(line, oldLineNo, newLineNo)
	}

	switch line[0] {
	case ' ':
		return p.createContextLine(line, oldLineNo, newLineNo)
	case '-':
		return p.createDeleteLine(line, oldLineNo)
	case '+':
		return p.createAddLine(line, newLineNo)
	default:
		return nil
	}
}

// createContextLine creates a context diff line.
func (p *DiffParser) createContextLine(content string, oldLineNo, newLineNo *int) *DiffLine {
	oldNo := *oldLineNo
	newNo := *newLineNo
	*oldLineNo++
	*newLineNo++
	return &DiffLine{
		Type:      lineTypeContext,
		Content:   content,
		OldLineNo: &oldNo,
		NewLineNo: &newNo,
	}
}

// createDeleteLine creates a delete diff line.
func (p *DiffParser) createDeleteLine(content string, oldLineNo *int) *DiffLine {
	oldNo := *oldLineNo
	*oldLineNo++
	return &DiffLine{
		Type:      lineTypeDelete,
		Content:   content,
		OldLineNo: &oldNo,
	}
}

// createAddLine creates an add diff line.
func (p *DiffParser) createAddLine(content string, newLineNo *int) *DiffLine {
	newNo := *newLineNo
	*newLineNo++
	return &DiffLine{
		Type:      lineTypeAdd,
		Content:   content,
		NewLineNo: &newNo,
	}
}

// Schema returns the JSON Schema for git diff output.
func (p *DiffParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *DiffParser) Matches(cmd string, subcommands []string) bool {
	if cmd != gitCommand {
		return false
	}
	if len(subcommands) == 0 {
		return false
	}
	return subcommands[0] == "diff"
}
