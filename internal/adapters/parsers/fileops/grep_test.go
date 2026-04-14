package fileops_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/fileops"
)

const (
	testFileMainGo      = "main.go"
	testContentFuncMain = "func main() {"
)

func TestGrepParser_WithLineNumbers(t *testing.T) {
	input := `main.go:10:func main() {
main.go:25:    fmt.Println("Hello")
util.go:5:func helper() {
`
	parser := fileops.NewGrepParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)

	output, ok := result.Data.(*fileops.GrepOutputCompact)
	require.True(t, ok, "result.Data is not *fileops.GrepOutputCompact, got %T", result.Data)

	assert.Equal(t, 3, output.Total, "Total should be 3")
	assert.Equal(t, 2, output.Files, "Files should be 2")
	require.Len(t, output.Results, 2, "Should have 2 file groups")

	// Check first file (main.go)
	assert.Equal(t, testFileMainGo, output.Results[0].Filename)
	assert.Equal(t, 2, output.Results[0].Count)
	require.Len(t, output.Results[0].Matches, 2)
	assert.Equal(t, 10, output.Results[0].Matches[0].Line)
	assert.Equal(t, testContentFuncMain, output.Results[0].Matches[0].Content)
}

func TestGrepParser_WithoutLineNumbers(t *testing.T) {
	input := `main.go:func main() {
util.go:func helper() {
`
	parser := fileops.NewGrepParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)

	output, ok := result.Data.(*fileops.GrepOutputCompact)
	require.True(t, ok, "result.Data is not *fileops.GrepOutputCompact, got %T", result.Data)

	assert.Equal(t, 2, output.Total, "Total should be 2")
	assert.Equal(t, 2, output.Files, "Files should be 2")
	require.Len(t, output.Results, 2, "Should have 2 file groups")

	// Check that line numbers are 0 when not provided
	assert.Equal(t, 0, output.Results[0].Matches[0].Line, "Line should be 0 when not provided")
	assert.Equal(t, testContentFuncMain, output.Results[0].Matches[0].Content)
}

func TestGrepParser_SingleFile(t *testing.T) {
	// grep without filename when searching single file
	input := `10:func main() {
25:    return nil
`
	parser := fileops.NewGrepParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)

	output, ok := result.Data.(*fileops.GrepOutputCompact)
	require.True(t, ok, "result.Data is not *fileops.GrepOutputCompact, got %T", result.Data)

	assert.Equal(t, 2, output.Total, "Total should be 2")
	assert.Equal(t, 1, output.Files, "Files should be 1")
	require.Len(t, output.Results, 1, "Should have 1 file group")
	require.Len(t, output.Results[0].Matches, 2, "File should have 2 matches")
	assert.Equal(t, 10, output.Results[0].Matches[0].Line, "First match line should be 10")
}

func TestGrepParser_NoResults(t *testing.T) {
	input := ``
	parser := fileops.NewGrepParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)

	output, ok := result.Data.(*fileops.GrepOutputCompact)
	require.True(t, ok, "result.Data is not *fileops.GrepOutputCompact, got %T", result.Data)

	assert.Equal(t, 0, output.Total, "Total should be 0")
	assert.Equal(t, 0, output.Files, "Files should be 0")
	assert.Len(t, output.Results, 0, "Results should be empty")
}

func TestGrepParser_Schema(t *testing.T) {
	parser := fileops.NewGrepParser()
	schema := parser.Schema()

	require.NotEmpty(t, schema.ID, "Schema().ID should not be empty")
	require.NotEmpty(t, schema.Title, "Schema().Title should not be empty")

	// Verify schema has the expected properties for compact format
	props := schema.Properties
	require.Contains(t, props, "total", "Schema should have 'total' property")
	require.Contains(t, props, "files", "Schema should have 'files' property")
	require.Contains(t, props, "results", "Schema should have 'results' property")
	require.Contains(t, props, "truncated", "Schema should have 'truncated' property")

	// Verify property types
	assert.Equal(t, "integer", props["total"].Type, "'total' should be integer")
	assert.Equal(t, "integer", props["files"].Type, "'files' should be integer")
	assert.Equal(t, "array", props["results"].Type, "'results' should be array")
	assert.Equal(t, "boolean", props["truncated"].Type, "'truncated' should be boolean")

	// Verify required fields
	require.Len(t, schema.Required, 4, "Should have 4 required fields")
	assert.Contains(t, schema.Required, "total")
	assert.Contains(t, schema.Required, "files")
	assert.Contains(t, schema.Required, "results")
	assert.Contains(t, schema.Required, "truncated")
}

