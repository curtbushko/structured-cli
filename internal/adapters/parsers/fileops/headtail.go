package fileops

import (
	"bufio"
	"io"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// HeadParser parses the output of 'head' command.
type HeadParser struct {
	schema domain.Schema
}

// NewHeadParser creates a new HeadParser with the head schema.
func NewHeadParser() *HeadParser {
	return &HeadParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/head.json",
			"Head Output",
			"object",
			map[string]domain.PropertySchema{
				"content":   {Type: "string", Description: "File content"},
				"lines":     {Type: "array", Description: "Lines read"},
				"lineCount": {Type: "integer", Description: "Number of lines returned"},
			},
			[]string{"content", "lines", "lineCount"},
		),
	}
}

// Parse reads head output and returns structured data.
func (p *HeadParser) Parse(r io.Reader) (domain.ParseResult, error) {
	return parseHeadTail(r)
}

// Schema returns the JSON Schema for head output.
func (p *HeadParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *HeadParser) Matches(cmd string, _ []string) bool {
	return cmd == "head"
}

// TailParser parses the output of 'tail' command.
type TailParser struct {
	schema domain.Schema
}

// NewTailParser creates a new TailParser with the tail schema.
func NewTailParser() *TailParser {
	return &TailParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/tail.json",
			"Tail Output",
			"object",
			map[string]domain.PropertySchema{
				"content":   {Type: "string", Description: "File content"},
				"lines":     {Type: "array", Description: "Lines read"},
				"lineCount": {Type: "integer", Description: "Number of lines returned"},
			},
			[]string{"content", "lines", "lineCount"},
		),
	}
}

// Parse reads tail output and returns structured data.
func (p *TailParser) Parse(r io.Reader) (domain.ParseResult, error) {
	return parseHeadTail(r)
}

// Schema returns the JSON Schema for tail output.
func (p *TailParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *TailParser) Matches(cmd string, _ []string) bool {
	return cmd == "tail"
}

// parseHeadTail is the shared implementation for head and tail parsing.
func parseHeadTail(r io.Reader) (domain.ParseResult, error) {
	output := &HeadTailOutput{
		Lines: []string{},
	}

	scanner := bufio.NewScanner(r)
	var rawBuilder strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		rawBuilder.WriteString(line)
		rawBuilder.WriteString("\n")

		output.Lines = append(output.Lines, line)
	}

	if err := scanner.Err(); err != nil {
		return domain.NewParseResultWithError(err, rawBuilder.String(), 0), nil
	}

	output.Content = rawBuilder.String()
	output.LineCount = len(output.Lines)

	return domain.NewParseResult(output, rawBuilder.String(), 0), nil
}
