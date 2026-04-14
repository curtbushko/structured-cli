package cargo

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: Old tests have been replaced with compact format tests below.
// The compact format tests cover the same functionality:
// - TestClippyParser_CompactFormat_EmptyOutput
// - TestClippyParser_CompactFormat_SingleWarning
// - TestClippyParser_CompactFormat_ClippyLint
// - TestClippyParser_CompactFormat_Error
// - TestClippyParser_CompactFormat_GroupedByFile (replaces MultipleWarnings)

func TestClippyParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches cargo clippy",
			cmd:         "cargo",
			subcommands: []string{"clippy"},
			want:        true,
		},
		{
			name:        "matches cargo clippy with flags",
			cmd:         "cargo",
			subcommands: []string{"clippy", "--", "-D", "warnings"},
			want:        true,
		},
		{
			name:        "does not match cargo build",
			cmd:         "cargo",
			subcommands: []string{"build"},
			want:        false,
		},
		{
			name:        "does not match cargo without subcommand",
			cmd:         "cargo",
			subcommands: []string{},
			want:        false,
		},
		{
			name:        "does not match empty command",
			cmd:         "",
			subcommands: []string{},
			want:        false,
		},
	}

	parser := NewClippyParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			assert.Equal(t, tt.want, got, "Matches(%q, %v)", tt.cmd, tt.subcommands)
		})
	}
}

func TestClippyParser_Schema(t *testing.T) {
	parser := NewClippyParser()
	schema := parser.Schema()

	assert.NotEmpty(t, schema.ID, "Schema.ID should not be empty")
	assert.NotEmpty(t, schema.Title, "Schema.Title should not be empty")
	assert.Equal(t, schemaTypeObject, schema.Type)

	requiredProps := []string{"total_issues", "files_with_issues", "severity_counts", "results", "truncated"}
	for _, prop := range requiredProps {
		_, ok := schema.Properties[prop]
		assert.True(t, ok, "Schema.Properties missing %q", prop)
	}
}

// Compact format tests

func TestClippyParser_CompactFormat_EmptyOutput(t *testing.T) {
	parser := NewClippyParser()
	result, err := parser.Parse(strings.NewReader(""))
	require.NoError(t, err)
	require.Nil(t, result.Error)

	got, ok := result.Data.(*ClippyResultCompact)
	require.True(t, ok, "ParseResult.Data type = %T, want *ClippyResultCompact", result.Data)

	assert.Equal(t, 0, got.TotalIssues)
	assert.Equal(t, 0, got.FilesWithIssues)
	assert.Empty(t, got.Results)
	assert.Equal(t, 0, got.Truncated)
}

func TestClippyParser_CompactFormat_SingleWarning(t *testing.T) {
	input := `{"reason":"compiler-message","package_id":"my_crate 0.1.0","manifest_path":"/path/Cargo.toml","target":{"kind":["lib"],"crate_types":["lib"],"name":"my_crate","src_path":"/path/src/lib.rs","edition":"2021"},"message":{"message":"unused variable: ` + "`x`" + `","code":{"code":"unused_variables","explanation":null},"level":"warning","spans":[{"file_name":"src/lib.rs","byte_start":100,"byte_end":101,"line_start":5,"line_end":5,"column_start":9,"column_end":10,"is_primary":true,"label":null}],"children":[],"rendered":"warning: unused variable"}}
{"reason":"build-finished","success":true}`

	parser := NewClippyParser()
	result, err := parser.Parse(strings.NewReader(input))
	require.NoError(t, err)
	require.Nil(t, result.Error)

	got, ok := result.Data.(*ClippyResultCompact)
	require.True(t, ok, "ParseResult.Data type = %T, want *ClippyResultCompact", result.Data)

	assert.Equal(t, 1, got.TotalIssues)
	assert.Equal(t, 1, got.FilesWithIssues)
	assert.Equal(t, 1, got.SeverityCounts[severityWarning])
	require.Len(t, got.Results, 1)

	// Verify file group
	fileGroup := got.Results[0]
	assert.Equal(t, "src/lib.rs", fileGroup[0])
	assert.Equal(t, 1, fileGroup[1])

	// Verify issue tuple: [line, severity, message, rule_id]
	issues := fileGroup[2].([]ClippyIssueTuple)
	require.Len(t, issues, 1)
	assert.Equal(t, 5, issues[0][0])                      // line
	assert.Equal(t, severityWarning, issues[0][1])        // severity
	assert.Equal(t, "unused variable: `x`", issues[0][2]) // message
	assert.Equal(t, "unused_variables", issues[0][3])     // rule_id (lint name)
}