func TestGrepParser_Matches(t *testing.T) {
	parser := fileops.NewGrepParser()

	tests := []struct {
		cmd         string
		subcommands []string
		want        bool
	}{
		{"grep", []string{}, true},
		{"grep", []string{"-r", "pattern"}, true},
		{"grep", []string{"-n", "-r", "TODO"}, true},
		{"egrep", []string{}, true},
		{"fgrep", []string{}, true},
		{"find", []string{}, false},
		{"rg", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.cmd+"_"+strings.Join(tt.subcommands, "_"), func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestGrepParser_BinaryFile(t *testing.T) {
	input := `Binary file image.png matches
main.go:10:func main() {
`
	parser := fileops.NewGrepParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)

	output, ok := result.Data.(*fileops.GrepOutputCompact)
	require.True(t, ok, "result.Data is not *fileops.GrepOutputCompact, got %T", result.Data)

	// Binary file match should be skipped, only 1 match
	assert.Equal(t, 1, output.Total, "Total should be 1 (binary files skipped)")
	assert.Equal(t, 1, output.Files, "Files should be 1")
	require.Len(t, output.Results, 1, "Should have 1 file group")
}

func TestGrepParser_ColonInContent(t *testing.T) {
	// File has colon in content
	input := `config.yaml:5:  key: value
`
	parser := fileops.NewGrepParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)

	output, ok := result.Data.(*fileops.GrepOutputCompact)
	require.True(t, ok, "result.Data is not *fileops.GrepOutputCompact, got %T", result.Data)

	assert.Equal(t, 1, output.Total, "Total should be 1")
	require.Len(t, output.Results, 1, "Should have 1 file group")
	require.Len(t, output.Results[0].Matches, 1, "File should have 1 match")
	assert.Equal(t, "  key: value", output.Results[0].Matches[0].Content)
}

func TestGrepParser_CompactFormat_NoResults(t *testing.T) {
	// Test empty grep results with the new compact format
	input := ``
	parser := fileops.NewGrepParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// This test expects the compact format output
	output, ok := result.Data.(*fileops.GrepOutputCompact)
	if !ok {
		t.Fatalf("result.Data is not *fileops.GrepOutputCompact, got %T", result.Data)
	}

	if output.Total != 0 {
		t.Errorf("Total = %d, want 0", output.Total)
	}
	if output.Files != 0 {
		t.Errorf("Files = %d, want 0", output.Files)
	}
	if len(output.Results) != 0 {
		t.Errorf("Results len = %d, want 0", len(output.Results))
	}
	if output.Truncated {
		t.Errorf("Truncated = %v, want false", output.Truncated)
	}
}

func TestGrepParser_CompactFormat_SingleFile(t *testing.T) {
	// Test grep results from a single file with multiple matches
	input := `main.go:10:func main() {
main.go:25:    fmt.Println("Hello")
main.go:30:    return nil
`
	parser := fileops.NewGrepParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)

	// This test expects the compact format output
	output, ok := result.Data.(*fileops.GrepOutputCompact)
	require.True(t, ok, "result.Data is not *fileops.GrepOutputCompact, got %T", result.Data)

	assert.Equal(t, 3, output.Total, "Total should be 3 matches")
	assert.Equal(t, 1, output.Files, "Files should be 1")
	require.Len(t, output.Results, 1, "Results should have 1 file group")

	// Check the file group
	fileGroup := output.Results[0]
	assert.Equal(t, "main.go", fileGroup.Filename)
	assert.Equal(t, 3, fileGroup.Count)
	require.Len(t, fileGroup.Matches, 3, "File group should have 3 matches")

	// Check individual matches
	assert.Equal(t, 10, fileGroup.Matches[0].Line)
	assert.Equal(t, "func main() {", fileGroup.Matches[0].Content)
	assert.Equal(t, 25, fileGroup.Matches[1].Line)
	assert.Equal(t, `    fmt.Println("Hello")`, fileGroup.Matches[1].Content)
	assert.Equal(t, 30, fileGroup.Matches[2].Line)
	assert.Equal(t, "    return nil", fileGroup.Matches[2].Content)
	assert.False(t, output.Truncated)
}

