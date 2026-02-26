package build

import (
	"strings"
	"testing"
)

func TestCargoParser_EmptyOutput(t *testing.T) {
	parser := NewCargoParser()
	result, err := parser.Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*CargoResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *CargoResult", result.Data)
	}

	if !got.Success {
		t.Error("CargoResult.Success = false, want true for empty output")
	}

	if len(got.Errors) != 0 {
		t.Errorf("CargoResult.Errors length = %d, want 0", len(got.Errors))
	}

	if len(got.Warnings) != 0 {
		t.Errorf("CargoResult.Warnings length = %d, want 0", len(got.Warnings))
	}
}

func TestCargoParser_BuildFinishedSuccess(t *testing.T) {
	input := `{"reason":"build-finished","success":true}`

	parser := NewCargoParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*CargoResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *CargoResult", result.Data)
	}

	if !got.Success {
		t.Error("CargoResult.Success = false, want true")
	}
}

func TestCargoParser_BuildFinishedFailure(t *testing.T) {
	input := `{"reason":"build-finished","success":false}`

	parser := NewCargoParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*CargoResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *CargoResult", result.Data)
	}

	if got.Success {
		t.Error("CargoResult.Success = true, want false")
	}
}

func TestCargoParser_CompilerArtifact(t *testing.T) {
	input := `{"reason":"compiler-artifact","package_id":"file:///path/to/my-package#0.1.0","manifest_path":"/path/to/my-package/Cargo.toml","target":{"kind":["lib"],"crate_types":["lib"],"name":"my_package","src_path":"/path/to/my-package/src/lib.rs","edition":"2018"},"profile":{"opt_level":"0","debuginfo":2,"debug_assertions":true,"overflow_checks":true,"test":false},"features":["feat1"],"filenames":["/path/to/my-package/target/debug/libmy_package.rlib"],"executable":null,"fresh":false}
{"reason":"build-finished","success":true}`

	parser := NewCargoParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*CargoResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *CargoResult", result.Data)
	}

	if !got.Success {
		t.Error("CargoResult.Success = false, want true")
	}

	if len(got.Artifacts) != 1 {
		t.Fatalf("CargoResult.Artifacts length = %d, want 1", len(got.Artifacts))
	}

	artifact := got.Artifacts[0]
	if artifact.PackageID != "file:///path/to/my-package#0.1.0" {
		t.Errorf("Artifact.PackageID = %q, want %q", artifact.PackageID, "file:///path/to/my-package#0.1.0")
	}

	if artifact.Target.Name != "my_package" {
		t.Errorf("Artifact.Target.Name = %q, want %q", artifact.Target.Name, "my_package")
	}

	if len(artifact.Filenames) != 1 {
		t.Fatalf("Artifact.Filenames length = %d, want 1", len(artifact.Filenames))
	}

	if artifact.Fresh {
		t.Error("Artifact.Fresh = true, want false")
	}
}

func TestCargoParser_CompilerWarning(t *testing.T) {
	input := `{"reason":"compiler-message","package_id":"file:///path/to/my-package#0.1.0","manifest_path":"/path/to/my-package/Cargo.toml","target":{"kind":["lib"],"crate_types":["lib"],"name":"my_package","src_path":"/path/to/my-package/src/lib.rs","edition":"2018"},"message":{"$message_type":"diagnostic","message":"unused variable: ` + "`x`" + `","code":{"code":"unused_variables","explanation":null},"level":"warning","spans":[{"file_name":"src/lib.rs","byte_start":21,"byte_end":22,"line_start":2,"line_end":2,"column_start":9,"column_end":10,"is_primary":true,"text":[{"text":"    let x = 123;","highlight_start":9,"highlight_end":10}],"label":null,"suggested_replacement":null,"suggestion_applicability":null,"expansion":null}],"children":[],"rendered":"warning: unused variable"}}
{"reason":"build-finished","success":true}`

	parser := NewCargoParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*CargoResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *CargoResult", result.Data)
	}

	// Warnings don't affect success
	if !got.Success {
		t.Error("CargoResult.Success = false, want true when only warnings present")
	}

	if len(got.Warnings) != 1 {
		t.Fatalf("CargoResult.Warnings length = %d, want 1", len(got.Warnings))
	}

	warn := got.Warnings[0]
	if warn.Message != "unused variable: `x`" {
		t.Errorf("Warning.Message = %q, want %q", warn.Message, "unused variable: `x`")
	}

	if warn.Code != "unused_variables" {
		t.Errorf("Warning.Code = %q, want %q", warn.Code, "unused_variables")
	}

	if warn.File != "src/lib.rs" {
		t.Errorf("Warning.File = %q, want %q", warn.File, "src/lib.rs")
	}

	if warn.Line != 2 {
		t.Errorf("Warning.Line = %d, want %d", warn.Line, 2)
	}

	if warn.Column != 9 {
		t.Errorf("Warning.Column = %d, want %d", warn.Column, 9)
	}
}