func TestClippyParser_CompactFormat_ClippyLint(t *testing.T) {
	// Clippy lints have codes like "clippy::unwrap_used"
	input := `{"reason":"compiler-message","package_id":"my_crate 0.1.0","manifest_path":"/path/Cargo.toml","target":{"kind":["lib"],"crate_types":["lib"],"name":"my_crate","src_path":"/path/src/lib.rs","edition":"2021"},"message":{"message":"using ` + "`unwrap()`" + ` on a Result value","code":{"code":"clippy::unwrap_used","explanation":null},"level":"warning","spans":[{"file_name":"src/lib.rs","byte_start":200,"byte_end":210,"line_start":10,"line_end":10,"column_start":5,"column_end":15,"is_primary":true,"label":null}],"children":[],"rendered":"warning: using unwrap()"}}
{"reason":"build-finished","success":true}`

	parser := NewClippyParser()
	result, err := parser.Parse(strings.NewReader(input))
	require.NoError(t, err)

	got, ok := result.Data.(*ClippyResultCompact)
	require.True(t, ok)

	require.Len(t, got.Results, 1)
	issues := got.Results[0][2].([]ClippyIssueTuple)
	require.Len(t, issues, 1)
	assert.Equal(t, "clippy::unwrap_used", issues[0][3]) // lint name preserved
}

func TestClippyParser_CompactFormat_Error(t *testing.T) {
	input := `{"reason":"compiler-message","package_id":"my_crate 0.1.0","manifest_path":"/path/Cargo.toml","target":{"kind":["lib"],"crate_types":["lib"],"name":"my_crate","src_path":"/path/src/lib.rs","edition":"2021"},"message":{"message":"cannot find value ` + "`foo`" + ` in this scope","code":{"code":"E0425","explanation":null},"level":"error","spans":[{"file_name":"src/lib.rs","byte_start":50,"byte_end":53,"line_start":3,"line_end":3,"column_start":5,"column_end":8,"is_primary":true,"label":"not found"}],"children":[],"rendered":"error[E0425]: cannot find value"}}
{"reason":"build-finished","success":false}`

	parser := NewClippyParser()
	result, err := parser.Parse(strings.NewReader(input))
	require.NoError(t, err)

	got, ok := result.Data.(*ClippyResultCompact)
	require.True(t, ok)

	assert.Equal(t, 1, got.TotalIssues)
	assert.Equal(t, 1, got.SeverityCounts[severityError])

	require.Len(t, got.Results, 1)
	issues := got.Results[0][2].([]ClippyIssueTuple)
	require.Len(t, issues, 1)
	assert.Equal(t, severityError, issues[0][1]) // severity mapped to error
}

func TestClippyParser_CompactFormat_GroupedByFile(t *testing.T) {
	// Two warnings in same file, one in different file
	input := `{"reason":"compiler-message","package_id":"pkg 0.1.0","manifest_path":"/Cargo.toml","target":{"kind":["lib"],"crate_types":["lib"],"name":"pkg","src_path":"/src/lib.rs","edition":"2021"},"message":{"message":"warning 1","code":{"code":"clippy::lint1","explanation":null},"level":"warning","spans":[{"file_name":"src/lib.rs","byte_start":10,"byte_end":11,"line_start":1,"line_end":1,"column_start":1,"column_end":2,"is_primary":true,"label":null}],"children":[],"rendered":"warning"}}
{"reason":"compiler-message","package_id":"pkg 0.1.0","manifest_path":"/Cargo.toml","target":{"kind":["lib"],"crate_types":["lib"],"name":"pkg","src_path":"/src/lib.rs","edition":"2021"},"message":{"message":"warning 2","code":{"code":"clippy::lint2","explanation":null},"level":"warning","spans":[{"file_name":"src/lib.rs","byte_start":20,"byte_end":21,"line_start":2,"line_end":2,"column_start":1,"column_end":2,"is_primary":true,"label":null}],"children":[],"rendered":"warning"}}
{"reason":"compiler-message","package_id":"pkg 0.1.0","manifest_path":"/Cargo.toml","target":{"kind":["lib"],"crate_types":["lib"],"name":"pkg","src_path":"/src/other.rs","edition":"2021"},"message":{"message":"warning 3","code":{"code":"clippy::lint3","explanation":null},"level":"warning","spans":[{"file_name":"src/other.rs","byte_start":30,"byte_end":31,"line_start":5,"line_end":5,"column_start":1,"column_end":2,"is_primary":true,"label":null}],"children":[],"rendered":"warning"}}
{"reason":"build-finished","success":true}`

	parser := NewClippyParser()
	result, err := parser.Parse(strings.NewReader(input))
	require.NoError(t, err)

	got, ok := result.Data.(*ClippyResultCompact)
	require.True(t, ok)

	assert.Equal(t, 3, got.TotalIssues)
	assert.Equal(t, 2, got.FilesWithIssues)
	require.Len(t, got.Results, 2)

	// Find src/lib.rs group
	var libGroup, otherGroup ClippyFileIssueGroup
	for _, g := range got.Results {
		switch g[0] {
		case "src/lib.rs":
			libGroup = g
		case "src/other.rs":
			otherGroup = g
		}
	}

	assert.Equal(t, 2, libGroup[1])   // 2 issues in lib.rs
	assert.Equal(t, 1, otherGroup[1]) // 1 issue in other.rs
}

