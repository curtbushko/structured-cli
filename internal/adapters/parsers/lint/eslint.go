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
				"success": {Type: "boolean", Description: "Whether linting completed without issues"},
				"issues":  {Type: "array", Description: "List of lint issues found"},
			},
			[]string{"success", "issues"},
		),
	}
}

// Parse reads eslint output and returns structured data.
// For clean code, eslint produces no output.
func (p *ESLintParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &ESLintResult{
		Success: true,
		Issues:  []ESLintIssue{},
	}

	result.Issues = parseESLintOutput(raw)

	if len(result.Issues) > 0 {
		result.Success = false
	}

	return domain.NewParseResult(result, raw, 0), nil
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
