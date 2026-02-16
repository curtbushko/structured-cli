package makejust

import (
	"strings"
	"testing"
)

func TestMakeParser_EmptyOutput(t *testing.T) {
	parser := NewMakeParser()
	result, err := parser.Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*Result)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *Result", result.Data)
	}

	if !got.Success {
		t.Error("Result.Success = false, want true for empty output")
	}
}

func TestMakeParser_SuccessfulBuild(t *testing.T) {
	input := `gcc -c main.c -o main.o
gcc main.o -o myapp`

	parser := NewMakeParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*Result)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *Result", result.Data)
	}

	if !got.Success {
		t.Error("Result.Success = false, want true")
	}

	if got.ExitCode != 0 {
		t.Errorf("Result.ExitCode = %d, want 0", got.ExitCode)
	}
}

func TestMakeParser_BuildFailure(t *testing.T) {
	input := `gcc -c main.c -o main.o
main.c:5:1: error: expected ';' before '}' token
make: *** [Makefile:3: main.o] Error 1`

	parser := NewMakeParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*Result)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *Result", result.Data)
	}

	if got.Success {
		t.Error("Result.Success = true, want false for build failure")
	}

	if got.Error == "" {
		t.Error("Result.Error should not be empty for build failure")
	}

	if got.ExitCode != 1 {
		t.Errorf("Result.ExitCode = %d, want 1", got.ExitCode)
	}
}

func TestMakeParser_NoRuleError(t *testing.T) {
	input := `make: *** No rule to make target 'nonexistent'. Stop.`

	parser := NewMakeParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*Result)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *Result", result.Data)
	}

	if got.Success {
		t.Error("Result.Success = true, want false for no rule error")
	}

	if !strings.Contains(got.Error, "No rule to make target") {
		t.Errorf("Result.Error = %q, should contain 'No rule to make target'", got.Error)
	}
}

func TestMakeParser_TargetListing(t *testing.T) {
	// Output from make with help targets or special parsing
	input := `Available targets:
  build        Build the application
  test         Run tests
  clean        Clean build artifacts
  install      Install the application`

	parser := NewMakeParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*Result)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *Result", result.Data)
	}

	if len(got.Targets) < 1 {
		t.Fatalf("Result.Targets length = %d, want >= 1", len(got.Targets))
	}

	// Check first target
	found := false
	for _, target := range got.Targets {
		if target.Name == "build" {
			found = true
			if target.Description != "Build the application" {
				t.Errorf("Target.Description = %q, want %q", target.Description, "Build the application")
			}
			break
		}
	}
	if !found {
		t.Error("Expected to find target 'build' in Targets list")
	}
}

func TestMakeParser_DryRun(t *testing.T) {
	input := `echo "Building..."
gcc -c main.c -o main.o
gcc main.o -o myapp
echo "Done!"`

	parser := NewMakeParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*Result)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *Result", result.Data)
	}

	if len(got.Commands) != 4 {
		t.Errorf("Result.Commands length = %d, want 4", len(got.Commands))
	}

	if len(got.Commands) > 0 && got.Commands[0] != `echo "Building..."` {
		t.Errorf("Result.Commands[0] = %q, want %q", got.Commands[0], `echo "Building..."`)
	}
}

func TestMakeParser_NothingToBeDone(t *testing.T) {
	input := `make: Nothing to be done for 'all'.`

	parser := NewMakeParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*Result)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *Result", result.Data)
	}

	if !got.Success {
		t.Error("Result.Success = false, want true for 'nothing to be done'")
	}
}

func TestMakeParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches make",
			cmd:         "make",
			subcommands: []string{},
			want:        true,
		},
		{
			name:        "matches make with target",
			cmd:         "make",
			subcommands: []string{"build"},
			want:        true,
		},
		{
			name:        "matches make with flags",
			cmd:         "make",
			subcommands: []string{"-j4", "all"},
			want:        true,
		},
		{
			name:        "matches gmake",
			cmd:         "gmake",
			subcommands: []string{},
			want:        true,
		},
		{
			name:        "does not match cmake",
			cmd:         "cmake",
			subcommands: []string{},
			want:        false,
		},
		{
			name:        "does not match other commands",
			cmd:         "gcc",
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

	parser := NewMakeParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestMakeParser_Schema(t *testing.T) {
	parser := NewMakeParser()
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

	requiredProps := []string{"success", "exit_code"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}

func TestMakeParser_RecursiveMakeError(t *testing.T) {
	input := `make[1]: Entering directory '/path/to/subdir'
make[1]: *** [Makefile:10: target] Error 2
make[1]: Leaving directory '/path/to/subdir'
make: *** [Makefile:5: recurse] Error 2`

	parser := NewMakeParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*Result)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *Result", result.Data)
	}

	if got.Success {
		t.Error("Result.Success = true, want false for recursive make error")
	}

	if got.ExitCode != 2 {
		t.Errorf("Result.ExitCode = %d, want 2", got.ExitCode)
	}
}
