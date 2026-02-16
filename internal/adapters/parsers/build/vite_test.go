package build

import (
	"strings"
	"testing"
)

func TestViteParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData ViteResult
	}{
		{
			name:  "empty output indicates successful build",
			input: "",
			wantData: ViteResult{
				Success:  true,
				Errors:   []ViteError{},
				Warnings: []ViteWarning{},
				Outputs:  []ViteOutput{},
				Duration: 0,
				Modules:  0,
			},
		},
		{
			name: "successful build with files and duration",
			input: `vite v6.3.2 building for production...
✓ 32 modules transformed.
dist/index.html                   0.46 kB │ gzip: 0.30 kB
dist/assets/react-CHdo91hT.svg    4.13 kB │ gzip: 2.05 kB
dist/assets/index-D8b4DHJx.css    1.39 kB │ gzip: 0.71 kB
dist/assets/index-9_sxcfan.js   188.05 kB │ gzip: 59.21 kB
✓ built in 1.90s`,
			wantData: ViteResult{
				Success:  true,
				Errors:   []ViteError{},
				Warnings: []ViteWarning{},
				Outputs: []ViteOutput{
					{Path: "dist/index.html", Size: 471, GzipSize: 307},
					{Path: "dist/assets/react-CHdo91hT.svg", Size: 4229, GzipSize: 2099},
					{Path: "dist/assets/index-D8b4DHJx.css", Size: 1423, GzipSize: 727},
					{Path: "dist/assets/index-9_sxcfan.js", Size: 192563, GzipSize: 60631},
				},
				Duration: 1900,
				Modules:  32,
			},
		},
		{
			name: "build with duration in milliseconds",
			input: `vite v4.2.1 building for production...
✓ 10 modules transformed.
dist/index.html   0.45 kB │ gzip: 0.38 kB
✓ built in 500ms`,
			wantData: ViteResult{
				Success:  true,
				Errors:   []ViteError{},
				Warnings: []ViteWarning{},
				Outputs: []ViteOutput{
					{Path: "dist/index.html", Size: 460, GzipSize: 389},
				},
				Duration: 500,
				Modules:  10,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewViteParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*ViteResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *ViteResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("ViteResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if len(got.Errors) != len(tt.wantData.Errors) {
				t.Errorf("ViteResult.Errors length = %d, want %d", len(got.Errors), len(tt.wantData.Errors))
			}

			if len(got.Warnings) != len(tt.wantData.Warnings) {
				t.Errorf("ViteResult.Warnings length = %d, want %d", len(got.Warnings), len(tt.wantData.Warnings))
			}

			if got.Duration != tt.wantData.Duration {
				t.Errorf("ViteResult.Duration = %v, want %v", got.Duration, tt.wantData.Duration)
			}

			if got.Modules != tt.wantData.Modules {
				t.Errorf("ViteResult.Modules = %v, want %v", got.Modules, tt.wantData.Modules)
			}

			if len(got.Outputs) != len(tt.wantData.Outputs) {
				t.Fatalf("ViteResult.Outputs length = %d, want %d", len(got.Outputs), len(tt.wantData.Outputs))
			}

			for i, wantOutput := range tt.wantData.Outputs {
				if got.Outputs[i].Path != wantOutput.Path {
					t.Errorf("ViteResult.Outputs[%d].Path = %v, want %v", i, got.Outputs[i].Path, wantOutput.Path)
				}
				if got.Outputs[i].Size != wantOutput.Size {
					t.Errorf("ViteResult.Outputs[%d].Size = %v, want %v", i, got.Outputs[i].Size, wantOutput.Size)
				}
				if got.Outputs[i].GzipSize != wantOutput.GzipSize {
					t.Errorf("ViteResult.Outputs[%d].GzipSize = %v, want %v", i, got.Outputs[i].GzipSize, wantOutput.GzipSize)
				}
			}
		})
	}
}

func TestViteParser_BuildError(t *testing.T) {
	input := `vite v6.3.2 building for production...
x Build failed in 57ms
error during build:
[vite:worker-import-meta-url] Invalid value "iife" for option "output.format" - UMD and IIFE output formats are not supported for code-splitting builds.`

	parser := NewViteParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*ViteResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ViteResult", result.Data)
	}

	if got.Success {
		t.Error("ViteResult.Success = true, want false when errors present")
	}

	if len(got.Errors) != 1 {
		t.Fatalf("ViteResult.Errors length = %d, want 1", len(got.Errors))
	}

	wantErr := ViteError{
		Plugin:  "vite:worker-import-meta-url",
		Message: `Invalid value "iife" for option "output.format" - UMD and IIFE output formats are not supported for code-splitting builds.`,
	}

	if got.Errors[0].Plugin != wantErr.Plugin {
		t.Errorf("ViteResult.Errors[0].Plugin = %q, want %q", got.Errors[0].Plugin, wantErr.Plugin)
	}

	if got.Errors[0].Message != wantErr.Message {
		t.Errorf("ViteResult.Errors[0].Message = %q, want %q", got.Errors[0].Message, wantErr.Message)
	}
}

