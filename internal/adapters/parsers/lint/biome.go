package lint

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// biomeIssuePattern matches "file:line:column category[/rule] message"
// Examples:
//   - src/index.js:10:5 lint/suspicious/noExplicitAny Unexpected any.
//   - src/index.js:10:5 format The file is not formatted.
var biomeIssuePattern = regexp.MustCompile(`^(.+):(\d+):(\d+)\s+(lint|format|parse)(?:/(\S+))?\s+(.+)$`)

// BiomeParser parses the output of 'biome check'.
// Biome outputs issues in format: file:line:column category/rule message
type BiomeParser struct {
	schema domain.Schema
}

// NewBiomeParser creates a new BiomeParser with the biome schema.
func NewBiomeParser() *BiomeParser {
	return &BiomeParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/biome.json",
			"Biome Check Output",
			"object",
			map[string]domain.PropertySchema{
				"success": {Type: "boolean", Description: "Whether the check completed without issues"},
				"issues":  {Type: "array", Description: "List of issues found"},
			},
			[]string{"success", "issues"},
		),
	}
}

// Parse reads biome check output and returns structured data.
func (p *BiomeParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &BiomeResult{
		Success: true,
		Issues:  []BiomeIssue{},
	}

	result.Issues = parseBiomeOutput(raw)

	if len(result.Issues) > 0 {
		result.Success = false
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// parseBiomeOutput extracts issues from biome output.
func parseBiomeOutput(output string) []BiomeIssue {
	var issues []BiomeIssue

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		if issue := parseBiomeIssueLine(line); issue != nil {
			issues = append(issues, *issue)
		}
	}

	return issues
}

// parseBiomeIssueLine attempts to parse a single line as a Biome issue.
// Returns nil if the line is not a recognized issue format.
func parseBiomeIssueLine(line string) *BiomeIssue {
	if matches := biomeIssuePattern.FindStringSubmatch(line); matches != nil {
		lineNum, _ := strconv.Atoi(matches[2])
		colNum, _ := strconv.Atoi(matches[3])
		return &BiomeIssue{
			File:     matches[1],
			Line:     lineNum,
			Column:   colNum,
			Severity: "error",
			Category: matches[4],
			Rule:     matches[5], // May be empty for format issues
			Message:  strings.TrimSpace(matches[6]),
		}
	}
	return nil
}

// Schema returns the JSON Schema for biome output.
func (p *BiomeParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *BiomeParser) Matches(cmd string, subcommands []string) bool {
	return cmd == "biome"
}
