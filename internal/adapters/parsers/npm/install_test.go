package npm

import (
	"strings"
	"testing"
)

func TestInstallParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData InstallResult
	}{
		{
			name:  "empty output indicates success",
			input: "",
			wantData: InstallResult{
				Success:  true,
				Warnings: []string{},
			},
		},
		{
			name: "simple install success",
			input: `
added 50 packages, and audited 51 packages in 2s

8 packages are looking for funding
  run ` + "`npm fund`" + ` for details

found 0 vulnerabilities`,
			wantData: InstallResult{
				Success:         true,
				PackagesAdded:   50,
				PackagesAudited: 51,
				Funding:         8,
				Vulnerabilities: VulnerabilitySummary{Total: 0},
				Warnings:        []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewInstallParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*InstallResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *InstallResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("InstallResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if got.PackagesAdded != tt.wantData.PackagesAdded {
				t.Errorf("InstallResult.PackagesAdded = %v, want %v", got.PackagesAdded, tt.wantData.PackagesAdded)
			}

			if got.PackagesAudited != tt.wantData.PackagesAudited {
				t.Errorf("InstallResult.PackagesAudited = %v, want %v", got.PackagesAudited, tt.wantData.PackagesAudited)
			}

			if got.Funding != tt.wantData.Funding {
				t.Errorf("InstallResult.Funding = %v, want %v", got.Funding, tt.wantData.Funding)
			}

			if got.Vulnerabilities.Total != tt.wantData.Vulnerabilities.Total {
				t.Errorf("InstallResult.Vulnerabilities.Total = %v, want %v", got.Vulnerabilities.Total, tt.wantData.Vulnerabilities.Total)
			}
		})
	}
}

func TestInstallParser_WithVulnerabilities(t *testing.T) {
	input := `
added 150 packages, removed 2 packages, changed 5 packages, and audited 155 packages in 5s

12 packages are looking for funding
  run ` + "`npm fund`" + ` for details

6 vulnerabilities (1 low, 2 moderate, 2 high, 1 critical)

To address all issues, run:
  npm audit fix`

	parser := NewInstallParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*InstallResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *InstallResult", result.Data)
	}

	if got.Success {
		t.Error("InstallResult.Success = true, want false when vulnerabilities present")
	}

	if got.PackagesAdded != 150 {
		t.Errorf("InstallResult.PackagesAdded = %d, want 150", got.PackagesAdded)
	}

	if got.PackagesRemoved != 2 {
		t.Errorf("InstallResult.PackagesRemoved = %d, want 2", got.PackagesRemoved)
	}

	if got.PackagesChanged != 5 {
		t.Errorf("InstallResult.PackagesChanged = %d, want 5", got.PackagesChanged)
	}

	if got.PackagesAudited != 155 {
		t.Errorf("InstallResult.PackagesAudited = %d, want 155", got.PackagesAudited)
	}

	if got.Funding != 12 {
		t.Errorf("InstallResult.Funding = %d, want 12", got.Funding)
	}

	wantVuln := VulnerabilitySummary{
		Total:    6,
		Low:      1,
		Moderate: 2,
		High:     2,
		Critical: 1,
	}

	if got.Vulnerabilities != wantVuln {
		t.Errorf("InstallResult.Vulnerabilities = %+v, want %+v", got.Vulnerabilities, wantVuln)
	}
}

func TestInstallParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches npm install",
			cmd:         "npm",
			subcommands: []string{"install"},
			want:        true,
		},
		{
			name:        "matches npm i",
			cmd:         "npm",
			subcommands: []string{"i"},
			want:        true,
		},
		{
			name:        "matches npm install with package",
			cmd:         "npm",
			subcommands: []string{"install", "lodash"},
			want:        true,
		},
		{
			name:        "matches npm ci",
			cmd:         "npm",
			subcommands: []string{"ci"},
			want:        true,
		},
		{
			name:        "does not match npm audit",
			cmd:         "npm",
			subcommands: []string{"audit"},
			want:        false,
		},
		{
			name:        "does not match yarn",
			cmd:         "yarn",
			subcommands: []string{"install"},
			want:        false,
		},
		{
			name:        "does not match empty",
			cmd:         "",
			subcommands: []string{},
			want:        false,
		},
	}

	parser := NewInstallParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestInstallParser_Schema(t *testing.T) {
	parser := NewInstallParser()
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

	requiredProps := []string{"success", "packages_added", "vulnerabilities"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}
