package golang

import (
	"strings"
	"testing"
)

const (
	schemaTypeObject = "object"
	testFileMainGo   = "main.go"
)

func TestFmtParser_AllFormatted(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData FmtResult
	}{
		{
			name:  "empty output indicates all files formatted",
			input: "",
			wantData: FmtResult{
				Unformatted: []string{},
			},
		},
		{
			name:  "whitespace only output indicates all files formatted",
			input: "   \n  \n",
			wantData: FmtResult{
				Unformatted: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewFmtParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
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

			if len(got.Unformatted) != len(tt.wantData.Unformatted) {
				t.Errorf("FmtResult.Unformatted length = %d, want %d", len(got.Unformatted), len(tt.wantData.Unformatted))
			}
		})
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
			name:        "matches gofmt",
			cmd:         "gofmt",
			subcommands: []string{},
			want:        true,
		},
		{
			name:        "matches gofmt with -l flag",
			cmd:         "gofmt",
			subcommands: []string{"-l", "."},
			want:        true,
		},
		{
			name:        "matches go fmt",
			cmd:         "go",
			subcommands: []string{"fmt"},
			want:        true,
		},
		{
			name:        "matches go fmt with path",
			cmd:         "go",
			subcommands: []string{"fmt", "./..."},
			want:        true,
		},
		{
			name:        "does not match go build",
			cmd:         "go",
			subcommands: []string{"build"},
			want:        false,
		},
		{
			name:        "does not match git",
			cmd:         "git",
			subcommands: []string{"fmt"},
			want:        false,
		},
		{
			name:        "does not match go without subcommand",
			cmd:         "go",
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

	if schema.Type != schemaTypeObject {
		t.Errorf("Schema.Type = %q, want %q", schema.Type, schemaTypeObject)
	}

	// Verify required properties exist
	requiredProps := []string{"unformatted"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}

func TestFmtParser_SingleFile(t *testing.T) {
	input := testFileMainGo

	parser := NewFmtParser()
	result, err := parser.Parse(strings.NewReader(input))
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

	if len(got.Unformatted) != 1 {
		t.Fatalf("FmtResult.Unformatted length = %d, want 1", len(got.Unformatted))
	}

	if got.Unformatted[0] != testFileMainGo {
		t.Errorf("FmtResult.Unformatted[0] = %q, want %q", got.Unformatted[0], testFileMainGo)
	}
}

func TestFmtParser_MultipleFiles(t *testing.T) {
	input := `main.go
utils.go
handler.go`

	parser := NewFmtParser()
	result, err := parser.Parse(strings.NewReader(input))
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

	if len(got.Unformatted) != 3 {
		t.Fatalf("FmtResult.Unformatted length = %d, want 3", len(got.Unformatted))
	}

	wantFiles := []string{"main.go", "utils.go", "handler.go"}
	for i, wantFile := range wantFiles {
		if got.Unformatted[i] != wantFile {
			t.Errorf("FmtResult.Unformatted[%d] = %q, want %q", i, got.Unformatted[i], wantFile)
		}
	}
}

func TestFmtParser_DifferentPaths(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantUnformatted []string
	}{
		{
			name:            "full path file",
			input:           "/home/user/project/pkg/handler.go",
			wantUnformatted: []string{"/home/user/project/pkg/handler.go"},
		},
		{
			name:            "relative path file",
			input:           "./internal/app/main.go",
			wantUnformatted: []string{"./internal/app/main.go"},
		},
		{
			name:            "mixed output with empty lines",
			input:           "\nmain.go\n\nutils.go\n",
			wantUnformatted: []string{"main.go", "utils.go"},
		},
		{
			name:            "paths with spaces trimmed",
			input:           "  main.go  \n  utils.go  ",
			wantUnformatted: []string{"main.go", "utils.go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewFmtParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
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

			if len(got.Unformatted) != len(tt.wantUnformatted) {
				t.Fatalf("FmtResult.Unformatted length = %d, want %d", len(got.Unformatted), len(tt.wantUnformatted))
			}

			for i, wantFile := range tt.wantUnformatted {
				if got.Unformatted[i] != wantFile {
					t.Errorf("FmtResult.Unformatted[%d] = %q, want %q", i, got.Unformatted[i], wantFile)
				}
			}
		})
	}
}
