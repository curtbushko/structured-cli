package lint

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGroupIssuesByFile(t *testing.T) {
	tests := []struct {
		name     string
		issues   []CompactIssue
		expected []FileIssueGroup
	}{
		{
			name:     "empty issues returns empty groups",
			issues:   []CompactIssue{},
			expected: []FileIssueGroup{},
		},
		{
			name: "single issue in one file",
			issues: []CompactIssue{
				{File: "src/main.go", Line: 10, Severity: SeverityError, Message: "error msg", RuleID: "rule1"},
			},
			expected: []FileIssueGroup{
				{"src/main.go", 1, []IssueTuple{{10, SeverityError, "error msg", "rule1"}}},
			},
		},
		{
			name: "multiple issues in same file",
			issues: []CompactIssue{
				{File: "src/main.go", Line: 10, Severity: SeverityError, Message: "error 1", RuleID: "rule1"},
				{File: "src/main.go", Line: 20, Severity: SeverityWarning, Message: "warning 1", RuleID: "rule2"},
			},
			expected: []FileIssueGroup{
				{"src/main.go", 2, []IssueTuple{
					{10, SeverityError, "error 1", "rule1"},
					{20, SeverityWarning, "warning 1", "rule2"},
				}},
			},
		},
		{
			name: "issues across multiple files",
			issues: []CompactIssue{
				{File: "src/a.go", Line: 5, Severity: SeverityError, Message: "err a", RuleID: "r1"},
				{File: "src/b.go", Line: 10, Severity: SeverityWarning, Message: "warn b", RuleID: "r2"},
				{File: "src/a.go", Line: 15, Severity: SeverityInfo, Message: "info a", RuleID: "r3"},
			},
			expected: []FileIssueGroup{
				{"src/a.go", 2, []IssueTuple{
					{5, SeverityError, "err a", "r1"},
					{15, SeverityInfo, "info a", "r3"},
				}},
				{"src/b.go", 1, []IssueTuple{
					{10, SeverityWarning, "warn b", "r2"},
				}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GroupIssuesByFile(tt.issues)
			assert.Equal(t, len(tt.expected), len(result))
			// Verify each group has correct structure
			for i, expected := range tt.expected {
				if i < len(result) {
					assert.Equal(t, expected[0], result[i][0], "filename mismatch")
					assert.Equal(t, expected[1], result[i][1], "issue count mismatch")
				}
			}
		})
	}
}

func TestTruncateIssues(t *testing.T) {
	tests := []struct {
		name              string
		issues            []CompactIssue
		expectedCount     int
		expectedTruncated int
	}{
		{
			name:              "empty issues returns empty result",
			issues:            []CompactIssue{},
			expectedCount:     0,
			expectedTruncated: 0,
		},
		{
			name:              "issues under limit are not truncated",
			issues:            makeIssues(10, "file.go", SeverityError),
			expectedCount:     10,
			expectedTruncated: 0,
		},
		{
			name:              "issues over per-file limit are truncated",
			issues:            makeIssues(30, "file.go", SeverityError),
			expectedCount:     MaxIssuesPerFile,
			expectedTruncated: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, truncated := TruncateIssues(tt.issues)
			assert.Equal(t, tt.expectedCount, len(result))
			assert.Equal(t, tt.expectedTruncated, truncated)
		})
	}
}

func TestTruncateIssues_TotalLimit(t *testing.T) {
	// Create issues across multiple files to test total limit (200)
	// With 15 files, each having 20 issues = 300 total
	// Per-file limit allows all 20, but total limit caps at 200
	const numFiles = 15
	const issuesPerFile = 20
	issues := make([]CompactIssue, 0, numFiles*issuesPerFile)
	for i := range numFiles {
		fileIssues := makeIssuesForFile(issuesPerFile, "file"+string(rune('A'+i))+".go", SeverityError)
		issues = append(issues, fileIssues...)
	}

	result, truncated := TruncateIssues(issues)
	assert.Equal(t, MaxTotalIssues, len(result))
	assert.Equal(t, 100, truncated) // 300 - 200 = 100 truncated
}

func TestTruncateIssues_SortsErrorsFirst(t *testing.T) {
	// Create issues across multiple files with different severity
	issues := []CompactIssue{
		{File: "warnings.go", Line: 1, Severity: SeverityWarning, Message: "warn1", RuleID: "r1"},
		{File: "warnings.go", Line: 2, Severity: SeverityWarning, Message: "warn2", RuleID: "r2"},
		{File: "errors.go", Line: 1, Severity: SeverityError, Message: "err1", RuleID: "r1"},
		{File: "errors.go", Line: 2, Severity: SeverityError, Message: "err2", RuleID: "r2"},
		{File: "info.go", Line: 1, Severity: SeverityInfo, Message: "info1", RuleID: "r1"},
	}

	result, _ := TruncateIssues(issues)
	require.NotEmpty(t, result)

	// After sorting by error count, errors.go should come first
	groups := GroupIssuesByFile(result)
	require.NotEmpty(t, groups)
	// First file should be the one with most errors
	assert.Equal(t, "errors.go", groups[0][0])
}

func TestStandardizeSeverity(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "error lowercase", input: "error", expected: SeverityError},
		{name: "Error uppercase", input: "Error", expected: SeverityError},
		{name: "ERROR all caps", input: "ERROR", expected: SeverityError},
		{name: "warning lowercase", input: "warning", expected: SeverityWarning},
		{name: "warn abbreviation", input: "warn", expected: SeverityWarning},
		{name: "info lowercase", input: "info", expected: SeverityInfo},
		{name: "information full", input: "information", expected: SeverityInfo},
		{name: "note as info", input: "note", expected: SeverityInfo},
		{name: "style lowercase", input: "style", expected: SeverityStyle},
		{name: "hint as style", input: "hint", expected: SeverityStyle},
		{name: "convention as style", input: "convention", expected: SeverityStyle},
		{name: "unknown defaults to warning", input: "unknown", expected: SeverityWarning},
		{name: "empty defaults to warning", input: "", expected: SeverityWarning},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StandardizeSeverity(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTruncateMessage(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "single line unchanged",
			input:    "This is a simple message",
			expected: "This is a simple message",
		},
		{
			name:     "multi-line truncated to first line with ellipsis",
			input:    "First line\nSecond line\nThird line",
			expected: "First line...",
		},
		{
			name:     "empty string unchanged",
			input:    "",
			expected: "",
		},
		{
			name:     "line with only newline at end",
			input:    "Message\n",
			expected: "Message...",
		},
		{
			name:     "carriage return handled",
			input:    "Line one\r\nLine two",
			expected: "Line one...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateMessage(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// makeIssues creates n issues for testing
func makeIssues(n int, file string, severity string) []CompactIssue {
	return makeIssuesForFile(n, file, severity)
}

// makeIssuesForFile creates n issues for a specific file
func makeIssuesForFile(n int, file string, severity string) []CompactIssue {
	issues := make([]CompactIssue, n)
	for i := range n {
		issues[i] = CompactIssue{
			File:     file,
			Line:     i + 1,
			Severity: severity,
			Message:  "test message",
			RuleID:   "test-rule",
		}
	}
	return issues
}
