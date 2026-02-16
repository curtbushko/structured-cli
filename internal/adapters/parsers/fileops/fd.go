package fileops

import (
	"bufio"
	"io"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// FDParser parses the output of 'fd' command.
type FDParser struct {
	schema domain.Schema
}

// NewFDParser creates a new FDParser with the fd schema.
func NewFDParser() *FDParser {
	return &FDParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/fd.json",
			"FD Output",
			"object",
			map[string]domain.PropertySchema{
				"files": {Type: "array", Description: "List of matching file paths"},
				"count": {Type: "integer", Description: "Number of matches found"},
			},
			[]string{"files", "count"},
		),
	}
}

// Parse reads fd output and returns structured data.
func (p *FDParser) Parse(r io.Reader) (domain.ParseResult, error) {
	output := &FDOutput{
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

// Schema returns the JSON Schema for fd output.
func (p *FDParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *FDParser) Matches(cmd string, _ []string) bool {
	// fd is sometimes installed as fdfind on Debian/Ubuntu
	return cmd == "fd" || cmd == "fdfind"
}
