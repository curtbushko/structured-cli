package fileops_test

import (
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/fileops"
)

func TestWCParser_SingleFile(t *testing.T) {
	// wc output format: lines words bytes filename
	input := `      10      50     500 file.txt
`
	parser := fileops.NewWCParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.WCOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.WCOutput, got %T", result.Data)
	}

	if len(output.Files) != 1 {
		t.Fatalf("Files len = %d, want 1", len(output.Files))
	}
	if output.Files[0].File != "file.txt" {
		t.Errorf("Files[0].File = %q, want %q", output.Files[0].File, "file.txt")
	}
	if output.Files[0].Lines != 10 {
		t.Errorf("Files[0].Lines = %d, want 10", output.Files[0].Lines)
	}
	if output.Files[0].Words != 50 {
		t.Errorf("Files[0].Words = %d, want 50", output.Files[0].Words)
	}
	if output.Files[0].Bytes != 500 {
		t.Errorf("Files[0].Bytes = %d, want 500", output.Files[0].Bytes)
	}
}

func TestWCParser_MultipleFiles(t *testing.T) {
	input := `      10      50     500 file1.txt
      20     100    1000 file2.txt
      30     150    1500 total
`
	parser := fileops.NewWCParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.WCOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.WCOutput, got %T", result.Data)
	}

	if len(output.Files) != 2 {
		t.Fatalf("Files len = %d, want 2", len(output.Files))
	}
	if output.Total == nil {
		t.Fatal("Total is nil, want non-nil")
	}
	if output.Total.Lines != 30 {
		t.Errorf("Total.Lines = %d, want 30", output.Total.Lines)
	}
	if output.Total.Words != 150 {
		t.Errorf("Total.Words = %d, want 150", output.Total.Words)
	}
	if output.Total.Bytes != 1500 {
		t.Errorf("Total.Bytes = %d, want 1500", output.Total.Bytes)
	}
}

func TestWCParser_StdinInput(t *testing.T) {
	// wc with no filename (reading from stdin)
	input := `      10      50     500
`
	parser := fileops.NewWCParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.WCOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.WCOutput, got %T", result.Data)
	}

	if len(output.Files) != 1 {
		t.Fatalf("Files len = %d, want 1", len(output.Files))
	}
	if output.Files[0].File != "" {
		t.Errorf("Files[0].File = %q, want empty", output.Files[0].File)
	}
	if output.Files[0].Lines != 10 {
		t.Errorf("Files[0].Lines = %d, want 10", output.Files[0].Lines)
	}
}

func TestWCParser_EmptyInput(t *testing.T) {
	input := ``
	parser := fileops.NewWCParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.WCOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.WCOutput, got %T", result.Data)
	}

	if len(output.Files) != 0 {
		t.Errorf("Files len = %d, want 0", len(output.Files))
	}
}

func TestWCParser_Schema(t *testing.T) {
	parser := fileops.NewWCParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema().ID is empty")
	}
	if schema.Title == "" {
		t.Error("Schema().Title is empty")
	}
}

func TestWCParser_Matches(t *testing.T) {
	parser := fileops.NewWCParser()

	tests := []struct {
		cmd         string
		subcommands []string
		want        bool
	}{
		{"wc", []string{}, true},
		{"wc", []string{"file.txt"}, true},
		{"wc", []string{"-l", "file.txt"}, true},
		{"wc", []string{"-w", "-c", "file.txt"}, true},
		{"cat", []string{}, false},
		{"head", []string{}, false},
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

func TestWCParser_LinesOnly(t *testing.T) {
	// wc -l output
	input := `      10 file.txt
`
	parser := fileops.NewWCParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.WCOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.WCOutput, got %T", result.Data)
	}

	if len(output.Files) != 1 {
		t.Fatalf("Files len = %d, want 1", len(output.Files))
	}
	if output.Files[0].Lines != 10 {
		t.Errorf("Files[0].Lines = %d, want 10", output.Files[0].Lines)
	}
	if output.Files[0].File != "file.txt" {
		t.Errorf("Files[0].File = %q, want %q", output.Files[0].File, "file.txt")
	}
}

func TestWCParser_CharCountWithFlag(t *testing.T) {
	// wc -m output (character count)
	input := `     100 file.txt
`
	parser := fileops.NewWCParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.WCOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.WCOutput, got %T", result.Data)
	}

	if len(output.Files) != 1 {
		t.Fatalf("Files len = %d, want 1", len(output.Files))
	}
	// With single value, we treat it as lines (most common use case)
	if output.Files[0].Lines != 100 {
		t.Errorf("Files[0].Lines = %d, want 100", output.Files[0].Lines)
	}
}
