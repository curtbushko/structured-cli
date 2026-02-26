package fileops_test

import (
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/fileops"
)

const (
	testFileMainGo      = "main.go"
	testContentFuncMain = "func main() {"
)

func TestGrepParser_WithLineNumbers(t *testing.T) {
	input := `main.go:10:func main() {
main.go:25:    fmt.Println("Hello")
util.go:5:func helper() {
`
	parser := fileops.NewGrepParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.GrepOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.GrepOutput, got %T", result.Data)
	}

	if len(output.Matches) != 3 {
		t.Fatalf("Matches len = %d, want 3", len(output.Matches))
	}
	if output.Count != 3 {
		t.Errorf("Count = %d, want 3", output.Count)
	}
	if output.FilesMatched != 2 {
		t.Errorf("FilesMatched = %d, want 2", output.FilesMatched)
	}

	// Check first match
	if output.Matches[0].File != testFileMainGo {
		t.Errorf("Matches[0].File = %q, want %q", output.Matches[0].File, testFileMainGo)
	}
	if output.Matches[0].Line != 10 {
		t.Errorf("Matches[0].Line = %d, want 10", output.Matches[0].Line)
	}
	if output.Matches[0].Content != testContentFuncMain {
		t.Errorf("Matches[0].Content = %q, want %q", output.Matches[0].Content, testContentFuncMain)
	}
}

func TestGrepParser_WithoutLineNumbers(t *testing.T) {
	input := `main.go:func main() {
util.go:func helper() {
`
	parser := fileops.NewGrepParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.GrepOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.GrepOutput, got %T", result.Data)
	}

	if len(output.Matches) != 2 {
		t.Fatalf("Matches len = %d, want 2", len(output.Matches))
	}

	// Check that line numbers are 0 when not provided
	if output.Matches[0].Line != 0 {
		t.Errorf("Matches[0].Line = %d, want 0", output.Matches[0].Line)
	}
	if output.Matches[0].Content != testContentFuncMain {
		t.Errorf("Matches[0].Content = %q, want %q", output.Matches[0].Content, testContentFuncMain)
	}
}

func TestGrepParser_SingleFile(t *testing.T) {
	// grep without filename when searching single file
	input := `10:func main() {
25:    return nil
`
	parser := fileops.NewGrepParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.GrepOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.GrepOutput, got %T", result.Data)
	}

	if len(output.Matches) != 2 {
		t.Fatalf("Matches len = %d, want 2", len(output.Matches))
	}
	if output.Matches[0].Line != 10 {
		t.Errorf("Matches[0].Line = %d, want 10", output.Matches[0].Line)
	}
}

func TestGrepParser_NoResults(t *testing.T) {
	input := ``
	parser := fileops.NewGrepParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.GrepOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.GrepOutput, got %T", result.Data)
	}

	if len(output.Matches) != 0 {
		t.Errorf("Matches len = %d, want 0", len(output.Matches))
	}
	if output.Count != 0 {
		t.Errorf("Count = %d, want 0", output.Count)
	}
}

func TestGrepParser_Schema(t *testing.T) {
	parser := fileops.NewGrepParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema().ID is empty")
	}
	if schema.Title == "" {
		t.Error("Schema().Title is empty")
	}
}

func TestGrepParser_Matches(t *testing.T) {
	parser := fileops.NewGrepParser()

	tests := []struct {
		cmd         string
		subcommands []string
		want        bool
	}{
		{"grep", []string{}, true},
		{"grep", []string{"-r", "pattern"}, true},
		{"grep", []string{"-n", "-r", "TODO"}, true},
		{"egrep", []string{}, true},
		{"fgrep", []string{}, true},
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

func TestGrepParser_BinaryFile(t *testing.T) {
	input := `Binary file image.png matches
main.go:10:func main() {
`
	parser := fileops.NewGrepParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.GrepOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.GrepOutput, got %T", result.Data)
	}

	// Binary file match should be skipped, only 1 match
	if len(output.Matches) != 1 {
		t.Fatalf("Matches len = %d, want 1", len(output.Matches))
	}
}

func TestGrepParser_ColonInContent(t *testing.T) {
	// File has colon in content
	input := `config.yaml:5:  key: value
`
	parser := fileops.NewGrepParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.GrepOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.GrepOutput, got %T", result.Data)
	}

	if len(output.Matches) != 1 {
		t.Fatalf("Matches len = %d, want 1", len(output.Matches))
	}
	if output.Matches[0].Content != "  key: value" {
		t.Errorf("Matches[0].Content = %q, want %q", output.Matches[0].Content, "  key: value")
	}
}
