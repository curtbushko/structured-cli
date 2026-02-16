package git

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// CommitOutput represents the structured output of 'git commit'.
// Named CommitOutput to avoid conflict with existing Commit type in types.go.
type CommitOutput struct {
	// Hash is the abbreviated commit hash.
	Hash string `json:"hash"`

	// Branch is the branch the commit was made on.
	Branch string `json:"branch"`

	// Message is the commit message.
	Message string `json:"message"`

	// FilesChanged is the number of files changed.
	FilesChanged int `json:"filesChanged"`

	// Insertions is the number of lines added.
	Insertions int `json:"insertions"`

	// Deletions is the number of lines removed.
	Deletions int `json:"deletions"`
}

// CommitParser parses the output of 'git commit'.
type CommitParser struct {
	schema domain.Schema
	// Regex to parse commit line: [branch hash] message
	// Also handles root-commit: [branch (root-commit) hash] message
	commitLineRe *regexp.Regexp
	// Regex to parse stats line: N file(s) changed, N insertions(+), N deletions(-)
	statsLineRe *regexp.Regexp
}

// NewCommitParser creates a new CommitParser with the git-commit schema.
func NewCommitParser() *CommitParser {
	return &CommitParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/git-commit.json",
			"Git Commit Output",
			"object",
			map[string]domain.PropertySchema{
				"hash":         {Type: "string", Description: "Abbreviated commit hash"},
				"branch":       {Type: "string", Description: "Branch name"},
				"message":      {Type: "string", Description: "Commit message"},
				"filesChanged": {Type: "integer", Description: "Number of files changed"},
				"insertions":   {Type: "integer", Description: "Lines added"},
				"deletions":    {Type: "integer", Description: "Lines removed"},
			},
			[]string{"hash", "branch", "message"},
		),
		// Match: [branch hash] or [branch (root-commit) hash]
		commitLineRe: regexp.MustCompile(`^\[([^\s\]]+)(?:\s+\(root-commit\))?\s+([a-f0-9]+)\]\s+(.+)$`),
		// Match: N file(s) changed, N insertion(s)(+), N deletion(s)(-)
		statsLineRe: regexp.MustCompile(`(\d+)\s+files?\s+changed`),
	}
}

// Parse reads git commit output and returns structured data.
func (p *CommitParser) Parse(r io.Reader) (domain.ParseResult, error) {
	commit := &CommitOutput{}

	scanner := bufio.NewScanner(r)
	var rawBuilder strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		rawBuilder.WriteString(line)
		rawBuilder.WriteString("\n")

		p.parseLine(line, commit)
	}

	if err := scanner.Err(); err != nil {
		return domain.NewParseResultWithError(err, rawBuilder.String(), 0), nil
	}

	return domain.NewParseResult(commit, rawBuilder.String(), 0), nil
}

// parseLine parses a single line of git commit output.
func (p *CommitParser) parseLine(line string, commit *CommitOutput) {
	if line == "" {
		return
	}

	// Try to match commit line
	if matches := p.commitLineRe.FindStringSubmatch(line); matches != nil {
		commit.Branch = matches[1]
		commit.Hash = matches[2]
		commit.Message = matches[3]
		return
	}

	// Try to match stats line
	if p.statsLineRe.MatchString(line) {
		p.parseStatsLine(line, commit)
	}
}

// parseStatsLine parses the files changed/insertions/deletions line.
func (p *CommitParser) parseStatsLine(line string, commit *CommitOutput) {
	// Parse files changed
	filesRe := regexp.MustCompile(`(\d+)\s+files?\s+changed`)
	if matches := filesRe.FindStringSubmatch(line); matches != nil {
		commit.FilesChanged, _ = strconv.Atoi(matches[1])
	}

	// Parse insertions
	insertRe := regexp.MustCompile(`(\d+)\s+insertions?\(\+\)`)
	if matches := insertRe.FindStringSubmatch(line); matches != nil {
		commit.Insertions, _ = strconv.Atoi(matches[1])
	}

	// Parse deletions
	deleteRe := regexp.MustCompile(`(\d+)\s+deletions?\(-\)`)
	if matches := deleteRe.FindStringSubmatch(line); matches != nil {
		commit.Deletions, _ = strconv.Atoi(matches[1])
	}
}

// Schema returns the JSON Schema for git commit output.
func (p *CommitParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *CommitParser) Matches(cmd string, subcommands []string) bool {
	if cmd != gitCommand {
		return false
	}
	if len(subcommands) == 0 {
		return false
	}
	return subcommands[0] == "commit"
}
