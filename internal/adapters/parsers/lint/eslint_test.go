package lint

import (
	"strings"
	"testing"
)

func TestESLintParser_Success(t *testing.T) {
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
			parser := NewESLintParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*ESLintResultCompact)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *ESLintResultCompact", result.Data)
			}

			if got.TotalIssues != tt.wantTotalIssues {
				t.Errorf("ESLintResultCompact.TotalIssues = %v, want %v", got.TotalIssues, tt.wantTotalIssues)
			}
		})
	}
}

func TestESLintParser_SingleIssue(t *testing.T) {
	// ESLint default formatter output: /path/file.js:line:column: message (rule)
	input := `/home/user/project/src/index.js
  10:5  error  'foo' is not defined  no-undef`

	parser := NewESLintParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*ESLintResultCompact)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ESLintResultCompact", result.Data)
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
	if !ok || fileName != "/home/user/project/src/index.js" {
		t.Errorf("FileGroup[0] = %v, want %q", fileGroup[0], "/home/user/project/src/index.js")
	}

	issueCount, ok := fileGroup[1].(int)
	if !ok || issueCount != 1 {
		t.Errorf("FileGroup[1] = %v, want 1", fileGroup[1])
	}

	issues, ok := fileGroup[2].([]IssueTuple)
	if !ok || len(issues) != 1 {
		t.Fatalf("FileGroup[2] type = %T, want []IssueTuple with 1 issue", fileGroup[2])
	}

	// Verify the issue tuple: [line, severity, message, rule_id]
	if issues[0][0] != 10 {
		t.Errorf("Issue line = %v, want 10", issues[0][0])
	}
	if issues[0][1] != SeverityError {
		t.Errorf("Issue severity = %v, want %q", issues[0][1], SeverityError)
	}
	if issues[0][2] != "'foo' is not defined" {
		t.Errorf("Issue message = %v, want \"'foo' is not defined\"", issues[0][2])
	}
	if issues[0][3] != "no-undef" {
		t.Errorf("Issue rule = %v, want \"no-undef\"", issues[0][3])
	}
}

func TestESLintParser_MultipleIssues(t *testing.T) {
	input := `/home/user/project/src/index.js
  10:5  error  'foo' is not defined  no-undef
  15:1  warning  Unexpected console statement  no-console

/home/user/project/src/utils.js
  5:10  error  'bar' is assigned a value but never used  no-unused-vars`

	parser := NewESLintParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*ESLintResultCompact)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ESLintResultCompact", result.Data)
	}

	if got.TotalIssues != 3 {
		t.Errorf("TotalIssues = %d, want 3", got.TotalIssues)
	}

	if got.FilesWithIssues != 2 {
		t.Errorf("FilesWithIssues = %d, want 2", got.FilesWithIssues)
	}

	// Verify severity counts
	if got.SeverityCounts[SeverityError] != 2 {
		t.Errorf("SeverityCounts[error] = %d, want 2", got.SeverityCounts[SeverityError])
	}
	if got.SeverityCounts[SeverityWarning] != 1 {
		t.Errorf("SeverityCounts[warning] = %d, want 1", got.SeverityCounts[SeverityWarning])
	}

	// Verify results are grouped by file (2 files)
	if len(got.Results) != 2 {
		t.Fatalf("Results length = %d, want 2", len(got.Results))
	}
}

func TestESLintParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches eslint with no subcommands",
			cmd:         "eslint",
			subcommands: []string{},
			want:        true,
		},
		{
			name:        "matches eslint with file path",
			cmd:         "eslint",
			subcommands: []string{"src/"},
			want:        true,
		},
		{
			name:        "matches eslint with flags",
			cmd:         "eslint",
			subcommands: []string{"--fix", "src/"},
			want:        true,
		},
		{
			name:        "does not match npx eslint (npx is the command)",
			cmd:         "npx",
			subcommands: []string{"eslint"},
			want:        false,
		},
		{
			name:        "does not match node",
			cmd:         "node",
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

	parser := NewESLintParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestESLintParser_Schema(t *testing.T) {
	parser := NewESLintParser()
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

func TestESLintParser_SummaryLine(t *testing.T) {
	// ESLint outputs a summary line at the end
	input := `/home/user/project/src/index.js
  10:5  error  'foo' is not defined  no-undef

✖ 1 problem (1 error, 0 warnings)`

	parser := NewESLintParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*ESLintResultCompact)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ESLintResultCompact", result.Data)
	}

	// Should parse exactly 1 issue, ignoring summary line
	if got.TotalIssues != 1 {
		t.Errorf("TotalIssues = %d, want 1", got.TotalIssues)
	}
}

