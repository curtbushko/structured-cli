package build

import (
	"strings"
	"testing"
)

const schemaTypeObject = "object"

func TestWebpackParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData WebpackResult
	}{
		{
			name:  "empty output indicates successful build",
			input: "",
			wantData: WebpackResult{
				Success:  true,
				Errors:   []WebpackError{},
				Warnings: []WebpackWarning{},
				Assets:   []WebpackAsset{},
				Chunks:   []WebpackChunk{},
				Modules:  0,
				Duration: 0,
			},
		},
		{
			name: "successful build with assets",
			input: `asset main.js 1.5 KiB [emitted] (name: main)
asset vendor.js 250 KiB [emitted] (name: vendor)
webpack 5.75.0 compiled successfully in 1234 ms`,
			wantData: WebpackResult{
				Success:  true,
				Errors:   []WebpackError{},
				Warnings: []WebpackWarning{},
				Assets: []WebpackAsset{
					{Name: "main.js", Size: 1536, Emitted: true, ChunkNames: []string{"main"}},
					{Name: "vendor.js", Size: 256000, Emitted: true, ChunkNames: []string{"vendor"}},
				},
				Chunks:   []WebpackChunk{},
				Modules:  0,
				Duration: 1234,
			},
		},
		{
			name: "build with modules count",
			input: `asset bundle.js 100 KiB [emitted] (name: main)
42 modules
webpack 5.75.0 compiled successfully in 500 ms`,
			wantData: WebpackResult{
				Success:  true,
				Errors:   []WebpackError{},
				Warnings: []WebpackWarning{},
				Assets: []WebpackAsset{
					{Name: "bundle.js", Size: 102400, Emitted: true, ChunkNames: []string{"main"}},
				},
				Chunks:   []WebpackChunk{},
				Modules:  42,
				Duration: 500,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewWebpackParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*WebpackResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *WebpackResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("WebpackResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if len(got.Errors) != len(tt.wantData.Errors) {
				t.Errorf("WebpackResult.Errors length = %d, want %d", len(got.Errors), len(tt.wantData.Errors))
			}

			if len(got.Warnings) != len(tt.wantData.Warnings) {
				t.Errorf("WebpackResult.Warnings length = %d, want %d", len(got.Warnings), len(tt.wantData.Warnings))
			}

			if len(got.Assets) != len(tt.wantData.Assets) {
				t.Errorf("WebpackResult.Assets length = %d, want %d", len(got.Assets), len(tt.wantData.Assets))
			}
			compareAssets(t, got.Assets, tt.wantData.Assets)

			if got.Modules != tt.wantData.Modules {
				t.Errorf("WebpackResult.Modules = %v, want %v", got.Modules, tt.wantData.Modules)
			}

			if got.Duration != tt.wantData.Duration {
				t.Errorf("WebpackResult.Duration = %v, want %v", got.Duration, tt.wantData.Duration)
			}
		})
	}
}

func TestWebpackParser_SingleError(t *testing.T) {
	input := `ERROR in ./src/index.js 10:5
Module build failed: SyntaxError: Unexpected token`

	parser := NewWebpackParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*WebpackResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *WebpackResult", result.Data)
	}

	if got.Success {
		t.Error("WebpackResult.Success = true, want false when errors present")
	}

	if len(got.Errors) != 1 {
		t.Fatalf("WebpackResult.Errors length = %d, want 1", len(got.Errors))
	}

	wantErr := WebpackError{
		File:    "./src/index.js",
		Line:    10,
		Column:  5,
		Message: "Module build failed: SyntaxError: Unexpected token",
	}

	if got.Errors[0].File != wantErr.File {
		t.Errorf("Error.File = %q, want %q", got.Errors[0].File, wantErr.File)
	}
	if got.Errors[0].Line != wantErr.Line {
		t.Errorf("Error.Line = %d, want %d", got.Errors[0].Line, wantErr.Line)
	}
	if got.Errors[0].Column != wantErr.Column {
		t.Errorf("Error.Column = %d, want %d", got.Errors[0].Column, wantErr.Column)
	}
	if got.Errors[0].Message != wantErr.Message {
		t.Errorf("Error.Message = %q, want %q", got.Errors[0].Message, wantErr.Message)
	}
}

func TestWebpackParser_Warning(t *testing.T) {
	input := `WARNING in ./src/utils.js 25:10
Module Warning: This import is unused
asset main.js 100 KiB [emitted] (name: main)
webpack 5.75.0 compiled with 1 warning in 500 ms`

	parser := NewWebpackParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*WebpackResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *WebpackResult", result.Data)
	}

	// Warnings don't affect success
	if !got.Success {
		t.Error("WebpackResult.Success = false, want true when only warnings present")
	}

	if len(got.Warnings) != 1 {
		t.Fatalf("WebpackResult.Warnings length = %d, want 1", len(got.Warnings))
	}

	if got.Warnings[0].File != "./src/utils.js" {
		t.Errorf("Warning.File = %q, want %q", got.Warnings[0].File, "./src/utils.js")
	}
	if got.Warnings[0].Line != 25 {
		t.Errorf("Warning.Line = %d, want %d", got.Warnings[0].Line, 25)
	}
	if got.Warnings[0].Column != 10 {
		t.Errorf("Warning.Column = %d, want %d", got.Warnings[0].Column, 10)
	}
	if got.Warnings[0].Message != "Module Warning: This import is unused" {
		t.Errorf("Warning.Message = %q, want %q", got.Warnings[0].Message, "Module Warning: This import is unused")
	}
}

