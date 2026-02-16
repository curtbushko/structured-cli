package fileops

import (
	"bufio"
	"io"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// RipgrepParser parses the output of 'rg' (ripgrep) command.
type RipgrepParser struct {
	schema domain.Schema
}

// NewRipgrepParser creates a new RipgrepParser with the ripgrep schema.
func NewRipgrepParser() *RipgrepParser {
	return &RipgrepParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/ripgrep.json",
			"Ripgrep Output",
			"object",
			map[string]domain.PropertySchema{
				"matches":      {Type: "array", Description: "List of matches found"},
				"count":        {Type: "integer", Description: "Total number of matches"},
				"filesMatched": {Type: "integer", Description: "Number of files with matches"},
				"stats":        {Type: "object", Description: "Search statistics"},
			},
			[]string{"matches", "count", "filesMatched"},
		),
	}
}

// Parse reads ripgrep output and returns structured data.
func (p *RipgrepParser) Parse(r io.Reader) (domain.ParseResult, error) {
	output := &RipgrepOutput{
		Matches: []RipgrepMatch{},
	}

	filesSet := make(map[string]struct{})
	scanner := bufio.NewScanner(r)
	var rawBuilder strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		rawBuilder.WriteString(line)
		rawBuilder.WriteString("\n")

		if line == "" {
			continue
		}

		match := p.parseLine(line)
		output.Matches = append(output.Matches, match)

		if match.File != "" {
			filesSet[match.File] = struct{}{}
		}
	}

	if err := scanner.Err(); err != nil {
		return domain.NewParseResultWithError(err, rawBuilder.String(), 0), nil
	}

	output.Count = len(output.Matches)
	output.FilesMatched = len(filesSet)

	return domain.NewParseResult(output, rawBuilder.String(), 0), nil
}

// Schema returns the JSON Schema for ripgrep output.
func (p *RipgrepParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *RipgrepParser) Matches(cmd string, _ []string) bool {
	return cmd == "rg" || cmd == "ripgrep"
}

// parseLine parses a single line of ripgrep output.
// Formats:
//   - file:line:column:content (with --column)
//   - file:line:content (default)
//   - file (with --files-with-matches)
func (p *RipgrepParser) parseLine(line string) RipgrepMatch {
	match := RipgrepMatch{}

	colonIdx := strings.Index(line, ":")
	if colonIdx == -1 {
		// No colons - file only (--files-with-matches output)
		match.File = line
		return match
	}

	// First part is always the filename
	match.File = line[:colonIdx]
	rest := line[colonIdx+1:]

	// Try to parse line number
	colonIdx2 := strings.Index(rest, ":")
	if colonIdx2 == -1 {
		// No more colons - rest is content
		match.Content = rest
		return match
	}

	lineNumStr := rest[:colonIdx2]
	lineNum, err := strconv.Atoi(lineNumStr)
	if err != nil {
		// Not a number, treat rest as content
		match.Content = rest
		return match
	}
	match.Line = lineNum
	rest = rest[colonIdx2+1:]

	// Try to parse column number
	colonIdx3 := strings.Index(rest, ":")
	if colonIdx3 != -1 {
		colNumStr := rest[:colonIdx3]
		colNum, err := strconv.Atoi(colNumStr)
		if err == nil {
			match.Column = colNum
			rest = rest[colonIdx3+1:]
		}
	}

	match.Content = rest
	return match
}