func TestGrepParser_CompactFormat_MultipleFiles(t *testing.T) {
	// Test grep results from multiple files - each file should be grouped separately
	input := `main.go:10:func main() {
main.go:25:    fmt.Println("Hello")
util.go:5:func helper() {
util.go:15:    return err
config.go:1:package config
`
	parser := fileops.NewGrepParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)

	// This test expects the compact format output
	output, ok := result.Data.(*fileops.GrepOutputCompact)
	require.True(t, ok, "result.Data is not *fileops.GrepOutputCompact, got %T", result.Data)

	assert.Equal(t, 5, output.Total, "Total should be 5 matches")
	assert.Equal(t, 3, output.Files, "Files should be 3")
	require.Len(t, output.Results, 3, "Results should have 3 file groups")

	// Check main.go group
	mainGroup := output.Results[0]
	assert.Equal(t, "main.go", mainGroup.Filename)
	assert.Equal(t, 2, mainGroup.Count)
	require.Len(t, mainGroup.Matches, 2)

	// Check util.go group
	utilGroup := output.Results[1]
	assert.Equal(t, "util.go", utilGroup.Filename)
	assert.Equal(t, 2, utilGroup.Count)
	require.Len(t, utilGroup.Matches, 2)

	// Check config.go group
	configGroup := output.Results[2]
	assert.Equal(t, "config.go", configGroup.Filename)
	assert.Equal(t, 1, configGroup.Count)
	require.Len(t, configGroup.Matches, 1)
	assert.Equal(t, 1, configGroup.Matches[0].Line)
	assert.Equal(t, "package config", configGroup.Matches[0].Content)

	assert.False(t, output.Truncated)
}

func TestGrepParser_CompactFormat_WithLineNumbers(t *testing.T) {
	// Test that line numbers are correctly extracted into tuples
	input := `file.go:42:answer := 42
file.go:100:    const maxRetries = 100
`
	parser := fileops.NewGrepParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)

	output, ok := result.Data.(*fileops.GrepOutputCompact)
	require.True(t, ok, "result.Data is not *fileops.GrepOutputCompact, got %T", result.Data)

	require.Len(t, output.Results, 1)
	fileGroup := output.Results[0]
	require.Len(t, fileGroup.Matches, 2)

	// Verify line numbers are in the tuples
	assert.Equal(t, 42, fileGroup.Matches[0].Line, "First match line should be 42")
	assert.Equal(t, "answer := 42", fileGroup.Matches[0].Content)
	assert.Equal(t, 100, fileGroup.Matches[1].Line, "Second match line should be 100")
	assert.Equal(t, "    const maxRetries = 100", fileGroup.Matches[1].Content)
}

func TestGrepParser_CompactFormat_WithoutLineNumbers(t *testing.T) {
	// Test grep output without -n flag (no line numbers)
	// Line number should be 0 when not provided
	input := `main.go:func main() {
util.go:func helper() {
`
	parser := fileops.NewGrepParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)

	output, ok := result.Data.(*fileops.GrepOutputCompact)
	require.True(t, ok, "result.Data is not *fileops.GrepOutputCompact, got %T", result.Data)

	assert.Equal(t, 2, output.Total)
	assert.Equal(t, 2, output.Files)
	require.Len(t, output.Results, 2)

	// Check that line numbers are 0 when not provided
	assert.Equal(t, 0, output.Results[0].Matches[0].Line, "Line should be 0 when no line number")
	assert.Equal(t, "func main() {", output.Results[0].Matches[0].Content)
	assert.Equal(t, 0, output.Results[1].Matches[0].Line, "Line should be 0 when no line number")
	assert.Equal(t, "func helper() {", output.Results[1].Matches[0].Content)
}

