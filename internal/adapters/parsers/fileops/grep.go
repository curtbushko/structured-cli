package fileops

import (
	"bufio"
	"io"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// GrepParser parses the output of 'grep' command.
type GrepParser struct {
	schema domain.Schema
}

// NewGrepParser creates a new GrepParser with the grep schema.
func NewGrepParser() *GrepParser {
	return &GrepParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/grep.json",
			"Grep Output",
			"object",
			map[string]domain.PropertySchema{
				"matches":      {Type: "array", Description: "List of matches found"},
				"count":        {Type: "integer", Description: "Total number of matches"},
				"filesMatched": {Type: "integer", Description: "Number of files with matches"},
			},
			[]string{"matches", "count", "filesMatched"},
		),
	}
}

// Parse reads grep output and returns structured data.
func (p *GrepParser) Parse(r io.Reader) (domain.ParseResult, error) {
	output := &GrepOutput{
		Matches: []GrepMatch{},
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

		// Skip "Binary file X matches" lines
		if strings.HasPrefix(line, "Binary file ") && strings.HasSuffix(line, " matches") {
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

// Schema returns the JSON Schema for grep output.
func (p *GrepParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *GrepParser) Matches(cmd string, _ []string) bool {
	return cmd == "grep" || cmd == "egrep" || cmd == "fgrep"
}

// parseLine parses a single line of grep output.
// Formats:
//   - file:line:content (with -n)
//   - file:content (without -n)
//   - line:content (single file with -n)
//   - content (single file without -n) - not distinguishable, treated as content
func (p *GrepParser) parseLine(line string) GrepMatch {
	match := GrepMatch{}

	// Try to parse as file:line:content
	colonIdx := strings.Index(line, ":")
	if colonIdx == -1 {
		// No colons - just content
		match.Content = line
		return match
	}

	firstPart := line[:colonIdx]
	rest := line[colonIdx+1:]

	// Check if first part is a number (line number for single file)
	if lineNum, err := strconv.Atoi(firstPart); err == nil {
		match.Line = lineNum
		match.Content = rest
		return match
	}

	// First part is a filename
	match.File = firstPart

	// Check for line number in rest
	colonIdx2 := strings.Index(rest, ":")
	if colonIdx2 != -1 {
		possibleLineNum := rest[:colonIdx2]
		if lineNum, err := strconv.Atoi(possibleLineNum); err == nil {
			match.Line = lineNum
			match.Content = rest[colonIdx2+1:]
			return match
		}
	}

	// No line number, rest is content
	match.Content = rest
	return match
}
