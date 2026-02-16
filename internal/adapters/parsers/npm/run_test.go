package npm

import (
	"strings"
	"testing"
)

func TestRunParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData RunResult
	}{
		{
			name:  "empty output",
			input: "",
			wantData: RunResult{
				Success:  true,
				Script:   "",
				Output:   "",
				ExitCode: 0,
			},
		},
		{
			name: "successful script run",
			input: `> myproject@1.0.0 build
> tsc

Compilation successful.
`,
			wantData: RunResult{
				Success:  true,
				Script:   "build",
				Output:   "> myproject@1.0.0 build\n> tsc\n\nCompilation successful.\n",
				ExitCode: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewRunParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
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

			if got.Success != tt.wantData.Success {
				t.Errorf("RunResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if got.Script != tt.wantData.Script {
				t.Errorf("RunResult.Script = %v, want %v", got.Script, tt.wantData.Script)
			}
		})
	}
}

func TestRunParser_WithError(t *testing.T) {
	input := `> myproject@1.0.0 build
> tsc

error TS2322: Type 'string' is not assignable to type 'number'.
npm ERR! code ELIFECYCLE
npm ERR! errno 2
`

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
		t.Error("RunResult.Success = true, want false when script failed")
	}

	if got.Script != "build" {
		t.Errorf("RunResult.Script = %q, want %q", got.Script, "build")
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
			name:        "matches npm run build",
			cmd:         "npm",
			subcommands: []string{"run", "build"},
			want:        true,
		},
		{
			name:        "matches npm run-script",
			cmd:         "npm",
			subcommands: []string{"run-script", "test"},
			want:        true,
		},
		{
			name:        "matches npm run",
			cmd:         "npm",
			subcommands: []string{"run"},
			want:        true,
		},
		{
			name:        "does not match npm install",
			cmd:         "npm",
			subcommands: []string{"install"},
			want:        false,
		},
		{
			name:        "does not match yarn run",
			cmd:         "yarn",
			subcommands: []string{"run"},
			want:        false,
		},
		{
			name:        "does not match empty",
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

	if schema.Type != "object" {
		t.Errorf("Schema.Type = %q, want %q", schema.Type, "object")
	}

	requiredProps := []string{"success", "script", "output"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}
