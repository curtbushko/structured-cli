package git

import (
	"bufio"
	"io"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Add represents the structured output of 'git add'.
type Add struct {
	// Staged contains paths of files that were staged.
	Staged []string `json:"staged"`

	// Errors contains any error messages from the add operation.
	Errors []string `json:"errors,omitempty"`
}

// AddParser parses the output of 'git add'.
// Use 'git add -v' for verbose output showing which files were added.
type AddParser struct {
	schema domain.Schema
}

// NewAddParser creates a new AddParser with the git-add schema.
func NewAddParser() *AddParser {
	return &AddParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/git-add.json",
			"Git Add Output",
			"object",
			map[string]domain.PropertySchema{
				"staged": {Type: "array", Description: "Files staged for commit"},
				"errors": {Type: "array", Description: "Error messages"},
			},
			[]string{"staged"},
		),
	}
}

// Parse reads git add output and returns structured data.
func (p *AddParser) Parse(r io.Reader) (domain.ParseResult, error) {
	add := &Add{
		Staged: []string{},
		Errors: []string{},
	}

	scanner := bufio.NewScanner(r)
	var rawBuilder strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		rawBuilder.WriteString(line)
		rawBuilder.WriteString("\n")

		p.parseLine(line, add)
	}

	if err := scanner.Err(); err != nil {
		return domain.NewParseResultWithError(err, rawBuilder.String(), 0), nil
	}

	return domain.NewParseResult(add, rawBuilder.String(), 0), nil
}

// parseLine parses a single line of git add output.
func (p *AddParser) parseLine(line string, add *Add) {
	if line == "" {
		return
	}

	switch {
	case strings.HasPrefix(line, "add '"):
		// Verbose output: add 'filename'
		file := strings.TrimPrefix(line, "add '")
		file = strings.TrimSuffix(file, "'")
		add.Staged = append(add.Staged, file)

	case strings.HasPrefix(line, "fatal:"):
		// Error message
		add.Errors = append(add.Errors, line)

	case strings.HasPrefix(line, "error:"):
		// Error message
		add.Errors = append(add.Errors, line)
	}
}

// Schema returns the JSON Schema for git add output.
func (p *AddParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *AddParser) Matches(cmd string, subcommands []string) bool {
	if cmd != gitCommand {
		return false
	}
	if len(subcommands) == 0 {
		return false
	}
	return subcommands[0] == "add"
}