func TestCargoParser_CompilerError(t *testing.T) {
	input := `{"reason":"compiler-message","package_id":"file:///path/to/my-package#0.1.0","manifest_path":"/path/to/my-package/Cargo.toml","target":{"kind":["lib"],"crate_types":["lib"],"name":"my_package","src_path":"/path/to/my-package/src/lib.rs","edition":"2018"},"message":{"$message_type":"diagnostic","message":"cannot find value ` + "`foo`" + ` in this scope","code":{"code":"E0425","explanation":"An unresolved name was used."},"level":"error","spans":[{"file_name":"src/lib.rs","byte_start":50,"byte_end":53,"line_start":5,"line_end":5,"column_start":5,"column_end":8,"is_primary":true,"text":[{"text":"    foo;","highlight_start":5,"highlight_end":8}],"label":"not found in this scope","suggested_replacement":null,"suggestion_applicability":null,"expansion":null}],"children":[],"rendered":"error[E0425]: cannot find value"}}
{"reason":"build-finished","success":false}`

	parser := NewCargoParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*CargoResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *CargoResult", result.Data)
	}

	if got.Success {
		t.Error("CargoResult.Success = true, want false when errors present")
	}

	if len(got.Errors) != 1 {
		t.Fatalf("CargoResult.Errors length = %d, want 1", len(got.Errors))
	}

	compErr := got.Errors[0]
	if compErr.Message != "cannot find value `foo` in this scope" {
		t.Errorf("Error.Message = %q, want %q", compErr.Message, "cannot find value `foo` in this scope")
	}

	if compErr.Code != "E0425" {
		t.Errorf("Error.Code = %q, want %q", compErr.Code, "E0425")
	}

	if compErr.File != "src/lib.rs" {
		t.Errorf("Error.File = %q, want %q", compErr.File, "src/lib.rs")
	}

	if compErr.Line != 5 {
		t.Errorf("Error.Line = %d, want %d", compErr.Line, 5)
	}

	if compErr.Column != 5 {
		t.Errorf("Error.Column = %d, want %d", compErr.Column, 5)
	}
}

