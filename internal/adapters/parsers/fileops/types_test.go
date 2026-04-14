package fileops_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/fileops"
)

func TestMatchTuple_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		tuple    fileops.MatchTuple
		expected string
	}{
		{
			name:     "basic tuple",
			tuple:    fileops.MatchTuple{Line: 10, Content: "func main() {"},
			expected: `[10,"func main() {"]`,
		},
		{
			name:     "zero line number",
			tuple:    fileops.MatchTuple{Line: 0, Content: "content"},
			expected: `[0,"content"]`,
		},
		{
			name:     "empty content",
			tuple:    fileops.MatchTuple{Line: 5, Content: ""},
			expected: `[5,""]`,
		},
		{
			name:     "content with special characters",
			tuple:    fileops.MatchTuple{Line: 42, Content: `key: "value"`},
			expected: `[42,"key: \"value\""]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.tuple)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(data))
		})
	}
}

func TestMatchTuple_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected fileops.MatchTuple
	}{
		{
			name:     "basic tuple",
			input:    `[10,"func main() {"]`,
			expected: fileops.MatchTuple{Line: 10, Content: "func main() {"},
		},
		{
			name:     "zero line number",
			input:    `[0,"content"]`,
			expected: fileops.MatchTuple{Line: 0, Content: "content"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var tuple fileops.MatchTuple
			err := json.Unmarshal([]byte(tt.input), &tuple)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, tuple)
		})
	}
}

func TestFileMatchGroup_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		group    fileops.FileMatchGroup
		expected string
	}{
		{
			name: "single match",
			group: fileops.FileMatchGroup{
				Filename: "main.go",
				Count:    1,
				Matches: []fileops.MatchTuple{
					{Line: 10, Content: "func main() {"},
				},
			},
			expected: `["main.go",1,[[10,"func main() {"]]]`,
		},
		{
			name: "multiple matches",
			group: fileops.FileMatchGroup{
				Filename: "util.go",
				Count:    2,
				Matches: []fileops.MatchTuple{
					{Line: 5, Content: "func helper() {"},
					{Line: 15, Content: "func another() {"},
				},
			},
			expected: `["util.go",2,[[5,"func helper() {"],[15,"func another() {"]]]`,
		},
		{
			name: "no matches",
			group: fileops.FileMatchGroup{
				Filename: "empty.go",
				Count:    0,
				Matches:  []fileops.MatchTuple{},
			},
			expected: `["empty.go",0,[]]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.group)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(data))
		})
	}
}

func TestFileMatchGroup_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected fileops.FileMatchGroup
	}{
		{
			name:  "single match",
			input: `["main.go",1,[[10,"func main() {"]]]`,
			expected: fileops.FileMatchGroup{
				Filename: "main.go",
				Count:    1,
				Matches: []fileops.MatchTuple{
					{Line: 10, Content: "func main() {"},
				},
			},
		},
		{
			name:  "multiple matches",
			input: `["util.go",2,[[5,"func helper() {"],[15,"func another() {"]]]`,
			expected: fileops.FileMatchGroup{
				Filename: "util.go",
				Count:    2,
				Matches: []fileops.MatchTuple{
					{Line: 5, Content: "func helper() {"},
					{Line: 15, Content: "func another() {"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var group fileops.FileMatchGroup
			err := json.Unmarshal([]byte(tt.input), &group)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, group)
		})
	}
}

func TestGrepOutputCompact_MarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		output   fileops.GrepOutputCompact
		expected string
	}{
		{
			name: "basic output",
			output: fileops.GrepOutputCompact{
				Total: 3,
				Files: 2,
				Results: []fileops.FileMatchGroup{
					{
						Filename: "main.go",
						Count:    2,
						Matches: []fileops.MatchTuple{
							{Line: 10, Content: "func main() {"},
							{Line: 25, Content: "return nil"},
						},
					},
					{
						Filename: "util.go",
						Count:    1,
						Matches: []fileops.MatchTuple{
							{Line: 5, Content: "func helper() {"},
						},
					},
				},
				Truncated: false,
			},
			expected: `{"total":3,"files":2,"results":[["main.go",2,[[10,"func main() {"],[25,"return nil"]]],["util.go",1,[[5,"func helper() {"]]]],"truncated":false}`,
		},
		{
			name: "empty output",
			output: fileops.GrepOutputCompact{
				Total:     0,
				Files:     0,
				Results:   []fileops.FileMatchGroup{},
				Truncated: false,
			},
			expected: `{"total":0,"files":0,"results":[],"truncated":false}`,
		},
		{
			name: "truncated output",
			output: fileops.GrepOutputCompact{
				Total: 100,
				Files: 10,
				Results: []fileops.FileMatchGroup{
					{
						Filename: "main.go",
						Count:    1,
						Matches: []fileops.MatchTuple{
							{Line: 1, Content: "package main"},
						},
					},
				},
				Truncated: true,
			},
			expected: `{"total":100,"files":10,"results":[["main.go",1,[[1,"package main"]]]],"truncated":true}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.output)
			require.NoError(t, err)
			assert.JSONEq(t, tt.expected, string(data))
		})
	}
}

func TestGrepOutputCompact_UnmarshalJSON(t *testing.T) {
	input := `{"total":3,"files":2,"results":[["main.go",2,[[10,"func main() {"],[25,"return nil"]]],["util.go",1,[[5,"func helper() {"]]]],"truncated":false}`

	var output fileops.GrepOutputCompact
	err := json.Unmarshal([]byte(input), &output)
	require.NoError(t, err)

	assert.Equal(t, 3, output.Total)
	assert.Equal(t, 2, output.Files)
	assert.Len(t, output.Results, 2)
	assert.False(t, output.Truncated)

	// Check first file group
	assert.Equal(t, "main.go", output.Results[0].Filename)
	assert.Equal(t, 2, output.Results[0].Count)
	assert.Len(t, output.Results[0].Matches, 2)
	assert.Equal(t, 10, output.Results[0].Matches[0].Line)
	assert.Equal(t, "func main() {", output.Results[0].Matches[0].Content)
}

// Test that existing types remain unchanged
func TestGrepOutput_BackwardCompatibility(t *testing.T) {
	output := fileops.GrepOutput{
		Matches: []fileops.GrepMatch{
			{File: "main.go", Line: 10, Content: "func main() {"},
		},
		Count:        1,
		FilesMatched: 1,
	}

	data, err := json.Marshal(output)
	require.NoError(t, err)

	expected := `{"matches":[{"file":"main.go","line":10,"content":"func main() {"}],"count":1,"filesMatched":1}`
	assert.JSONEq(t, expected, string(data))
}

func TestGrepMatch_BackwardCompatibility(t *testing.T) {
	match := fileops.GrepMatch{
		File:    "test.go",
		Line:    42,
		Content: "test content",
		Column:  5,
	}

	data, err := json.Marshal(match)
	require.NoError(t, err)

	expected := `{"file":"test.go","line":42,"content":"test content","column":5}`
	assert.JSONEq(t, expected, string(data))
}
