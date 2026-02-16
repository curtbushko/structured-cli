package test

import (
	"strings"
	"testing"
)

func TestJestParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData JestResult
	}{
		{
			name:  "empty output indicates no tests",
			input: "",
			wantData: JestResult{
				Passed:   0,
				Failed:   0,
				Skipped:  0,
				Total:    0,
				Duration: 0,
				Suites:   []JestSuite{},
			},
		},
		{
			name: "all tests passed",
			input: `PASS  src/utils.test.js
  Utils
    ✓ should add numbers (2 ms)
    ✓ should multiply numbers (1 ms)

Test Suites: 1 passed, 1 total
Tests:       2 passed, 2 total
Snapshots:   0 total
Time:        1.234 s`,
			wantData: JestResult{
				Passed:   2,
				Failed:   0,
				Skipped:  0,
				Total:    2,
				Duration: 1.234,
				Suites:   []JestSuite{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewJestParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*JestResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *JestResult", result.Data)
			}

			if got.Passed != tt.wantData.Passed {
				t.Errorf("JestResult.Passed = %d, want %d", got.Passed, tt.wantData.Passed)
			}

			if got.Failed != tt.wantData.Failed {
				t.Errorf("JestResult.Failed = %d, want %d", got.Failed, tt.wantData.Failed)
			}

			if got.Total != tt.wantData.Total {
				t.Errorf("JestResult.Total = %d, want %d", got.Total, tt.wantData.Total)
			}
		})
	}
}

func TestJestParser_Failures(t *testing.T) {
	input := `FAIL  src/utils.test.js
  Utils
    ✓ should add numbers (2 ms)
    ✕ should multiply numbers (5 ms)

  ● Utils › should multiply numbers

    expect(received).toBe(expected)

    Expected: 6
    Received: 5

      at Object.<anonymous> (src/utils.test.js:10:23)

Test Suites: 1 failed, 1 total
Tests:       1 failed, 1 passed, 2 total
Snapshots:   0 total
Time:        1.567 s`

	parser := NewJestParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*JestResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *JestResult", result.Data)
	}

	if got.Passed != 1 {
		t.Errorf("JestResult.Passed = %d, want 1", got.Passed)
	}

	if got.Failed != 1 {
		t.Errorf("JestResult.Failed = %d, want 1", got.Failed)
	}

	if got.Total != 2 {
		t.Errorf("JestResult.Total = %d, want 2", got.Total)
	}
}

func TestJestParser_Skipped(t *testing.T) {
	input := `PASS  src/utils.test.js
  Utils
    ✓ should add numbers (2 ms)
    ○ skipped should multiply numbers
    ○ skipped should divide numbers

Test Suites: 1 passed, 1 total
Tests:       2 skipped, 1 passed, 3 total
Snapshots:   0 total
Time:        0.892 s`

	parser := NewJestParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*JestResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *JestResult", result.Data)
	}

	if got.Passed != 1 {
		t.Errorf("JestResult.Passed = %d, want 1", got.Passed)
	}

	if got.Skipped != 2 {
		t.Errorf("JestResult.Skipped = %d, want 2", got.Skipped)
	}

	if got.Total != 3 {
		t.Errorf("JestResult.Total = %d, want 3", got.Total)
	}
}

func TestJestParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches jest",
			cmd:         "jest",
			subcommands: []string{},
			want:        true,
		},
		{
			name:        "matches jest with path",
			cmd:         "jest",
			subcommands: []string{"--coverage"},
			want:        true,
		},
		{
			name:        "matches npx jest",
			cmd:         "npx",
			subcommands: []string{"jest"},
			want:        true,
		},
		{
			name:        "matches yarn jest",
			cmd:         "yarn",
			subcommands: []string{"jest"},
			want:        true,
		},
		{
			name:        "matches npm test (common alias for jest)",
			cmd:         "npm",
			subcommands: []string{"test"},
			want:        false,
		},
		{
			name:        "does not match empty command",
			cmd:         "",
			subcommands: []string{},
			want:        false,
		},
	}

	parser := NewJestParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestJestParser_Schema(t *testing.T) {
	parser := NewJestParser()
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

	requiredProps := []string{"passed", "failed", "skipped", "total", "suites"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}

func TestJestParser_TimeFormats(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantTime float64
	}{
		{
			name: "time in seconds",
			input: `Test Suites: 1 passed, 1 total
Tests:       2 passed, 2 total
Time:        1.234 s`,
			wantTime: 1.234,
		},
		{
			name: "time in milliseconds",
			input: `Test Suites: 1 passed, 1 total
Tests:       2 passed, 2 total
Time:        234 ms`,
			wantTime: 0.234,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewJestParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			got, ok := result.Data.(*JestResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *JestResult", result.Data)
			}

			if got.Duration != tt.wantTime {
				t.Errorf("JestResult.Duration = %f, want %f", got.Duration, tt.wantTime)
			}
		})
	}
}