func TestCargoParser_MultipleErrorsAndWarnings(t *testing.T) {
	input := `{"reason":"compiler-message","package_id":"pkg#0.1.0","manifest_path":"/Cargo.toml","target":{"kind":["lib"],"crate_types":["lib"],"name":"pkg","src_path":"/src/lib.rs","edition":"2021"},"message":{"message":"unused variable: ` + "`a`" + `","code":{"code":"unused_variables","explanation":null},"level":"warning","spans":[{"file_name":"src/lib.rs","byte_start":10,"byte_end":11,"line_start":1,"line_end":1,"column_start":5,"column_end":6,"is_primary":true,"text":[],"label":null,"suggested_replacement":null,"suggestion_applicability":null,"expansion":null}],"children":[],"rendered":"warning"}}
{"reason":"compiler-message","package_id":"pkg#0.1.0","manifest_path":"/Cargo.toml","target":{"kind":["lib"],"crate_types":["lib"],"name":"pkg","src_path":"/src/lib.rs","edition":"2021"},"message":{"message":"cannot find value ` + "`x`" + `","code":{"code":"E0425","explanation":null},"level":"error","spans":[{"file_name":"src/lib.rs","byte_start":20,"byte_end":21,"line_start":2,"line_end":2,"column_start":1,"column_end":2,"is_primary":true,"text":[],"label":null,"suggested_replacement":null,"suggestion_applicability":null,"expansion":null}],"children":[],"rendered":"error"}}
{"reason":"compiler-message","package_id":"pkg#0.1.0","manifest_path":"/Cargo.toml","target":{"kind":["lib"],"crate_types":["lib"],"name":"pkg","src_path":"/src/lib.rs","edition":"2021"},"message":{"message":"expected ` + "`;`" + `","code":{"code":"E0001","explanation":null},"level":"error","spans":[{"file_name":"src/lib.rs","byte_start":30,"byte_end":31,"line_start":3,"line_end":3,"column_start":10,"column_end":11,"is_primary":true,"text":[],"label":null,"suggested_replacement":null,"suggestion_applicability":null,"expansion":null}],"children":[],"rendered":"error"}}
{"reason":"build-finished","success":false}`

	parser := NewCargoParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*CargoResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *CargoResult", result.Data)
	}

	if got.Success {
		t.Error("CargoResult.Success = true, want false when errors present")
	}

	if len(got.Errors) != 2 {
		t.Errorf("CargoResult.Errors length = %d, want 2", len(got.Errors))
	}

	if len(got.Warnings) != 1 {
		t.Errorf("CargoResult.Warnings length = %d, want 1", len(got.Warnings))
	}
}

func TestCargoParser_BuildScriptExecuted(t *testing.T) {
	input := `{"reason":"build-script-executed","package_id":"file:///path/to/pkg#0.1.0","linked_libs":["foo","static=bar"],"linked_paths":["/some/path"],"cfgs":["feature=\"test\""],"env":[["SOME_KEY","some value"]],"out_dir":"/target/debug/build/pkg-abc/out"}
{"reason":"build-finished","success":true}`

	parser := NewCargoParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*CargoResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *CargoResult", result.Data)
	}

	if !got.Success {
		t.Error("CargoResult.Success = false, want true")
	}

	if len(got.BuildScripts) != 1 {
		t.Fatalf("CargoResult.BuildScripts length = %d, want 1", len(got.BuildScripts))
	}

	bs := got.BuildScripts[0]
	if bs.PackageID != "file:///path/to/pkg#0.1.0" {
		t.Errorf("BuildScript.PackageID = %q, want %q", bs.PackageID, "file:///path/to/pkg#0.1.0")
	}

	if len(bs.LinkedLibs) != 2 {
		t.Errorf("BuildScript.LinkedLibs length = %d, want 2", len(bs.LinkedLibs))
	}

	if bs.OutDir != "/target/debug/build/pkg-abc/out" {
		t.Errorf("BuildScript.OutDir = %q, want %q", bs.OutDir, "/target/debug/build/pkg-abc/out")
	}
}

func TestCargoParser_NoSpans(t *testing.T) {
	// Some messages may not have spans (e.g., aborting due to previous errors)
	input := `{"reason":"compiler-message","package_id":"pkg#0.1.0","manifest_path":"/Cargo.toml","target":{"kind":["lib"],"crate_types":["lib"],"name":"pkg","src_path":"/src/lib.rs","edition":"2021"},"message":{"message":"aborting due to 2 previous errors","code":null,"level":"error","spans":[],"children":[],"rendered":"error: aborting"}}
{"reason":"build-finished","success":false}`

	parser := NewCargoParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*CargoResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *CargoResult", result.Data)
	}

	if got.Success {
		t.Error("CargoResult.Success = true, want false")
	}

	if len(got.Errors) != 1 {
		t.Fatalf("CargoResult.Errors length = %d, want 1", len(got.Errors))
	}

	// Error without spans should have empty file/line/column
	compErr := got.Errors[0]
	if compErr.File != "" {
		t.Errorf("Error.File = %q, want empty string", compErr.File)
	}
	if compErr.Line != 0 {
		t.Errorf("Error.Line = %d, want 0", compErr.Line)
	}
}

