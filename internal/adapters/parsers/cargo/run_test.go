package cargo

import (
	"strings"
	"testing"
)

const errorCodeE0425 = "E0425"

func TestRunParser_EmptyOutput(t *testing.T) {
	parser := NewRunParser()
	result, err := parser.Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*RunResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *RunResult", result.Data)
	}

	if !got.Success {
		t.Error("RunResult.Success = false, want true for empty output")
	}

	if len(got.Errors) != 0 {
		t.Errorf("RunResult.Errors length = %d, want 0", len(got.Errors))
	}
}

func TestRunParser_BuildAndRunSuccess(t *testing.T) {
	input := `{"reason":"compiler-artifact","package_id":"my_app 0.1.0","manifest_path":"/path/Cargo.toml","target":{"kind":["bin"],"crate_types":["bin"],"name":"my_app","src_path":"/path/src/main.rs","edition":"2021"},"profile":{"opt_level":"0","debuginfo":2,"debug_assertions":true,"overflow_checks":true,"test":false},"features":[],"filenames":["/path/target/debug/my_app"],"executable":"/path/target/debug/my_app","fresh":false}
{"reason":"build-finished","success":true}
Hello, World!`

	parser := NewRunParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*RunResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *RunResult", result.Data)
	}

	if !got.Success {
		t.Error("RunResult.Success = false, want true")
	}

	if !got.BuildSuccess {
		t.Error("RunResult.BuildSuccess = false, want true")
	}

	if got.Executable != "/path/target/debug/my_app" {
		t.Errorf("RunResult.Executable = %q, want %q", got.Executable, "/path/target/debug/my_app")
	}
}

func TestRunParser_BuildFailure(t *testing.T) {
	input := `{"reason":"compiler-message","package_id":"my_app 0.1.0","manifest_path":"/path/Cargo.toml","target":{"kind":["bin"],"crate_types":["bin"],"name":"my_app","src_path":"/path/src/main.rs","edition":"2021"},"message":{"message":"cannot find value ` + "`foo`" + ` in this scope","code":{"code":"E0425","explanation":null},"level":"error","spans":[{"file_name":"src/main.rs","byte_start":50,"byte_end":53,"line_start":5,"line_end":5,"column_start":5,"column_end":8,"is_primary":true,"label":"not found"}],"children":[],"rendered":"error[E0425]: cannot find value"}}
{"reason":"build-finished","success":false}`

	parser := NewRunParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*RunResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *RunResult", result.Data)
	}

	if got.Success {
		t.Error("RunResult.Success = true, want false")
	}

	if got.BuildSuccess {
		t.Error("RunResult.BuildSuccess = true, want false")
	}

	if len(got.Errors) != 1 {
		t.Fatalf("RunResult.Errors length = %d, want 1", len(got.Errors))
	}

	if got.Errors[0].Code != errorCodeE0425 {
		t.Errorf("Error.Code = %q, want %q", got.Errors[0].Code, errorCodeE0425)
	}
}

func TestRunParser_ProgramOutput(t *testing.T) {
	input := `{"reason":"build-finished","success":true}
Line 1
Line 2
Line 3`

	parser := NewRunParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*RunResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *RunResult", result.Data)
	}

	if !got.Success {
		t.Error("RunResult.Success = false, want true")
	}

	// Output should contain the non-JSON lines
	if !strings.Contains(got.Output, "Line 1") {
		t.Errorf("RunResult.Output should contain 'Line 1', got %q", got.Output)
	}
}

func TestRunParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches cargo run",
			cmd:         "cargo",
			subcommands: []string{"run"},
			want:        true,
		},
		{
			name:        "matches cargo run with flags",
			cmd:         "cargo",
			subcommands: []string{"run", "--release"},
			want:        true,
		},
		{
			name:        "matches cargo r (alias)",
			cmd:         "cargo",
			subcommands: []string{"r"},
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

	parser := NewRunParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestRunParser_Schema(t *testing.T) {
	parser := NewRunParser()
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

	requiredProps := []string{"success", "build_success", "errors"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}