func TestGrepParser_CompactFormat_BinaryFiles(t *testing.T) {
	// Test that "Binary file X matches" lines are skipped in compact format
	input := `Binary file image.png matches
main.go:10:func main() {
Binary file archive.tar.gz matches
util.go:5:func helper() {
Binary file data.bin matches
`
	parser := fileops.NewGrepParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)

	output, ok := result.Data.(*fileops.GrepOutputCompact)
	require.True(t, ok, "result.Data is not *fileops.GrepOutputCompact, got %T", result.Data)

	// Binary file matches should be skipped - only 2 real matches
	assert.Equal(t, 2, output.Total, "Total should be 2 (binary files skipped)")
	assert.Equal(t, 2, output.Files, "Files should be 2 (binary files skipped)")
	require.Len(t, output.Results, 2, "Results should have 2 file groups")

	// Verify the actual matches are present
	assert.Equal(t, "main.go", output.Results[0].Filename)
	assert.Equal(t, 10, output.Results[0].Matches[0].Line)
	assert.Equal(t, "func main() {", output.Results[0].Matches[0].Content)

	assert.Equal(t, "util.go", output.Results[1].Filename)
	assert.Equal(t, 5, output.Results[1].Matches[0].Line)
	assert.Equal(t, "func helper() {", output.Results[1].Matches[0].Content)
}

func TestGrepParser_CompactFormat_DirectoryWarnings(t *testing.T) {
	// Test that "X is a directory" warnings are skipped in compact format
	input := `grep: src: Is a directory
main.go:10:func main() {
grep: node_modules: Is a directory
util.go:5:func helper() {
grep: .git: Is a directory
`
	parser := fileops.NewGrepParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)

	output, ok := result.Data.(*fileops.GrepOutputCompact)
	require.True(t, ok, "result.Data is not *fileops.GrepOutputCompact, got %T", result.Data)

	// Directory warnings should be skipped - only 2 real matches
	assert.Equal(t, 2, output.Total, "Total should be 2 (directory warnings skipped)")
	assert.Equal(t, 2, output.Files, "Files should be 2 (directory warnings skipped)")
	require.Len(t, output.Results, 2, "Results should have 2 file groups")

	// Verify the actual matches are present
	assert.Equal(t, "main.go", output.Results[0].Filename)
	assert.Equal(t, 10, output.Results[0].Matches[0].Line)
	assert.Equal(t, "func main() {", output.Results[0].Matches[0].Content)

	assert.Equal(t, "util.go", output.Results[1].Filename)
	assert.Equal(t, 5, output.Results[1].Matches[0].Line)
	assert.Equal(t, "func helper() {", output.Results[1].Matches[0].Content)
}

func TestGrepParser_CompactFormat_PermissionErrors(t *testing.T) {
	// Test that "grep: X: Permission denied" lines are skipped in compact format
	input := `grep: /etc/shadow: Permission denied
main.go:10:func main() {
grep: /root/.bashrc: Permission denied
util.go:5:func helper() {
grep: /var/log/secure: Permission denied
`
	parser := fileops.NewGrepParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)

	output, ok := result.Data.(*fileops.GrepOutputCompact)
	require.True(t, ok, "result.Data is not *fileops.GrepOutputCompact, got %T", result.Data)

	// Permission denied errors should be skipped - only 2 real matches
	assert.Equal(t, 2, output.Total, "Total should be 2 (permission errors skipped)")
	assert.Equal(t, 2, output.Files, "Files should be 2 (permission errors skipped)")
	require.Len(t, output.Results, 2, "Results should have 2 file groups")

	// Verify the actual matches are present
	assert.Equal(t, "main.go", output.Results[0].Filename)
	assert.Equal(t, 10, output.Results[0].Matches[0].Line)
	assert.Equal(t, "func main() {", output.Results[0].Matches[0].Content)

	assert.Equal(t, "util.go", output.Results[1].Filename)
	assert.Equal(t, 5, output.Results[1].Matches[0].Line)
	assert.Equal(t, "func helper() {", output.Results[1].Matches[0].Content)
}

