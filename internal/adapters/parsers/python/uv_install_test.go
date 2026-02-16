package python

import (
	"strings"
	"testing"
)

func TestUVInstallParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData UVInstallResult
	}{
		{
			name:  "empty output indicates success",
			input: "",
			wantData: UVInstallResult{
				Success:             true,
				PackagesInstalled:   []InstalledPackage{},
				PackagesUninstalled: []string{},
				Warnings:            []string{},
			},
		},
		{
			name:  "already satisfied",
			input: `Audited 1 package in 5ms`,
			wantData: UVInstallResult{
				Success:             true,
				PackagesInstalled:   []InstalledPackage{},
				PackagesUninstalled: []string{},
				Warnings:            []string{},
				AlreadySatisfied:    true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewUVInstallParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*UVInstallResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *UVInstallResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("UVInstallResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if len(got.PackagesInstalled) != len(tt.wantData.PackagesInstalled) {
				t.Errorf("UVInstallResult.PackagesInstalled length = %d, want %d", len(got.PackagesInstalled), len(tt.wantData.PackagesInstalled))
			}
		})
	}
}

func TestUVInstallParser_SimpleInstall(t *testing.T) {
	input := `Resolved 3 packages in 145ms
Prepared 3 packages in 50ms
Installed 3 packages in 12ms
 + requests==2.31.0
 + urllib3==2.0.4
 + certifi==2023.7.22`

	parser := NewUVInstallParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*UVInstallResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *UVInstallResult", result.Data)
	}

	if !got.Success {
		t.Error("UVInstallResult.Success = false, want true")
	}

	if len(got.PackagesInstalled) != 3 {
		t.Fatalf("UVInstallResult.PackagesInstalled length = %d, want 3", len(got.PackagesInstalled))
	}

	wantPackages := []InstalledPackage{
		{Name: "requests", Version: "2.31.0"},
		{Name: "urllib3", Version: "2.0.4"},
		{Name: "certifi", Version: "2023.7.22"},
	}

	for _, want := range wantPackages {
		found := false
		for _, got := range got.PackagesInstalled {
			if got.Name == want.Name && got.Version == want.Version {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Package %s-%s not found in installed packages", want.Name, want.Version)
		}
	}
}

func TestUVInstallParser_WithUninstall(t *testing.T) {
	input := `Resolved 2 packages in 80ms
Uninstalled 1 package in 5ms
Installed 2 packages in 10ms
 - old-package==1.0.0
 + new-package==2.0.0
 + another==1.5.0`

	parser := NewUVInstallParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*UVInstallResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *UVInstallResult", result.Data)
	}

	if !got.Success {
		t.Error("UVInstallResult.Success = false, want true")
	}

	if len(got.PackagesInstalled) != 2 {
		t.Errorf("UVInstallResult.PackagesInstalled length = %d, want 2", len(got.PackagesInstalled))
	}

	if len(got.PackagesUninstalled) != 1 {
		t.Errorf("UVInstallResult.PackagesUninstalled length = %d, want 1", len(got.PackagesUninstalled))
	}

	if got.PackagesUninstalled[0] != "old-package" {
		t.Errorf("UVInstallResult.PackagesUninstalled[0] = %q, want %q", got.PackagesUninstalled[0], "old-package")
	}
}

func TestUVInstallParser_CachedPackages(t *testing.T) {
	input := `Resolved 2 packages in 50ms
Installed 2 packages in 5ms
 + requests==2.31.0 (cached)
 + urllib3==2.0.4`

	parser := NewUVInstallParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*UVInstallResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *UVInstallResult", result.Data)
	}

	if got.Cached != 1 {
		t.Errorf("UVInstallResult.Cached = %d, want 1", got.Cached)
	}
}

func TestUVInstallParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches uv pip install",
			cmd:         "uv",
			subcommands: []string{"pip", "install"},
			want:        true,
		},
		{
			name:        "matches uv pip install with package",
			cmd:         "uv",
			subcommands: []string{"pip", "install", "requests"},
			want:        true,
		},
		{
			name:        "does not match uv run",
			cmd:         "uv",
			subcommands: []string{"run"},
			want:        false,
		},
		{
			name:        "does not match pip",
			cmd:         "pip",
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

	parser := NewUVInstallParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestUVInstallParser_Schema(t *testing.T) {
	parser := NewUVInstallParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema.ID should not be empty")
	}

	if schema.Title == "" {
		t.Error("Schema.Title should not be empty")
	}

	if schema.Type != "object" {
		t.Errorf("Schema.Type = %q, want %q", schema.Type, "object")
	}

	requiredProps := []string{"success", "packages_installed"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}
