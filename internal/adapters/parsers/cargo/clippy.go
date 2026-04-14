package cargo

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"sort"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Severity level constants for Clippy issues.
const (
	severityError   = "error"
	severityWarning = "warning"
	severityInfo    = "info"
)

// Truncation limit constants for compact output format.
const (
	maxTotalIssues   = 200
	maxIssuesPerFile = 20
)

// ClippyParser parses the JSON output of 'cargo clippy --message-format=json'.
type ClippyParser struct {
	schema domain.Schema
}

// NewClippyParser creates a new ClippyParser with the cargo clippy schema.
func NewClippyParser() *ClippyParser {
	return &ClippyParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/cargo-clippy.json",
			"Cargo Clippy Output",
			"object",
			map[string]domain.PropertySchema{
				"total_issues":      {Type: "integer", Description: "Total count of all lint issues"},
				"files_with_issues": {Type: "integer", Description: "Count of files with at least one issue"},
				"severity_counts":   {Type: "object", Description: "Maps severity levels to their counts"},
				"results":           {Type: "array", Description: "Grouped issues by file in compact tuple format"},
				"truncated":         {Type: "integer", Description: "Number of issues omitted due to truncation"},
			},
			[]string{"total_issues", "files_with_issues", "severity_counts", "results", "truncated"},
		),
	}
}

// clippyMessage represents a single JSON message from clippy output.
type clippyMessage struct {
	Reason  string           `json:"reason"`
	Success *bool            `json:"success,omitempty"`
	Message *clippyDiagMsg   `json:"message,omitempty"`
	Target  *clippyTargetMsg `json:"target,omitempty"`
}

// clippyTargetMsg represents target information in clippy JSON output.
type clippyTargetMsg struct {
	Kind       []string `json:"kind"`
	CrateTypes []string `json:"crate_types"`
	Name       string   `json:"name"`
	SrcPath    string   `json:"src_path"`
	Edition    string   `json:"edition"`
}

// clippyDiagMsg represents a diagnostic message from clippy.
type clippyDiagMsg struct {
	Message  string          `json:"message"`
	Code     *clippyCodeMsg  `json:"code,omitempty"`
	Level    string          `json:"level"`
	Spans    []clippySpanMsg `json:"spans"`
	Children []clippyDiagMsg `json:"children,omitempty"`
	Rendered string          `json:"rendered,omitempty"`
}

// clippyCodeMsg represents the diagnostic code.
type clippyCodeMsg struct {
	Code        string `json:"code"`
	Explanation string `json:"explanation,omitempty"`
}

// clippySpanMsg represents a source code span in a diagnostic.
type clippySpanMsg struct {
	FileName    string `json:"file_name"`
	ByteStart   int    `json:"byte_start"`
	ByteEnd     int    `json:"byte_end"`
	LineStart   int    `json:"line_start"`
	LineEnd     int    `json:"line_end"`
	ColumnStart int    `json:"column_start"`
	ColumnEnd   int    `json:"column_end"`
	IsPrimary   bool   `json:"is_primary"`
	Label       string `json:"label,omitempty"`
}

// clippyCompactIssue represents an intermediate issue for processing.
type clippyCompactIssue struct {
	File     string
	Line     int
	Severity string
	Message  string
	RuleID   string
}

// Parse reads cargo clippy JSON output and returns structured data in compact format.
func (p *ClippyParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)
	issues := parseClippyIssues(raw)

	// Apply truncation
	truncatedIssues, truncatedCount := truncateClippyIssues(issues)

	// Count severities
	severityCounts := make(map[string]int)
	for _, issue := range truncatedIssues {
		severityCounts[issue.Severity]++
	}

	// Group issues by file
	results := groupClippyIssuesByFile(truncatedIssues)

	result := &ClippyResultCompact{
		TotalIssues:     len(issues),
		FilesWithIssues: len(results),
		SeverityCounts:  severityCounts,
		Results:         results,
		Truncated:       truncatedCount,
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// parseClippyIssues extracts all issues from raw clippy JSON output.
func parseClippyIssues(raw string) []clippyCompactIssue {
	var issues []clippyCompactIssue

	scanner := bufio.NewScanner(bytes.NewReader([]byte(raw)))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var msg clippyMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}

		if msg.Reason == reasonCompilerMessage && msg.Message != nil {
			issue := extractClippyIssue(msg.Message)
			if issue != nil {
				issues = append(issues, *issue)
			}
		}
	}

	return issues
}