func TestGrepParser_CompactFormat_NoSuchFileErrors(t *testing.T) {
	// Test that "grep: X: No such file or directory" lines are skipped in compact format
	input := `grep: /missing/file.txt: No such file or directory
main.go:10:func main() {
grep: /nonexistent/path: No such file or directory
util.go:5:func helper() {
grep: deleted_file.go: No such file or directory
`
	parser := fileops.NewGrepParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)

	output, ok := result.Data.(*fileops.GrepOutputCompact)
	require.True(t, ok, "result.Data is not *fileops.GrepOutputCompact, got %T", result.Data)

	// "No such file or directory" errors should be skipped - only 2 real matches
	assert.Equal(t, 2, output.Total, "Total should be 2 (no such file errors skipped)")
	assert.Equal(t, 2, output.Files, "Files should be 2 (no such file errors skipped)")
	require.Len(t, output.Results, 2, "Results should have 2 file groups")

	// Verify the actual matches are present
	assert.Equal(t, "main.go", output.Results[0].Filename)
	assert.Equal(t, 10, output.Results[0].Matches[0].Line)
	assert.Equal(t, "func main() {", output.Results[0].Matches[0].Content)

	assert.Equal(t, "util.go", output.Results[1].Filename)
	assert.Equal(t, 5, output.Results[1].Matches[0].Line)
	assert.Equal(t, "func helper() {", output.Results[1].Matches[0].Content)
}

func TestGrepParser_CompactFormat_ColonInContent(t *testing.T) {
	// Test that colons in content are preserved correctly in compact format
	input := `config.yaml:5:  server: localhost:8080
config.yaml:10:  redis: redis://host:6379
main.go:25:    url := "http://example.com:443/path"
`
	parser := fileops.NewGrepParser()
	result, err := parser.Parse(strings.NewReader(input))

	require.NoError(t, err)

	output, ok := result.Data.(*fileops.GrepOutputCompact)
	require.True(t, ok, "result.Data is not *fileops.GrepOutputCompact, got %T", result.Data)

	assert.Equal(t, 3, output.Total, "Total should be 3 matches")
	assert.Equal(t, 2, output.Files, "Files should be 2")
	require.Len(t, output.Results, 2, "Results should have 2 file groups")

	// Check config.yaml matches - colons in YAML content should be preserved
	configGroup := output.Results[0]
	assert.Equal(t, "config.yaml", configGroup.Filename)
	require.Len(t, configGroup.Matches, 2)
	assert.Equal(t, 5, configGroup.Matches[0].Line)
	assert.Equal(t, "  server: localhost:8080", configGroup.Matches[0].Content)
	assert.Equal(t, 10, configGroup.Matches[1].Line)
	assert.Equal(t, "  redis: redis://host:6379", configGroup.Matches[1].Content)

	// Check main.go match - colons in URL should be preserved
	mainGroup := output.Results[1]
	assert.Equal(t, "main.go", mainGroup.Filename)
	require.Len(t, mainGroup.Matches, 1)
	assert.Equal(t, 25, mainGroup.Matches[0].Line)
	assert.Equal(t, `    url := "http://example.com:443/path"`, mainGroup.Matches[0].Content)
}

func TestGrepParser_CompactFormat_GlobalTruncation(t *testing.T) {
	// Test that total matches are capped at 200 globally
	// Generate input with 250 matches across 25 files (10 matches per file)
	var builder strings.Builder
	for fileNum := range 25 {
		filename := fmt.Sprintf("file%03d.go", fileNum)
		for lineNum := range 10 {
			builder.WriteString(fmt.Sprintf("%s:%d:func example%d() {}\n", filename, lineNum+1, lineNum))
		}
	}

	parser := fileops.NewGrepParser()
	result, err := parser.Parse(strings.NewReader(builder.String()))

	require.NoError(t, err)

	output, ok := result.Data.(*fileops.GrepOutputCompact)
	require.True(t, ok, "result.Data is not *fileops.GrepOutputCompact, got %T", result.Data)

	// Global limit of 200 matches should be enforced
	totalMatches := 0
	for _, group := range output.Results {
		totalMatches += len(group.Matches)
	}
	assert.LessOrEqual(t, totalMatches, 200, "Total matches should be capped at 200")
	assert.True(t, output.Truncated, "Truncated should be true when over 200 matches")
}

