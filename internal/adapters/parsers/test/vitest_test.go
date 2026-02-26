package test

import (
	"strings"
	"testing"
)

const schemaTypeObject = "object"

func TestVitestParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData VitestResult
	}{
		{
			name:  "empty output indicates no tests",
			input: "",
			wantData: VitestResult{
				Passed:   0,
				Failed:   0,
				Skipped:  0,
				Duration: 0,
				Files:    []VitestFile{},
			},
		},
		{
			name: "all tests passed",
			input: ` ✓ src/utils.test.ts (2)
   ✓ add numbers
   ✓ multiply numbers

 Test Files  1 passed (1)
      Tests  2 passed (2)
   Start at  10:30:00
   Duration  1.23s`,
			wantData: VitestResult{
				Passed:   2,
				Failed:   0,
				Skipped:  0,
				Duration: 1.23,
				Files:    []VitestFile{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewVitestParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*VitestResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *VitestResult", result.Data)
			}

			if got.Passed != tt.wantData.Passed {
				t.Errorf("VitestResult.Passed = %d, want %d", got.Passed, tt.wantData.Passed)
			}

			if got.Failed != tt.wantData.Failed {
				t.Errorf("VitestResult.Failed = %d, want %d", got.Failed, tt.wantData.Failed)
			}

			if got.Skipped != tt.wantData.Skipped {
				t.Errorf("VitestResult.Skipped = %d, want %d", got.Skipped, tt.wantData.Skipped)
			}
		})
	}
}

func TestVitestParser_Failures(t *testing.T) {
	input := ` ✓ src/utils.test.ts (2)
   ✓ add numbers
   ✗ multiply numbers

 ⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯ Failed Tests 1 ⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯

 FAIL  src/utils.test.ts > multiply numbers
AssertionError: expected 5 to be 6

 Test Files  1 failed (1)
      Tests  1 failed | 1 passed (2)
   Duration  1.56s`

	parser := NewVitestParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*VitestResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *VitestResult", result.Data)
	}

	if got.Passed != 1 {
		t.Errorf("VitestResult.Passed = %d, want 1", got.Passed)
	}

	if got.Failed != 1 {
		t.Errorf("VitestResult.Failed = %d, want 1", got.Failed)
	}
}

func TestVitestParser_Skipped(t *testing.T) {
	input := ` ✓ src/utils.test.ts (3)
   ✓ add numbers
   ↓ multiply numbers [skipped]
   ↓ divide numbers [skipped]

 Test Files  1 passed (1)
      Tests  2 skipped | 1 passed (3)
   Duration  0.89s`

	parser := NewVitestParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*VitestResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *VitestResult", result.Data)
	}

	if got.Passed != 1 {
		t.Errorf("VitestResult.Passed = %d, want 1", got.Passed)
	}

	if got.Skipped != 2 {
		t.Errorf("VitestResult.Skipped = %d, want 2", got.Skipped)
	}
}

func TestVitestParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches vitest",
			cmd:         "vitest",
			subcommands: []string{},
			want:        true,
		},
		{
			name:        "matches vitest run",
			cmd:         "vitest",
			subcommands: []string{"run"},
			want:        true,
		},
		{
			name:        "matches npx vitest",
			cmd:         "npx",
			subcommands: []string{"vitest"},
			want:        true,
		},
		{
			name:        "matches pnpm vitest",
			cmd:         "pnpm",
			subcommands: []string{"vitest"},
			want:        true,
		},
		{
			name:        "does not match vite",
			cmd:         "vite",
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

	parser := NewVitestParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestVitestParser_Schema(t *testing.T) {
	parser := NewVitestParser()
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

	requiredProps := []string{"passed", "failed", "skipped", "files"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}

func TestVitestParser_DurationFormats(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantTime float64
	}{
		{
			name: "time in seconds",
			input: ` Test Files  1 passed (1)
      Tests  2 passed (2)
   Duration  1.23s`,
			wantTime: 1.23,
		},
		{
			name: "time in milliseconds",
			input: ` Test Files  1 passed (1)
      Tests  2 passed (2)
   Duration  234ms`,
			wantTime: 0.234,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewVitestParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			got, ok := result.Data.(*VitestResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *VitestResult", result.Data)
			}

			if got.Duration != tt.wantTime {
				t.Errorf("VitestResult.Duration = %f, want %f", got.Duration, tt.wantTime)
			}
		})
	}
}