// extractClippyIssue extracts a clippyCompactIssue from a clippy diagnostic message.
// Returns nil for diagnostics that should be skipped (e.g., notes without location).
func extractClippyIssue(diag *clippyDiagMsg) *clippyCompactIssue {
	// Skip diagnostics that aren't errors or warnings
	if diag.Level != levelError && diag.Level != levelErrorICE && diag.Level != levelWarning && diag.Level != "note" {
		return nil
	}

	// Find primary span for location
	var file string
	var line int
	for _, span := range diag.Spans {
		if span.IsPrimary {
			file = span.FileName
			line = span.LineStart
			break
		}
	}

	// Skip issues without a file location
	if file == "" {
		return nil
	}

	// Extract lint code (rule ID)
	var code string
	if diag.Code != nil {
		code = diag.Code.Code
	}

	// Map Clippy severity to standard severity
	severity := mapClippySeverity(diag.Level)

	// Use only the main message, excluding help text from children
	message := truncateClippyMessage(diag.Message)

	return &clippyCompactIssue{
		File:     file,
		Line:     line,
		Severity: severity,
		Message:  message,
		RuleID:   code,
	}
}

// mapClippySeverity maps Clippy diagnostic levels to standard severity.
// Clippy uses: error, warning, note, help
// We map: error -> error, warning -> warning, note -> info
func mapClippySeverity(level string) string {
	switch level {
	case levelError, levelErrorICE:
		return severityError
	case levelWarning:
		return severityWarning
	case "note":
		return severityInfo
	default:
		return severityWarning
	}
}

// truncateClippyMessage truncates multi-line messages to the first line.
func truncateClippyMessage(msg string) string {
	if msg == "" {
		return msg
	}
	idx := strings.IndexAny(msg, "\r\n")
	if idx == -1 {
		return msg
	}
	return msg[:idx] + "..."
}

// truncateClippyIssues applies truncation limits to a slice of issues.
func truncateClippyIssues(issues []clippyCompactIssue) ([]clippyCompactIssue, int) {
	if len(issues) == 0 {
		return []clippyCompactIssue{}, 0
	}

	// Count errors per file for sorting
	fileErrorCounts := make(map[string]int)
	fileIssues := make(map[string][]clippyCompactIssue)

	for _, issue := range issues {
		fileIssues[issue.File] = append(fileIssues[issue.File], issue)
		if issue.Severity == severityError {
			fileErrorCounts[issue.File]++
		}
	}

	// Sort files by error count (descending)
	files := make([]string, 0, len(fileIssues))
	for file := range fileIssues {
		files = append(files, file)
	}
	sort.Slice(files, func(i, j int) bool {
		return fileErrorCounts[files[i]] > fileErrorCounts[files[j]]
	})

	// Apply per-file limit and total limit
	var result []clippyCompactIssue
	truncated := 0
	totalCount := 0

	for _, file := range files {
		fileIssueList := fileIssues[file]
		perFileCount := 0

		for _, issue := range fileIssueList {
			if totalCount >= maxTotalIssues {
				truncated++
				continue
			}
			if perFileCount >= maxIssuesPerFile {
				truncated++
				continue
			}
			result = append(result, issue)
			totalCount++
			perFileCount++
		}
	}

	return result, truncated
}

// groupClippyIssuesByFile groups issues by their file path.
func groupClippyIssuesByFile(issues []clippyCompactIssue) []ClippyFileIssueGroup {
	if len(issues) == 0 {
		return []ClippyFileIssueGroup{}
	}

	// Use a map to group issues by file, maintaining insertion order
	fileOrder := []string{}
	fileMap := make(map[string][]ClippyIssueTuple)

	for _, issue := range issues {
		if _, exists := fileMap[issue.File]; !exists {
			fileOrder = append(fileOrder, issue.File)
			fileMap[issue.File] = []ClippyIssueTuple{}
		}
		tuple := ClippyIssueTuple{issue.Line, issue.Severity, issue.Message, issue.RuleID}
		fileMap[issue.File] = append(fileMap[issue.File], tuple)
	}

	// Build result in insertion order
	result := make([]ClippyFileIssueGroup, 0, len(fileOrder))
	for _, file := range fileOrder {
		tuples := fileMap[file]
		group := ClippyFileIssueGroup{file, len(tuples), tuples}
		result = append(result, group)
	}

	return result
}

// Schema returns the JSON Schema for cargo clippy output.
func (p *ClippyParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *ClippyParser) Matches(cmd string, subcommands []string) bool {
	if cmd != cmdCargo {
		return false
	}
	if len(subcommands) == 0 {
		return false
	}
	return subcommands[0] == "clippy"
}
