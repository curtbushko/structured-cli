package test

import (
	"strings"
	"testing"
)

func TestCargoTestParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData CargoTestResult
	}{
		{
			name:  "empty output indicates no tests",
			input: "",
			wantData: CargoTestResult{
				Passed:   0,
				Failed:   0,
				Ignored:  0,
				Measured: 0,
				Filtered: 0,
				Duration: 0,
				Tests:    []CargoTestCase{},
			},
		},
		{
			name: "all tests passed",
			input: `running 3 tests
test utils::test_add ... ok
test utils::test_multiply ... ok
test utils::test_divide ... ok

test result: ok. 3 passed; 0 failed; 0 ignored; 0 measured; 0 filtered out; finished in 0.12s`,
			wantData: CargoTestResult{
				Passed:   3,
				Failed:   0,
				Ignored:  0,
				Measured: 0,
				Filtered: 0,
				Duration: 0.12,
				Tests:    []CargoTestCase{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewCargoTestParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*CargoTestResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *CargoTestResult", result.Data)
			}

			if got.Passed != tt.wantData.Passed {
				t.Errorf("CargoTestResult.Passed = %d, want %d", got.Passed, tt.wantData.Passed)
			}

			if got.Failed != tt.wantData.Failed {
				t.Errorf("CargoTestResult.Failed = %d, want %d", got.Failed, tt.wantData.Failed)
			}

			if got.Ignored != tt.wantData.Ignored {
				t.Errorf("CargoTestResult.Ignored = %d, want %d", got.Ignored, tt.wantData.Ignored)
			}
		})
	}
}

func TestCargoTestParser_Failures(t *testing.T) {
	input := `running 3 tests
test utils::test_add ... ok
test utils::test_multiply ... FAILED
test utils::test_divide ... ok

failures:

---- utils::test_multiply stdout ----
thread 'utils::test_multiply' panicked at 'assertion failed: (left == right)
  left: 5,
 right: 6', src/utils.rs:15:9

failures:
    utils::test_multiply

test result: FAILED. 2 passed; 1 failed; 0 ignored; 0 measured; 0 filtered out; finished in 0.15s`

	parser := NewCargoTestParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*CargoTestResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *CargoTestResult", result.Data)
	}

	if got.Passed != 2 {
		t.Errorf("CargoTestResult.Passed = %d, want 2", got.Passed)
	}

	if got.Failed != 1 {
		t.Errorf("CargoTestResult.Failed = %d, want 1", got.Failed)
	}
}

func TestCargoTestParser_Ignored(t *testing.T) {
	input := `running 3 tests
test utils::test_add ... ok
test utils::test_multiply ... ignored
test utils::test_divide ... ignored

test result: ok. 1 passed; 0 failed; 2 ignored; 0 measured; 0 filtered out; finished in 0.08s`

	parser := NewCargoTestParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*CargoTestResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *CargoTestResult", result.Data)
	}

	if got.Passed != 1 {
		t.Errorf("CargoTestResult.Passed = %d, want 1", got.Passed)
	}

	if got.Ignored != 2 {
		t.Errorf("CargoTestResult.Ignored = %d, want 2", got.Ignored)
	}
}

func TestCargoTestParser_Filtered(t *testing.T) {
	input := `running 1 test
test utils::test_add ... ok

test result: ok. 1 passed; 0 failed; 0 ignored; 0 measured; 5 filtered out; finished in 0.05s`

	parser := NewCargoTestParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*CargoTestResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *CargoTestResult", result.Data)
	}

	if got.Passed != 1 {
		t.Errorf("CargoTestResult.Passed = %d, want 1", got.Passed)
	}

	if got.Filtered != 5 {
		t.Errorf("CargoTestResult.Filtered = %d, want 5", got.Filtered)
	}
}

func TestCargoTestParser_IndividualTests(t *testing.T) {
	input := `running 2 tests
test utils::test_add ... ok
test utils::test_multiply ... FAILED

test result: FAILED. 1 passed; 1 failed; 0 ignored; 0 measured; 0 filtered out; finished in 0.10s`

	parser := NewCargoTestParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*CargoTestResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *CargoTestResult", result.Data)
	}

	if len(got.Tests) != 2 {
		t.Fatalf("CargoTestResult.Tests length = %d, want 2", len(got.Tests))
	}

	if got.Tests[0].Name != "utils::test_add" {
		t.Errorf("CargoTestResult.Tests[0].Name = %q, want %q", got.Tests[0].Name, "utils::test_add")
	}

	if got.Tests[0].Status != "ok" {
		t.Errorf("CargoTestResult.Tests[0].Status = %q, want %q", got.Tests[0].Status, "ok")
	}

	if got.Tests[1].Name != "utils::test_multiply" {
		t.Errorf("CargoTestResult.Tests[1].Name = %q, want %q", got.Tests[1].Name, "utils::test_multiply")
	}

	if got.Tests[1].Status != "FAILED" {
		t.Errorf("CargoTestResult.Tests[1].Status = %q, want %q", got.Tests[1].Status, "FAILED")
	}
}

func TestCargoTestParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches cargo test",
			cmd:         "cargo",
			subcommands: []string{"test"},
			want:        true,
		},
		{
			name:        "matches cargo test with flags",
			cmd:         "cargo",
			subcommands: []string{"test", "--", "--nocapture"},
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

	parser := NewCargoTestParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestCargoTestParser_Schema(t *testing.T) {
	parser := NewCargoTestParser()
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

	requiredProps := []string{"passed", "failed", "ignored", "tests"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}