func TestGrepParser_CompactFormat_PerFileTruncation(t *testing.T) {
	// Test that matches per file are capped at 10
	// Generate input with 25 matches in a single file
	var builder strings.Builder
	for lineNum := range 25 {
		builder.WriteString(fmt.Sprintf("main.go:%d:func example%d() {}\n", lineNum+1, lineNum))
	}

	parser := fileops.NewGrepParser()
	result, err := parser.Parse(strings.NewReader(builder.String()))

	require.NoError(t, err)

	output, ok := result.Data.(*fileops.GrepOutputCompact)
	require.True(t, ok, "result.Data is not *fileops.GrepOutputCompact, got %T", result.Data)

	// Per-file limit of 10 matches should be enforced
	require.Len(t, output.Results, 1, "Should have 1 file group")
	assert.LessOrEqual(t, len(output.Results[0].Matches), 10, "Matches per file should be capped at 10")
	assert.True(t, output.Truncated, "Truncated should be true when per-file limit exceeded")
}

func TestGrepParser_CompactFormat_TruncationCount(t *testing.T) {
	// Test that Total still reflects the original count before truncation
	// Generate input with 50 matches in a single file
	var builder strings.Builder
	for lineNum := range 50 {
		builder.WriteString(fmt.Sprintf("main.go:%d:func example%d() {}\n", lineNum+1, lineNum))
	}

	parser := fileops.NewGrepParser()
	result, err := parser.Parse(strings.NewReader(builder.String()))

	require.NoError(t, err)

	output, ok := result.Data.(*fileops.GrepOutputCompact)
	require.True(t, ok, "result.Data is not *fileops.GrepOutputCompact, got %T", result.Data)

	// Total should reflect the original 50 matches
	assert.Equal(t, 50, output.Total, "Total should reflect all 50 matches before truncation")

	// But actual matches returned should be limited to 10
	require.Len(t, output.Results, 1)
	assert.LessOrEqual(t, len(output.Results[0].Matches), 10, "Returned matches should be capped at 10")

	// Count in the file group should also reflect original count
	assert.Equal(t, 50, output.Results[0].Count, "File group count should reflect all 50 matches")

	assert.True(t, output.Truncated, "Truncated should be true")
}

func TestGrepParser_CompactFormat_NoTruncation(t *testing.T) {
	// Test that truncated=false when under all limits
	// Generate input with exactly 5 matches in a single file (under 10 per file and under 200 total)
	var builder strings.Builder
	for lineNum := range 5 {
		builder.WriteString(fmt.Sprintf("main.go:%d:func example%d() {}\n", lineNum+1, lineNum))
	}

	parser := fileops.NewGrepParser()
	result, err := parser.Parse(strings.NewReader(builder.String()))

	require.NoError(t, err)

	output, ok := result.Data.(*fileops.GrepOutputCompact)
	require.True(t, ok, "result.Data is not *fileops.GrepOutputCompact, got %T", result.Data)

	assert.Equal(t, 5, output.Total, "Total should be 5")
	assert.Equal(t, 1, output.Files, "Files should be 1")
	require.Len(t, output.Results, 1)
	assert.Len(t, output.Results[0].Matches, 5, "Should have all 5 matches")
	assert.False(t, output.Truncated, "Truncated should be false when under all limits")
}

// TestGrepParser_TokenSavings_SixMatches validates the 25% token savings claim
// using 6 matches (like the phase doc example).
// Old format: ~147 tokens, New format: ~110 tokens
func TestGrepParser_TokenSavings_SixMatches(t *testing.T) {
	// Generate test input with 6 matches across 2 files (like phase doc example)
	input := `CLAUDE.md:127:### 3. Write tests
CLAUDE.md:129:Location: ` + "`internal/adapters/parsers/{category}/{command}_test.go`" + `
CLAUDE.md:152:# Run all tests with race detection
README.md:20:go install github.com/curtbushko/structured-cli/cmd/structured-cli@latest
README.md:104:| ` + "`go test`" + ` | Test results with pass/fail/skip counts |
README.md:200:For more information, see the documentation.
`

	parser := fileops.NewGrepParser()
	result, err := parser.Parse(strings.NewReader(input))
	require.NoError(t, err)

	output, ok := result.Data.(*fileops.GrepOutputCompact)
	require.True(t, ok, "result.Data is not *fileops.GrepOutputCompact, got %T", result.Data)

	// Verify correct parsing
	assert.Equal(t, 6, output.Total, "Total should be 6 matches")
	assert.Equal(t, 2, output.Files, "Files should be 2")
	require.Len(t, output.Results, 2, "Should have 2 file groups")

	// Calculate token counts for both formats
	// Token estimation: 1 token per ~4 chars (GPT-style tokenization approximation)
	compactJSON, err := json.Marshal(output)
	require.NoError(t, err)
	compactTokens := estimateTokenCount(string(compactJSON))

	// Build old format for comparison
	oldFormat := buildOldFormat(output)
	oldJSON, err := json.Marshal(oldFormat)
	require.NoError(t, err)
	oldTokens := estimateTokenCount(string(oldJSON))

	// Verify token savings
	savings := float64(oldTokens-compactTokens) / float64(oldTokens) * 100
	t.Logf("Old format tokens: %d, Compact format tokens: %d, Savings: %.1f%%", oldTokens, compactTokens, savings)
	t.Logf("Old JSON: %s", string(oldJSON))
	t.Logf("Compact JSON: %s", string(compactJSON))

	// The compact format should show positive token savings
	assert.Greater(t, savings, 15.0, "Token savings should be at least 15%% (actual: %.1f%%)", savings)
	// And ideally close to 25%
	assert.Greater(t, savings, 20.0, "Token savings should be at least 20%% for 6 matches (actual: %.1f%%)", savings)
}

