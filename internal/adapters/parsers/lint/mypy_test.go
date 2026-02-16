package lint

import (
	"strings"
	"testing"
)

func TestMypyParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData MypyResult
	}{
		{
			name:  "empty output indicates clean type check",
			input: "",
			wantData: MypyResult{
				Success: true,
				Errors:  []MypyError{},
				Summary: "",
			},
		},
		{
			name:  "success message with no errors",
			input: "Success: no issues found in 10 source files",
			wantData: MypyResult{
				Success: true,
				Errors:  []MypyError{},
				Summary: "Success: no issues found in 10 source files",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewMypyParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*MypyResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *MypyResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("MypyResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if len(got.Errors) != len(tt.wantData.Errors) {
				t.Errorf("MypyResult.Errors length = %d, want %d", len(got.Errors), len(tt.wantData.Errors))
			}

			if got.Summary != tt.wantData.Summary {
				t.Errorf("MypyResult.Summary = %q, want %q", got.Summary, tt.wantData.Summary)
			}
		})
	}
}

func TestMypyParser_SingleError(t *testing.T) {
	// Mypy outputs errors in format: file:line: severity: message [error-code]
	input := `main.py:10: error: Argument 1 to "foo" has incompatible type "str"; expected "int"  [arg-type]`

	parser := NewMypyParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*MypyResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *MypyResult", result.Data)
	}

	if got.Success {
		t.Error("MypyResult.Success = true, want false when errors present")
	}

	if len(got.Errors) != 1 {
		t.Fatalf("MypyResult.Errors length = %d, want 1", len(got.Errors))
	}

	wantError := MypyError{
		File:     "main.py",
		Line:     10,
		Severity: "error",
		Message:  `Argument 1 to "foo" has incompatible type "str"; expected "int"`,
		Code:     "arg-type",
	}

	if got.Errors[0] != wantError {
		t.Errorf("MypyResult.Errors[0] = %+v, want %+v", got.Errors[0], wantError)
	}
}

func TestMypyParser_MultipleErrors(t *testing.T) {
	input := `main.py:10: error: Argument 1 to "foo" has incompatible type "str"; expected "int"  [arg-type]
utils.py:25: error: Function is missing a return type annotation  [no-untyped-def]
app.py:42: warning: "Optional[str]" is deprecated; use "str | None" instead  [deprecated]
Found 3 errors in 3 files (checked 10 source files)`

	parser := NewMypyParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*MypyResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *MypyResult", result.Data)
	}

	if got.Success {
		t.Error("MypyResult.Success = true, want false when errors present")
	}

	if len(got.Errors) != 3 {
		t.Fatalf("MypyResult.Errors length = %d, want 3", len(got.Errors))
	}

	wantErrors := []MypyError{
		{File: "main.py", Line: 10, Severity: "error", Message: `Argument 1 to "foo" has incompatible type "str"; expected "int"`, Code: "arg-type"},
		{File: "utils.py", Line: 25, Severity: "error", Message: "Function is missing a return type annotation", Code: "no-untyped-def"},
		{File: "app.py", Line: 42, Severity: "warning", Message: `"Optional[str]" is deprecated; use "str | None" instead`, Code: "deprecated"},
	}

	for i, wantError := range wantErrors {
		if got.Errors[i] != wantError {
			t.Errorf("MypyResult.Errors[%d] = %+v, want %+v", i, got.Errors[i], wantError)
		}
	}

	if got.Summary != "Found 3 errors in 3 files (checked 10 source files)" {
		t.Errorf("MypyResult.Summary = %q, want %q", got.Summary, "Found 3 errors in 3 files (checked 10 source files)")
	}
}

func TestMypyParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches mypy with no subcommands",
			cmd:         "mypy",
			subcommands: []string{},
			want:        true,
		},
		{
			name:        "matches mypy with path",
			cmd:         "mypy",
			subcommands: []string{"src/"},
			want:        true,
		},
		{
			name:        "matches mypy with flags",
			cmd:         "mypy",
			subcommands: []string{"--strict", "src/"},
			want:        true,
		},
		{
			name:        "does not match python mypy",
			cmd:         "python",
			subcommands: []string{"-m", "mypy"},
			want:        false,
		},
		{
			name:        "does not match empty command",
			cmd:         "",
			subcommands: []string{},
			want:        false,
		},
	}

	parser := NewMypyParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestMypyParser_Schema(t *testing.T) {
	parser := NewMypyParser()
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
	requiredProps := []string{"success", "errors", "summary"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}

func TestMypyParser_ErrorWithoutCode(t *testing.T) {
	// Some mypy errors don't have error codes
	input := `main.py:10: error: Cannot find implementation or library stub for module named "nonexistent"`

	parser := NewMypyParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*MypyResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *MypyResult", result.Data)
	}

	if len(got.Errors) != 1 {
		t.Fatalf("MypyResult.Errors length = %d, want 1", len(got.Errors))
	}

	wantError := MypyError{
		File:     "main.py",
		Line:     10,
		Severity: "error",
		Message:  `Cannot find implementation or library stub for module named "nonexistent"`,
		Code:     "",
	}

	if got.Errors[0] != wantError {
		t.Errorf("MypyResult.Errors[0] = %+v, want %+v", got.Errors[0], wantError)
	}
}

func TestMypyParser_NoteMessage(t *testing.T) {
	input := `main.py:10: note: See https://mypy.readthedocs.io/en/stable/running_mypy.html#missing-imports`

	parser := NewMypyParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*MypyResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *MypyResult", result.Data)
	}

	if len(got.Errors) != 1 {
		t.Fatalf("MypyResult.Errors length = %d, want 1", len(got.Errors))
	}

	if got.Errors[0].Severity != "note" {
		t.Errorf("Error.Severity = %q, want %q", got.Errors[0].Severity, "note")
	}
}

func TestMypyParser_DifferentFormats(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantErrors []MypyError
	}{
		{
			name:  "full path file",
			input: "/home/user/project/src/main.py:10: error: Type error  [type-arg]",
			wantErrors: []MypyError{
				{File: "/home/user/project/src/main.py", Line: 10, Severity: "error", Message: "Type error", Code: "type-arg"},
			},
		},
		{
			name:  "relative path file",
			input: "./src/main.py:15: error: Missing return statement  [return]",
			wantErrors: []MypyError{
				{File: "./src/main.py", Line: 15, Severity: "error", Message: "Missing return statement", Code: "return"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewMypyParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*MypyResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *MypyResult", result.Data)
			}

			if len(got.Errors) != len(tt.wantErrors) {
				t.Fatalf("MypyResult.Errors length = %d, want %d", len(got.Errors), len(tt.wantErrors))
			}

			for i, wantError := range tt.wantErrors {
				if got.Errors[i] != wantError {
					t.Errorf("MypyResult.Errors[%d] = %+v, want %+v", i, got.Errors[i], wantError)
				}
			}
		})
	}
}
