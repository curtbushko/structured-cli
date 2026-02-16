package golang

import (
	"strings"
	"testing"
)

func TestGenerateParser_NoGenerators(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData GenerateResult
	}{
		{
			name:  "empty output indicates no generators",
			input: "",
			wantData: GenerateResult{
				Success:   true,
				Generated: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewGenerateParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*GenerateResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *GenerateResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("GenerateResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if len(got.Generated) != len(tt.wantData.Generated) {
				t.Errorf("GenerateResult.Generated length = %d, want %d", len(got.Generated), len(tt.wantData.Generated))
			}
		})
	}
}

func TestGenerateParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches go generate",
			cmd:         "go",
			subcommands: []string{"generate"},
			want:        true,
		},
		{
			name:        "matches go generate with path",
			cmd:         "go",
			subcommands: []string{"generate", "./..."},
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
			subcommands: []string{"generate"},
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
			subcommands: []string{"generate"},
			want:        false,
		},
	}

	parser := NewGenerateParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestGenerateParser_Schema(t *testing.T) {
	parser := NewGenerateParser()
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
	requiredProps := []string{"success", "generated"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}

func TestGenerateParser_SingleGenerate(t *testing.T) {
	// go generate -v outputs the file and directive being run
	input := "mypackage/generate.go:3: running stringer"

	parser := NewGenerateParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*GenerateResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *GenerateResult", result.Data)
	}

	if !got.Success {
		t.Error("GenerateResult.Success = false, want true")
	}

	if len(got.Generated) != 1 {
		t.Fatalf("GenerateResult.Generated length = %d, want 1", len(got.Generated))
	}

	want := "mypackage/generate.go"
	if got.Generated[0] != want {
		t.Errorf("GenerateResult.Generated[0] = %q, want %q", got.Generated[0], want)
	}
}

func TestGenerateParser_MultipleGenerates(t *testing.T) {
	// go generate -v output with multiple files
	input := `pkg/types/generate.go:3: running stringer
pkg/models/generate.go:5: running mockgen
internal/service/generate.go:7: running enumer`

	parser := NewGenerateParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*GenerateResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *GenerateResult", result.Data)
	}

	if !got.Success {
		t.Error("GenerateResult.Success = false, want true")
	}

	wantFiles := []string{
		"pkg/types/generate.go",
		"pkg/models/generate.go",
		"internal/service/generate.go",
	}

	if len(got.Generated) != len(wantFiles) {
		t.Fatalf("GenerateResult.Generated length = %d, want %d", len(got.Generated), len(wantFiles))
	}

	for i, want := range wantFiles {
		if got.Generated[i] != want {
			t.Errorf("GenerateResult.Generated[%d] = %q, want %q", i, got.Generated[i], want)
		}
	}
}

func TestGenerateParser_DifferentFormats(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantGenerated []string
	}{
		{
			name:          "directive with full path",
			input:         "/home/user/project/pkg/generate.go:10: running stringer",
			wantGenerated: []string{"/home/user/project/pkg/generate.go"},
		},
		{
			name:          "directive with relative path",
			input:         "./internal/types/generate.go:3: running mockgen",
			wantGenerated: []string{"./internal/types/generate.go"},
		},
		{
			name:          "mixed output with empty lines",
			input:         "\npkg/a/gen.go:3: running stringer\n\npkg/b/gen.go:5: running enumer\n",
			wantGenerated: []string{"pkg/a/gen.go", "pkg/b/gen.go"},
		},
		{
			name:          "empty output",
			input:         "",
			wantGenerated: []string{},
		},
		{
			name:          "whitespace only",
			input:         "   \n\t\n  ",
			wantGenerated: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewGenerateParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*GenerateResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *GenerateResult", result.Data)
			}

			if len(got.Generated) != len(tt.wantGenerated) {
				t.Fatalf("GenerateResult.Generated length = %d, want %d", len(got.Generated), len(tt.wantGenerated))
			}

			for i, want := range tt.wantGenerated {
				if got.Generated[i] != want {
					t.Errorf("GenerateResult.Generated[%d] = %q, want %q", i, got.Generated[i], want)
				}
			}
		})
	}
}
