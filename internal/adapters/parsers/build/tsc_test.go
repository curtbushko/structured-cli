package build

import (
	"strings"
	"testing"
)

func TestTSCParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData TSCResult
	}{
		{
			name:  "empty output indicates successful build",
			input: "",
			wantData: TSCResult{
				Success: true,
				Errors:  []TSCError{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewTSCParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*TSCResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *TSCResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("TSCResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if len(got.Errors) != len(tt.wantData.Errors) {
				t.Errorf("TSCResult.Errors length = %d, want %d", len(got.Errors), len(tt.wantData.Errors))
			}
		})
	}
}

func TestTSCParser_SingleError(t *testing.T) {
	input := "src/index.ts(10,5): error TS2322: Type 'string' is not assignable to type 'number'."

	parser := NewTSCParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*TSCResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *TSCResult", result.Data)
	}

	if got.Success {
		t.Error("TSCResult.Success = true, want false when errors present")
	}

	if len(got.Errors) != 1 {
		t.Fatalf("TSCResult.Errors length = %d, want 1", len(got.Errors))
	}

	wantErr := TSCError{
		File:    "src/index.ts",
		Line:    10,
		Column:  5,
		Code:    "TS2322",
		Message: "Type 'string' is not assignable to type 'number'.",
	}

	if got.Errors[0] != wantErr {
		t.Errorf("TSCResult.Errors[0] = %+v, want %+v", got.Errors[0], wantErr)
	}
}

func TestTSCParser_MultipleErrors(t *testing.T) {
	input := `src/index.ts(10,5): error TS2322: Type 'string' is not assignable to type 'number'.
src/utils.ts(25,10): error TS2304: Cannot find name 'foo'.
src/app.ts(42,3): error TS7006: Parameter 'x' implicitly has an 'any' type.`

	parser := NewTSCParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*TSCResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *TSCResult", result.Data)
	}

	if got.Success {
		t.Error("TSCResult.Success = true, want false when errors present")
	}

	if len(got.Errors) != 3 {
		t.Fatalf("TSCResult.Errors length = %d, want 3", len(got.Errors))
	}

	wantErrors := []TSCError{
		{File: "src/index.ts", Line: 10, Column: 5, Code: "TS2322", Message: "Type 'string' is not assignable to type 'number'."},
		{File: "src/utils.ts", Line: 25, Column: 10, Code: "TS2304", Message: "Cannot find name 'foo'."},
		{File: "src/app.ts", Line: 42, Column: 3, Code: "TS7006", Message: "Parameter 'x' implicitly has an 'any' type."},
	}

	for i, wantErr := range wantErrors {
		if got.Errors[i] != wantErr {
			t.Errorf("TSCResult.Errors[%d] = %+v, want %+v", i, got.Errors[i], wantErr)
		}
	}
}

func TestTSCParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches tsc with no subcommands",
			cmd:         "tsc",
			subcommands: []string{},
			want:        true,
		},
		{
			name:        "matches tsc with --noEmit flag",
			cmd:         "tsc",
			subcommands: []string{"--noEmit"},
			want:        true,
		},
		{
			name:        "matches tsc with --build flag",
			cmd:         "tsc",
			subcommands: []string{"--build"},
			want:        true,
		},
		{
			name:        "does not match npx tsc",
			cmd:         "npx",
			subcommands: []string{"tsc"},
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

	parser := NewTSCParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestTSCParser_Schema(t *testing.T) {
	parser := NewTSCParser()
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
	requiredProps := []string{"success", "errors"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}

func TestTSCParser_WindowsPath(t *testing.T) {
	// Test parsing Windows-style paths
	input := `C:\project\src\index.ts(10,5): error TS2322: Type 'string' is not assignable to type 'number'.`

	parser := NewTSCParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*TSCResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *TSCResult", result.Data)
	}

	if got.Success {
		t.Error("TSCResult.Success = true, want false when errors present")
	}

	if len(got.Errors) != 1 {
		t.Fatalf("TSCResult.Errors length = %d, want 1", len(got.Errors))
	}

	wantErr := TSCError{
		File:    `C:\project\src\index.ts`,
		Line:    10,
		Column:  5,
		Code:    "TS2322",
		Message: "Type 'string' is not assignable to type 'number'.",
	}

	if got.Errors[0] != wantErr {
		t.Errorf("TSCResult.Errors[0] = %+v, want %+v", got.Errors[0], wantErr)
	}
}

func TestTSCParser_MixedContent(t *testing.T) {
	// Test parsing output with some non-error lines
	input := `Some informational message
src/index.ts(10,5): error TS2322: Type 'string' is not assignable to type 'number'.
Another info line`

	parser := NewTSCParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*TSCResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *TSCResult", result.Data)
	}

	if got.Success {
		t.Error("TSCResult.Success = true, want false when errors present")
	}

	if len(got.Errors) != 1 {
		t.Fatalf("TSCResult.Errors length = %d, want 1", len(got.Errors))
	}
}
