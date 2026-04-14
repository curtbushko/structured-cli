package lint

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMypyParser_EmptyOutput(t *testing.T) {
	parser := NewMypyParser()
	result, err := parser.Parse(strings.NewReader(""))
	require.NoError(t, err)
	require.Nil(t, result.Error)

	got, ok := result.Data.(*MypyResultCompact)
	require.True(t, ok, "expected *MypyResultCompact, got %T", result.Data)

	assert.Equal(t, 0, got.TotalIssues)
	assert.Equal(t, 0, got.FilesWithIssues)
	assert.Empty(t, got.Results)
	assert.Equal(t, 0, got.Truncated)
	assert.Empty(t, got.Summary)
}

func TestMypyParser_Success(t *testing.T) {
	input := "Success: no issues found in 10 source files"

	parser := NewMypyParser()
	result, err := parser.Parse(strings.NewReader(input))
	require.NoError(t, err)
	require.Nil(t, result.Error)

	got, ok := result.Data.(*MypyResultCompact)
	require.True(t, ok, "expected *MypyResultCompact, got %T", result.Data)

	assert.Equal(t, 0, got.TotalIssues)
	assert.Equal(t, 0, got.FilesWithIssues)
	assert.Empty(t, got.Results)
	assert.Equal(t, 0, got.Truncated)
	assert.Equal(t, "Success: no issues found in 10 source files", got.Summary)
}

func TestMypyParser_SingleError(t *testing.T) {
	input := `main.py:10: error: Argument 1 to "foo" has incompatible type "str"; expected "int"  [arg-type]`

	parser := NewMypyParser()
	result, err := parser.Parse(strings.NewReader(input))
	require.NoError(t, err)
	require.Nil(t, result.Error)

	got, ok := result.Data.(*MypyResultCompact)
	require.True(t, ok, "expected *MypyResultCompact, got %T", result.Data)

	assert.Equal(t, 1, got.TotalIssues)
	assert.Equal(t, 1, got.FilesWithIssues)
	assert.Len(t, got.Results, 1)

	// Check severity counts
	assert.Equal(t, 1, got.SeverityCounts[SeverityError])

	// Check grouped result: FileIssueGroup is [filename, count, issues]
	fileGroup := got.Results[0]
	assert.Equal(t, "main.py", fileGroup[0])
	assert.Equal(t, 1, fileGroup[1])

	// Issues is []IssueTuple, each tuple is [line, severity, message, rule_id]
	issues, ok := fileGroup[2].([]IssueTuple)
	require.True(t, ok)
	require.Len(t, issues, 1)

	issue := issues[0]
	assert.Equal(t, 10, issue[0])                                                                // line
	assert.Equal(t, SeverityError, issue[1])                                                     // severity
	assert.Equal(t, `Argument 1 to "foo" has incompatible type "str"; expected "int"`, issue[2]) // message
	assert.Equal(t, "arg-type", issue[3])                                                        // rule_id (error code)
}

func TestMypyParser_MultipleErrors(t *testing.T) {
	input := `main.py:10: error: Argument 1 to "foo" has incompatible type "str"; expected "int"  [arg-type]
utils.py:25: error: Function is missing a return type annotation  [no-untyped-def]
app.py:42: warning: "Optional[str]" is deprecated; use "str | None" instead  [deprecated]
Found 3 errors in 3 files (checked 10 source files)`

	parser := NewMypyParser()
	result, err := parser.Parse(strings.NewReader(input))
	require.NoError(t, err)
	require.Nil(t, result.Error)

	got, ok := result.Data.(*MypyResultCompact)
	require.True(t, ok, "expected *MypyResultCompact, got %T", result.Data)

	assert.Equal(t, 3, got.TotalIssues)
	assert.Equal(t, 3, got.FilesWithIssues)
	assert.Len(t, got.Results, 3)

	// Check severity counts
	assert.Equal(t, 2, got.SeverityCounts[SeverityError])
	assert.Equal(t, 1, got.SeverityCounts[SeverityWarning])

	// Check summary
	assert.Equal(t, "Found 3 errors in 3 files (checked 10 source files)", got.Summary)
}

func TestMypyParser_SeverityMapping(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedSeverity string
	}{
		{
			name:             "error maps to error",
			input:            `main.py:10: error: Type error  [type-arg]`,
			expectedSeverity: SeverityError,
		},
		{
			name:             "warning maps to warning",
			input:            `main.py:10: warning: Deprecated  [deprecated]`,
			expectedSeverity: SeverityWarning,
		},
		{
			name:             "note maps to info",
			input:            `main.py:10: note: See documentation  [help]`,
			expectedSeverity: SeverityInfo,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewMypyParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			require.NoError(t, err)

			got, ok := result.Data.(*MypyResultCompact)
			require.True(t, ok)

			assert.Equal(t, 1, got.SeverityCounts[tt.expectedSeverity])
		})
	}
}

func TestMypyParser_ErrorWithoutCode(t *testing.T) {
	input := `main.py:10: error: Cannot find implementation or library stub for module named "nonexistent"`

	parser := NewMypyParser()
	result, err := parser.Parse(strings.NewReader(input))
	require.NoError(t, err)

	got, ok := result.Data.(*MypyResultCompact)
	require.True(t, ok, "expected *MypyResultCompact, got %T", result.Data)

	require.Len(t, got.Results, 1)
	fileGroup := got.Results[0]
	issues, ok := fileGroup[2].([]IssueTuple)
	require.True(t, ok)
	require.Len(t, issues, 1)

	issue := issues[0]
	assert.Equal(t, 10, issue[0])                                                                          // line
	assert.Equal(t, SeverityError, issue[1])                                                               // severity
	assert.Equal(t, `Cannot find implementation or library stub for module named "nonexistent"`, issue[2]) // message
	assert.Equal(t, "", issue[3])                                                                          // no error code
}

