package lint

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// mypyErrorWithCodePattern matches "file:line: severity: message  [error-code]"
var mypyErrorWithCodePattern = regexp.MustCompile(`^(.+):(\d+):\s*(error|warning|note):\s*(.+?)\s+\[([^\]]+)\]$`)

// mypyErrorNoCodePattern matches "file:line: severity: message"
var mypyErrorNoCodePattern = regexp.MustCompile(`^(.+):(\d+):\s*(error|warning|note):\s*(.+)$`)

// mypySummaryPattern matches "Found X errors" or "Success: no issues"
var mypySummaryPattern = regexp.MustCompile(`^(Found \d+ error|Success: no issues)`)

// MypyParser parses the output of 'mypy'.
// Mypy outputs errors in format: file:line: severity: message [error-code]
type MypyParser struct {
	schema domain.Schema
}

// NewMypyParser creates a new MypyParser with the mypy schema.
func NewMypyParser() *MypyParser {
	return &MypyParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/mypy.json",
			"Mypy Type Check Output",
			"object",
			map[string]domain.PropertySchema{
				"total_issues":      {Type: "integer", Description: "Total count of all type errors"},
				"files_with_issues": {Type: "integer", Description: "Count of files with at least one error"},
				"severity_counts":   {Type: "object", Description: "Map of severity levels to their counts"},
				"results":           {Type: "array", Description: "Errors grouped by file in compact tuple format"},
				"truncated":         {Type: "integer", Description: "Count of errors omitted due to truncation limits"},
				"summary":           {Type: "string", Description: "Summary line from mypy output"},
			},
			[]string{"total_issues", "files_with_issues", "severity_counts", "results", "truncated", "summary"},
		),
	}
}

// maxMypyMessageLength is the maximum length for mypy messages (truncate verbose type signatures)
const maxMypyMessageLength = 100

// Parse reads mypy output and returns structured data in compact format.
func (p *MypyParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	// Parse errors using existing logic
	mypyErrors, summary := parseMypyOutput(raw)

	// Convert MypyError to CompactIssue for helper functions
	compactIssues := make([]CompactIssue, 0, len(mypyErrors))
	for _, mypyErr := range mypyErrors {
		// Map mypy severity: error->error, warning->warning, note->info
		severity := mypySeverityToStandard(mypyErr.Severity)

		// Truncate verbose type signatures in messages
		message := TruncateMessageLength(mypyErr.Message, maxMypyMessageLength)

		compactIssues = append(compactIssues, CompactIssue{
			File:     mypyErr.File,
			Line:     mypyErr.Line,
			Severity: severity,
			Message:  message,
			RuleID:   mypyErr.Code, // Use error code as rule ID
		})
	}

	// Apply truncation
	truncatedIssues, truncatedCount := TruncateIssues(compactIssues)

	// Count severities
	severityCounts := make(map[string]int)
	for _, issue := range truncatedIssues {
		severityCounts[issue.Severity]++
	}

	// Group errors by file
	results := GroupIssuesByFile(truncatedIssues)

	result := &MypyResultCompact{
		TotalIssues:     len(mypyErrors),
		FilesWithIssues: len(results),
		SeverityCounts:  severityCounts,
		Results:         results,
		Truncated:       truncatedCount,
		Summary:         summary,
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// mypySeverityToStandard maps mypy severity to standard severity levels.
func mypySeverityToStandard(severity string) string {
	switch severity {
	case "error":
		return SeverityError
	case "warning":
		return SeverityWarning
	case "note":
		return SeverityInfo
	default:
		return SeverityWarning
	}
}

// parseMypyOutput extracts errors and summary from mypy output.
func parseMypyOutput(output string) ([]MypyError, string) {
	var errors []MypyError
	var summary string

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Check if this is a summary line
		if mypySummaryPattern.MatchString(line) {
			summary = line
			continue
		}

		if mypyErr := parseMypyErrorLine(line); mypyErr != nil {
			errors = append(errors, *mypyErr)
		}
	}

	return errors, summary
}

// parseMypyErrorLine attempts to parse a single line as a mypy error.
// Returns nil if the line is not a recognized error format.
func parseMypyErrorLine(line string) *MypyError {
	// Try file:line: severity: message [error-code] format first
	if matches := mypyErrorWithCodePattern.FindStringSubmatch(line); matches != nil {
		lineNum, _ := strconv.Atoi(matches[2])
		return &MypyError{
			File:     matches[1],
			Line:     lineNum,
			Severity: matches[3],
			Message:  strings.TrimSpace(matches[4]),
			Code:     matches[5],
		}
	}

	// Try file:line: severity: message format (no code)
	if matches := mypyErrorNoCodePattern.FindStringSubmatch(line); matches != nil {
		lineNum, _ := strconv.Atoi(matches[2])
		return &MypyError{
			File:     matches[1],
			Line:     lineNum,
			Severity: matches[3],
			Message:  strings.TrimSpace(matches[4]),
			Code:     "",
		}
	}

	return nil
}

// Schema returns the JSON Schema for mypy output.
func (p *MypyParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *MypyParser) Matches(cmd string, subcommands []string) bool {
	return cmd == "mypy"
}
