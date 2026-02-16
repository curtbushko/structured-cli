package cargo

import (
	"strings"
	"testing"
)

func TestFmtParser_EmptyOutput(t *testing.T) {
	parser := NewFmtParser()
	result, err := parser.Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*FmtResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *FmtResult", result.Data)
	}

	if !got.Success {
		t.Error("FmtResult.Success = false, want true for empty output")
	}

	if len(got.Files) != 0 {
		t.Errorf("FmtResult.Files length = %d, want 0", len(got.Files))
	}
}

func TestFmtParser_AllFormatted(t *testing.T) {
	// When all files are formatted, cargo fmt --check produces no output
	input := ``

	parser := NewFmtParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*FmtResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *FmtResult", result.Data)
	}

	if !got.Success {
		t.Error("FmtResult.Success = false, want true")
	}
}

func TestFmtParser_UnformattedFile(t *testing.T) {
	input := `Diff in /path/to/project/src/lib.rs:2:
 fn main() {
-let x=1;
+    let x = 1;
 }
`

	parser := NewFmtParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*FmtResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *FmtResult", result.Data)
	}

	if got.Success {
		t.Error("FmtResult.Success = true, want false")
	}

	if len(got.Files) != 1 {
		t.Fatalf("FmtResult.Files length = %d, want 1", len(got.Files))
	}

	if got.Files[0].Path != "/path/to/project/src/lib.rs" {
		t.Errorf("File.Path = %q, want %q", got.Files[0].Path, "/path/to/project/src/lib.rs")
	}
}

func TestFmtParser_MultipleUnformattedFiles(t *testing.T) {
	input := `Diff in /path/src/lib.rs:2:
-let x=1;
+    let x = 1;

Diff in /path/src/main.rs:5:
-let y=2;
+    let y = 2;
`

	parser := NewFmtParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*FmtResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *FmtResult", result.Data)
	}

	if got.Success {
		t.Error("FmtResult.Success = true, want false")
	}

	if len(got.Files) != 2 {
		t.Errorf("FmtResult.Files length = %d, want 2", len(got.Files))
	}
}

func TestFmtParser_WouldFormat(t *testing.T) {
	// cargo fmt --check outputs file paths that would be formatted
	input := `src/lib.rs
src/main.rs
`

	parser := NewFmtParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*FmtResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *FmtResult", result.Data)
	}

	if got.Success {
		t.Error("FmtResult.Success = true, want false when files need formatting")
	}

	if len(got.Files) != 2 {
		t.Errorf("FmtResult.Files length = %d, want 2", len(got.Files))
	}
}

func TestFmtParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches cargo fmt",
			cmd:         "cargo",
			subcommands: []string{"fmt"},
			want:        true,
		},
		{
			name:        "matches cargo fmt with check",
			cmd:         "cargo",
			subcommands: []string{"fmt", "--check"},
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

	parser := NewFmtParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestFmtParser_Schema(t *testing.T) {
	parser := NewFmtParser()
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

	requiredProps := []string{"success", "files"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}