// TestGrepParser_TokenSavings_FiftyMatches validates savings scale with more matches.
func TestGrepParser_TokenSavings_FiftyMatches(t *testing.T) {
	// Generate 50 matches across 10 files (5 per file)
	var builder strings.Builder
	for fileNum := range 10 {
		filename := fmt.Sprintf("file%03d.go", fileNum)
		for lineNum := range 5 {
			builder.WriteString(fmt.Sprintf("%s:%d:func example%d() { return nil }\n", filename, (lineNum+1)*10, lineNum))
		}
	}

	parser := fileops.NewGrepParser()
	result, err := parser.Parse(strings.NewReader(builder.String()))
	require.NoError(t, err)

	output, ok := result.Data.(*fileops.GrepOutputCompact)
	require.True(t, ok, "result.Data is not *fileops.GrepOutputCompact, got %T", result.Data)

	assert.Equal(t, 50, output.Total, "Total should be 50 matches")
	assert.Equal(t, 10, output.Files, "Files should be 10")

	// Calculate token counts
	compactJSON, err := json.Marshal(output)
	require.NoError(t, err)
	compactTokens := estimateTokenCount(string(compactJSON))

	oldFormat := buildOldFormat(output)
	oldJSON, err := json.Marshal(oldFormat)
	require.NoError(t, err)
	oldTokens := estimateTokenCount(string(oldJSON))

	savings := float64(oldTokens-compactTokens) / float64(oldTokens) * 100
	t.Logf("50 matches - Old: %d tokens, Compact: %d tokens, Savings: %.1f%%", oldTokens, compactTokens, savings)

	// More matches should show better savings due to reduced per-match overhead
	assert.Greater(t, savings, 20.0, "Token savings for 50 matches should be at least 20%% (actual: %.1f%%)", savings)
}

// TestGrepParser_TokenSavings_TwoHundredPlusMatches validates truncation works and savings.
func TestGrepParser_TokenSavings_TwoHundredPlusMatches(t *testing.T) {
	// Generate 300 matches across 30 files (10 per file)
	var builder strings.Builder
	for fileNum := range 30 {
		filename := fmt.Sprintf("file%03d.go", fileNum)
		for lineNum := range 10 {
			builder.WriteString(fmt.Sprintf("%s:%d:func example%d() { return nil }\n", filename, (lineNum+1)*10, lineNum))
		}
	}

	parser := fileops.NewGrepParser()
	result, err := parser.Parse(strings.NewReader(builder.String()))
	require.NoError(t, err)

	output, ok := result.Data.(*fileops.GrepOutputCompact)
	require.True(t, ok, "result.Data is not *fileops.GrepOutputCompact, got %T", result.Data)

	// Total reflects original count before truncation
	assert.Equal(t, 300, output.Total, "Total should reflect all 300 matches")
	assert.True(t, output.Truncated, "Truncated should be true for 300 matches")

	// Calculate actual matches returned (should be capped at 200)
	actualMatches := 0
	for _, group := range output.Results {
		actualMatches += len(group.Matches)
	}
	assert.LessOrEqual(t, actualMatches, 200, "Actual matches returned should be capped at 200")

	// Calculate token counts
	compactJSON, err := json.Marshal(output)
	require.NoError(t, err)
	compactTokens := estimateTokenCount(string(compactJSON))

	// For truncated output, compare against what old format would have been for same data
	oldFormat := buildOldFormat(output)
	oldJSON, err := json.Marshal(oldFormat)
	require.NoError(t, err)
	oldTokens := estimateTokenCount(string(oldJSON))

	savings := float64(oldTokens-compactTokens) / float64(oldTokens) * 100
	t.Logf("200+ matches - Old: %d tokens, Compact: %d tokens, Savings: %.1f%% (truncated to %d matches)",
		oldTokens, compactTokens, savings, actualMatches)

	// Truncation + compact format should still show savings
	assert.Greater(t, savings, 20.0, "Token savings for truncated output should be at least 20%% (actual: %.1f%%)", savings)
}

