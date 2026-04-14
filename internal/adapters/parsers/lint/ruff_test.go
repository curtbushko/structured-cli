package lint

import (
	"fmt"
	"strings"
	"testing"
)

const schemaTypeObject = "object"

func TestRuffParser_Success(t *testing.T) {
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
		{
			name:            "all checks passed message",
			input:           "All checks passed!",
			wantTotalIssues: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewRuffParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*RuffResultCompact)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *RuffResultCompact", result.Data)
			}

			if got.TotalIssues != tt.wantTotalIssues {
				t.Errorf("RuffResultCompact.TotalIssues = %v, want %v", got.TotalIssues, tt.wantTotalIssues)
			}
		})
	}
}

func TestRuffParser_SingleIssue(t *testing.T) {
	// Ruff outputs issues in format: file:line:column: CODE message
	input := `main.py:10:1: F401 ` + "`" + `os` + "`" + ` imported but unused`

	parser := NewRuffParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*RuffResultCompact)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *RuffResultCompact", result.Data)
	}

	if got.TotalIssues != 1 {
		t.Errorf("TotalIssues = %d, want 1", got.TotalIssues)
	}

	if got.FilesWithIssues != 1 {
		t.Errorf("FilesWithIssues = %d, want 1", got.FilesWithIssues)
	}

	// F401 should map to warning
	if got.SeverityCounts[SeverityWarning] != 1 {
		t.Errorf("SeverityCounts[warning] = %d, want 1", got.SeverityCounts[SeverityWarning])
	}

	if len(got.Results) != 1 {
		t.Fatalf("Results length = %d, want 1", len(got.Results))
	}

	// Verify the file group
	fileGroup := got.Results[0]
	fileName, ok := fileGroup[0].(string)
	if !ok || fileName != "main.py" {
		t.Errorf("FileGroup[0] = %v, want %q", fileGroup[0], "main.py")
	}

	issues, ok := fileGroup[2].([]IssueTuple)
	if !ok || len(issues) != 1 {
		t.Fatalf("FileGroup[2] type = %T, want []IssueTuple with 1 issue", fileGroup[2])
	}

	// Verify the issue tuple: [line, severity, message, rule_code]
	if issues[0][0] != 10 {
		t.Errorf("Issue line = %v, want 10", issues[0][0])
	}
	if issues[0][1] != SeverityWarning {
		t.Errorf("Issue severity = %v, want %q", issues[0][1], SeverityWarning)
	}
	if issues[0][3] != "F401" {
		t.Errorf("Issue rule = %v, want %q", issues[0][3], "F401")
	}
}

func TestRuffParser_MultipleIssues(t *testing.T) {
	input := `main.py:10:1: F401 ` + "`" + `os` + "`" + ` imported but unused
utils.py:25:5: E501 Line too long (120 > 88 characters)
app.py:42:1: I001 Import block is un-sorted or un-formatted`

	parser := NewRuffParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*RuffResultCompact)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *RuffResultCompact", result.Data)
	}

	if got.TotalIssues != 3 {
		t.Errorf("TotalIssues = %d, want 3", got.TotalIssues)
	}

	if got.FilesWithIssues != 3 {
		t.Errorf("FilesWithIssues = %d, want 3", got.FilesWithIssues)
	}

	// Verify severity counts: F401->warning, E501->error, I001->info
	if got.SeverityCounts[SeverityWarning] != 1 {
		t.Errorf("SeverityCounts[warning] = %d, want 1", got.SeverityCounts[SeverityWarning])
	}
	if got.SeverityCounts[SeverityError] != 1 {
		t.Errorf("SeverityCounts[error] = %d, want 1", got.SeverityCounts[SeverityError])
	}
	if got.SeverityCounts[SeverityInfo] != 1 {
		t.Errorf("SeverityCounts[info] = %d, want 1", got.SeverityCounts[SeverityInfo])
	}

	// Verify results are grouped by file (3 files)
	if len(got.Results) != 3 {
		t.Fatalf("Results length = %d, want 3", len(got.Results))
	}
}

func TestRuffParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches ruff check",
			cmd:         "ruff",
			subcommands: []string{"check"},
			want:        true,
		},
		{
			name:        "matches ruff check with path",
			cmd:         "ruff",
			subcommands: []string{"check", "."},
			want:        true,
		},
		{
			name:        "matches ruff with no subcommands",
			cmd:         "ruff",
			subcommands: []string{},
			want:        true,
		},
		{
			name:        "does not match python",
			cmd:         "python",
			subcommands: []string{"-m", "ruff"},
			want:        false,
		},
		{
			name:        "does not match empty command",
			cmd:         "",
			subcommands: []string{},
			want:        false,
		},
	}

	parser := NewRuffParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestRuffParser_Schema(t *testing.T) {
	parser := NewRuffParser()
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

func TestRuffParser_DifferentFormats(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantTotalIssues int
		wantFiles       int
	}{
		{
			name:            "full path file",
			input:           "/home/user/project/src/main.py:10:1: F401 unused import",
			wantTotalIssues: 1,
			wantFiles:       1,
		},
		{
			name:            "relative path file",
			input:           "./src/main.py:15:2: E999 SyntaxError",
			wantTotalIssues: 1,
			wantFiles:       1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewRuffParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*RuffResultCompact)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *RuffResultCompact", result.Data)
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

func TestRuffParser_FoundIssuesSummary(t *testing.T) {
	// Ruff may output a summary line at the end
	input := `main.py:10:1: F401 ` + "`" + `os` + "`" + ` imported but unused
Found 1 error.`

	parser := NewRuffParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*RuffResultCompact)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *RuffResultCompact", result.Data)
	}

	// Should parse exactly 1 issue, ignoring summary line
	if got.TotalIssues != 1 {
		t.Errorf("TotalIssues = %d, want 1", got.TotalIssues)
	}
}

// TestRuffParser_RuleCodeExtraction tests that rule codes are included in tuples
func TestRuffParser_RuleCodeExtraction(t *testing.T) {
	input := `main.py:10:1: E501 Line too long
main.py:15:1: F401 unused import
main.py:20:1: W503 line break after operator
main.py:25:1: I001 imports unsorted`

	parser := NewRuffParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*RuffResultCompact)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *RuffResultCompact", result.Data)
	}

	if len(got.Results) != 1 {
		t.Fatalf("Results length = %d, want 1", len(got.Results))
	}

	issues, ok := got.Results[0][2].([]IssueTuple)
	if !ok {
		t.Fatalf("FileGroup[2] type = %T, want []IssueTuple", got.Results[0][2])
	}

	expectedCodes := []string{"E501", "F401", "W503", "I001"}
	for i, issue := range issues {
		if issue[3] != expectedCodes[i] {
			t.Errorf("Issue[%d] code = %v, want %q", i, issue[3], expectedCodes[i])
		}
	}
}

// TestRuffParser_SeverityMapping tests rule code prefix to severity mapping
func TestRuffParser_SeverityMapping(t *testing.T) {
	input := `a.py:1:1: E501 error code
b.py:1:1: W503 warning code
c.py:1:1: F401 flake8 warning
d.py:1:1: I001 info code`

	parser := NewRuffParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*RuffResultCompact)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *RuffResultCompact", result.Data)
	}

	// E* -> error
	if got.SeverityCounts[SeverityError] != 1 {
		t.Errorf("SeverityCounts[error] = %d, want 1 (E*)", got.SeverityCounts[SeverityError])
	}

	// W*, F* -> warning (2 total)
	if got.SeverityCounts[SeverityWarning] != 2 {
		t.Errorf("SeverityCounts[warning] = %d, want 2 (W*, F*)", got.SeverityCounts[SeverityWarning])
	}

	// I* -> info
	if got.SeverityCounts[SeverityInfo] != 1 {
		t.Errorf("SeverityCounts[info] = %d, want 1 (I*)", got.SeverityCounts[SeverityInfo])
	}
}

// TestRuffParser_MassiveTruncation tests truncation with 500 issues
func TestRuffParser_MassiveTruncation(t *testing.T) {
	var sb strings.Builder
	totalIssues := 500
	for i := range totalIssues {
		file := fmt.Sprintf("file%d.py", i%25)
		sb.WriteString(fmt.Sprintf("%s:%d:1: E501 error message %d\n", file, i+1, i))
	}

	parser := NewRuffParser()
	result, err := parser.Parse(strings.NewReader(sb.String()))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*RuffResultCompact)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *RuffResultCompact", result.Data)
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

	// Truncated count should be 300 (500 - 200)
	expectedTruncated := totalIssues - MaxTotalIssues
	if got.Truncated != expectedTruncated {
		t.Errorf("Truncated = %d, want %d", got.Truncated, expectedTruncated)
	}
}

// TestRuffParser_NoIssues tests compact format with no issues
func TestRuffParser_NoIssues(t *testing.T) {
	parser := NewRuffParser()
	result, err := parser.Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*RuffResultCompact)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *RuffResultCompact", result.Data)
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
