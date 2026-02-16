package python

import (
	"strings"
	"testing"
)

func TestPipInstallParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData PipInstallResult
	}{
		{
			name:  "empty output indicates success",
			input: "",
			wantData: PipInstallResult{
				Success:           true,
				PackagesInstalled: []InstalledPackage{},
				Warnings:          []string{},
				AlreadySatisfied:  []string{},
			},
		},
		{
			name: "simple install success",
			input: `Collecting requests
  Downloading requests-2.31.0-py3-none-any.whl (62 kB)
Installing collected packages: requests
Successfully installed requests-2.31.0`,
			wantData: PipInstallResult{
				Success: true,
				PackagesInstalled: []InstalledPackage{
					{Name: "requests", Version: "2.31.0"},
				},
				Warnings:         []string{},
				AlreadySatisfied: []string{},
			},
		},
		{
			name:  "requirement already satisfied",
			input: `Requirement already satisfied: requests in /usr/lib/python3/dist-packages (2.31.0)`,
			wantData: PipInstallResult{
				Success:           true,
				PackagesInstalled: []InstalledPackage{},
				Warnings:          []string{},
				AlreadySatisfied:  []string{"requests"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewPipInstallParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*PipInstallResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *PipInstallResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("PipInstallResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if len(got.PackagesInstalled) != len(tt.wantData.PackagesInstalled) {
				t.Errorf("PipInstallResult.PackagesInstalled length = %d, want %d", len(got.PackagesInstalled), len(tt.wantData.PackagesInstalled))
			}

			for i, wantPkg := range tt.wantData.PackagesInstalled {
				if i < len(got.PackagesInstalled) && got.PackagesInstalled[i] != wantPkg {
					t.Errorf("PipInstallResult.PackagesInstalled[%d] = %+v, want %+v", i, got.PackagesInstalled[i], wantPkg)
				}
			}

			if len(got.AlreadySatisfied) != len(tt.wantData.AlreadySatisfied) {
				t.Errorf("PipInstallResult.AlreadySatisfied length = %d, want %d", len(got.AlreadySatisfied), len(tt.wantData.AlreadySatisfied))
			}
		})
	}
}

func TestPipInstallParser_MultiplePackages(t *testing.T) {
	input := `Collecting flask
  Downloading flask-2.3.3-py3-none-any.whl (96 kB)
Collecting werkzeug>=2.3.7
  Downloading werkzeug-2.3.7-py3-none-any.whl (242 kB)
Collecting jinja2>=3.1.2
  Downloading Jinja2-3.1.2-py3-none-any.whl (133 kB)
Installing collected packages: werkzeug, jinja2, flask
Successfully installed flask-2.3.3 jinja2-3.1.2 werkzeug-2.3.7`

	parser := NewPipInstallParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*PipInstallResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *PipInstallResult", result.Data)
	}

	if !got.Success {
		t.Error("PipInstallResult.Success = false, want true")
	}

	if len(got.PackagesInstalled) != 3 {
		t.Fatalf("PipInstallResult.PackagesInstalled length = %d, want 3", len(got.PackagesInstalled))
	}

	// Packages are listed in the order they appear in "Successfully installed"
	wantPackages := []InstalledPackage{
		{Name: "flask", Version: "2.3.3"},
		{Name: "jinja2", Version: "3.1.2"},
		{Name: "werkzeug", Version: "2.3.7"},
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

func TestPipInstallParser_WithWarnings(t *testing.T) {
	input := `WARNING: pip is configured with locations that require TLS/SSL, however the ssl module in Python is not available.
Collecting requests
  Downloading requests-2.31.0-py3-none-any.whl (62 kB)
WARNING: Retrying (Retry(total=4, connect=None, read=None, redirect=None, status=None)) after connection broken
Installing collected packages: requests
Successfully installed requests-2.31.0`

	parser := NewPipInstallParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*PipInstallResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *PipInstallResult", result.Data)
	}

	if !got.Success {
		t.Error("PipInstallResult.Success = false, want true")
	}

	if len(got.Warnings) != 2 {
		t.Errorf("PipInstallResult.Warnings length = %d, want 2", len(got.Warnings))
	}
}

func TestPipInstallParser_FromRequirements(t *testing.T) {
	input := `Collecting requests==2.31.0 (from -r requirements.txt (line 1))
  Downloading requests-2.31.0-py3-none-any.whl (62 kB)
Installing collected packages: requests
Successfully installed requests-2.31.0`

	parser := NewPipInstallParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*PipInstallResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *PipInstallResult", result.Data)
	}

	if got.RequirementsFile != "requirements.txt" {
		t.Errorf("PipInstallResult.RequirementsFile = %q, want %q", got.RequirementsFile, "requirements.txt")
	}
}

func TestPipInstallParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches pip install",
			cmd:         "pip",
			subcommands: []string{"install"},
			want:        true,
		},
		{
			name:        "matches pip install with package",
			cmd:         "pip",
			subcommands: []string{"install", "requests"},
			want:        true,
		},
		{
			name:        "matches pip3 install",
			cmd:         "pip3",
			subcommands: []string{"install"},
			want:        true,
		},
		{
			name:        "does not match pip uninstall",
			cmd:         "pip",
			subcommands: []string{"uninstall"},
			want:        false,
		},
		{
			name:        "does not match npm",
			cmd:         "npm",
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

	parser := NewPipInstallParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestPipInstallParser_Schema(t *testing.T) {
	parser := NewPipInstallParser()
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
