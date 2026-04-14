package lint

import (
	"io"
	"sort"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Severity-related string constants for standardization.
const (
	severityNote = "note"
)

// CompactIssue represents a generic lint issue with fields common to all linters.
// This type is used as input for the helper functions that convert linter-specific
// issues to the compact output format.
type CompactIssue struct {
	// File is the source file where the issue was found.
	File string

	// Line is the line number where the issue was found.
	Line int

	// Severity is the severity level of the issue.
	Severity string

	// Message is the description of the issue.
	Message string

	// RuleID is the linter rule that was violated.
	RuleID string
}

// GroupIssuesByFile groups a slice of issues by their file path.
// It returns a slice of FileIssueGroup where each group contains
// all issues for a single file in compact tuple format.
func GroupIssuesByFile(issues []CompactIssue) []FileIssueGroup {
	if len(issues) == 0 {
		return []FileIssueGroup{}
	}

	// Use a map to group issues by file, maintaining insertion order
	fileOrder := []string{}
	fileMap := make(map[string][]IssueTuple)

	for _, issue := range issues {
		if _, exists := fileMap[issue.File]; !exists {
			fileOrder = append(fileOrder, issue.File)
			fileMap[issue.File] = []IssueTuple{}
		}
		tuple := IssueTuple{issue.Line, issue.Severity, issue.Message, issue.RuleID}
		fileMap[issue.File] = append(fileMap[issue.File], tuple)
	}

	// Build result in insertion order
	result := make([]FileIssueGroup, 0, len(fileOrder))
	for _, file := range fileOrder {
		tuples := fileMap[file]
		group := FileIssueGroup{file, len(tuples), tuples}
		result = append(result, group)
	}

	return result
}

// TruncateIssues applies truncation limits to a slice of issues.
// It sorts files by error count (errors first), limits to MaxTotalIssues total,
// and limits to MaxIssuesPerFile per file.
// Returns the truncated issues and the count of issues that were omitted.
func TruncateIssues(issues []CompactIssue) ([]CompactIssue, int) {
	if len(issues) == 0 {
		return []CompactIssue{}, 0
	}

	// Count errors per file for sorting
	fileErrorCounts := make(map[string]int)
	fileIssues := make(map[string][]CompactIssue)

	for _, issue := range issues {
		fileIssues[issue.File] = append(fileIssues[issue.File], issue)
		if issue.Severity == SeverityError {
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
	var result []CompactIssue
	truncated := 0
	totalCount := 0

	for _, file := range files {
		fileIssueList := fileIssues[file]
		perFileCount := 0

		for _, issue := range fileIssueList {
			if totalCount >= MaxTotalIssues {
				truncated++
				continue
			}
			if perFileCount >= MaxIssuesPerFile {
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

// StandardizeSeverity maps linter-specific severity strings to standard values.
// It normalizes various severity names used by different linters to one of:
// SeverityError, SeverityWarning, SeverityInfo, or SeverityStyle.
// Unknown values default to SeverityWarning.
func StandardizeSeverity(severity string) string {
	lower := strings.ToLower(severity)

	switch lower {
	case "error", "err", "fatal", "critical":
		return SeverityError
	case "warning", "warn":
		return SeverityWarning
	case "info", "information", severityNote, "message":
		return SeverityInfo
	case "style", "hint", "convention", "refactor":
		return SeverityStyle
	default:
		// Default to warning for unknown severity levels
		return SeverityWarning
	}
}

// TruncateMessage truncates multi-line messages to the first line.
// If the message contains multiple lines, returns the first line with "..." appended.
// Single-line messages are returned unchanged.
func TruncateMessage(msg string) string {
	if msg == "" {
		return msg
	}

	// Check for newline (handle both Unix and Windows line endings)
	idx := strings.IndexAny(msg, "\r\n")
	if idx == -1 {
		return msg
	}

	return msg[:idx] + "..."
}

// RuffCodeToSeverity maps Ruff rule codes to standard severity levels.
// E* = error, W* = warning, F* = warning (flake8), I* = info
func RuffCodeToSeverity(code string) string {
	if len(code) == 0 {
		return SeverityWarning
	}

	switch code[0] {
	case 'E':
		return SeverityError
	case 'W', 'F':
		return SeverityWarning
	case 'I':
		return SeverityInfo
	default:
		return SeverityWarning
	}
}

// TruncateMessageLength truncates a message to a maximum length.
// If truncated, appends "..." to indicate truncation.
func TruncateMessageLength(msg string, maxLen int) string {
	if len(msg) <= maxLen {
		return msg
	}
	if maxLen <= 3 {
		return "..."
	}
	return msg[:maxLen-3] + "..."
}

// CompactResultData holds the computed data for compact format output.
// It is returned by ProcessCompactResult and can be used to build
// linter-specific result structs.
type CompactResultData struct {
	TotalIssues     int
	FilesWithIssues int
	SeverityCounts  map[string]int
	Results         []FileIssueGroup
	Truncated       int
}

// ProcessCompactResult reads input, converts issues to compact format, and returns the computed data.
// The parseFunc parameter extracts linter-specific issues and converts them to CompactIssue.
// This function handles the common pattern of truncation, severity counting, and file grouping.
func ProcessCompactResult(r io.Reader, parseFunc func(raw string) []CompactIssue) (CompactResultData, string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return CompactResultData{}, "", err
	}

	raw := string(data)
	compactIssues := parseFunc(raw)

	// Apply truncation
	truncatedIssues, truncatedCount := TruncateIssues(compactIssues)

	// Count severities
	severityCounts := make(map[string]int)
	for _, issue := range truncatedIssues {
		severityCounts[issue.Severity]++
	}

	// Group issues by file
	results := GroupIssuesByFile(truncatedIssues)

	return CompactResultData{
		TotalIssues:     len(compactIssues),
		FilesWithIssues: len(results),
		SeverityCounts:  severityCounts,
		Results:         results,
		Truncated:       truncatedCount,
	}, raw, nil
}

// BuildCompactParseResult is a convenience function that builds a domain.ParseResult
// from CompactResultData using the provided result builder function.
func BuildCompactParseResult[T any](r io.Reader, parseFunc func(raw string) []CompactIssue, buildResult func(data CompactResultData) *T) (domain.ParseResult, error) {
	compactData, raw, err := ProcessCompactResult(r, parseFunc)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	result := buildResult(compactData)
	return domain.NewParseResult(result, raw, 0), nil
}
