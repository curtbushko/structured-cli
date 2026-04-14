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
				"total_issues":      {Type: "integer", Description: "Total count of all lint issues"},
				"files_with_issues": {Type: "integer", Description: "Count of files with at least one issue"},
				"severity_counts":   {Type: "object", Description: "Map of severity levels to their counts"},
				"results":           {Type: "array", Description: "Issues grouped by file in compact tuple format"},
				"truncated":         {Type: "integer", Description: "Count of issues omitted due to truncation limits"},
			},
			[]string{"total_issues", "files_with_issues", "severity_counts", "results", "truncated"},
		),
	}
}

// Parse reads golangci-lint run output and returns structured data in compact format.
func (p *GolangCILintParser) Parse(r io.Reader) (domain.ParseResult, error) {
	return BuildCompactParseResult(r, golangciLintToCompactIssues, func(data CompactResultData) *GolangCILintResultCompact {
		return &GolangCILintResultCompact{
			TotalIssues:     data.TotalIssues,
			FilesWithIssues: data.FilesWithIssues,
			SeverityCounts:  data.SeverityCounts,
			Results:         data.Results,
			Truncated:       data.Truncated,
		}
	})
}

// golangciLintToCompactIssues parses golangci-lint output and converts to CompactIssue.
func golangciLintToCompactIssues(raw string) []CompactIssue {
	golangciIssues := parseGolangCILintOutput(raw)
	compactIssues := make([]CompactIssue, 0, len(golangciIssues))
	for _, issue := range golangciIssues {
		compactIssues = append(compactIssues, CompactIssue{
			File:     issue.File,
			Line:     issue.Line,
			Severity: StandardizeSeverity(issue.Severity),
			Message:  TruncateMessage(issue.Message),
			RuleID:   issue.Linter, // Use linter name as rule ID
		})
	}
	return compactIssues
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