func TestWebpackParser_MultipleErrorsAndWarnings(t *testing.T) {
	input := `ERROR in ./src/index.js 10:5
Module build failed: SyntaxError: Unexpected token
WARNING in ./src/utils.js 25:10
Module Warning: This import is unused
ERROR in ./src/app.js 42:3
Module not found: Error: Can't resolve 'missing-module'`

	parser := NewWebpackParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*WebpackResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *WebpackResult", result.Data)
	}

	if got.Success {
		t.Error("WebpackResult.Success = true, want false when errors present")
	}

	if len(got.Errors) != 2 {
		t.Fatalf("WebpackResult.Errors length = %d, want 2", len(got.Errors))
	}

	if len(got.Warnings) != 1 {
		t.Fatalf("WebpackResult.Warnings length = %d, want 1", len(got.Warnings))
	}
}

func TestWebpackParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches webpack with no subcommands",
			cmd:         "webpack",
			subcommands: []string{},
			want:        true,
		},
		{
			name:        "matches webpack with --mode production",
			cmd:         "webpack",
			subcommands: []string{"--mode", "production"},
			want:        true,
		},
		{
			name:        "matches webpack with --config flag",
			cmd:         "webpack",
			subcommands: []string{"--config", "webpack.config.js"},
			want:        true,
		},
		{
			name:        "does not match npx webpack",
			cmd:         "npx",
			subcommands: []string{"webpack"},
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

	parser := NewWebpackParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestWebpackParser_Schema(t *testing.T) {
	parser := NewWebpackParser()
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

	// Verify required properties exist
	requiredProps := []string{"success", "errors", "warnings", "assets"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}

func TestWebpackParser_Chunks(t *testing.T) {
	input := `chunk (runtime: main) main.js (main) 1.5 KiB [entry] [rendered]
chunk (runtime: main) vendor.js (vendor) 250 KiB [initial] [rendered]
asset main.js 1.5 KiB [emitted] (name: main)
webpack 5.75.0 compiled successfully in 500 ms`

	parser := NewWebpackParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*WebpackResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *WebpackResult", result.Data)
	}

	if len(got.Chunks) != 2 {
		t.Fatalf("WebpackResult.Chunks length = %d, want 2", len(got.Chunks))
	}

	// Verify first chunk
	if got.Chunks[0].Name != "main.js" {
		t.Errorf("Chunk[0].Name = %q, want %q", got.Chunks[0].Name, "main.js")
	}
	if !got.Chunks[0].Entry {
		t.Error("Chunk[0].Entry = false, want true")
	}
}

func TestWebpackParser_AssetSizeUnits(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantSize int64
	}{
		{
			name:     "KiB unit",
			input:    "asset bundle.js 100 KiB [emitted] (name: main)",
			wantSize: 102400, // 100 * 1024
		},
		{
			name:     "MiB unit",
			input:    "asset bundle.js 2.5 MiB [emitted] (name: main)",
			wantSize: 2621440, // 2.5 * 1024 * 1024
		},
		{
			name:     "bytes unit",
			input:    "asset bundle.js 512 bytes [emitted] (name: main)",
			wantSize: 512,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewWebpackParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			got, ok := result.Data.(*WebpackResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *WebpackResult", result.Data)
			}

			if len(got.Assets) != 1 {
				t.Fatalf("Assets length = %d, want 1", len(got.Assets))
			}

			if got.Assets[0].Size != tt.wantSize {
				t.Errorf("Asset.Size = %d, want %d", got.Assets[0].Size, tt.wantSize)
			}
		})
	}
}

func TestWebpackParser_FailedCompilation(t *testing.T) {
	input := `ERROR in ./src/index.js
Module build failed: SyntaxError: Unexpected token

webpack 5.75.0 compiled with 1 error in 234 ms`

	parser := NewWebpackParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*WebpackResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *WebpackResult", result.Data)
	}

	if got.Success {
		t.Error("WebpackResult.Success = true, want false when errors present")
	}

	if got.Duration != 234 {
		t.Errorf("WebpackResult.Duration = %v, want 234", got.Duration)
	}

	if len(got.Errors) < 1 {
		t.Error("WebpackResult.Errors should contain at least one error")
	}
}

func TestWebpackParser_MixedContent(t *testing.T) {
	input := `[webpack-cli] Compilation starting...
[webpack-cli] Compilation finished
asset main.js 1.5 KiB [emitted] (name: main)
./src/index.js 1.2 KiB [built]
webpack 5.75.0 compiled successfully in 100 ms`

	parser := NewWebpackParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*WebpackResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *WebpackResult", result.Data)
	}

	if !got.Success {
		t.Error("WebpackResult.Success = false, want true for successful build")
	}

	if len(got.Assets) != 1 {
		t.Errorf("WebpackResult.Assets length = %d, want 1", len(got.Assets))
	}

	if got.Duration != 100 {
		t.Errorf("WebpackResult.Duration = %v, want 100", got.Duration)
	}
}

// compareAssets is a test helper that compares two asset slices.
func compareAssets(t *testing.T, got, want []WebpackAsset) {
	t.Helper()
	for i, wantAsset := range want {
		if i >= len(got) {
			return
		}
		if got[i].Name != wantAsset.Name {
			t.Errorf("Assets[%d].Name = %q, want %q", i, got[i].Name, wantAsset.Name)
		}
		if got[i].Size != wantAsset.Size {
			t.Errorf("Assets[%d].Size = %d, want %d", i, got[i].Size, wantAsset.Size)
		}
		if got[i].Emitted != wantAsset.Emitted {
			t.Errorf("Assets[%d].Emitted = %v, want %v", i, got[i].Emitted, wantAsset.Emitted)
		}
	}
}
