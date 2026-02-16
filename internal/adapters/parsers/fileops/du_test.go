package fileops_test

import (
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/fileops"
)

func TestDUParser_BasicOutput(t *testing.T) {
	// du output format: size\tpath
	input := `4096	./src
8192	./bin
12288	.
`
	parser := fileops.NewDUParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.DUOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.DUOutput, got %T", result.Data)
	}

	if len(output.Entries) != 3 {
		t.Fatalf("Entries len = %d, want 3", len(output.Entries))
	}
	if output.Entries[0].Path != "./src" {
		t.Errorf("Entries[0].Path = %q, want %q", output.Entries[0].Path, "./src")
	}
	if output.Entries[0].Size != 4096 {
		t.Errorf("Entries[0].Size = %d, want 4096", output.Entries[0].Size)
	}
}

func TestDUParser_HumanReadable(t *testing.T) {
	// du -h output
	input := `4.0K	./src
8.0K	./bin
12K	.
`
	parser := fileops.NewDUParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.DUOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.DUOutput, got %T", result.Data)
	}

	if len(output.Entries) != 3 {
		t.Fatalf("Entries len = %d, want 3", len(output.Entries))
	}
	if output.Entries[0].SizeHuman != "4.0K" {
		t.Errorf("Entries[0].SizeHuman = %q, want %q", output.Entries[0].SizeHuman, "4.0K")
	}
	// Human-readable should still parse to bytes
	if output.Entries[0].Size != 4096 {
		t.Errorf("Entries[0].Size = %d, want 4096", output.Entries[0].Size)
	}
}

func TestDUParser_WithTotal(t *testing.T) {
	// du -c output
	input := `4096	./src
8192	./bin
12288	total
`
	parser := fileops.NewDUParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.DUOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.DUOutput, got %T", result.Data)
	}

	if len(output.Entries) != 2 {
		t.Fatalf("Entries len = %d, want 2", len(output.Entries))
	}
	if output.Total != 12288 {
		t.Errorf("Total = %d, want 12288", output.Total)
	}
}

func TestDUParser_EmptyInput(t *testing.T) {
	input := ``
	parser := fileops.NewDUParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.DUOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.DUOutput, got %T", result.Data)
	}

	if len(output.Entries) != 0 {
		t.Errorf("Entries len = %d, want 0", len(output.Entries))
	}
}

func TestDUParser_Schema(t *testing.T) {
	parser := fileops.NewDUParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema().ID is empty")
	}
	if schema.Title == "" {
		t.Error("Schema().Title is empty")
	}
}

func TestDUParser_Matches(t *testing.T) {
	parser := fileops.NewDUParser()

	tests := []struct {
		cmd         string
		subcommands []string
		want        bool
	}{
		{"du", []string{}, true},
		{"du", []string{"-h"}, true},
		{"du", []string{"-sh", "."}, true},
		{"du", []string{"-c", "--max-depth=1"}, true},
		{"df", []string{}, false},
		{"ls", []string{}, false},
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

func TestDUParser_LargeSizes(t *testing.T) {
	input := `1073741824	./large-dir
10737418240	.
`
	parser := fileops.NewDUParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.DUOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.DUOutput, got %T", result.Data)
	}

	if len(output.Entries) != 2 {
		t.Fatalf("Entries len = %d, want 2", len(output.Entries))
	}
	if output.Entries[0].Size != 1073741824 {
		t.Errorf("Entries[0].Size = %d, want 1073741824", output.Entries[0].Size)
	}
}

func TestDUParser_HumanReadableSizes(t *testing.T) {
	input := `1.5M	./medium
2.3G	./large
`
	parser := fileops.NewDUParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.DUOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.DUOutput, got %T", result.Data)
	}

	if len(output.Entries) != 2 {
		t.Fatalf("Entries len = %d, want 2", len(output.Entries))
	}

	// Check M suffix parsing (1.5M = 1572864 bytes)
	if output.Entries[0].Size != 1572864 {
		t.Errorf("Entries[0].Size = %d, want 1572864", output.Entries[0].Size)
	}
	if output.Entries[0].SizeHuman != "1.5M" {
		t.Errorf("Entries[0].SizeHuman = %q, want %q", output.Entries[0].SizeHuman, "1.5M")
	}

	// Check G suffix parsing (2.3G bytes approximately)
	var expectedGigaSize int64 = 2469606195 // 2.3 * 1024^3
	if output.Entries[1].Size != expectedGigaSize {
		t.Errorf("Entries[1].Size = %d, want %d", output.Entries[1].Size, expectedGigaSize)
	}
}