func TestClippyParser_CompactFormat_SeverityMapping(t *testing.T) {
	// Test that Clippy severity levels are mapped correctly
	// deny = error, warn = warning, allow = info
	tests := []struct {
		level    string
		expected string
	}{
		{"error", severityError},
		{"warning", severityWarning},
		{"note", severityInfo},
	}

	for _, tt := range tests {
		t.Run(tt.level, func(t *testing.T) {
			input := `{"reason":"compiler-message","package_id":"pkg 0.1.0","manifest_path":"/Cargo.toml","target":{"kind":["lib"],"crate_types":["lib"],"name":"pkg","src_path":"/src/lib.rs","edition":"2021"},"message":{"message":"test message","code":{"code":"test_lint","explanation":null},"level":"` + tt.level + `","spans":[{"file_name":"src/lib.rs","byte_start":10,"byte_end":11,"line_start":1,"line_end":1,"column_start":1,"column_end":2,"is_primary":true,"label":null}],"children":[],"rendered":"test"}}
{"reason":"build-finished","success":true}`

			parser := NewClippyParser()
			result, err := parser.Parse(strings.NewReader(input))
			require.NoError(t, err)

			got, ok := result.Data.(*ClippyResultCompact)
			require.True(t, ok)

			require.Len(t, got.Results, 1)
			issues := got.Results[0][2].([]ClippyIssueTuple)
			require.Len(t, issues, 1)
			assert.Equal(t, tt.expected, issues[0][1], "severity for level %q", tt.level)
		})
	}
}

func TestClippyParser_CompactFormat_HelpTextSeparated(t *testing.T) {
	// Clippy includes help text in children - should be separated from main message
	input := `{"reason":"compiler-message","package_id":"my_crate 0.1.0","manifest_path":"/path/Cargo.toml","target":{"kind":["lib"],"crate_types":["lib"],"name":"my_crate","src_path":"/path/src/lib.rs","edition":"2021"},"message":{"message":"unused variable: ` + "`x`" + `","code":{"code":"unused_variables","explanation":null},"level":"warning","spans":[{"file_name":"src/lib.rs","byte_start":100,"byte_end":101,"line_start":5,"line_end":5,"column_start":9,"column_end":10,"is_primary":true,"label":null}],"children":[{"message":"if this is intentional, prefix it with an underscore: ` + "`_x`" + `","code":null,"level":"help","spans":[],"children":[],"rendered":null}],"rendered":"warning: unused variable"}}
{"reason":"build-finished","success":true}`

	parser := NewClippyParser()
	result, err := parser.Parse(strings.NewReader(input))
	require.NoError(t, err)

	got, ok := result.Data.(*ClippyResultCompact)
	require.True(t, ok)

	require.Len(t, got.Results, 1)
	issues := got.Results[0][2].([]ClippyIssueTuple)
	require.Len(t, issues, 1)

	// Message should NOT include the help text
	message := issues[0][2].(string)
	assert.Equal(t, "unused variable: `x`", message)
	assert.NotContains(t, message, "if this is intentional")
}

func TestClippyParser_LintNameExtraction(t *testing.T) {
	// Test that various lint names are correctly extracted and included in tuples
	tests := []struct {
		name         string
		lintCode     string
		expectedRule string
	}{
		{
			name:         "needless_borrow",
			lintCode:     "clippy::needless_borrow",
			expectedRule: "clippy::needless_borrow",
		},
		{
			name:         "unused_variables",
			lintCode:     "unused_variables",
			expectedRule: "unused_variables",
		},
		{
			name:         "dead_code",
			lintCode:     "dead_code",
			expectedRule: "dead_code",
		},
		{
			name:         "unwrap_used",
			lintCode:     "clippy::unwrap_used",
			expectedRule: "clippy::unwrap_used",
		},
		{
			name:         "expect_used",
			lintCode:     "clippy::expect_used",
			expectedRule: "clippy::expect_used",
		},
		{
			name:         "E0425 error code",
			lintCode:     "E0425",
			expectedRule: "E0425",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := `{"reason":"compiler-message","package_id":"pkg 0.1.0","manifest_path":"/Cargo.toml","target":{"kind":["lib"],"crate_types":["lib"],"name":"pkg","src_path":"/src/lib.rs","edition":"2021"},"message":{"message":"lint message","code":{"code":"` + tt.lintCode + `","explanation":null},"level":"warning","spans":[{"file_name":"src/lib.rs","byte_start":10,"byte_end":11,"line_start":1,"line_end":1,"column_start":1,"column_end":2,"is_primary":true,"label":null}],"children":[],"rendered":"warning"}}
{"reason":"build-finished","success":true}`

			parser := NewClippyParser()
			result, err := parser.Parse(strings.NewReader(input))
			require.NoError(t, err)

			got, ok := result.Data.(*ClippyResultCompact)
			require.True(t, ok)

			require.Len(t, got.Results, 1)
			issues := got.Results[0][2].([]ClippyIssueTuple)
			require.Len(t, issues, 1)
			assert.Equal(t, tt.expectedRule, issues[0][3], "lint name should be extracted")
		})
	}
}

