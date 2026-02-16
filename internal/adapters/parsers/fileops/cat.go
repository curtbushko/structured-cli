package fileops

import (
	"io"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// CatParser parses the output of 'cat' command.
type CatParser struct {
	schema domain.Schema
}

// NewCatParser creates a new CatParser with the cat schema.
func NewCatParser() *CatParser {
	return &CatParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/cat.json",
			"Cat Output",
			"object",
			map[string]domain.PropertySchema{
				"content": {Type: "string", Description: "File content"},
				"lines":   {Type: "integer", Description: "Number of lines"},
				"bytes":   {Type: "integer", Description: "Number of bytes"},
			},
			[]string{"content", "bytes"},
		),
	}
}

// Parse reads cat output and returns structured data.
func (p *CatParser) Parse(r io.Reader) (domain.ParseResult, error) {
	content, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	contentStr := string(content)
	output := &CatOutput{
		Content: contentStr,
		Bytes:   len(content),
	}

	// Count lines
	if len(contentStr) > 0 {
		output.Lines = strings.Count(contentStr, "\n")
		// Add 1 if content doesn't end with newline but has content
		if len(contentStr) > 0 && !strings.HasSuffix(contentStr, "\n") {
			output.Lines++
		}
	}

	return domain.NewParseResult(output, contentStr, 0), nil
}

// Schema returns the JSON Schema for cat output.
func (p *CatParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *CatParser) Matches(cmd string, _ []string) bool {
	return cmd == "cat"
}
