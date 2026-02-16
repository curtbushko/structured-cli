package fileops

import (
	"bufio"
	"io"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// FindParser parses the output of 'find' command.
type FindParser struct {
	schema domain.Schema
}

// NewFindParser creates a new FindParser with the find schema.
func NewFindParser() *FindParser {
	return &FindParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/find.json",
			"Find Output",
			"object",
			map[string]domain.PropertySchema{
				"files": {Type: "array", Description: "List of matching file paths"},
				"count": {Type: "integer", Description: "Number of matches found"},
			},
			[]string{"files", "count"},
		),
	}
}

// Parse reads find output and returns structured data.
func (p *FindParser) Parse(r io.Reader) (domain.ParseResult, error) {
	output := &FindOutput{
		Files: []string{},
	}

	scanner := bufio.NewScanner(r)
	var rawBuilder strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		rawBuilder.WriteString(line)
		rawBuilder.WriteString("\n")

		if line == "" {
			continue
		}

		output.Files = append(output.Files, line)
	}

	if err := scanner.Err(); err != nil {
		return domain.NewParseResultWithError(err, rawBuilder.String(), 0), nil
	}

	output.Count = len(output.Files)

	return domain.NewParseResult(output, rawBuilder.String(), 0), nil
}

// Schema returns the JSON Schema for find output.
func (p *FindParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *FindParser) Matches(cmd string, _ []string) bool {
	return cmd == "find"
}