func TestMypyParser_NoteMessage(t *testing.T) {
	input := `main.py:10: note: See https://mypy.readthedocs.io/en/stable/running_mypy.html#missing-imports`

	parser := NewMypyParser()
	result, err := parser.Parse(strings.NewReader(input))
	require.NoError(t, err)

	got, ok := result.Data.(*MypyResultCompact)
	require.True(t, ok, "expected *MypyResultCompact, got %T", result.Data)

	// Note severity maps to info
	assert.Equal(t, 1, got.SeverityCounts[SeverityInfo])
}

func TestMypyParser_VerboseMessageTruncation(t *testing.T) {
	// Create a very long message (over 100 chars)
	longMessage := strings.Repeat("x", 150)
	input := `main.py:10: error: ` + longMessage + `  [arg-type]`

	parser := NewMypyParser()
	result, err := parser.Parse(strings.NewReader(input))
	require.NoError(t, err)

	got, ok := result.Data.(*MypyResultCompact)
	require.True(t, ok, "expected *MypyResultCompact, got %T", result.Data)

	require.Len(t, got.Results, 1)
	fileGroup := got.Results[0]
	issues, ok := fileGroup[2].([]IssueTuple)
	require.True(t, ok)
	require.Len(t, issues, 1)

	message, ok := issues[0][2].(string)
	require.True(t, ok)
	// Message should be truncated to 100 chars with ...
	assert.LessOrEqual(t, len(message), maxMypyMessageLength)
	assert.True(t, strings.HasSuffix(message, "..."))
}

func TestMypyParser_GroupingByFile(t *testing.T) {
	input := `main.py:10: error: Error 1  [arg-type]
main.py:20: error: Error 2  [return-value]
utils.py:5: warning: Warning 1  [deprecated]`

	parser := NewMypyParser()
	result, err := parser.Parse(strings.NewReader(input))
	require.NoError(t, err)

	got, ok := result.Data.(*MypyResultCompact)
	require.True(t, ok)

	assert.Equal(t, 3, got.TotalIssues)
	assert.Equal(t, 2, got.FilesWithIssues)

	// Files with errors should come first (sorted by error count)
	// main.py has 2 errors, utils.py has 1 warning
	firstFile := got.Results[0][0]
	assert.Equal(t, "main.py", firstFile)

	firstFileIssues, ok := got.Results[0][2].([]IssueTuple)
	require.True(t, ok)
	assert.Len(t, firstFileIssues, 2)
}

func TestMypyParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches mypy with no subcommands",
			cmd:         "mypy",
			subcommands: []string{},
			want:        true,
		},
		{
			name:        "matches mypy with path",
			cmd:         "mypy",
			subcommands: []string{"src/"},
			want:        true,
		},
		{
			name:        "matches mypy with flags",
			cmd:         "mypy",
			subcommands: []string{"--strict", "src/"},
			want:        true,
		},
		{
			name:        "does not match python mypy",
			cmd:         "python",
			subcommands: []string{"-m", "mypy"},
			want:        false,
		},
		{
			name:        "does not match empty command",
			cmd:         "",
			subcommands: []string{},
			want:        false,
		},
	}

	parser := NewMypyParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMypyParser_Schema(t *testing.T) {
	parser := NewMypyParser()
	schema := parser.Schema()

	assert.NotEmpty(t, schema.ID)
	assert.NotEmpty(t, schema.Title)
	assert.Equal(t, "object", schema.Type)

	// Verify compact format required properties exist
	requiredProps := []string{"total_issues", "files_with_issues", "severity_counts", "results", "truncated", "summary"}
	for _, prop := range requiredProps {
		_, ok := schema.Properties[prop]
		assert.True(t, ok, "Schema.Properties missing %q", prop)
	}
}

func TestMypyParser_DifferentFormats(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedFile string
		expectedLine int
	}{
		{
			name:         "full path file",
			input:        "/home/user/project/src/main.py:10: error: Type error  [type-arg]",
			expectedFile: "/home/user/project/src/main.py",
			expectedLine: 10,
		},
		{
			name:         "relative path file",
			input:        "./src/main.py:15: error: Missing return statement  [return]",
			expectedFile: "./src/main.py",
			expectedLine: 15,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewMypyParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			require.NoError(t, err)
			require.Nil(t, result.Error)

			got, ok := result.Data.(*MypyResultCompact)
			require.True(t, ok, "expected *MypyResultCompact, got %T", result.Data)

			require.Len(t, got.Results, 1)
			assert.Equal(t, tt.expectedFile, got.Results[0][0])

			issues, ok := got.Results[0][2].([]IssueTuple)
			require.True(t, ok)
			require.Len(t, issues, 1)
			assert.Equal(t, tt.expectedLine, issues[0][0])
		})
	}
}

func TestMypyParser_ErrorCodeInTuple(t *testing.T) {
	// Verify that error codes like "arg-type" and "return-value" are in the tuple
	input := `main.py:10: error: Type mismatch  [arg-type]
main.py:20: error: Missing return  [return-value]`

	parser := NewMypyParser()
	result, err := parser.Parse(strings.NewReader(input))
	require.NoError(t, err)

	got, ok := result.Data.(*MypyResultCompact)
	require.True(t, ok)

	require.Len(t, got.Results, 1)
	issues, ok := got.Results[0][2].([]IssueTuple)
	require.True(t, ok)
	require.Len(t, issues, 2)

	// Check error codes are included in the tuple
	assert.Equal(t, "arg-type", issues[0][3])
	assert.Equal(t, "return-value", issues[1][3])
}
