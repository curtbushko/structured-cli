package fileops_test

import (
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/fileops"
)

func TestFindParser_MultipleFiles(t *testing.T) {
	input := `/home/user/project/src/main.go
/home/user/project/src/util.go
/home/user/project/test/main_test.go
`
	parser := fileops.NewFindParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.FindOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.FindOutput, got %T", result.Data)
	}

	if len(output.Files) != 3 {
		t.Fatalf("Files len = %d, want 3", len(output.Files))
	}
	if output.Count != 3 {
		t.Errorf("Count = %d, want 3", output.Count)
	}
	if output.Files[0] != "/home/user/project/src/main.go" {
		t.Errorf("Files[0] = %q, want %q", output.Files[0], "/home/user/project/src/main.go")
	}
}

func TestFindParser_SingleFile(t *testing.T) {
	input := `./config.yaml
`
	parser := fileops.NewFindParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.FindOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.FindOutput, got %T", result.Data)
	}

	if len(output.Files) != 1 {
		t.Fatalf("Files len = %d, want 1", len(output.Files))
	}
	if output.Count != 1 {
		t.Errorf("Count = %d, want 1", output.Count)
	}
}

func TestFindParser_NoResults(t *testing.T) {
	input := ``
	parser := fileops.NewFindParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.FindOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.FindOutput, got %T", result.Data)
	}

	if len(output.Files) != 0 {
		t.Errorf("Files len = %d, want 0", len(output.Files))
	}
	if output.Count != 0 {
		t.Errorf("Count = %d, want 0", output.Count)
	}
}

func TestFindParser_Schema(t *testing.T) {
	parser := fileops.NewFindParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema().ID is empty")
	}
	if schema.Title == "" {
		t.Error("Schema().Title is empty")
	}
}

func TestFindParser_Matches(t *testing.T) {
	parser := fileops.NewFindParser()

	tests := []struct {
		cmd         string
		subcommands []string
		want        bool
	}{
		{"find", []string{}, true},
		{"find", []string{".", "-name", "*.go"}, true},
		{"find", []string{"/home", "-type", "f"}, true},
		{"ls", []string{}, false},
		{"fd", []string{}, false},
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

func TestFindParser_FilesWithSpaces(t *testing.T) {
	input := `/home/user/My Documents/file.txt
/home/user/path with spaces/data.csv
`
	parser := fileops.NewFindParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.FindOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.FindOutput, got %T", result.Data)
	}

	if len(output.Files) != 2 {
		t.Fatalf("Files len = %d, want 2", len(output.Files))
	}
	if output.Files[0] != "/home/user/My Documents/file.txt" {
		t.Errorf("Files[0] = %q, want %q", output.Files[0], "/home/user/My Documents/file.txt")
	}
}
