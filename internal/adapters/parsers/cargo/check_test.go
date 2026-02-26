package cargo

import (
	"strings"
	"testing"
)

const (
	schemaTypeObject = "object"
	testFileSrcLibRs = "src/lib.rs"
)

func TestCheckParser_EmptyOutput(t *testing.T) {
	parser := NewCheckParser()
	result, err := parser.Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*CheckResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *CheckResult", result.Data)
	}

	if !got.Success {
		t.Error("CheckResult.Success = false, want true for empty output")
	}

	if len(got.Errors) != 0 {
		t.Errorf("CheckResult.Errors length = %d, want 0", len(got.Errors))
	}

	if len(got.Warnings) != 0 {
		t.Errorf("CheckResult.Warnings length = %d, want 0", len(got.Warnings))
	}
}

func TestCheckParser_Success(t *testing.T) {
	input := `{"reason":"build-finished","success":true}`

	parser := NewCheckParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*CheckResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *CheckResult", result.Data)
	}

	if !got.Success {
		t.Error("CheckResult.Success = false, want true")
	}
}

func TestCheckParser_Warning(t *testing.T) {
	input := `{"reason":"compiler-message","package_id":"my_crate 0.1.0","manifest_path":"/path/Cargo.toml","target":{"kind":["lib"],"crate_types":["lib"],"name":"my_crate","src_path":"/path/src/lib.rs","edition":"2021"},"message":{"message":"unused variable: ` + "`x`" + `","code":{"code":"unused_variables","explanation":null},"level":"warning","spans":[{"file_name":"src/lib.rs","byte_start":100,"byte_end":101,"line_start":5,"line_end":5,"column_start":9,"column_end":10,"is_primary":true,"label":null}],"children":[],"rendered":"warning: unused variable"}}
{"reason":"build-finished","success":true}`

	parser := NewCheckParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*CheckResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *CheckResult", result.Data)
	}

	if !got.Success {
		t.Error("CheckResult.Success = false, want true")
	}

	if len(got.Warnings) != 1 {
		t.Fatalf("CheckResult.Warnings length = %d, want 1", len(got.Warnings))
	}

	warn := got.Warnings[0]
	if warn.Message != "unused variable: `x`" {
		t.Errorf("Warning.Message = %q, want %q", warn.Message, "unused variable: `x`")
	}

	if warn.Code != "unused_variables" {
		t.Errorf("Warning.Code = %q, want %q", warn.Code, "unused_variables")
	}

	if warn.File != testFileSrcLibRs {
		t.Errorf("Warning.File = %q, want %q", warn.File, testFileSrcLibRs)
	}

	if warn.Line != 5 {
		t.Errorf("Warning.Line = %d, want 5", warn.Line)
	}

	if warn.Column != 9 {
		t.Errorf("Warning.Column = %d, want 9", warn.Column)
	}
}

func TestCheckParser_Error(t *testing.T) {
	input := `{"reason":"compiler-message","package_id":"my_crate 0.1.0","manifest_path":"/path/Cargo.toml","target":{"kind":["lib"],"crate_types":["lib"],"name":"my_crate","src_path":"/path/src/lib.rs","edition":"2021"},"message":{"message":"cannot find value ` + "`foo`" + ` in this scope","code":{"code":"E0425","explanation":"An unresolved name was used."},"level":"error","spans":[{"file_name":"src/lib.rs","byte_start":50,"byte_end":53,"line_start":3,"line_end":3,"column_start":5,"column_end":8,"is_primary":true,"label":"not found"}],"children":[],"rendered":"error[E0425]: cannot find value"}}
{"reason":"build-finished","success":false}`

	parser := NewCheckParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*CheckResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *CheckResult", result.Data)
	}

	if got.Success {
		t.Error("CheckResult.Success = true, want false")
	}

	if len(got.Errors) != 1 {
		t.Fatalf("CheckResult.Errors length = %d, want 1", len(got.Errors))
	}

	compErr := got.Errors[0]
	if compErr.Message != "cannot find value `foo` in this scope" {
		t.Errorf("Error.Message = %q, want %q", compErr.Message, "cannot find value `foo` in this scope")
	}

	if compErr.Code != "E0425" {
		t.Errorf("Error.Code = %q, want %q", compErr.Code, "E0425")
	}

	if compErr.File != testFileSrcLibRs {
		t.Errorf("Error.File = %q, want %q", compErr.File, testFileSrcLibRs)
	}

	if compErr.Line != 3 {
		t.Errorf("Error.Line = %d, want 3", compErr.Line)
	}

	if compErr.Column != 5 {
		t.Errorf("Error.Column = %d, want 5", compErr.Column)
	}
}