func TestClippyParser_Truncation(t *testing.T) {
	// Generate 250 warnings to test truncation
	// maxTotalIssues = 200, so 50 should be truncated
	var builder strings.Builder

	// Generate 250 warnings spread across 5 files (50 per file)
	// maxIssuesPerFile = 20, so we'll hit per-file limit too
	for fileNum := 1; fileNum <= 5; fileNum++ {
		fileName := "src/file" + string(rune('0'+fileNum)) + ".rs"
		for i := 0; i < 50; i++ {
			line := i + 1
			builder.WriteString(`{"reason":"compiler-message","package_id":"pkg 0.1.0","manifest_path":"/Cargo.toml","target":{"kind":["lib"],"crate_types":["lib"],"name":"pkg","src_path":"/src/lib.rs","edition":"2021"},"message":{"message":"warning ` + string(rune('0'+fileNum)) + `-` + string(rune('0'+i%10)) + `","code":{"code":"clippy::test_lint","explanation":null},"level":"warning","spans":[{"file_name":"` + fileName + `","byte_start":10,"byte_end":11,"line_start":` + strings.TrimSpace(strings.Repeat("", 0)+itoa(line)) + `,"line_end":` + itoa(line) + `,"column_start":1,"column_end":2,"is_primary":true,"label":null}],"children":[],"rendered":"warning"}}`)
			builder.WriteString("\n")
		}
	}
	builder.WriteString(`{"reason":"build-finished","success":true}`)

	parser := NewClippyParser()
	result, err := parser.Parse(strings.NewReader(builder.String()))
	require.NoError(t, err)

	got, ok := result.Data.(*ClippyResultCompact)
	require.True(t, ok)

	// TotalIssues should reflect the actual count before truncation
	assert.Equal(t, 250, got.TotalIssues, "total should reflect original count")

	// Count issues actually included in results
	var includedCount int
	for _, group := range got.Results {
		issues := group[2].([]ClippyIssueTuple)
		includedCount += len(issues)
	}

	// With 5 files and max 20 per file, we get at most 100 issues
	// But total limit is 200, so per-file limit kicks in first
	assert.LessOrEqual(t, includedCount, maxTotalIssues, "should not exceed total limit")

	// Truncated should be 250 - included issues
	assert.Equal(t, 250-includedCount, got.Truncated, "truncated count should match")

	// Each file should have at most maxIssuesPerFile issues
	for _, group := range got.Results {
		issues := group[2].([]ClippyIssueTuple)
		assert.LessOrEqual(t, len(issues), maxIssuesPerFile, "file should not exceed per-file limit")
	}
}

// itoa converts an int to string (simple helper to avoid import)
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

func TestClippyParser_NoWarnings(t *testing.T) {
	// Test that no warnings produces correct compact format
	input := `{"reason":"compiler-artifact","package_id":"my_crate 0.1.0","manifest_path":"/path/Cargo.toml","target":{"kind":["lib"],"crate_types":["lib"],"name":"my_crate","src_path":"/path/src/lib.rs","edition":"2021"},"profile":{"opt_level":"0","debuginfo":2,"debug_assertions":true,"overflow_checks":true,"test":false},"features":[],"filenames":["/path/target/debug/libmy_crate.rlib"],"executable":null,"fresh":false}
{"reason":"build-finished","success":true}`

	parser := NewClippyParser()
	result, err := parser.Parse(strings.NewReader(input))
	require.NoError(t, err)
	require.Nil(t, result.Error)

	got, ok := result.Data.(*ClippyResultCompact)
	require.True(t, ok, "ParseResult.Data type = %T, want *ClippyResultCompact", result.Data)

	assert.Equal(t, 0, got.TotalIssues)
	assert.Equal(t, 0, got.FilesWithIssues)
	assert.Empty(t, got.Results)
	assert.Empty(t, got.SeverityCounts)
	assert.Equal(t, 0, got.Truncated)
}
