package git

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Reflog represents the structured output of 'git reflog'.
type Reflog struct {
	// Entries contains the reflog entries.
	Entries []ReflogEntry `json:"entries"`
}

// ReflogEntry represents a single entry in the reflog.
type ReflogEntry struct {
	// Hash is the commit hash.
	Hash string `json:"hash"`

	// Index is the reflog index (e.g., 0 for HEAD@{0}).
	Index int `json:"index"`

	// Action is the action that was performed (e.g., commit, checkout, merge).
	Action string `json:"action"`

	// Message is the reflog message.
	Message string `json:"message"`

	// Date is the date of the action (optional, depends on format).
	Date string `json:"date,omitempty"`
}

// ReflogParser parses the output of 'git reflog'.
type ReflogParser struct {
	schema domain.Schema
	// Regex to parse reflog line: <hash> HEAD@{<index>} <action>: <message>
	entryRe *regexp.Regexp
}

// NewReflogParser creates a new ReflogParser with the git-reflog schema.
func NewReflogParser() *ReflogParser {
	return &ReflogParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/git-reflog.json",
			"Git Reflog Output",
			"object",
			map[string]domain.PropertySchema{
				"entries": {Type: "array", Description: "Reflog entries"},
			},
			[]string{"entries"},
		),
		// Match: <hash> HEAD@{<index>}: <action>: <message>
		entryRe: regexp.MustCompile(`^([a-f0-9]+)\s+HEAD@\{(\d+)\}:\s+([^:]+):\s*(.*)$`),
	}
}

// Parse reads git reflog output and returns structured data.
func (p *ReflogParser) Parse(r io.Reader) (domain.ParseResult, error) {
	reflog := &Reflog{
		Entries: []ReflogEntry{},
	}

	scanner := bufio.NewScanner(r)
	var rawBuilder strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		rawBuilder.WriteString(line)
		rawBuilder.WriteString("\n")

		if entry := p.parseLine(line); entry != nil {
			reflog.Entries = append(reflog.Entries, *entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return domain.NewParseResultWithError(err, rawBuilder.String(), 0), nil
	}

	return domain.NewParseResult(reflog, rawBuilder.String(), 0), nil
}

// parseLine parses a single line of git reflog output.
func (p *ReflogParser) parseLine(line string) *ReflogEntry {
	if line == "" {
		return nil
	}

	matches := p.entryRe.FindStringSubmatch(line)
	if matches == nil {
		return nil
	}

	index, _ := strconv.Atoi(matches[2])

	return &ReflogEntry{
		Hash:    matches[1],
		Index:   index,
		Action:  matches[3],
		Message: matches[4],
	}
}

// Schema returns the JSON Schema for git reflog output.
func (p *ReflogParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *ReflogParser) Matches(cmd string, subcommands []string) bool {
	if cmd != gitCommand {
		return false
	}
	if len(subcommands) == 0 {
		return false
	}
	return subcommands[0] == "reflog"
}
