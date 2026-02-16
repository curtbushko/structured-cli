package git

import (
	"bufio"
	"io"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

const (
	commitStartMarker = "COMMIT_START"
	commitEndMarker   = "COMMIT_END"
)

// LogParser parses the output of 'git log' with a custom format.
// Expected format:
//
//	git log --format='COMMIT_START%n%H%n%h%n%an%n%ae%n%aI%n%s%n%b%nCOMMIT_END' --numstat
type LogParser struct {
	schema domain.Schema
}

// NewLogParser creates a new LogParser with the git-log schema.
func NewLogParser() *LogParser {
	return &LogParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/git-log.json",
			"Git Log Output",
			"object",
			map[string]domain.PropertySchema{
				"commits": {Type: "array", Description: "List of commits"},
			},
			[]string{"commits"},
		),
	}
}

// Parse reads git log output and returns structured data.
func (p *LogParser) Parse(r io.Reader) (domain.ParseResult, error) {
	log := &Log{
		Commits: []Commit{},
	}

	scanner := bufio.NewScanner(r)
	var rawBuilder strings.Builder
	var currentCommit *Commit
	var inCommit bool
	var bodyLines []string
	var lineIndex int

	for scanner.Scan() {
		line := scanner.Text()
		rawBuilder.WriteString(line)
		rawBuilder.WriteString("\n")

		switch {
		case line == commitStartMarker:
			inCommit = true
			currentCommit = &Commit{
				Files: []FileChange{},
			}
			bodyLines = nil
			lineIndex = 0

		case line == commitEndMarker:
			if currentCommit != nil {
				// Join body lines and trim
				currentCommit.Body = strings.TrimSpace(strings.Join(bodyLines, "\n"))
				// Build full message
				if currentCommit.Body != "" {
					currentCommit.Message = currentCommit.Subject + "\n\n" + currentCommit.Body
				} else {
					currentCommit.Message = currentCommit.Subject
				}
				log.Commits = append(log.Commits, *currentCommit)
			}
			inCommit = false
			currentCommit = nil

		case inCommit && currentCommit != nil:
			p.parseCommitLine(line, currentCommit, &bodyLines, &lineIndex)

		case !inCommit && currentCommit == nil && len(log.Commits) > 0:
			// This is a numstat line after COMMIT_END
			p.parseNumstatLine(line, &log.Commits[len(log.Commits)-1])
		}
	}

	if err := scanner.Err(); err != nil {
		return domain.NewParseResultWithError(err, rawBuilder.String(), 0), nil
	}

	return domain.NewParseResult(log, rawBuilder.String(), 0), nil
}

// parseCommitLine parses a line within the commit block.
func (p *LogParser) parseCommitLine(line string, commit *Commit, bodyLines *[]string, lineIndex *int) {
	switch *lineIndex {
	case 0:
		commit.Hash = line
	case 1:
		commit.AbbrevHash = line
	case 2:
		commit.Author = line
	case 3:
		commit.Email = line
	case 4:
		commit.Date = line
	case 5:
		commit.Subject = line
	default:
		// Lines after subject are body (index 6+)
		// Skip the blank line that separates subject from body
		if *lineIndex == 6 && line == "" {
			// Skip the blank line after subject
		} else {
			*bodyLines = append(*bodyLines, line)
		}
	}
	*lineIndex++
}

// parseNumstatLine parses a numstat line (additions deletions path).
func (p *LogParser) parseNumstatLine(line string, commit *Commit) {
	if line == "" {
		return
	}

	parts := strings.Split(line, "\t")
	if len(parts) != 3 {
		return
	}

	additionsStr := parts[0]
	deletionsStr := parts[1]
	path := parts[2]

	var additions, deletions int

	// Binary files show "-" instead of numbers
	if additionsStr == "-" {
		additions = -1
	} else {
		additions, _ = strconv.Atoi(additionsStr)
	}

	if deletionsStr == "-" {
		deletions = -1
	} else {
		deletions, _ = strconv.Atoi(deletionsStr)
	}

	fileChange := FileChange{
		Path:      path,
		Additions: additions,
		Deletions: deletions,
	}

	commit.Files = append(commit.Files, fileChange)

	// Update totals (skip binary files marked as -1)
	if additions > 0 {
		commit.Insertions += additions
	}
	if deletions > 0 {
		commit.Deletions += deletions
	}
}

// Schema returns the JSON Schema for git log output.
func (p *LogParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *LogParser) Matches(cmd string, subcommands []string) bool {
	if cmd != "git" {
		return false
	}
	if len(subcommands) == 0 {
		return false
	}
	return subcommands[0] == "log"
}
