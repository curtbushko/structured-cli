package lint

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// golangcilintIssueWithColumnPattern matches "file:line:column: message (linter)"
var golangcilintIssueWithColumnPattern = regexp.MustCompile(`^(.+):(\d+):(\d+):\s*(.+?)\s+\((\w+)\)$`)

// golangcilintIssueNoColumnPattern matches "file:line: message (linter)"
var golangcilintIssueNoColumnPattern = regexp.MustCompile(`^(.+):(\d+):\s*(.+?)\s+\((\w+)\)$`)

// GolangCILintParser parses the output of 'golangci-lint run'.
// golangci-lint outputs issues in format: file:line:column: message (linter)
type GolangCILintParser struct {
	schema domain.Schema
}

// NewGolangCILintParser creates a new GolangCILintParser with the golangci-lint schema.
func NewGolangCILintParser() *GolangCILintParser {
	return &GolangCILintParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/golangci-lint.json",
			"golangci-lint Output",
			"object",
			map[string]domain.PropertySchema{
				"success": {Type: "boolean", Description: "Whether linting completed without issues"},
				"issues":  {Type: "array", Description: "List of lint issues found"},
			},
			[]string{"success", "issues"},
		),
	}
}

// Parse reads golangci-lint run output and returns structured data.
func (p *GolangCILintParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &GolangCILintResult{
		Success: true,
		Issues:  []GolangCILintIssue{},
	}

	result.Issues = parseGolangCILintOutput(raw)

	if len(result.Issues) > 0 {
		result.Success = false
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// parseGolangCILintOutput extracts issues from golangci-lint output.
func parseGolangCILintOutput(output string) []GolangCILintIssue {
	var issues []GolangCILintIssue

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		if issue := parseGolangCILintIssueLine(line); issue != nil {
			issues = append(issues, *issue)
		}
	}

	return issues
}

// parseGolangCILintIssueLine attempts to parse a single line as a golangci-lint issue.
// Returns nil if the line is not a recognized issue format.
func parseGolangCILintIssueLine(line string) *GolangCILintIssue {
	// Try file:line:column: message (linter) format first
	if matches := golangcilintIssueWithColumnPattern.FindStringSubmatch(line); matches != nil {
		lineNum, _ := strconv.Atoi(matches[2])
		colNum, _ := strconv.Atoi(matches[3])
		return &GolangCILintIssue{
			File:     matches[1],
			Line:     lineNum,
			Column:   colNum,
			Message:  strings.TrimSpace(matches[4]),
			Linter:   matches[5],
			Severity: "error",
		}
	}

	// Try file:line: message (linter) format (no column)
	if matches := golangcilintIssueNoColumnPattern.FindStringSubmatch(line); matches != nil {
		lineNum, _ := strconv.Atoi(matches[2])
		return &GolangCILintIssue{
			File:     matches[1],
			Line:     lineNum,
			Column:   0,
			Message:  strings.TrimSpace(matches[3]),
			Linter:   matches[4],
			Severity: "error",
		}
	}

	return nil
}

// Schema returns the JSON Schema for golangci-lint output.
func (p *GolangCILintParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *GolangCILintParser) Matches(cmd string, subcommands []string) bool {
	return cmd == "golangci-lint"
}
