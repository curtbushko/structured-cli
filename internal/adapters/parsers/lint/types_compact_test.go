package lint

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSeverityConstants(t *testing.T) {
	// Verify severity constants have correct values
	assert.Equal(t, "error", SeverityError)
	assert.Equal(t, "warning", SeverityWarning)
	assert.Equal(t, "info", SeverityInfo)
	assert.Equal(t, "style", SeverityStyle)
}

func TestTruncationLimitConstants(t *testing.T) {
	// Verify truncation limit constants have correct values
	assert.Equal(t, 200, MaxTotalIssues)
	assert.Equal(t, 20, MaxIssuesPerFile)
}

func TestOutputCompact_Fields(t *testing.T) {
	// Test that OutputCompact has the required fields
	compact := OutputCompact{
		TotalIssues:     10,
		FilesWithIssues: 3,
		SeverityCounts: map[string]int{
			SeverityError:   5,
			SeverityWarning: 3,
			SeverityInfo:    2,
		},
		Results:   []FileIssueGroup{},
		Truncated: 0,
	}

	assert.Equal(t, 10, compact.TotalIssues)
	assert.Equal(t, 3, compact.FilesWithIssues)
	assert.Equal(t, 5, compact.SeverityCounts[SeverityError])
	assert.Equal(t, 3, compact.SeverityCounts[SeverityWarning])
	assert.Equal(t, 2, compact.SeverityCounts[SeverityInfo])
	assert.Empty(t, compact.Results)
	assert.Equal(t, 0, compact.Truncated)
}

func TestOutputCompact_JSONSerialization(t *testing.T) {
	// Verify JSON serialization uses correct field names
	compact := OutputCompact{
		TotalIssues:     5,
		FilesWithIssues: 2,
		SeverityCounts: map[string]int{
			SeverityError:   3,
			SeverityWarning: 2,
		},
		Results:   []FileIssueGroup{},
		Truncated: 1,
	}

	data, err := json.Marshal(compact)
	require.NoError(t, err)

	var parsed map[string]interface{}
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, float64(5), parsed["total_issues"])
	assert.Equal(t, float64(2), parsed["files_with_issues"])
	assert.NotNil(t, parsed["severity_counts"])
	assert.NotNil(t, parsed["results"])
	assert.Equal(t, float64(1), parsed["truncated"])
}

func TestFileIssueGroup_Structure(t *testing.T) {
	// FileIssueGroup is [3]interface{} representing [filename, issue_count, issues_array]
	issues := []IssueTuple{
		{10, SeverityError, "undefined variable 'foo'", "no-undef"},
	}
	group := FileIssueGroup{"src/index.js", 1, issues}

	assert.Equal(t, "src/index.js", group[0])
	assert.Equal(t, 1, group[1])
	assert.Equal(t, issues, group[2])
}

func TestFileIssueGroup_JSONSerialization(t *testing.T) {
	issues := []IssueTuple{
		{10, SeverityError, "test message", "test-rule"},
	}
	group := FileIssueGroup{"test.js", 1, issues}

	data, err := json.Marshal(group)
	require.NoError(t, err)

	// Should serialize as an array: ["test.js", 1, [[10, "error", "test message", "test-rule"]]]
	var parsed []interface{}
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Len(t, parsed, 3)
	assert.Equal(t, "test.js", parsed[0])
	assert.Equal(t, float64(1), parsed[1])
}

func TestIssueTuple_Structure(t *testing.T) {
	// IssueTuple is [4]interface{} representing [line, severity, message, rule_id]
	tuple := IssueTuple{42, SeverityWarning, "unused variable", "no-unused-vars"}

	assert.Equal(t, 42, tuple[0])
	assert.Equal(t, SeverityWarning, tuple[1])
	assert.Equal(t, "unused variable", tuple[2])
	assert.Equal(t, "no-unused-vars", tuple[3])
}

func TestIssueTuple_JSONSerialization(t *testing.T) {
	tuple := IssueTuple{15, SeverityError, "syntax error", "parse-error"}

	data, err := json.Marshal(tuple)
	require.NoError(t, err)

	// Should serialize as an array: [15, "error", "syntax error", "parse-error"]
	var parsed []interface{}
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Len(t, parsed, 4)
	assert.Equal(t, float64(15), parsed[0])
	assert.Equal(t, SeverityError, parsed[1])
	assert.Equal(t, "syntax error", parsed[2])
	assert.Equal(t, "parse-error", parsed[3])
}

func TestOutputCompact_WithResults(t *testing.T) {
	// Test complete structure with nested results
	issues1 := []IssueTuple{
		{10, SeverityError, "error 1", "rule-1"},
		{20, SeverityWarning, "warning 1", "rule-2"},
	}
	issues2 := []IssueTuple{
		{5, SeverityInfo, "info 1", "rule-3"},
	}

	compact := OutputCompact{
		TotalIssues:     3,
		FilesWithIssues: 2,
		SeverityCounts: map[string]int{
			SeverityError:   1,
			SeverityWarning: 1,
			SeverityInfo:    1,
		},
		Results: []FileIssueGroup{
			{"file1.js", 2, issues1},
			{"file2.js", 1, issues2},
		},
		Truncated: 0,
	}

	// Serialize and deserialize to verify structure
	data, err := json.Marshal(compact)
	require.NoError(t, err)

	var roundTrip OutputCompact
	err = json.Unmarshal(data, &roundTrip)
	require.NoError(t, err)

	assert.Equal(t, compact.TotalIssues, roundTrip.TotalIssues)
	assert.Equal(t, compact.FilesWithIssues, roundTrip.FilesWithIssues)
	assert.Equal(t, compact.SeverityCounts, roundTrip.SeverityCounts)
	assert.Len(t, roundTrip.Results, 2)
}
