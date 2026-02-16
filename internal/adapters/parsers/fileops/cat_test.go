package fileops_test

import (
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/fileops"
)

func TestCatParser_SimpleContent(t *testing.T) {
	input := `Hello, World!
This is line 2.
And line 3.
`
	parser := fileops.NewCatParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.CatOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.CatOutput, got %T", result.Data)
	}

	if output.Content != input {
		t.Errorf("Content = %q, want %q", output.Content, input)
	}
	if output.Lines != 3 {
		t.Errorf("Lines = %d, want 3", output.Lines)
	}
	if output.Bytes != len(input) {
		t.Errorf("Bytes = %d, want %d", output.Bytes, len(input))
	}
}

func TestCatParser_EmptyFile(t *testing.T) {
	input := ``
	parser := fileops.NewCatParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.CatOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.CatOutput, got %T", result.Data)
	}

	if output.Content != "" {
		t.Errorf("Content = %q, want empty", output.Content)
	}
	if output.Lines != 0 {
		t.Errorf("Lines = %d, want 0", output.Lines)
	}
	if output.Bytes != 0 {
		t.Errorf("Bytes = %d, want 0", output.Bytes)
	}
}

func TestCatParser_SingleLine(t *testing.T) {
	input := `Single line without newline`
	parser := fileops.NewCatParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.CatOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.CatOutput, got %T", result.Data)
	}

	if output.Content != input {
		t.Errorf("Content = %q, want %q", output.Content, input)
	}
	if output.Lines != 1 {
		t.Errorf("Lines = %d, want 1", output.Lines)
	}
}

func TestCatParser_Schema(t *testing.T) {
	parser := fileops.NewCatParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema().ID is empty")
	}
	if schema.Title == "" {
		t.Error("Schema().Title is empty")
	}
}

func TestCatParser_Matches(t *testing.T) {
	parser := fileops.NewCatParser()

	tests := []struct {
		cmd         string
		subcommands []string
		want        bool
	}{
		{"cat", []string{}, true},
		{"cat", []string{"file.txt"}, true},
		{"cat", []string{"-n", "file.txt"}, true},
		{"head", []string{}, false},
		{"tail", []string{}, false},
		{"less", []string{}, false},
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

func TestCatParser_BinaryContent(t *testing.T) {
	// Binary content is passed through as-is
	input := "binary\x00content\xFFhere"
	parser := fileops.NewCatParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.CatOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.CatOutput, got %T", result.Data)
	}

	if output.Content != input {
		t.Errorf("Content mismatch for binary data")
	}
	if output.Bytes != len(input) {
		t.Errorf("Bytes = %d, want %d", output.Bytes, len(input))
	}
}

func TestCatParser_MultipleNewlines(t *testing.T) {
	input := "line1\n\nline3\n"
	parser := fileops.NewCatParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.CatOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.CatOutput, got %T", result.Data)
	}

	if output.Lines != 3 {
		t.Errorf("Lines = %d, want 3", output.Lines)
	}
}
