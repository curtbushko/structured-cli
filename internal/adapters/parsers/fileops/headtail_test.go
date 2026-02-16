package fileops_test

import (
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/fileops"
)

func TestHeadParser_MultipleLines(t *testing.T) {
	input := `Line 1
Line 2
Line 3
Line 4
Line 5
`
	parser := fileops.NewHeadParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.HeadTailOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.HeadTailOutput, got %T", result.Data)
	}

	if output.LineCount != 5 {
		t.Errorf("LineCount = %d, want 5", output.LineCount)
	}
	if len(output.Lines) != 5 {
		t.Fatalf("Lines len = %d, want 5", len(output.Lines))
	}
	if output.Lines[0] != "Line 1" {
		t.Errorf("Lines[0] = %q, want %q", output.Lines[0], "Line 1")
	}
	if output.Lines[4] != "Line 5" {
		t.Errorf("Lines[4] = %q, want %q", output.Lines[4], "Line 5")
	}
}

func TestHeadParser_EmptyInput(t *testing.T) {
	input := ``
	parser := fileops.NewHeadParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.HeadTailOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.HeadTailOutput, got %T", result.Data)
	}

	if output.LineCount != 0 {
		t.Errorf("LineCount = %d, want 0", output.LineCount)
	}
	if len(output.Lines) != 0 {
		t.Errorf("Lines len = %d, want 0", len(output.Lines))
	}
}

func TestHeadParser_Schema(t *testing.T) {
	parser := fileops.NewHeadParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema().ID is empty")
	}
	if schema.Title == "" {
		t.Error("Schema().Title is empty")
	}
}

func TestHeadParser_Matches(t *testing.T) {
	parser := fileops.NewHeadParser()

	tests := []struct {
		cmd         string
		subcommands []string
		want        bool
	}{
		{"head", []string{}, true},
		{"head", []string{"file.txt"}, true},
		{"head", []string{"-n", "20", "file.txt"}, true},
		{"tail", []string{}, false},
		{"cat", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.cmd+"_"+strings.Join(tt.subcommands, "_"), func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestTailParser_MultipleLines(t *testing.T) {
	input := `Line 1
Line 2
Line 3
Line 4
Line 5
`
	parser := fileops.NewTailParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.HeadTailOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.HeadTailOutput, got %T", result.Data)
	}

	if output.LineCount != 5 {
		t.Errorf("LineCount = %d, want 5", output.LineCount)
	}
	if len(output.Lines) != 5 {
		t.Fatalf("Lines len = %d, want 5", len(output.Lines))
	}
}

func TestTailParser_EmptyInput(t *testing.T) {
	input := ``
	parser := fileops.NewTailParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.HeadTailOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.HeadTailOutput, got %T", result.Data)
	}

	if output.LineCount != 0 {
		t.Errorf("LineCount = %d, want 0", output.LineCount)
	}
}

func TestTailParser_Schema(t *testing.T) {
	parser := fileops.NewTailParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema().ID is empty")
	}
	if schema.Title == "" {
		t.Error("Schema().Title is empty")
	}
}

func TestTailParser_Matches(t *testing.T) {
	parser := fileops.NewTailParser()

	tests := []struct {
		cmd         string
		subcommands []string
		want        bool
	}{
		{"tail", []string{}, true},
		{"tail", []string{"file.txt"}, true},
		{"tail", []string{"-n", "20", "file.txt"}, true},
		{"tail", []string{"-f", "log.txt"}, true},
		{"head", []string{}, false},
		{"cat", []string{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.cmd+"_"+strings.Join(tt.subcommands, "_"), func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestHeadParser_LineWithoutNewline(t *testing.T) {
	input := `Line without newline`
	parser := fileops.NewHeadParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.HeadTailOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.HeadTailOutput, got %T", result.Data)
	}

	if output.LineCount != 1 {
		t.Errorf("LineCount = %d, want 1", output.LineCount)
	}
	if output.Lines[0] != "Line without newline" {
		t.Errorf("Lines[0] = %q, want %q", output.Lines[0], "Line without newline")
	}
}

func TestHeadParser_ContentPreserved(t *testing.T) {
	input := `Line 1
Line 2
`
	parser := fileops.NewHeadParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.HeadTailOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.HeadTailOutput, got %T", result.Data)
	}

	expectedContent := "Line 1\nLine 2\n"
	if output.Content != expectedContent {
		t.Errorf("Content = %q, want %q", output.Content, expectedContent)
	}
}
