package python

import (
	"strings"
	"testing"
)

const schemaTypeObject = "object"

func TestUVRunParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData UVRunResult
	}{
		{
			name:  "empty output indicates success",
			input: "",
			wantData: UVRunResult{
				Success: true,
			},
		},
		{
			name:  "simple script output",
			input: `Hello, World!`,
			wantData: UVRunResult{
				Success: true,
				Output:  "Hello, World!",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewUVRunParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*UVRunResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *UVRunResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("UVRunResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if tt.wantData.Output != "" && got.Output != tt.wantData.Output {
				t.Errorf("UVRunResult.Output = %q, want %q", got.Output, tt.wantData.Output)
			}
		})
	}
}

func TestUVRunParser_WithPackageInstall(t *testing.T) {
	input := `Resolved 2 packages in 100ms
Prepared 2 packages in 50ms
Installed 2 packages in 10ms
 + requests==2.31.0
 + urllib3==2.0.4
Running script...
Done!`

	parser := NewUVRunParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*UVRunResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *UVRunResult", result.Data)
	}

	if !got.Success {
		t.Error("UVRunResult.Success = false, want true")
	}

	if len(got.InstalledPackages) != 2 {
		t.Errorf("UVRunResult.InstalledPackages length = %d, want 2", len(got.InstalledPackages))
	}

	wantPackages := []InstalledPackage{
		{Name: "requests", Version: "2.31.0"},
		{Name: "urllib3", Version: "2.0.4"},
	}

	for _, want := range wantPackages {
		found := false
		for _, pkg := range got.InstalledPackages {
			if pkg.Name == want.Name && pkg.Version == want.Version {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Package %s-%s not found in installed packages", want.Name, want.Version)
		}
	}
}

func TestUVRunParser_ScriptWithError(t *testing.T) {
	input := `Traceback (most recent call last):
  File "script.py", line 5, in <module>
    raise ValueError("Something went wrong")
ValueError: Something went wrong`

	parser := NewUVRunParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*UVRunResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *UVRunResult", result.Data)
	}

	// Parser cannot determine success from output alone - it captures output
	// The exit code would determine actual success
	if !strings.Contains(got.Output, "Traceback") {
		t.Error("UVRunResult.Output should contain traceback")
	}
}

func TestUVRunParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches uv run",
			cmd:         "uv",
			subcommands: []string{"run"},
			want:        true,
		},
		{
			name:        "matches uv run with script",
			cmd:         "uv",
			subcommands: []string{"run", "script.py"},
			want:        true,
		},
		{
			name:        "matches uv run with flags",
			cmd:         "uv",
			subcommands: []string{"run", "--with", "requests", "script.py"},
			want:        true,
		},
		{
			name:        "does not match uv pip",
			cmd:         "uv",
			subcommands: []string{"pip", "install"},
			want:        false,
		},
		{
			name:        "does not match python",
			cmd:         "python",
			subcommands: []string{"script.py"},
			want:        false,
		},
		{
			name:        "does not match empty",
			cmd:         "",
			subcommands: []string{},
			want:        false,
		},
	}

	parser := NewUVRunParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestUVRunParser_Schema(t *testing.T) {
	parser := NewUVRunParser()
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

	requiredProps := []string{"success", "output"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}
