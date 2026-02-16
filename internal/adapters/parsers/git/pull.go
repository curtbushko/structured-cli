package git

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Pull represents the structured output of 'git pull'.
type Pull struct {
	// Success indicates if the pull was successful.
	Success bool `json:"success"`

	// Commits is the number of commits pulled.
	Commits int `json:"commits"`

	// FilesChanged is the number of files changed.
	FilesChanged int `json:"filesChanged"`

	// Insertions is the number of lines added.
	Insertions int `json:"insertions"`

	// Deletions is the number of lines removed.
	Deletions int `json:"deletions"`

	// FastForward indicates if it was a fast-forward merge.
	FastForward bool `json:"fastForward"`

	// Conflicts contains paths of files with merge conflicts.
	Conflicts []string `json:"conflicts,omitempty"`
}

// PullParser parses the output of 'git pull'.
type PullParser struct {
	schema domain.Schema
	// Regex to parse conflict line
	conflictRe *regexp.Regexp
	// Regex to parse stats line
	statsRe *regexp.Regexp
}

// NewPullParser creates a new PullParser with the git-pull schema.
func NewPullParser() *PullParser {
	return &PullParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/git-pull.json",
			"Git Pull Output",
			"object",
			map[string]domain.PropertySchema{
				"success":      {Type: "boolean", Description: "Whether pull was successful"},
				"commits":      {Type: "integer", Description: "Number of commits pulled"},
				"filesChanged": {Type: "integer", Description: "Number of files changed"},
				"insertions":   {Type: "integer", Description: "Lines added"},
				"deletions":    {Type: "integer", Description: "Lines removed"},
				"fastForward":  {Type: "boolean", Description: "Whether it was fast-forward"},
				"conflicts":    {Type: "array", Description: "Files with conflicts"},
			},
			[]string{"success"},
		),
		// Match: CONFLICT (type): Merge conflict in <file>
		conflictRe: regexp.MustCompile(`^CONFLICT\s+\([^)]+\):\s+Merge conflict in\s+(.+)$`),
		// Match: N file(s) changed
		statsRe: regexp.MustCompile(`(\d+)\s+files?\s+changed`),
	}
}

// Parse reads git pull output and returns structured data.
func (p *PullParser) Parse(r io.Reader) (domain.ParseResult, error) {
	pull := &Pull{
		Success:   true, // Assume success unless we see conflicts
		Conflicts: []string{},
	}

	scanner := bufio.NewScanner(r)
	var rawBuilder strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		rawBuilder.WriteString(line)
		rawBuilder.WriteString("\n")

		p.parseLine(line, pull)
	}

	if err := scanner.Err(); err != nil {
		return domain.NewParseResultWithError(err, rawBuilder.String(), 0), nil
	}

	// If there are conflicts, the pull was not fully successful
	if len(pull.Conflicts) > 0 {
		pull.Success = false
	}

	return domain.NewParseResult(pull, rawBuilder.String(), 0), nil
}

// parseLine parses a single line of git pull output.
func (p *PullParser) parseLine(line string, pull *Pull) {
	if line == "" {
		return
	}

	switch {
	case line == "Fast-forward":
		pull.FastForward = true

	case strings.HasPrefix(line, "Already up to date"):
		pull.Success = true

	case strings.HasPrefix(line, "CONFLICT"):
		if matches := p.conflictRe.FindStringSubmatch(line); matches != nil {
			pull.Conflicts = append(pull.Conflicts, matches[1])
		}

	case strings.HasPrefix(line, "Automatic merge failed"):
		pull.Success = false

	case p.statsRe.MatchString(line):
		p.parseStatsLine(line, pull)
	}
}

// parseStatsLine parses the files changed/insertions/deletions line.
func (p *PullParser) parseStatsLine(line string, pull *Pull) {
	// Parse files changed
	filesRe := regexp.MustCompile(`(\d+)\s+files?\s+changed`)
	if matches := filesRe.FindStringSubmatch(line); matches != nil {
		pull.FilesChanged, _ = strconv.Atoi(matches[1])
	}

	// Parse insertions
	insertRe := regexp.MustCompile(`(\d+)\s+insertions?\(\+\)`)
	if matches := insertRe.FindStringSubmatch(line); matches != nil {
		pull.Insertions, _ = strconv.Atoi(matches[1])
	}

	// Parse deletions
	deleteRe := regexp.MustCompile(`(\d+)\s+deletions?\(-\)`)
	if matches := deleteRe.FindStringSubmatch(line); matches != nil {
		pull.Deletions, _ = strconv.Atoi(matches[1])
	}
}

// Schema returns the JSON Schema for git pull output.
func (p *PullParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *PullParser) Matches(cmd string, subcommands []string) bool {
	if cmd != gitCommand {
		return false
	}
	if len(subcommands) == 0 {
		return false
	}
	return subcommands[0] == "pull"
}
