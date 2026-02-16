package lint

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// ruffIssuePattern matches "file:line:column: CODE message"
var ruffIssuePattern = regexp.MustCompile(`^(.+):(\d+):(\d+):\s*([A-Z]+\d+)\s+(.+)$`)

// RuffParser parses the output of 'ruff check'.
// Ruff outputs issues in format: file:line:column: CODE message
type RuffParser struct {
	schema domain.Schema
}

// NewRuffParser creates a new RuffParser with the ruff schema.
func NewRuffParser() *RuffParser {
	return &RuffParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/ruff.json",
			"Ruff Check Output",
			"object",
			map[string]domain.PropertySchema{
				"success": {Type: "boolean", Description: "Whether linting completed without issues"},
				"issues":  {Type: "array", Description: "List of lint issues found"},
			},
			[]string{"success", "issues"},
		),
	}
}

// Parse reads ruff check output and returns structured data.
func (p *RuffParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &RuffResult{
		Success: true,
		Issues:  []RuffIssue{},
	}

	result.Issues = parseRuffOutput(raw)

	if len(result.Issues) > 0 {
		result.Success = false
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// parseRuffOutput extracts issues from ruff output.
func parseRuffOutput(output string) []RuffIssue {
	var issues []RuffIssue

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		if issue := parseRuffIssueLine(line); issue != nil {
			issues = append(issues, *issue)
		}
	}

	return issues
}

// parseRuffIssueLine attempts to parse a single line as a Ruff issue.
// Returns nil if the line is not a recognized issue format.
func parseRuffIssueLine(line string) *RuffIssue {
	if matches := ruffIssuePattern.FindStringSubmatch(line); matches != nil {
		lineNum, _ := strconv.Atoi(matches[2])
		colNum, _ := strconv.Atoi(matches[3])
		return &RuffIssue{
			File:    matches[1],
			Line:    lineNum,
			Column:  colNum,
			Code:    matches[4],
			Message: strings.TrimSpace(matches[5]),
		}
	}
	return nil
}

// Schema returns the JSON Schema for ruff output.
func (p *RuffParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *RuffParser) Matches(cmd string, subcommands []string) bool {
	return cmd == "ruff"
}
