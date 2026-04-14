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

// Parse reads ruff check output and returns structured data in compact format.
func (p *RuffParser) Parse(r io.Reader) (domain.ParseResult, error) {
	return BuildCompactParseResult(r, ruffToCompactIssues, func(data CompactResultData) *RuffResultCompact {
		return &RuffResultCompact{
			TotalIssues:     data.TotalIssues,
			FilesWithIssues: data.FilesWithIssues,
			SeverityCounts:  data.SeverityCounts,
			Results:         data.Results,
			Truncated:       data.Truncated,
		}
	})
}

// ruffToCompactIssues parses ruff output and converts to CompactIssue.
func ruffToCompactIssues(raw string) []CompactIssue {
	ruffIssues := parseRuffOutput(raw)
	compactIssues := make([]CompactIssue, 0, len(ruffIssues))
	for _, issue := range ruffIssues {
		compactIssues = append(compactIssues, CompactIssue{
			File:     issue.File,
			Line:     issue.Line,
			Severity: RuffCodeToSeverity(issue.Code),
			Message:  TruncateMessage(issue.Message),
			RuleID:   issue.Code, // Use rule code as rule ID
		})
	}
	return compactIssues
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