func TestCargoParser_InvalidJSON(t *testing.T) {
	// Parser should skip invalid JSON lines gracefully
	input := `not valid json
{"reason":"build-finished","success":true}`

	parser := NewCargoParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*CargoResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *CargoResult", result.Data)
	}

	// Should still succeed if build-finished says success
	if !got.Success {
		t.Error("CargoResult.Success = false, want true")
	}
}

func TestCargoParser_FreshArtifact(t *testing.T) {
	input := `{"reason":"compiler-artifact","package_id":"file:///path/to/pkg#0.1.0","manifest_path":"/Cargo.toml","target":{"kind":["lib"],"crate_types":["lib"],"name":"pkg","src_path":"/src/lib.rs","edition":"2021"},"profile":{"opt_level":"3","debuginfo":0,"debug_assertions":false,"overflow_checks":false,"test":false},"features":[],"filenames":["/target/release/libpkg.rlib"],"executable":null,"fresh":true}
{"reason":"build-finished","success":true}`

	parser := NewCargoParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*CargoResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *CargoResult", result.Data)
	}

	if len(got.Artifacts) != 1 {
		t.Fatalf("CargoResult.Artifacts length = %d, want 1", len(got.Artifacts))
	}

	if !got.Artifacts[0].Fresh {
		t.Error("Artifact.Fresh = false, want true")
	}
}

func TestCargoParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches cargo build",
			cmd:         "cargo",
			subcommands: []string{"build"},
			want:        true,
		},
		{
			name:        "matches cargo build with flags",
			cmd:         "cargo",
			subcommands: []string{"build", "--release"},
			want:        true,
		},
		{
			name:        "matches cargo b (alias)",
			cmd:         "cargo",
			subcommands: []string{"b"},
			want:        true,
		},
		{
			name:        "does not match cargo test",
			cmd:         "cargo",
			subcommands: []string{"test"},
			want:        false,
		},
		{
			name:        "does not match cargo without subcommand",
			cmd:         "cargo",
			subcommands: []string{},
			want:        false,
		},
		{
			name:        "does not match rustc",
			cmd:         "rustc",
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

	parser := NewCargoParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestCargoParser_Schema(t *testing.T) {
	parser := NewCargoParser()
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

	requiredProps := []string{"success", "errors", "warnings", "artifacts"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}

func TestCargoParser_ExecutableArtifact(t *testing.T) {
	input := `{"reason":"compiler-artifact","package_id":"file:///path/to/my-bin#0.1.0","manifest_path":"/path/to/my-bin/Cargo.toml","target":{"kind":["bin"],"crate_types":["bin"],"name":"my_bin","src_path":"/path/to/my-bin/src/main.rs","edition":"2021"},"profile":{"opt_level":"0","debuginfo":2,"debug_assertions":true,"overflow_checks":true,"test":false},"features":[],"filenames":["/path/to/my-bin/target/debug/my_bin"],"executable":"/path/to/my-bin/target/debug/my_bin","fresh":false}
{"reason":"build-finished","success":true}`

	parser := NewCargoParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*CargoResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *CargoResult", result.Data)
	}

	if len(got.Artifacts) != 1 {
		t.Fatalf("CargoResult.Artifacts length = %d, want 1", len(got.Artifacts))
	}

	artifact := got.Artifacts[0]
	if artifact.Executable != "/path/to/my-bin/target/debug/my_bin" {
		t.Errorf("Artifact.Executable = %q, want %q", artifact.Executable, "/path/to/my-bin/target/debug/my_bin")
	}

	if len(artifact.Target.Kind) != 1 || artifact.Target.Kind[0] != "bin" {
		t.Errorf("Artifact.Target.Kind = %v, want [bin]", artifact.Target.Kind)
	}
}
