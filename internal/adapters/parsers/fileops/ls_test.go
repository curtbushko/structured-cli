package fileops_test

import (
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/fileops"
)

func TestLSParser_SimpleList(t *testing.T) {
	input := `file1.txt
file2.txt
directory1
`
	parser := fileops.NewLSParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.LSOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.LSOutput, got %T", result.Data)
	}

	if len(output.Entries) != 3 {
		t.Fatalf("Entries len = %d, want 3", len(output.Entries))
	}
	if output.Entries[0].Name != "file1.txt" {
		t.Errorf("Entries[0].Name = %q, want %q", output.Entries[0].Name, "file1.txt")
	}
}

func TestLSParser_LongFormat(t *testing.T) {
	input := `total 32
drwxr-xr-x  5 user group  4096 2024-01-15 10:30 directory
-rw-r--r--  1 user group  1234 2024-01-15 09:00 file.txt
lrwxrwxrwx  1 user group    11 2024-01-15 08:00 link -> target.txt
`
	parser := fileops.NewLSParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.LSOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.LSOutput, got %T", result.Data)
	}

	if output.Total != 32 {
		t.Errorf("Total = %d, want 32", output.Total)
	}

	if len(output.Entries) != 3 {
		t.Fatalf("Entries len = %d, want 3", len(output.Entries))
	}

	// Check directory entry
	if output.Entries[0].Name != "directory" {
		t.Errorf("Entries[0].Name = %q, want %q", output.Entries[0].Name, "directory")
	}
	if output.Entries[0].Type != fileops.TypeDirectory {
		t.Errorf("Entries[0].Type = %q, want %q", output.Entries[0].Type, fileops.TypeDirectory)
	}
	if output.Entries[0].Permissions != "rwxr-xr-x" {
		t.Errorf("Entries[0].Permissions = %q, want %q", output.Entries[0].Permissions, "rwxr-xr-x")
	}
	if output.Entries[0].Owner != "user" {
		t.Errorf("Entries[0].Owner = %q, want %q", output.Entries[0].Owner, "user")
	}
	if output.Entries[0].Group != "group" {
		t.Errorf("Entries[0].Group = %q, want %q", output.Entries[0].Group, "group")
	}

	// Check file entry
	if output.Entries[1].Name != "file.txt" {
		t.Errorf("Entries[1].Name = %q, want %q", output.Entries[1].Name, "file.txt")
	}
	if output.Entries[1].Type != fileops.TypeFile {
		t.Errorf("Entries[1].Type = %q, want %q", output.Entries[1].Type, fileops.TypeFile)
	}
	if output.Entries[1].Size != 1234 {
		t.Errorf("Entries[1].Size = %d, want 1234", output.Entries[1].Size)
	}

	// Check symlink entry
	if output.Entries[2].Name != "link" {
		t.Errorf("Entries[2].Name = %q, want %q", output.Entries[2].Name, "link")
	}
	if output.Entries[2].Type != fileops.TypeSymlink {
		t.Errorf("Entries[2].Type = %q, want %q", output.Entries[2].Type, fileops.TypeSymlink)
	}
	if output.Entries[2].Target != "target.txt" {
		t.Errorf("Entries[2].Target = %q, want %q", output.Entries[2].Target, "target.txt")
	}
}

func TestLSParser_EmptyDirectory(t *testing.T) {
	input := ``
	parser := fileops.NewLSParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.LSOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.LSOutput, got %T", result.Data)
	}

	if len(output.Entries) != 0 {
		t.Errorf("Entries len = %d, want 0", len(output.Entries))
	}
}

func TestLSParser_Schema(t *testing.T) {
	parser := fileops.NewLSParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema().ID is empty")
	}
	if schema.Title == "" {
		t.Error("Schema().Title is empty")
	}
}

func TestLSParser_Matches(t *testing.T) {
	parser := fileops.NewLSParser()

	tests := []struct {
		cmd         string
		subcommands []string
		want        bool
	}{
		{"ls", []string{}, true},
		{"ls", []string{"-l"}, true},
		{"ls", []string{"-la"}, true},
		{"ls", []string{"/tmp"}, true},
		{"find", []string{}, false},
		{"dir", []string{}, false},
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

func TestLSParser_SpecialFileTypes(t *testing.T) {
	input := `total 0
srwxrwxrwx  1 user group     0 2024-01-15 10:30 socket.sock
prw-r--r--  1 user group     0 2024-01-15 10:30 named_pipe
brw-r-----  1 root disk  8, 0 2024-01-15 10:30 block_device
crw-rw-rw-  1 root root  1, 3 2024-01-15 10:30 char_device
`
	parser := fileops.NewLSParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.LSOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.LSOutput, got %T", result.Data)
	}

	if len(output.Entries) != 4 {
		t.Fatalf("Entries len = %d, want 4", len(output.Entries))
	}

	if output.Entries[0].Type != fileops.TypeSocket {
		t.Errorf("Entries[0].Type = %q, want %q", output.Entries[0].Type, fileops.TypeSocket)
	}
	if output.Entries[1].Type != fileops.TypeFIFO {
		t.Errorf("Entries[1].Type = %q, want %q", output.Entries[1].Type, fileops.TypeFIFO)
	}
	if output.Entries[2].Type != fileops.TypeBlock {
		t.Errorf("Entries[2].Type = %q, want %q", output.Entries[2].Type, fileops.TypeBlock)
	}
	if output.Entries[3].Type != fileops.TypeChar {
		t.Errorf("Entries[3].Type = %q, want %q", output.Entries[3].Type, fileops.TypeChar)
	}
}
