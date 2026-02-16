package lint

import (
	"strings"
	"testing"
)

func TestPrettierParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData PrettierResult
	}{
		{
			name:  "empty output indicates all files formatted",
			input: "",
			wantData: PrettierResult{
				Success:     true,
				Unformatted: []string{},
			},
		},
		{
			name:  "checking message with no files indicates success",
			input: "Checking formatting...",
			wantData: PrettierResult{
				Success:     true,
				Unformatted: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewPrettierParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*PrettierResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *PrettierResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("PrettierResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if len(got.Unformatted) != len(tt.wantData.Unformatted) {
				t.Errorf("PrettierResult.Unformatted length = %d, want %d", len(got.Unformatted), len(tt.wantData.Unformatted))
			}
		})
	}
}

func TestPrettierParser_SingleUnformattedFile(t *testing.T) {
	// Prettier --check outputs files that differ from prettier formatting
	input := `Checking formatting...
[warn] src/index.js
[warn] Code style issues found in the above file. Run Prettier to fix.`

	parser := NewPrettierParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*PrettierResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *PrettierResult", result.Data)
	}

	if got.Success {
		t.Error("PrettierResult.Success = true, want false when unformatted files present")
	}

	if len(got.Unformatted) != 1 {
		t.Fatalf("PrettierResult.Unformatted length = %d, want 1", len(got.Unformatted))
	}

	if got.Unformatted[0] != "src/index.js" {
		t.Errorf("PrettierResult.Unformatted[0] = %q, want %q", got.Unformatted[0], "src/index.js")
	}
}

func TestPrettierParser_MultipleUnformattedFiles(t *testing.T) {
	input := `Checking formatting...
[warn] src/index.js
[warn] src/utils.ts
[warn] src/app.tsx
[warn] Code style issues found in 3 files. Run Prettier to fix.`

	parser := NewPrettierParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*PrettierResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *PrettierResult", result.Data)
	}

	if got.Success {
		t.Error("PrettierResult.Success = true, want false when unformatted files present")
	}

	if len(got.Unformatted) != 3 {
		t.Fatalf("PrettierResult.Unformatted length = %d, want 3", len(got.Unformatted))
	}

	wantFiles := []string{"src/index.js", "src/utils.ts", "src/app.tsx"}
	for i, wantFile := range wantFiles {
		if got.Unformatted[i] != wantFile {
			t.Errorf("PrettierResult.Unformatted[%d] = %q, want %q", i, got.Unformatted[i], wantFile)
		}
	}
}

func TestPrettierParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches prettier with --check",
			cmd:         "prettier",
			subcommands: []string{"--check"},
			want:        true,
		},
		{
			name:        "matches prettier with --check and path",
			cmd:         "prettier",
			subcommands: []string{"--check", "src/"},
			want:        true,
		},
		{
			name:        "matches prettier without subcommands",
			cmd:         "prettier",
			subcommands: []string{},
			want:        true,
		},
		{
			name:        "does not match npx prettier",
			cmd:         "npx",
			subcommands: []string{"prettier"},
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

	parser := NewPrettierParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestPrettierParser_Schema(t *testing.T) {
	parser := NewPrettierParser()
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
	requiredProps := []string{"success", "unformatted"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}

func TestPrettierParser_AllFilesFormatted(t *testing.T) {
	// When all files are formatted, prettier --check outputs success message
	input := `Checking formatting...
All matched files use Prettier code style!`

	parser := NewPrettierParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*PrettierResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *PrettierResult", result.Data)
	}

	if !got.Success {
		t.Error("PrettierResult.Success = false, want true when all files formatted")
	}

	if len(got.Unformatted) != 0 {
		t.Errorf("PrettierResult.Unformatted length = %d, want 0", len(got.Unformatted))
	}
}
