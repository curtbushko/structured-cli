package fileops_test

import (
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/fileops"
)

func TestRipgrepParser_StandardOutput(t *testing.T) {
	input := `main.go:10:5:func main() {
main.go:25:10:    fmt.Println("Hello")
util.go:5:1:func helper() {
`
	parser := fileops.NewRipgrepParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.RipgrepOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.RipgrepOutput, got %T", result.Data)
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
	if output.Matches[0].File != "main.go" {
		t.Errorf("Matches[0].File = %q, want %q", output.Matches[0].File, "main.go")
	}
	if output.Matches[0].Line != 10 {
		t.Errorf("Matches[0].Line = %d, want 10", output.Matches[0].Line)
	}
	if output.Matches[0].Column != 5 {
		t.Errorf("Matches[0].Column = %d, want 5", output.Matches[0].Column)
	}
	if output.Matches[0].Content != "func main() {" {
		t.Errorf("Matches[0].Content = %q, want %q", output.Matches[0].Content, "func main() {")
	}
}

func TestRipgrepParser_NoColumnNumbers(t *testing.T) {
	// rg without --column flag
	input := `main.go:10:func main() {
util.go:5:func helper() {
`
	parser := fileops.NewRipgrepParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.RipgrepOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.RipgrepOutput, got %T", result.Data)
	}

	if len(output.Matches) != 2 {
		t.Fatalf("Matches len = %d, want 2", len(output.Matches))
	}

	if output.Matches[0].Line != 10 {
		t.Errorf("Matches[0].Line = %d, want 10", output.Matches[0].Line)
	}
	if output.Matches[0].Column != 0 {
		t.Errorf("Matches[0].Column = %d, want 0", output.Matches[0].Column)
	}
}

func TestRipgrepParser_NoResults(t *testing.T) {
	input := ``
	parser := fileops.NewRipgrepParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.RipgrepOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.RipgrepOutput, got %T", result.Data)
	}

	if len(output.Matches) != 0 {
		t.Errorf("Matches len = %d, want 0", len(output.Matches))
	}
	if output.Count != 0 {
		t.Errorf("Count = %d, want 0", output.Count)
	}
}

func TestRipgrepParser_Schema(t *testing.T) {
	parser := fileops.NewRipgrepParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema().ID is empty")
	}
	if schema.Title == "" {
		t.Error("Schema().Title is empty")
	}
}

func TestRipgrepParser_Matches(t *testing.T) {
	parser := fileops.NewRipgrepParser()

	tests := []struct {
		cmd         string
		subcommands []string
		want        bool
	}{
		{"rg", []string{}, true},
		{"rg", []string{"pattern"}, true},
		{"rg", []string{"-n", "--column", "TODO"}, true},
		{"ripgrep", []string{}, true},
		{"grep", []string{}, false},
		{"find", []string{}, false},
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

func TestRipgrepParser_ColonInContent(t *testing.T) {
	input := `config.yaml:5:1:  key: value
`
	parser := fileops.NewRipgrepParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.RipgrepOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.RipgrepOutput, got %T", result.Data)
	}

	if len(output.Matches) != 1 {
		t.Fatalf("Matches len = %d, want 1", len(output.Matches))
	}
	if output.Matches[0].Content != "  key: value" {
		t.Errorf("Matches[0].Content = %q, want %q", output.Matches[0].Content, "  key: value")
	}
}

func TestRipgrepParser_FilesOnly(t *testing.T) {
	// rg --files-with-matches or rg -l
	input := `main.go
util.go
config.yaml
`
	parser := fileops.NewRipgrepParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.RipgrepOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.RipgrepOutput, got %T", result.Data)
	}

	// Files-only output becomes matches with just file names
	if len(output.Matches) != 3 {
		t.Fatalf("Matches len = %d, want 3", len(output.Matches))
	}
	if output.Matches[0].File != "main.go" {
		t.Errorf("Matches[0].File = %q, want %q", output.Matches[0].File, "main.go")
	}
}
