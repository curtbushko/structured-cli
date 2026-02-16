package build

import (
	"strings"
	"testing"
)

func TestESBuildParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData ESBuildResult
	}{
		{
			name:  "empty output indicates successful build",
			input: "",
			wantData: ESBuildResult{
				Success:  true,
				Errors:   []ESBuildError{},
				Warnings: []ESBuildWarning{},
				Outputs:  []ESBuildOutput{},
				Duration: 0,
			},
		},
		{
			name:  "build with duration",
			input: "Done in 42ms",
			wantData: ESBuildResult{
				Success:  true,
				Errors:   []ESBuildError{},
				Warnings: []ESBuildWarning{},
				Outputs:  []ESBuildOutput{},
				Duration: 42,
			},
		},
		{
			name:  "build with duration in seconds",
			input: "Done in 1.5s",
			wantData: ESBuildResult{
				Success:  true,
				Errors:   []ESBuildError{},
				Warnings: []ESBuildWarning{},
				Outputs:  []ESBuildOutput{},
				Duration: 1500,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewESBuildParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*ESBuildResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *ESBuildResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("ESBuildResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if len(got.Errors) != len(tt.wantData.Errors) {
				t.Errorf("ESBuildResult.Errors length = %d, want %d", len(got.Errors), len(tt.wantData.Errors))
			}

			if len(got.Warnings) != len(tt.wantData.Warnings) {
				t.Errorf("ESBuildResult.Warnings length = %d, want %d", len(got.Warnings), len(tt.wantData.Warnings))
			}

			if got.Duration != tt.wantData.Duration {
				t.Errorf("ESBuildResult.Duration = %v, want %v", got.Duration, tt.wantData.Duration)
			}
		})
	}
}

func TestESBuildParser_SingleError(t *testing.T) {
	input := `> src/index.ts:10:5: error: Could not resolve "missing-module"`

	parser := NewESBuildParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*ESBuildResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ESBuildResult", result.Data)
	}

	if got.Success {
		t.Error("ESBuildResult.Success = true, want false when errors present")
	}

	if len(got.Errors) != 1 {
		t.Fatalf("ESBuildResult.Errors length = %d, want 1", len(got.Errors))
	}

	wantErr := ESBuildError{
		File:    "src/index.ts",
		Line:    10,
		Column:  5,
		Message: `Could not resolve "missing-module"`,
	}

	if got.Errors[0] != wantErr {
		t.Errorf("ESBuildResult.Errors[0] = %+v, want %+v", got.Errors[0], wantErr)
	}
}

func TestESBuildParser_Warning(t *testing.T) {
	input := `> src/utils.ts:25:10: warning: This import is unused`

	parser := NewESBuildParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*ESBuildResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ESBuildResult", result.Data)
	}

	// Warnings don't affect success
	if !got.Success {
		t.Error("ESBuildResult.Success = false, want true when only warnings present")
	}

	if len(got.Warnings) != 1 {
		t.Fatalf("ESBuildResult.Warnings length = %d, want 1", len(got.Warnings))
	}

	wantWarn := ESBuildWarning{
		File:    "src/utils.ts",
		Line:    25,
		Column:  10,
		Message: "This import is unused",
	}

	if got.Warnings[0] != wantWarn {
		t.Errorf("ESBuildResult.Warnings[0] = %+v, want %+v", got.Warnings[0], wantWarn)
	}
}

func TestESBuildParser_MultipleErrorsAndWarnings(t *testing.T) {
	input := `> src/index.ts:10:5: error: Could not resolve "missing-module"
> src/utils.ts:25:10: warning: This import is unused
> src/app.ts:42:3: error: Unexpected token`

	parser := NewESBuildParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*ESBuildResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ESBuildResult", result.Data)
	}

	if got.Success {
		t.Error("ESBuildResult.Success = true, want false when errors present")
	}

	if len(got.Errors) != 2 {
		t.Fatalf("ESBuildResult.Errors length = %d, want 2", len(got.Errors))
	}

	if len(got.Warnings) != 1 {
		t.Fatalf("ESBuildResult.Warnings length = %d, want 1", len(got.Warnings))
	}

	wantErrors := []ESBuildError{
		{File: "src/index.ts", Line: 10, Column: 5, Message: `Could not resolve "missing-module"`},
		{File: "src/app.ts", Line: 42, Column: 3, Message: "Unexpected token"},
	}

	for i, wantErr := range wantErrors {
		if got.Errors[i] != wantErr {
			t.Errorf("ESBuildResult.Errors[%d] = %+v, want %+v", i, got.Errors[i], wantErr)
		}
	}
}

func TestESBuildParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches esbuild with no subcommands",
			cmd:         "esbuild",
			subcommands: []string{},
			want:        true,
		},
		{
			name:        "matches esbuild with src/index.ts --bundle",
			cmd:         "esbuild",
			subcommands: []string{"src/index.ts", "--bundle"},
			want:        true,
		},
		{
			name:        "matches esbuild with --minify flag",
			cmd:         "esbuild",
			subcommands: []string{"--minify"},
			want:        true,
		},
		{
			name:        "does not match npx esbuild",
			cmd:         "npx",
			subcommands: []string{"esbuild"},
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

	parser := NewESBuildParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestESBuildParser_Schema(t *testing.T) {
	parser := NewESBuildParser()
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

func TestESBuildParser_MixedContent(t *testing.T) {
	// Test parsing output with some non-error lines
	input := `Some informational message
> src/index.ts:10:5: error: Could not resolve "missing-module"
Another info line
Done in 100ms`

	parser := NewESBuildParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*ESBuildResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ESBuildResult", result.Data)
	}

	if got.Success {
		t.Error("ESBuildResult.Success = true, want false when errors present")
	}

	if len(got.Errors) != 1 {
		t.Fatalf("ESBuildResult.Errors length = %d, want 1", len(got.Errors))
	}

	if got.Duration != 100 {
		t.Errorf("ESBuildResult.Duration = %v, want 100", got.Duration)
	}
}

func TestESBuildParser_WindowsPath(t *testing.T) {
	// Test parsing Windows-style paths
	input := `> C:\project\src\index.ts:10:5: error: Could not resolve "missing-module"`

	parser := NewESBuildParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*ESBuildResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ESBuildResult", result.Data)
	}

	if got.Success {
		t.Error("ESBuildResult.Success = true, want false when errors present")
	}

	if len(got.Errors) != 1 {
		t.Fatalf("ESBuildResult.Errors length = %d, want 1", len(got.Errors))
	}

	wantErr := ESBuildError{
		File:    `C:\project\src\index.ts`,
		Line:    10,
		Column:  5,
		Message: `Could not resolve "missing-module"`,
	}

	if got.Errors[0] != wantErr {
		t.Errorf("ESBuildResult.Errors[0] = %+v, want %+v", got.Errors[0], wantErr)
	}
}