func TestESLintParser_WarningOnly(t *testing.T) {
	input := `/home/user/project/src/index.js
  10:5  warning  Unexpected console statement  no-console`

	parser := NewESLintParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*ESLintResultCompact)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ESLintResultCompact", result.Data)
	}

	if got.TotalIssues != 1 {
		t.Errorf("TotalIssues = %d, want 1", got.TotalIssues)
	}

	if got.SeverityCounts[SeverityWarning] != 1 {
		t.Errorf("SeverityCounts[warning] = %d, want 1", got.SeverityCounts[SeverityWarning])
	}
}

// TestESLintParser_CompactFormat tests the new compact output format
func TestESLintParser_CompactFormat(t *testing.T) {
	input := `/home/user/project/src/index.js
  10:5  error  'foo' is not defined  no-undef
  15:1  warning  Unexpected console statement  no-console

/home/user/project/src/utils.js
  5:10  error  'bar' is assigned a value but never used  no-unused-vars`

	parser := NewESLintParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*ESLintResultCompact)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ESLintResultCompact", result.Data)
	}

	// Verify compact format fields
	if got.TotalIssues != 3 {
		t.Errorf("TotalIssues = %d, want 3", got.TotalIssues)
	}

	if got.FilesWithIssues != 2 {
		t.Errorf("FilesWithIssues = %d, want 2", got.FilesWithIssues)
	}

	// Verify severity counts
	if got.SeverityCounts[SeverityError] != 2 {
		t.Errorf("SeverityCounts[error] = %d, want 2", got.SeverityCounts[SeverityError])
	}
	if got.SeverityCounts[SeverityWarning] != 1 {
		t.Errorf("SeverityCounts[warning] = %d, want 1", got.SeverityCounts[SeverityWarning])
	}

	// Verify results are grouped by file
	if len(got.Results) != 2 {
		t.Fatalf("Results length = %d, want 2", len(got.Results))
	}

	// Verify truncated is 0 (no truncation for small dataset)
	if got.Truncated != 0 {
		t.Errorf("Truncated = %d, want 0", got.Truncated)
	}
}

// TestESLintParser_CompactFormat_EmptyInput tests compact format with no issues
func TestESLintParser_CompactFormat_EmptyInput(t *testing.T) {
	parser := NewESLintParser()
	result, err := parser.Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*ESLintResultCompact)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ESLintResultCompact", result.Data)
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
}

// TestESLintParser_CompactFormat_SeverityMapping tests ESLint severity mapping
func TestESLintParser_CompactFormat_SeverityMapping(t *testing.T) {
	input := `/home/user/project/src/index.js
  10:5  error  Error message  rule1
  15:1  warning  Warning message  rule2`

	parser := NewESLintParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*ESLintResultCompact)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ESLintResultCompact", result.Data)
	}

	// ESLint severity "error" and "warning" should map to standard "error" and "warning"
	if got.SeverityCounts[SeverityError] != 1 {
		t.Errorf("SeverityCounts[error] = %d, want 1", got.SeverityCounts[SeverityError])
	}
	if got.SeverityCounts[SeverityWarning] != 1 {
		t.Errorf("SeverityCounts[warning] = %d, want 1", got.SeverityCounts[SeverityWarning])
	}

	// Check the issue tuples have standardized severity
	if len(got.Results) != 1 {
		t.Fatalf("Results length = %d, want 1", len(got.Results))
	}

	fileGroup := got.Results[0]
	issues, ok := fileGroup[2].([]IssueTuple)
	if !ok {
		t.Fatalf("FileGroup[2] type = %T, want []IssueTuple", fileGroup[2])
	}

	if len(issues) != 2 {
		t.Fatalf("issues length = %d, want 2", len(issues))
	}

	// First issue should be error
	if issues[0][1] != SeverityError {
		t.Errorf("issues[0] severity = %v, want %q", issues[0][1], SeverityError)
	}
	// Second issue should be warning
	if issues[1][1] != SeverityWarning {
		t.Errorf("issues[1] severity = %v, want %q", issues[1][1], SeverityWarning)
	}
}