func TestCheckParser_MultipleErrorsAndWarnings(t *testing.T) {
	input := `{"reason":"compiler-message","package_id":"pkg 0.1.0","manifest_path":"/Cargo.toml","target":{"kind":["lib"],"crate_types":["lib"],"name":"pkg","src_path":"/src/lib.rs","edition":"2021"},"message":{"message":"warning 1","code":{"code":"W001","explanation":null},"level":"warning","spans":[{"file_name":"src/lib.rs","byte_start":10,"byte_end":11,"line_start":1,"line_end":1,"column_start":1,"column_end":2,"is_primary":true,"label":null}],"children":[],"rendered":"warning"}}
{"reason":"compiler-message","package_id":"pkg 0.1.0","manifest_path":"/Cargo.toml","target":{"kind":["lib"],"crate_types":["lib"],"name":"pkg","src_path":"/src/lib.rs","edition":"2021"},"message":{"message":"error 1","code":{"code":"E0001","explanation":null},"level":"error","spans":[{"file_name":"src/lib.rs","byte_start":20,"byte_end":21,"line_start":2,"line_end":2,"column_start":1,"column_end":2,"is_primary":true,"label":null}],"children":[],"rendered":"error"}}
{"reason":"compiler-message","package_id":"pkg 0.1.0","manifest_path":"/Cargo.toml","target":{"kind":["lib"],"crate_types":["lib"],"name":"pkg","src_path":"/src/lib.rs","edition":"2021"},"message":{"message":"error 2","code":{"code":"E0002","explanation":null},"level":"error","spans":[{"file_name":"src/lib.rs","byte_start":30,"byte_end":31,"line_start":3,"line_end":3,"column_start":1,"column_end":2,"is_primary":true,"label":null}],"children":[],"rendered":"error"}}
{"reason":"build-finished","success":false}`

	parser := NewCheckParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*CheckResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *CheckResult", result.Data)
	}

	if got.Success {
		t.Error("CheckResult.Success = true, want false")
	}

	if len(got.Errors) != 2 {
		t.Errorf("CheckResult.Errors length = %d, want 2", len(got.Errors))
	}

	if len(got.Warnings) != 1 {
		t.Errorf("CheckResult.Warnings length = %d, want 1", len(got.Warnings))
	}
}

func TestCheckParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches cargo check",
			cmd:         "cargo",
			subcommands: []string{"check"},
			want:        true,
		},
		{
			name:        "matches cargo check with flags",
			cmd:         "cargo",
			subcommands: []string{"check", "--all-targets"},
			want:        true,
		},
		{
			name:        "matches cargo c (alias)",
			cmd:         "cargo",
			subcommands: []string{"c"},
			want:        true,
		},
		{
			name:        "does not match cargo build",
			cmd:         "cargo",
			subcommands: []string{"build"},
			want:        false,
		},
		{
			name:        "does not match cargo without subcommand",
			cmd:         "cargo",
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

	parser := NewCheckParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestCheckParser_Schema(t *testing.T) {
	parser := NewCheckParser()
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

	requiredProps := []string{"success", "errors", "warnings"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}
