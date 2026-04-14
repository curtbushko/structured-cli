package lint

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// eslintIssuePattern matches "  line:column  severity  message  rule"
// ESLint default formatter output format with leading spaces
var eslintIssuePattern = regexp.MustCompile(`^\s+(\d+):(\d+)\s+(error|warning)\s+(.+?)\s{2,}(\S+)$`)

// eslintFilePathPattern matches a file path line (starts with / or letter:\ for Windows)
var eslintFilePathPattern = regexp.MustCompile(`^([/A-Za-z].*)$`)

// ESLintParser parses the output of 'eslint'.
// ESLint outputs issues grouped by file with format:
//
//	/path/to/file.js
//	  line:column  severity  message  rule
type ESLintParser struct {
	schema domain.Schema
}

// NewESLintParser creates a new ESLintParser with the eslint schema.
func NewESLintParser() *ESLintParser {
	return &ESLintParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/eslint.json",
			"ESLint Output",
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

// Parse reads eslint output and returns structured data in compact format.
// For clean code, eslint produces no output.
func (p *ESLintParser) Parse(r io.Reader) (domain.ParseResult, error) {
	return BuildCompactParseResult(r, eslintToCompactIssues, func(data CompactResultData) *ESLintResultCompact {
		return &ESLintResultCompact{
			TotalIssues:     data.TotalIssues,
			FilesWithIssues: data.FilesWithIssues,
			SeverityCounts:  data.SeverityCounts,
			Results:         data.Results,
			Truncated:       data.Truncated,
		}
	})
}

// eslintToCompactIssues parses eslint output and converts to CompactIssue.
func eslintToCompactIssues(raw string) []CompactIssue {
	eslintIssues := parseESLintOutput(raw)
	compactIssues := make([]CompactIssue, 0, len(eslintIssues))
	for _, issue := range eslintIssues {
		compactIssues = append(compactIssues, CompactIssue{
			File:     issue.File,
			Line:     issue.Line,
			Severity: StandardizeSeverity(issue.Severity),
			Message:  TruncateMessage(issue.Message),
			RuleID:   issue.Rule,
		})
	}
	return compactIssues
}

// parseESLintOutput parses the ESLint output and extracts issues.
func parseESLintOutput(output string) []ESLintIssue {
	var issues []ESLintIssue
	var currentFile string

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Skip summary lines (start with special characters)
		if strings.HasPrefix(line, "✖") || strings.HasPrefix(line, "✓") {
			continue
		}

		// Check if this is a file path line
		if eslintFilePathPattern.MatchString(line) && !strings.HasPrefix(strings.TrimSpace(line), " ") {
			// Only treat as file path if it doesn't start with whitespace
			trimmed := strings.TrimSpace(line)
			if trimmed == line {
				currentFile = line
				continue
			}
		}

		// Try to parse as an issue line
		if issue := parseESLintIssueLine(line, currentFile); issue != nil {
			issues = append(issues, *issue)
		}
	}

	return issues
}

// parseESLintIssueLine attempts to parse a single line as an ESLint issue.
// Returns nil if the line is not a recognized issue format.
func parseESLintIssueLine(line, currentFile string) *ESLintIssue {
	if matches := eslintIssuePattern.FindStringSubmatch(line); matches != nil {
		lineNum, _ := strconv.Atoi(matches[1])
		colNum, _ := strconv.Atoi(matches[2])
		return &ESLintIssue{
			File:     currentFile,
			Line:     lineNum,
			Column:   colNum,
			Severity: matches[3],
			Message:  strings.TrimSpace(matches[4]),
			Rule:     matches[5],
		}
	}
	return nil
}

// Schema returns the JSON Schema for eslint output.
func (p *ESLintParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *ESLintParser) Matches(cmd string, subcommands []string) bool {
	return cmd == "eslint"
}