// TestESLintParser_CompactFormat_SchemaUpdated tests that schema has compact format fields
func TestESLintParser_CompactFormat_SchemaUpdated(t *testing.T) {
	parser := NewESLintParser()
	schema := parser.Schema()

	// Verify compact format properties exist
	requiredProps := []string{"total_issues", "files_with_issues", "severity_counts", "results", "truncated"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}

// TestESLintParser_Truncation tests that 250 issues are truncated to 200
func TestESLintParser_Truncation(t *testing.T) {
	// Generate 250 issues across multiple files (13 files with ~20 issues each)
	var sb strings.Builder
	totalIssues := 250
	issuesPerFile := 20
	numFiles := (totalIssues + issuesPerFile - 1) / issuesPerFile

	for f := range numFiles {
		sb.WriteString("/home/user/project/src/file" + string(rune('A'+f)) + ".js\n")
		issuesInThisFile := issuesPerFile
		if f == numFiles-1 {
			issuesInThisFile = totalIssues - (numFiles-1)*issuesPerFile
		}
		for i := range issuesInThisFile {
			sb.WriteString("  " + string(rune('1'+i%9)) + "0:5  error  Error message " + string(rune('a'+i%26)) + "  no-undef\n")
		}
		sb.WriteString("\n")
	}

	parser := NewESLintParser()
	result, err := parser.Parse(strings.NewReader(sb.String()))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*ESLintResultCompact)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ESLintResultCompact", result.Data)
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
		t.Errorf("Returned issues = %d, want <= %d (MaxTotalIssues)", actualIssueCount, MaxTotalIssues)
	}

	// Truncated count should be 50 (250 - 200)
	expectedTruncated := totalIssues - MaxTotalIssues
	if got.Truncated != expectedTruncated {
		t.Errorf("Truncated = %d, want %d", got.Truncated, expectedTruncated)
	}
}

// TestESLintParser_PerFileTruncation tests that 30 issues in single file are truncated to 20
func TestESLintParser_PerFileTruncation(t *testing.T) {
	// Generate 30 issues in a single file
	var sb strings.Builder
	sb.WriteString("/home/user/project/src/bigfile.js\n")
	for i := range 30 {
		sb.WriteString("  " + string(rune('1'+i%9)) + string(rune('0'+i%10)) + ":5  error  Error message " + string(rune('a'+i%26)) + "  no-undef\n")
	}

	parser := NewESLintParser()
	result, err := parser.Parse(strings.NewReader(sb.String()))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*ESLintResultCompact)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ESLintResultCompact", result.Data)
	}

	// TotalIssues should reflect original count
	if got.TotalIssues != 30 {
		t.Errorf("TotalIssues = %d, want 30", got.TotalIssues)
	}

	// Should have 1 file
	if len(got.Results) != 1 {
		t.Fatalf("Results length = %d, want 1", len(got.Results))
	}

	// Issues in that file should be truncated to MaxIssuesPerFile (20)
	issues, ok := got.Results[0][2].([]IssueTuple)
	if !ok {
		t.Fatalf("FileGroup[2] type = %T, want []IssueTuple", got.Results[0][2])
	}

	if len(issues) > MaxIssuesPerFile {
		t.Errorf("Issues in file = %d, want <= %d (MaxIssuesPerFile)", len(issues), MaxIssuesPerFile)
	}

	// Truncated count should be 10 (30 - 20)
	expectedTruncated := 30 - MaxIssuesPerFile
	if got.Truncated != expectedTruncated {
		t.Errorf("Truncated = %d, want %d", got.Truncated, expectedTruncated)
	}
}