// TestGrepParser_TokenSavings_FiveMatches validates savings for small outputs.
func TestGrepParser_TokenSavings_FiveMatches(t *testing.T) {
	// Generate 5 matches in a single file
	input := `main.go:10:func main() {
main.go:25:    fmt.Println("Hello, World!")
main.go:30:    return nil
main.go:45:func helper() {
main.go:50:    return true
`

	parser := fileops.NewGrepParser()
	result, err := parser.Parse(strings.NewReader(input))
	require.NoError(t, err)

	output, ok := result.Data.(*fileops.GrepOutputCompact)
	require.True(t, ok, "result.Data is not *fileops.GrepOutputCompact, got %T", result.Data)

	assert.Equal(t, 5, output.Total, "Total should be 5 matches")
	assert.Equal(t, 1, output.Files, "Files should be 1")
	assert.False(t, output.Truncated, "Should not be truncated for 5 matches")

	// Calculate token counts
	compactJSON, err := json.Marshal(output)
	require.NoError(t, err)
	compactTokens := estimateTokenCount(string(compactJSON))

	oldFormat := buildOldFormat(output)
	oldJSON, err := json.Marshal(oldFormat)
	require.NoError(t, err)
	oldTokens := estimateTokenCount(string(oldJSON))

	savings := float64(oldTokens-compactTokens) / float64(oldTokens) * 100
	t.Logf("5 matches - Old: %d tokens, Compact: %d tokens, Savings: %.1f%%", oldTokens, compactTokens, savings)

	// Even small outputs should show positive savings
	assert.Greater(t, savings, 10.0, "Token savings for 5 matches should be positive (actual: %.1f%%)", savings)
}

// estimateTokenCount provides a rough token count estimation.
// Uses ~4 characters per token as a reasonable approximation for GPT-style tokenization.
func estimateTokenCount(s string) int {
	// More accurate estimation: count word boundaries and punctuation
	tokens := 0
	inWord := false
	for _, r := range s {
		switch r {
		case ' ', '\n', '\t':
			if inWord {
				tokens++
				inWord = false
			}
		case '{', '}', '[', ']', ':', ',', '"':
			if inWord {
				tokens++
				inWord = false
			}
			tokens++ // Punctuation is typically its own token
		default:
			inWord = true
		}
	}
	if inWord {
		tokens++
	}
	return tokens
}

// OldGrepOutput represents the old verbose format for token comparison.
type OldGrepOutput struct {
	Matches      []OldGrepMatch `json:"matches"`
	Count        int            `json:"count"`
	FilesMatched int            `json:"filesMatched"`
}

// OldGrepMatch represents a single match in the old verbose format.
type OldGrepMatch struct {
	File    string `json:"file"`
	Line    int    `json:"line"`
	Content string `json:"content"`
}

// buildOldFormat converts compact format to old verbose format for comparison.
func buildOldFormat(compact *fileops.GrepOutputCompact) *OldGrepOutput {
	old := &OldGrepOutput{
		Matches:      make([]OldGrepMatch, 0),
		Count:        0,
		FilesMatched: compact.Files,
	}

	for _, group := range compact.Results {
		for _, match := range group.Matches {
			old.Matches = append(old.Matches, OldGrepMatch{
				File:    group.Filename,
				Line:    match.Line,
				Content: match.Content,
			})
			old.Count++
		}
	}

	return old
}
