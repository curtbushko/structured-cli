package git

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// authorEmailRegex matches "Author: Name <email>" format.
var authorEmailRegex = regexp.MustCompile(`^Author:\s*(.+?)\s*<(.+?)>`)

// ShowParser parses the output of 'git show'.
// Expected format is the default git show format which combines commit info with diff.
type ShowParser struct {
	diffParser *DiffParser
	schema     domain.Schema
}

// NewShowParser creates a new ShowParser with the git-show schema.
func NewShowParser() *ShowParser {
	return &ShowParser{
		diffParser: NewDiffParser(),
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/git-show.json",
			"Git Show Output",
			"object",
			map[string]domain.PropertySchema{
				"commit": {Type: "object", Description: "Commit details"},
				"diff":   {Type: "object", Description: "Diff changes"},
			},
			[]string{"commit", "diff"},
		),
	}
}

// Parse reads git show output and returns structured data.
func (p *ShowParser) Parse(r io.Reader) (domain.ParseResult, error) {
	show := &Show{
		Commit: Commit{},
		Diff:   Diff{Files: []DiffFile{}},
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
		return domain.NewParseResultWithError(err, rawBuilder.String(), 0), nil
	}

	// Parse the commit header and message, collecting diff lines
	diffLines := p.parseCommitSection(lines, &show.Commit)

	// Parse the diff section if present
	if len(diffLines) > 0 {
		diffInput := strings.Join(diffLines, "\n")
		diffResult, err := p.diffParser.Parse(strings.NewReader(diffInput))
		if err != nil {
			return domain.NewParseResultWithError(err, rawBuilder.String(), 0), nil
		}
		if diff, ok := diffResult.Data.(*Diff); ok {
			show.Diff = *diff
		}
	}

	return domain.NewParseResult(show, rawBuilder.String(), 0), nil
}

// showParseState tracks the parsing state.
type showParseState int

const (
	stateHeader showParseState = iota
	stateMessage
	stateDiff
)

// parseCommitSection parses commit header and message, returns remaining diff lines.
func (p *ShowParser) parseCommitSection(lines []string, commit *Commit) []string {
	state := stateHeader
	var messageLines []string
	var diffStartIndex int

	for i, line := range lines {
		switch state {
		case stateHeader:
			if p.parseHeaderLine(line, commit) {
				continue
			}
			// Empty line after header marks start of message
			if line == "" {
				state = stateMessage
				continue
			}

		case stateMessage:
			// Check if we've reached the diff section
			if strings.HasPrefix(line, "diff --git ") {
				state = stateDiff
				diffStartIndex = i
				break
			}
			// Message lines are indented with 4 spaces
			if strings.HasPrefix(line, "    ") {
				messageLines = append(messageLines, strings.TrimPrefix(line, "    "))
			} else if line == "" {
				// Keep blank lines in message
				messageLines = append(messageLines, "")
			}
		}

		if state == stateDiff {
			break
		}
	}

	// Build subject and body from message lines
	p.buildCommitMessage(messageLines, commit)

	// Return diff lines
	if diffStartIndex > 0 && diffStartIndex < len(lines) {
		return lines[diffStartIndex:]
	}
	return nil
}

// parseHeaderLine parses a header line and updates the commit.
// Returns true if the line was a header line.
func (p *ShowParser) parseHeaderLine(line string, commit *Commit) bool {
	switch {
	case strings.HasPrefix(line, "commit "):
		commit.Hash = strings.TrimPrefix(line, "commit ")
		// Set abbreviated hash (first 7 characters)
		if len(commit.Hash) >= 7 {
			commit.AbbrevHash = commit.Hash[:7]
		} else {
			commit.AbbrevHash = commit.Hash
		}
		return true

	case strings.HasPrefix(line, "Author:"):
		p.parseAuthorLine(line, commit)
		return true

	case strings.HasPrefix(line, "Date:"):
		commit.Date = strings.TrimSpace(strings.TrimPrefix(line, "Date:"))
		return true
	}
	return false
}

// parseAuthorLine parses "Author: Name <email>" line.
func (p *ShowParser) parseAuthorLine(line string, commit *Commit) {
	matches := authorEmailRegex.FindStringSubmatch(line)
	if matches != nil {
		commit.Author = matches[1]
		commit.Email = matches[2]
	} else {
		// Fallback: just take everything after "Author:"
		commit.Author = strings.TrimSpace(strings.TrimPrefix(line, "Author:"))
	}
}

// buildCommitMessage constructs Subject, Body, and Message from message lines.
func (p *ShowParser) buildCommitMessage(messageLines []string, commit *Commit) {
	// Remove leading and trailing empty lines
	messageLines = trimEmptyLines(messageLines)

	if len(messageLines) == 0 {
		return
	}

	// First non-empty line is the subject
	commit.Subject = messageLines[0]

	// Rest is body (skip the blank line separator if present)
	if len(messageLines) > 1 {
		bodyStart := 1
		// Skip blank line after subject
		if bodyStart < len(messageLines) && messageLines[bodyStart] == "" {
			bodyStart++
		}
		if bodyStart < len(messageLines) {
			bodyLines := trimEmptyLines(messageLines[bodyStart:])
			commit.Body = strings.Join(bodyLines, "\n")
		}
	}

	// Build full message
	if commit.Body != "" {
		commit.Message = commit.Subject + "\n\n" + commit.Body
	} else {
		commit.Message = commit.Subject
	}
}

// trimEmptyLines removes leading and trailing empty strings from a slice.
func trimEmptyLines(lines []string) []string {
	start := 0
	for start < len(lines) && lines[start] == "" {
		start++
	}
	end := len(lines)
	for end > start && lines[end-1] == "" {
		end--
	}
	if start >= end {
		return nil
	}
	return lines[start:end]
}

// Schema returns the JSON Schema for git show output.
func (p *ShowParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *ShowParser) Matches(cmd string, subcommands []string) bool {
	if cmd != gitCommand {
		return false
	}
	if len(subcommands) == 0 {
		return false
	}
	return subcommands[0] == "show"
}
