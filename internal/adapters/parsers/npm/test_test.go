package npm

import (
	"strings"
	"testing"
)

func TestTestParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData TestResult
	}{
		{
			name:  "empty output",
			input: "",
			wantData: TestResult{
				Success:  true,
				Output:   "",
				ExitCode: 0,
			},
		},
		{
			name: "successful test run",
			input: `> myproject@1.0.0 test
> jest

PASS  src/index.test.js
  Calculator
    add
      ✓ adds 1 + 2 to equal 3 (2 ms)

Test Suites: 1 passed, 1 total
Tests:       1 passed, 1 total
`,
			wantData: TestResult{
				Success:  true,
				Output:   "> myproject@1.0.0 test\n> jest\n\nPASS  src/index.test.js\n  Calculator\n    add\n      ✓ adds 1 + 2 to equal 3 (2 ms)\n\nTest Suites: 1 passed, 1 total\nTests:       1 passed, 1 total\n",
				ExitCode: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewTestParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*TestResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *TestResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("TestResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}
		})
	}
}

func TestTestParser_WithFailure(t *testing.T) {
	input := `> myproject@1.0.0 test
> jest

FAIL  src/index.test.js
  Calculator
    add
      ✕ adds 1 + 2 to equal 3 (5 ms)

  ● Calculator › add › adds 1 + 2 to equal 3

    expect(received).toBe(expected)

    Expected: 3
    Received: 4

Test Suites: 1 failed, 1 total
Tests:       1 failed, 1 total
npm ERR! code ELIFECYCLE
`

	parser := NewTestParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*TestResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *TestResult", result.Data)
	}

	if got.Success {
		t.Error("TestResult.Success = true, want false when tests failed")
	}
}

func TestTestParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches npm test",
			cmd:         "npm",
			subcommands: []string{"test"},
			want:        true,
		},
		{
			name:        "matches npm t",
			cmd:         "npm",
			subcommands: []string{"t"},
			want:        true,
		},
		{
			name:        "matches npm tst",
			cmd:         "npm",
			subcommands: []string{"tst"},
			want:        true,
		},
		{
			name:        "does not match npm install",
			cmd:         "npm",
			subcommands: []string{"install"},
			want:        false,
		},
		{
			name:        "does not match yarn test",
			cmd:         "yarn",
			subcommands: []string{"test"},
			want:        false,
		},
		{
			name:        "does not match empty",
			cmd:         "",
			subcommands: []string{},
			want:        false,
		},
	}

	parser := NewTestParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestTestParser_Schema(t *testing.T) {
	parser := NewTestParser()
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

	requiredProps := []string{"success", "output"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}