func TestViteParser_GenericError(t *testing.T) {
	input := `vite v6.3.2 building for production...
x Build failed in 100ms
error during build:
Could not resolve entry module "src/main.ts".`

	parser := NewViteParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*ViteResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ViteResult", result.Data)
	}

	if got.Success {
		t.Error("ViteResult.Success = true, want false when errors present")
	}

	if len(got.Errors) != 1 {
		t.Fatalf("ViteResult.Errors length = %d, want 1", len(got.Errors))
	}

	if got.Errors[0].Message != `Could not resolve entry module "src/main.ts".` {
		t.Errorf("ViteResult.Errors[0].Message = %q, want %q", got.Errors[0].Message, `Could not resolve entry module "src/main.ts".`)
	}
}

func TestViteParser_RollupError(t *testing.T) {
	input := `vite v6.3.2 building for production...
[ERROR] With statements cannot be used with the "esm" output format due to strict mode`

	parser := NewViteParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*ViteResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ViteResult", result.Data)
	}

	if got.Success {
		t.Error("ViteResult.Success = true, want false when errors present")
	}

	if len(got.Errors) < 1 {
		t.Fatalf("ViteResult.Errors length = %d, want at least 1", len(got.Errors))
	}
}

func TestViteParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches vite build",
			cmd:         "vite",
			subcommands: []string{"build"},
			want:        true,
		},
		{
			name:        "matches vite build with flags",
			cmd:         "vite",
			subcommands: []string{"build", "--mode", "production"},
			want:        true,
		},
		{
			name:        "does not match vite without build",
			cmd:         "vite",
			subcommands: []string{},
			want:        false,
		},
		{
			name:        "does not match vite dev",
			cmd:         "vite",
			subcommands: []string{"dev"},
			want:        false,
		},
		{
			name:        "does not match vite preview",
			cmd:         "vite",
			subcommands: []string{"preview"},
			want:        false,
		},
		{
			name:        "does not match npx vite build",
			cmd:         "npx",
			subcommands: []string{"vite", "build"},
			want:        false,
		},
		{
			name:        "does not match other commands",
			cmd:         "npm",
			subcommands: []string{"run", "build"},
			want:        false,
		},
		{
			name:        "does not match empty command",
			cmd:         "",
			subcommands: []string{},
			want:        false,
		},
	}

	parser := NewViteParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestViteParser_Schema(t *testing.T) {
	parser := NewViteParser()
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

	// Verify required properties exist
	requiredProps := []string{"success", "errors", "warnings", "outputs"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}

func TestViteParser_MixedContent(t *testing.T) {
	// Test parsing output with some informational lines
	input := `
> react-ts-vite-app@0.0.0 build
> tsc -b && vite build

vite v6.3.2 building for production...
✓ 32 modules transformed.
dist/index.html                0.46 kB │ gzip: 0.30 kB
dist/assets/index-D8b4DHJx.js 188.05 kB │ gzip: 59.21 kB
✓ built in 1.90s`

	parser := NewViteParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*ViteResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ViteResult", result.Data)
	}

	if !got.Success {
		t.Error("ViteResult.Success = false, want true for successful build")
	}

	if len(got.Outputs) != 2 {
		t.Fatalf("ViteResult.Outputs length = %d, want 2", len(got.Outputs))
	}

	if got.Duration != 1900 {
		t.Errorf("ViteResult.Duration = %v, want 1900", got.Duration)
	}

	if got.Modules != 32 {
		t.Errorf("ViteResult.Modules = %v, want 32", got.Modules)
	}
}

func TestViteParser_LibraryBuild(t *testing.T) {
	// Test library mode output format (simpler, may not have gzip info in some cases)
	input := `vite v4.2.1 building for production...
✓ 5 modules transformed.
dist/my-lib.js      0.08 kB / gzip: 0.07 kB
dist/my-lib.umd.cjs 0.30 kB / gzip: 0.16 kB
✓ built in 0.5s`

	parser := NewViteParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*ViteResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ViteResult", result.Data)
	}

	if !got.Success {
		t.Error("ViteResult.Success = false, want true")
	}

	if len(got.Outputs) != 2 {
		t.Fatalf("ViteResult.Outputs length = %d, want 2", len(got.Outputs))
	}

	// Duration of 0.5s = 500ms
	if got.Duration != 500 {
		t.Errorf("ViteResult.Duration = %v, want 500", got.Duration)
	}

	if got.Modules != 5 {
		t.Errorf("ViteResult.Modules = %v, want 5", got.Modules)
	}
}
