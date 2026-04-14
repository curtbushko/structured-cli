package lint

import (
	"fmt"
	"strings"
	"testing"
)

func TestGolangCILintParser_Success(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantTotalIssues int
	}{
		{
			name:            "empty output indicates clean lint",
			input:           "",
			wantTotalIssues: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewGolangCILintParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*GolangCILintResultCompact)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *GolangCILintResultCompact", result.Data)
			}

			if got.TotalIssues != tt.wantTotalIssues {
				t.Errorf("GolangCILintResultCompact.TotalIssues = %v, want %v", got.TotalIssues, tt.wantTotalIssues)
			}
		})
	}
}

func TestGolangCILintParser_SingleIssue(t *testing.T) {
	// golangci-lint outputs issues in format: file:line:column: message (linter)
	input := `main.go:10:5: Error return value of 'foo' is not checked (errcheck)`

	parser := NewGolangCILintParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*GolangCILintResultCompact)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *GolangCILintResultCompact", result.Data)
	}

	if got.TotalIssues != 1 {
		t.Errorf("TotalIssues = %d, want 1", got.TotalIssues)
	}

	if got.FilesWithIssues != 1 {
		t.Errorf("FilesWithIssues = %d, want 1", got.FilesWithIssues)
	}

	if got.SeverityCounts[SeverityError] != 1 {
		t.Errorf("SeverityCounts[error] = %d, want 1", got.SeverityCounts[SeverityError])
	}

	if len(got.Results) != 1 {
		t.Fatalf("Results length = %d, want 1", len(got.Results))
	}

	// Verify the file group
	fileGroup := got.Results[0]
	fileName, ok := fileGroup[0].(string)
	if !ok || fileName != "main.go" {
		t.Errorf("FileGroup[0] = %v, want %q", fileGroup[0], "main.go")
	}

	issues, ok := fileGroup[2].([]IssueTuple)
	if !ok || len(issues) != 1 {
		t.Fatalf("FileGroup[2] type = %T, want []IssueTuple with 1 issue", fileGroup[2])
	}

	// Verify the issue tuple: [line, severity, message, linter]
	if issues[0][0] != 10 {
		t.Errorf("Issue line = %v, want 10", issues[0][0])
	}
	if issues[0][1] != SeverityError {
		t.Errorf("Issue severity = %v, want %q", issues[0][1], SeverityError)
	}
	if issues[0][2] != "Error return value of 'foo' is not checked" {
		t.Errorf("Issue message = %v, want %q", issues[0][2], "Error return value of 'foo' is not checked")
	}
	if issues[0][3] != "errcheck" {
		t.Errorf("Issue linter = %v, want %q", issues[0][3], "errcheck")
	}
}

func TestGolangCILintParser_MultipleIssues(t *testing.T) {
	input := `main.go:10:5: Error return value of 'foo' is not checked (errcheck)
utils.go:25:10: S1000: should use a simple channel send/receive instead of select (gosimple)
handler.go:50:3: printf: fmt.Printf format %s has arg x of wrong type int (govet)`

	parser := NewGolangCILintParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*GolangCILintResultCompact)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *GolangCILintResultCompact", result.Data)
	}

	if got.TotalIssues != 3 {
		t.Errorf("TotalIssues = %d, want 3", got.TotalIssues)
	}

	if got.FilesWithIssues != 3 {
		t.Errorf("FilesWithIssues = %d, want 3", got.FilesWithIssues)
	}

	// Verify severity counts (all are errors in golangci-lint)
	if got.SeverityCounts[SeverityError] != 3 {
		t.Errorf("SeverityCounts[error] = %d, want 3", got.SeverityCounts[SeverityError])
	}

	// Verify results are grouped by file (3 files)
	if len(got.Results) != 3 {
		t.Fatalf("Results length = %d, want 3", len(got.Results))
	}
}

func TestGolangCILintParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches golangci-lint run",
			cmd:         "golangci-lint",
			subcommands: []string{"run"},
			want:        true,
		},
		{
			name:        "matches golangci-lint with path",
			cmd:         "golangci-lint",
			subcommands: []string{"run", "./..."},
			want:        true,
		},
		{
			name:        "matches golangci-lint with no subcommands",
			cmd:         "golangci-lint",
			subcommands: []string{},
			want:        true,
		},
		{
			name:        "does not match go tool",
			cmd:         "go",
			subcommands: []string{"vet"},
			want:        false,
		},
		{
			name:        "does not match empty command",
			cmd:         "",
			subcommands: []string{},
			want:        false,
		},
	}

	parser := NewGolangCILintParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestGolangCILintParser_Schema(t *testing.T) {
	parser := NewGolangCILintParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema.ID should not be empty")
	}

	if schema.Title == "" {
		t.Error("Schema.Title should not be empty")
	}

	if schema.Type != schemaTypeObject {
		t.Errorf("Schema.Type = %q, want %q", schema.Type, schemaTypeObject)
	}

	// Verify compact format properties exist
	requiredProps := []string{"total_issues", "files_with_issues", "severity_counts", "results", "truncated"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}

