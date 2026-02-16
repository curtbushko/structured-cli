package cargo

import (
	"strings"
	"testing"
)

func TestDocParser_EmptyOutput(t *testing.T) {
	parser := NewDocParser()
	result, err := parser.Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*DocResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *DocResult", result.Data)
	}

	if !got.Success {
		t.Error("DocResult.Success = false, want true for empty output")
	}

	if len(got.Warnings) != 0 {
		t.Errorf("DocResult.Warnings length = %d, want 0", len(got.Warnings))
	}

	if len(got.Errors) != 0 {
		t.Errorf("DocResult.Errors length = %d, want 0", len(got.Errors))
	}
}

func TestDocParser_Success(t *testing.T) {
	input := `{"reason":"build-finished","success":true}`

	parser := NewDocParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*DocResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *DocResult", result.Data)
	}

	if !got.Success {
		t.Error("DocResult.Success = false, want true")
	}
}

func TestDocParser_Warning(t *testing.T) {
	input := `{"reason":"compiler-message","package_id":"my_crate 0.1.0","manifest_path":"/path/Cargo.toml","target":{"kind":["lib"],"crate_types":["lib"],"name":"my_crate","src_path":"/path/src/lib.rs","edition":"2021"},"message":{"message":"missing documentation for a function","code":{"code":"missing_docs","explanation":null},"level":"warning","spans":[{"file_name":"src/lib.rs","byte_start":100,"byte_end":110,"line_start":5,"line_end":5,"column_start":1,"column_end":11,"is_primary":true,"label":null}],"children":[],"rendered":"warning: missing documentation"}}
{"reason":"build-finished","success":true}`

	parser := NewDocParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*DocResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *DocResult", result.Data)
	}

	if !got.Success {
		t.Error("DocResult.Success = false, want true")
	}

	if len(got.Warnings) != 1 {
		t.Fatalf("DocResult.Warnings length = %d, want 1", len(got.Warnings))
	}

	warn := got.Warnings[0]
	if warn.Message != "missing documentation for a function" {
		t.Errorf("Warning.Message = %q, want %q", warn.Message, "missing documentation for a function")
	}

	if warn.File != "src/lib.rs" {
		t.Errorf("Warning.File = %q, want %q", warn.File, "src/lib.rs")
	}

	if warn.Line != 5 {
		t.Errorf("Warning.Line = %d, want 5", warn.Line)
	}
}

func TestDocParser_Error(t *testing.T) {
	input := `{"reason":"compiler-message","package_id":"my_crate 0.1.0","manifest_path":"/path/Cargo.toml","target":{"kind":["lib"],"crate_types":["lib"],"name":"my_crate","src_path":"/path/src/lib.rs","edition":"2021"},"message":{"message":"unresolved link to ` + "`foo`" + `","code":{"code":"rustdoc::broken_intra_doc_links","explanation":null},"level":"error","spans":[{"file_name":"src/lib.rs","byte_start":50,"byte_end":53,"line_start":3,"line_end":3,"column_start":5,"column_end":8,"is_primary":true,"label":null}],"children":[],"rendered":"error: unresolved link"}}
{"reason":"build-finished","success":false}`

	parser := NewDocParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*DocResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *DocResult", result.Data)
	}

	if got.Success {
		t.Error("DocResult.Success = true, want false")
	}

	if len(got.Errors) != 1 {
		t.Fatalf("DocResult.Errors length = %d, want 1", len(got.Errors))
	}

	docErr := got.Errors[0]
	if docErr.Code != "rustdoc::broken_intra_doc_links" {
		t.Errorf("Error.Code = %q, want %q", docErr.Code, "rustdoc::broken_intra_doc_links")
	}
}

func TestDocParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches cargo doc",
			cmd:         "cargo",
			subcommands: []string{"doc"},
			want:        true,
		},
		{
			name:        "matches cargo doc with flags",
			cmd:         "cargo",
			subcommands: []string{"doc", "--no-deps"},
			want:        true,
		},
		{
			name:        "matches cargo d (alias)",
			cmd:         "cargo",
			subcommands: []string{"d"},
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

	parser := NewDocParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestDocParser_Schema(t *testing.T) {
	parser := NewDocParser()
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

	requiredProps := []string{"success", "warnings", "errors"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}
