package python

import (
	"strings"
	"testing"
)

func TestBlackParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData BlackResult
	}{
		{
			name:  "empty output indicates success",
			input: "",
			wantData: BlackResult{
				Success:            true,
				FilesWouldReformat: []string{},
				Errors:             []BlackError{},
			},
		},
		{
			name: "all files clean",
			input: `All done! ✨ 🍰 ✨
3 files would be left unchanged.`,
			wantData: BlackResult{
				Success:            true,
				FilesChecked:       3,
				FilesUnchanged:     3,
				FilesWouldReformat: []string{},
				Errors:             []BlackError{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewBlackParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*BlackResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *BlackResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("BlackResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if len(got.FilesWouldReformat) != len(tt.wantData.FilesWouldReformat) {
				t.Errorf("BlackResult.FilesWouldReformat length = %d, want %d", len(got.FilesWouldReformat), len(tt.wantData.FilesWouldReformat))
			}
		})
	}
}

func TestBlackParser_FilesNeedReformat(t *testing.T) {
	input := `would reformat main.py
would reformat utils.py
Oh no! 💥 💔 💥
2 files would be reformatted, 3 files would be left unchanged.`

	parser := NewBlackParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*BlackResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *BlackResult", result.Data)
	}

	if got.Success {
		t.Error("BlackResult.Success = true, want false when files need reformatting")
	}

	if len(got.FilesWouldReformat) != 2 {
		t.Fatalf("BlackResult.FilesWouldReformat length = %d, want 2", len(got.FilesWouldReformat))
	}

	wantFiles := []string{"main.py", "utils.py"}
	for i, want := range wantFiles {
		if i < len(got.FilesWouldReformat) && got.FilesWouldReformat[i] != want {
			t.Errorf("BlackResult.FilesWouldReformat[%d] = %q, want %q", i, got.FilesWouldReformat[i], want)
		}
	}

	if got.FilesUnchanged != 3 {
		t.Errorf("BlackResult.FilesUnchanged = %d, want 3", got.FilesUnchanged)
	}
}

func TestBlackParser_WithErrors(t *testing.T) {
	input := `error: cannot format broken.py: Cannot parse: 1:0:
would reformat main.py
Oh no! 💥 💔 💥
1 file would be reformatted, 1 file would fail to reformat.`

	parser := NewBlackParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*BlackResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *BlackResult", result.Data)
	}

	if got.Success {
		t.Error("BlackResult.Success = true, want false when errors present")
	}

	if len(got.Errors) != 1 {
		t.Fatalf("BlackResult.Errors length = %d, want 1", len(got.Errors))
	}

	if got.Errors[0].File != "broken.py" {
		t.Errorf("BlackResult.Errors[0].File = %q, want %q", got.Errors[0].File, "broken.py")
	}

	if len(got.FilesWouldReformat) != 1 {
		t.Errorf("BlackResult.FilesWouldReformat length = %d, want 1", len(got.FilesWouldReformat))
	}
}

func TestBlackParser_DifferentPaths(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantFiles []string
	}{
		{
			name: "relative paths",
			input: `would reformat src/main.py
would reformat tests/test_main.py
Oh no! 💥 💔 💥
2 files would be reformatted.`,
			wantFiles: []string{"src/main.py", "tests/test_main.py"},
		},
		{
			name: "absolute paths",
			input: `would reformat /home/user/project/main.py
Oh no! 💥 💔 💥
1 file would be reformatted.`,
			wantFiles: []string{"/home/user/project/main.py"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewBlackParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			got, ok := result.Data.(*BlackResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *BlackResult", result.Data)
			}

			if len(got.FilesWouldReformat) != len(tt.wantFiles) {
				t.Fatalf("BlackResult.FilesWouldReformat length = %d, want %d", len(got.FilesWouldReformat), len(tt.wantFiles))
			}

			for i, want := range tt.wantFiles {
				if got.FilesWouldReformat[i] != want {
					t.Errorf("BlackResult.FilesWouldReformat[%d] = %q, want %q", i, got.FilesWouldReformat[i], want)
				}
			}
		})
	}
}

func TestBlackParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches black --check",
			cmd:         "black",
			subcommands: []string{"--check"},
			want:        true,
		},
		{
			name:        "matches black --check with path",
			cmd:         "black",
			subcommands: []string{"--check", "."},
			want:        true,
		},
		{
			name:        "matches black with check flag",
			cmd:         "black",
			subcommands: []string{"--diff", "--check", "src/"},
			want:        true,
		},
		{
			name:        "does not match black without check",
			cmd:         "black",
			subcommands: []string{"."},
			want:        false,
		},
		{
			name:        "does not match black alone",
			cmd:         "black",
			subcommands: []string{},
			want:        false,
		},
		{
			name:        "does not match other commands",
			cmd:         "ruff",
			subcommands: []string{"--check"},
			want:        false,
		},
		{
			name:        "does not match empty",
			cmd:         "",
			subcommands: []string{},
			want:        false,
		},
	}

	parser := NewBlackParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestBlackParser_Schema(t *testing.T) {
	parser := NewBlackParser()
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

	requiredProps := []string{"success", "files_would_reformat"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}
