package python

import (
	"strings"
	"testing"
)

func TestPipAuditParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData PipAuditResult
	}{
		{
			name:  "empty output indicates clean audit",
			input: "",
			wantData: PipAuditResult{
				Success:         true,
				Vulnerabilities: []PipVulnerability{},
			},
		},
		{
			name:  "no vulnerabilities found",
			input: `No known vulnerabilities found`,
			wantData: PipAuditResult{
				Success:         true,
				Vulnerabilities: []PipVulnerability{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewPipAuditParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*PipAuditResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *PipAuditResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("PipAuditResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if len(got.Vulnerabilities) != len(tt.wantData.Vulnerabilities) {
				t.Errorf("PipAuditResult.Vulnerabilities length = %d, want %d", len(got.Vulnerabilities), len(tt.wantData.Vulnerabilities))
			}
		})
	}
}

func TestPipAuditParser_WithVulnerabilities(t *testing.T) {
	// pip-audit outputs in tabular format by default
	input := `Name       Version ID                  Fix Versions
---------- ------- ------------------- ------------
requests   2.25.0  PYSEC-2021-123      2.25.1
urllib3    1.26.4  GHSA-1234-5678-abcd 1.26.5

Found 2 known vulnerabilities in 2 packages`

	parser := NewPipAuditParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*PipAuditResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *PipAuditResult", result.Data)
	}

	if got.Success {
		t.Error("PipAuditResult.Success = true, want false when vulnerabilities present")
	}

	if len(got.Vulnerabilities) != 2 {
		t.Fatalf("PipAuditResult.Vulnerabilities length = %d, want 2", len(got.Vulnerabilities))
	}

	wantVulns := []PipVulnerability{
		{Name: "requests", Version: "2.25.0", ID: "PYSEC-2021-123", FixVersions: []string{"2.25.1"}},
		{Name: "urllib3", Version: "1.26.4", ID: "GHSA-1234-5678-abcd", FixVersions: []string{"1.26.5"}},
	}

	for i, want := range wantVulns {
		got := got.Vulnerabilities[i]
		if got.Name != want.Name {
			t.Errorf("Vulnerability[%d].Name = %q, want %q", i, got.Name, want.Name)
		}
		if got.Version != want.Version {
			t.Errorf("Vulnerability[%d].Version = %q, want %q", i, got.Version, want.Version)
		}
		if got.ID != want.ID {
			t.Errorf("Vulnerability[%d].ID = %q, want %q", i, got.ID, want.ID)
		}
	}

	if got.Summary.Total != 2 {
		t.Errorf("PipAuditResult.Summary.Total = %d, want 2", got.Summary.Total)
	}
}

func TestPipAuditParser_JSONFormat(t *testing.T) {
	// pip-audit can output JSON with --format json
	input := `[
  {
    "name": "requests",
    "version": "2.25.0",
    "vulns": [
      {
        "id": "PYSEC-2021-123",
        "fix_versions": ["2.25.1"],
        "aliases": ["CVE-2021-12345"],
        "description": "A security vulnerability in requests"
      }
    ]
  }
]`

	parser := NewPipAuditParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*PipAuditResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *PipAuditResult", result.Data)
	}

	if got.Success {
		t.Error("PipAuditResult.Success = true, want false when vulnerabilities present")
	}

	if len(got.Vulnerabilities) != 1 {
		t.Fatalf("PipAuditResult.Vulnerabilities length = %d, want 1", len(got.Vulnerabilities))
	}

	vuln := got.Vulnerabilities[0]
	if vuln.Name != "requests" {
		t.Errorf("Vulnerability.Name = %q, want %q", vuln.Name, "requests")
	}
	if vuln.ID != "PYSEC-2021-123" {
		t.Errorf("Vulnerability.ID = %q, want %q", vuln.ID, "PYSEC-2021-123")
	}
	if vuln.Description != "A security vulnerability in requests" {
		t.Errorf("Vulnerability.Description = %q, want %q", vuln.Description, "A security vulnerability in requests")
	}
}

func TestPipAuditParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches pip-audit",
			cmd:         "pip-audit",
			subcommands: []string{},
			want:        true,
		},
		{
			name:        "matches pip-audit with flags",
			cmd:         "pip-audit",
			subcommands: []string{"--format", "json"},
			want:        true,
		},
		{
			name:        "does not match pip",
			cmd:         "pip",
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

	parser := NewPipAuditParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestPipAuditParser_Schema(t *testing.T) {
	parser := NewPipAuditParser()
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

	requiredProps := []string{"success", "vulnerabilities"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}
