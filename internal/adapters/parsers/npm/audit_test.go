package npm

import (
	"strings"
	"testing"
)

func TestAuditParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData AuditResult
	}{
		{
			name:  "no vulnerabilities",
			input: "found 0 vulnerabilities",
			wantData: AuditResult{
				Success:         true,
				Vulnerabilities: []Vulnerability{},
				Summary:         VulnerabilitySummary{Total: 0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewAuditParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*AuditResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *AuditResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("AuditResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if got.Summary.Total != tt.wantData.Summary.Total {
				t.Errorf("AuditResult.Summary.Total = %v, want %v", got.Summary.Total, tt.wantData.Summary.Total)
			}
		})
	}
}

func TestAuditParser_WithVulnerabilities(t *testing.T) {
	input := `# npm audit report

lodash  <4.17.21
Severity: high
Prototype Pollution - https://npmjs.com/advisories/1065
fix available via ` + "`npm audit fix`" + `
node_modules/lodash

minimist  <1.2.6
Severity: critical
Prototype Pollution - https://npmjs.com/advisories/1179
fix available via ` + "`npm audit fix --force`" + `
node_modules/minimist

3 vulnerabilities (1 moderate, 1 high, 1 critical)

To address all issues, run:
  npm audit fix`

	parser := NewAuditParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*AuditResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *AuditResult", result.Data)
	}

	if got.Success {
		t.Error("AuditResult.Success = true, want false when vulnerabilities present")
	}

	if got.Summary.Total != 3 {
		t.Errorf("AuditResult.Summary.Total = %d, want 3", got.Summary.Total)
	}

	if got.Summary.Moderate != 1 {
		t.Errorf("AuditResult.Summary.Moderate = %d, want 1", got.Summary.Moderate)
	}

	if got.Summary.High != 1 {
		t.Errorf("AuditResult.Summary.High = %d, want 1", got.Summary.High)
	}

	if got.Summary.Critical != 1 {
		t.Errorf("AuditResult.Summary.Critical = %d, want 1", got.Summary.Critical)
	}

	if len(got.Vulnerabilities) != 2 {
		t.Fatalf("AuditResult.Vulnerabilities length = %d, want 2", len(got.Vulnerabilities))
	}

	// Check first vulnerability
	if got.Vulnerabilities[0].Name != "lodash" {
		t.Errorf("Vulnerability[0].Name = %q, want %q", got.Vulnerabilities[0].Name, "lodash")
	}

	if got.Vulnerabilities[0].Severity != "high" {
		t.Errorf("Vulnerability[0].Severity = %q, want %q", got.Vulnerabilities[0].Severity, "high")
	}
}

func TestAuditParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches npm audit",
			cmd:         "npm",
			subcommands: []string{"audit"},
			want:        true,
		},
		{
			name:        "matches npm audit with fix",
			cmd:         "npm",
			subcommands: []string{"audit", "fix"},
			want:        true,
		},
		{
			name:        "does not match npm install",
			cmd:         "npm",
			subcommands: []string{"install"},
			want:        false,
		},
		{
			name:        "does not match yarn audit",
			cmd:         "yarn",
			subcommands: []string{"audit"},
			want:        false,
		},
		{
			name:        "does not match empty",
			cmd:         "",
			subcommands: []string{},
			want:        false,
		},
	}

	parser := NewAuditParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestAuditParser_Schema(t *testing.T) {
	parser := NewAuditParser()
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

	requiredProps := []string{"success", "vulnerabilities", "summary"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}
