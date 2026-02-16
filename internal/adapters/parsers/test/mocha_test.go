package test

import (
	"strings"
	"testing"
)

func TestMochaParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData MochaResult
	}{
		{
			name:  "empty output indicates no tests",
			input: "",
			wantData: MochaResult{
				Passed:   0,
				Failed:   0,
				Pending:  0,
				Duration: 0,
				Suites:   []MochaSuite{},
			},
		},
		{
			name: "all tests passed",
			input: `  Utils
    ✓ should add numbers (2ms)
    ✓ should multiply numbers (1ms)

  2 passing (10ms)`,
			wantData: MochaResult{
				Passed:   2,
				Failed:   0,
				Pending:  0,
				Duration: 10,
				Suites:   []MochaSuite{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewMochaParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*MochaResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *MochaResult", result.Data)
			}

			if got.Passed != tt.wantData.Passed {
				t.Errorf("MochaResult.Passed = %d, want %d", got.Passed, tt.wantData.Passed)
			}

			if got.Failed != tt.wantData.Failed {
				t.Errorf("MochaResult.Failed = %d, want %d", got.Failed, tt.wantData.Failed)
			}

			if got.Pending != tt.wantData.Pending {
				t.Errorf("MochaResult.Pending = %d, want %d", got.Pending, tt.wantData.Pending)
			}
		})
	}
}

func TestMochaParser_Failures(t *testing.T) {
	input := `  Utils
    ✓ should add numbers (2ms)
    1) should multiply numbers

  1 passing (15ms)
  1 failing

  1) Utils
       should multiply numbers:

      AssertionError: expected 5 to equal 6
      + expected - actual

      -5
      +6

      at Context.<anonymous> (test/utils.test.js:15:14)`

	parser := NewMochaParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*MochaResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *MochaResult", result.Data)
	}

	if got.Passed != 1 {
		t.Errorf("MochaResult.Passed = %d, want 1", got.Passed)
	}

	if got.Failed != 1 {
		t.Errorf("MochaResult.Failed = %d, want 1", got.Failed)
	}
}

func TestMochaParser_Pending(t *testing.T) {
	input := `  Utils
    ✓ should add numbers (2ms)
    - should multiply numbers
    - should divide numbers

  1 passing (8ms)
  2 pending`

	parser := NewMochaParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*MochaResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *MochaResult", result.Data)
	}

	if got.Passed != 1 {
		t.Errorf("MochaResult.Passed = %d, want 1", got.Passed)
	}

	if got.Pending != 2 {
		t.Errorf("MochaResult.Pending = %d, want 2", got.Pending)
	}
}

func TestMochaParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches mocha",
			cmd:         "mocha",
			subcommands: []string{},
			want:        true,
		},
		{
			name:        "matches mocha with path",
			cmd:         "mocha",
			subcommands: []string{"test/"},
			want:        true,
		},
		{
			name:        "matches npx mocha",
			cmd:         "npx",
			subcommands: []string{"mocha"},
			want:        true,
		},
		{
			name:        "matches yarn mocha",
			cmd:         "yarn",
			subcommands: []string{"mocha"},
			want:        true,
		},
		{
			name:        "does not match empty command",
			cmd:         "",
			subcommands: []string{},
			want:        false,
		},
	}

	parser := NewMochaParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestMochaParser_Schema(t *testing.T) {
	parser := NewMochaParser()
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

	requiredProps := []string{"passed", "failed", "pending", "suites"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}

func TestMochaParser_DurationFormats(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantTime float64
	}{
		{
			name:     "time in milliseconds",
			input:    `  2 passing (234ms)`,
			wantTime: 234,
		},
		{
			name:     "time in seconds",
			input:    `  2 passing (1s)`,
			wantTime: 1000,
		},
		{
			name:     "time in minutes",
			input:    `  2 passing (1m)`,
			wantTime: 60000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewMochaParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			got, ok := result.Data.(*MochaResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *MochaResult", result.Data)
			}

			if got.Duration != tt.wantTime {
				t.Errorf("MochaResult.Duration = %f, want %f", got.Duration, tt.wantTime)
			}
		})
	}
}
