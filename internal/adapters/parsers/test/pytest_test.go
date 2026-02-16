package test

import (
	"strings"
	"testing"
)

func TestPytestParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData PytestResult
	}{
		{
			name:  "empty output indicates no tests",
			input: "",
			wantData: PytestResult{
				Passed:   0,
				Failed:   0,
				Skipped:  0,
				Errors:   0,
				Duration: 0,
				Tests:    []PytestCase{},
			},
		},
		{
			name: "all tests passed short form",
			input: `============================= test session starts ==============================
collected 3 items

test_example.py ...                                                      [100%]

============================== 3 passed in 0.12s ===============================`,
			wantData: PytestResult{
				Passed:   3,
				Failed:   0,
				Skipped:  0,
				Errors:   0,
				Duration: 0.12,
				Tests:    []PytestCase{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewPytestParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*PytestResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *PytestResult", result.Data)
			}

			if got.Passed != tt.wantData.Passed {
				t.Errorf("PytestResult.Passed = %d, want %d", got.Passed, tt.wantData.Passed)
			}

			if got.Failed != tt.wantData.Failed {
				t.Errorf("PytestResult.Failed = %d, want %d", got.Failed, tt.wantData.Failed)
			}

			if got.Skipped != tt.wantData.Skipped {
				t.Errorf("PytestResult.Skipped = %d, want %d", got.Skipped, tt.wantData.Skipped)
			}
		})
	}
}

func TestPytestParser_Failures(t *testing.T) {
	input := `============================= test session starts ==============================
collected 3 items

test_example.py .F.                                                      [100%]

=================================== FAILURES ===================================
________________________________ test_example __________________________________

    def test_example():
>       assert 1 == 2
E       assert 1 == 2

test_example.py:5: AssertionError
=========================== short test summary info ============================
FAILED test_example.py::test_example - assert 1 == 2
========================= 2 passed, 1 failed in 0.15s ==========================`

	parser := NewPytestParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*PytestResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *PytestResult", result.Data)
	}

	if got.Passed != 2 {
		t.Errorf("PytestResult.Passed = %d, want 2", got.Passed)
	}

	if got.Failed != 1 {
		t.Errorf("PytestResult.Failed = %d, want 1", got.Failed)
	}

	if got.Duration != 0.15 {
		t.Errorf("PytestResult.Duration = %f, want 0.15", got.Duration)
	}
}

func TestPytestParser_Skipped(t *testing.T) {
	input := `============================= test session starts ==============================
collected 5 items

test_example.py ..ss.                                                    [100%]

========================= 3 passed, 2 skipped in 0.08s =========================`

	parser := NewPytestParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*PytestResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *PytestResult", result.Data)
	}

	if got.Passed != 3 {
		t.Errorf("PytestResult.Passed = %d, want 3", got.Passed)
	}

	if got.Skipped != 2 {
		t.Errorf("PytestResult.Skipped = %d, want 2", got.Skipped)
	}
}

func TestPytestParser_VerboseOutput(t *testing.T) {
	input := `============================= test session starts ==============================
collected 2 items

test_example.py::test_one PASSED                                         [ 50%]
test_example.py::test_two PASSED                                         [100%]

============================== 2 passed in 0.05s ===============================`

	parser := NewPytestParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*PytestResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *PytestResult", result.Data)
	}

	if got.Passed != 2 {
		t.Errorf("PytestResult.Passed = %d, want 2", got.Passed)
	}

	if len(got.Tests) != 2 {
		t.Fatalf("PytestResult.Tests length = %d, want 2", len(got.Tests))
	}

	if got.Tests[0].Name != "test_one" {
		t.Errorf("PytestResult.Tests[0].Name = %q, want %q", got.Tests[0].Name, "test_one")
	}

	if got.Tests[0].Outcome != "passed" {
		t.Errorf("PytestResult.Tests[0].Outcome = %q, want %q", got.Tests[0].Outcome, "passed")
	}
}

func TestPytestParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches pytest",
			cmd:         "pytest",
			subcommands: []string{},
			want:        true,
		},
		{
			name:        "matches pytest with path",
			cmd:         "pytest",
			subcommands: []string{"tests/"},
			want:        true,
		},
		{
			name:        "matches python -m pytest",
			cmd:         "python",
			subcommands: []string{"-m", "pytest"},
			want:        true,
		},
		{
			name:        "matches python3 -m pytest",
			cmd:         "python3",
			subcommands: []string{"-m", "pytest"},
			want:        true,
		},
		{
			name:        "does not match python without pytest",
			cmd:         "python",
			subcommands: []string{"script.py"},
			want:        false,
		},
		{
			name:        "does not match empty command",
			cmd:         "",
			subcommands: []string{},
			want:        false,
		},
	}

	parser := NewPytestParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestPytestParser_Schema(t *testing.T) {
	parser := NewPytestParser()
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

	requiredProps := []string{"passed", "failed", "skipped", "tests"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}

func TestPytestParser_Errors(t *testing.T) {
	input := `============================= test session starts ==============================
collected 2 items

test_example.py E.                                                       [100%]

=================================== ERRORS =====================================
_____________________ ERROR at setup of test_error _____________________________

    @pytest.fixture
    def broken_fixture():
>       raise ValueError("setup failed")
E       ValueError: setup failed

test_example.py:10: ValueError
========================= 1 passed, 1 error in 0.10s ===========================`

	parser := NewPytestParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*PytestResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *PytestResult", result.Data)
	}

	if got.Passed != 1 {
		t.Errorf("PytestResult.Passed = %d, want 1", got.Passed)
	}

	if got.Errors != 1 {
		t.Errorf("PytestResult.Errors = %d, want 1", got.Errors)
	}
}
