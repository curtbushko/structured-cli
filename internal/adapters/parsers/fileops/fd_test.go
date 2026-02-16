package fileops_test

import (
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/fileops"
)

func TestFDParser_MultipleFiles(t *testing.T) {
	input := `src/main.go
src/util.go
test/main_test.go
`
	parser := fileops.NewFDParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.FDOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.FDOutput, got %T", result.Data)
	}

	if len(output.Files) != 3 {
		t.Fatalf("Files len = %d, want 3", len(output.Files))
	}
	if output.Count != 3 {
		t.Errorf("Count = %d, want 3", output.Count)
	}
	if output.Files[0] != "src/main.go" {
		t.Errorf("Files[0] = %q, want %q", output.Files[0], "src/main.go")
	}
}

func TestFDParser_AbsolutePaths(t *testing.T) {
	input := `/home/user/project/main.go
/home/user/project/util.go
`
	parser := fileops.NewFDParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.FDOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.FDOutput, got %T", result.Data)
	}

	if len(output.Files) != 2 {
		t.Fatalf("Files len = %d, want 2", len(output.Files))
	}
	if output.Files[0] != "/home/user/project/main.go" {
		t.Errorf("Files[0] = %q, want %q", output.Files[0], "/home/user/project/main.go")
	}
}

func TestFDParser_NoResults(t *testing.T) {
	input := ``
	parser := fileops.NewFDParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.FDOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.FDOutput, got %T", result.Data)
	}

	if len(output.Files) != 0 {
		t.Errorf("Files len = %d, want 0", len(output.Files))
	}
	if output.Count != 0 {
		t.Errorf("Count = %d, want 0", output.Count)
	}
}

func TestFDParser_Schema(t *testing.T) {
	parser := fileops.NewFDParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema().ID is empty")
	}
	if schema.Title == "" {
		t.Error("Schema().Title is empty")
	}
}

func TestFDParser_Matches(t *testing.T) {
	parser := fileops.NewFDParser()

	tests := []struct {
		cmd         string
		subcommands []string
		want        bool
	}{
		{"fd", []string{}, true},
		{"fd", []string{"pattern"}, true},
		{"fd", []string{"-e", "go", "main"}, true},
		{"fdfind", []string{}, true},
		{"find", []string{}, false},
		{"rg", []string{}, false},
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

func TestFDParser_FilesWithSpaces(t *testing.T) {
	input := `My Documents/file.txt
path with spaces/data.csv
`
	parser := fileops.NewFDParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.FDOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.FDOutput, got %T", result.Data)
	}

	if len(output.Files) != 2 {
		t.Fatalf("Files len = %d, want 2", len(output.Files))
	}
	if output.Files[0] != "My Documents/file.txt" {
		t.Errorf("Files[0] = %q, want %q", output.Files[0], "My Documents/file.txt")
	}
}