func TestGolangCILintParser_DifferentFormats(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantTotalIssues int
		wantFiles       int
	}{
		{
			name:            "issue without column (file:line: message)",
			input:           "main.go:10: unused variable (deadcode)",
			wantTotalIssues: 1,
			wantFiles:       1,
		},
		{
			name:            "full path file",
			input:           "/home/user/project/pkg/handler.go:42:8: ineffectual assignment (ineffassign)",
			wantTotalIssues: 1,
			wantFiles:       1,
		},
		{
			name:            "relative path file",
			input:           "./internal/app/main.go:15:2: exported function without comment (golint)",
			wantTotalIssues: 1,
			wantFiles:       1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewGolangCILintParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*GolangCILintResultCompact)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *GolangCILintResultCompact", result.Data)
			}

			if got.TotalIssues != tt.wantTotalIssues {
				t.Errorf("TotalIssues = %d, want %d", got.TotalIssues, tt.wantTotalIssues)
			}

			if got.FilesWithIssues != tt.wantFiles {
				t.Errorf("FilesWithIssues = %d, want %d", got.FilesWithIssues, tt.wantFiles)
			}
		})
	}
}

// TestGolangCILintParser_MultiLinter tests parsing output with multiple linters
func TestGolangCILintParser_MultiLinter(t *testing.T) {
	input := `main.go:10:5: Error return value of 'foo' is not checked (errcheck)
main.go:15:3: S1000: should use a simple channel send/receive instead of select (gosimple)
main.go:20:1: printf: fmt.Printf format %s has arg x of wrong type int (govet)`

	parser := NewGolangCILintParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*GolangCILintResultCompact)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *GolangCILintResultCompact", result.Data)
	}

	if got.TotalIssues != 3 {
		t.Errorf("TotalIssues = %d, want 3", got.TotalIssues)
	}

	// All issues are in one file
	if got.FilesWithIssues != 1 {
		t.Errorf("FilesWithIssues = %d, want 1", got.FilesWithIssues)
	}

	// Verify linter names are in tuples
	if len(got.Results) != 1 {
		t.Fatalf("Results length = %d, want 1", len(got.Results))
	}

	issues, ok := got.Results[0][2].([]IssueTuple)
	if !ok {
		t.Fatalf("FileGroup[2] type = %T, want []IssueTuple", got.Results[0][2])
	}

	expectedLinters := []string{"errcheck", "gosimple", "govet"}
	for i, issue := range issues {
		if issue[3] != expectedLinters[i] {
			t.Errorf("Issue[%d] linter = %v, want %q", i, issue[3], expectedLinters[i])
		}
	}
}

// TestGolangCILintParser_Truncation tests truncation with 250 issues
func TestGolangCILintParser_Truncation(t *testing.T) {
	var sb strings.Builder
	totalIssues := 250
	for i := range totalIssues {
		file := fmt.Sprintf("file%d.go", i%15)
		sb.WriteString(fmt.Sprintf("%s:%d:1: test message %d (errcheck)\n", file, i+1, i))
	}

	parser := NewGolangCILintParser()
	result, err := parser.Parse(strings.NewReader(sb.String()))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*GolangCILintResultCompact)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *GolangCILintResultCompact", result.Data)
	}

	// TotalIssues should reflect original count
	if got.TotalIssues != totalIssues {
		t.Errorf("TotalIssues = %d, want %d", got.TotalIssues, totalIssues)
	}

	// Count actual issues returned
	actualIssueCount := 0
	for _, fileGroup := range got.Results {
		issues, ok := fileGroup[2].([]IssueTuple)
		if ok {
			actualIssueCount += len(issues)
		}
	}

	// Should be truncated to MaxTotalIssues (200)
	if actualIssueCount > MaxTotalIssues {
		t.Errorf("Returned issues = %d, want <= %d", actualIssueCount, MaxTotalIssues)
	}

	// Truncated count should be 50
	expectedTruncated := totalIssues - MaxTotalIssues
	if got.Truncated != expectedTruncated {
		t.Errorf("Truncated = %d, want %d", got.Truncated, expectedTruncated)
	}
}

// TestGolangCILintParser_FileGrouping tests proper grouping by file
func TestGolangCILintParser_FileGrouping(t *testing.T) {
	input := `file_a.go:10:1: error 1 (errcheck)
file_a.go:20:1: error 2 (errcheck)
file_b.go:5:1: error 3 (govet)
file_a.go:30:1: error 4 (gosimple)`

	parser := NewGolangCILintParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*GolangCILintResultCompact)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *GolangCILintResultCompact", result.Data)
	}

	if got.TotalIssues != 4 {
		t.Errorf("TotalIssues = %d, want 4", got.TotalIssues)
	}

	if got.FilesWithIssues != 2 {
		t.Errorf("FilesWithIssues = %d, want 2", got.FilesWithIssues)
	}

	// Verify grouping - file_a should have 3 issues, file_b should have 1
	if len(got.Results) != 2 {
		t.Fatalf("Results length = %d, want 2", len(got.Results))
	}
}

// TestGolangCILintParser_NoIssues tests compact format with no issues
func TestGolangCILintParser_NoIssues(t *testing.T) {
	parser := NewGolangCILintParser()
	result, err := parser.Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*GolangCILintResultCompact)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *GolangCILintResultCompact", result.Data)
	}

	if got.TotalIssues != 0 {
		t.Errorf("TotalIssues = %d, want 0", got.TotalIssues)
	}

	if got.FilesWithIssues != 0 {
		t.Errorf("FilesWithIssues = %d, want 0", got.FilesWithIssues)
	}

	if len(got.Results) != 0 {
		t.Errorf("Results length = %d, want 0", len(got.Results))
	}

	if got.Truncated != 0 {
		t.Errorf("Truncated = %d, want 0", got.Truncated)
	}
}
