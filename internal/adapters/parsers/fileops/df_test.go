package fileops_test

import (
	"strings"
	"testing"

	"github.com/curtbushko/structured-cli/internal/adapters/parsers/fileops"
)

func TestDFParser_BasicOutput(t *testing.T) {
	// df output format: Filesystem 1K-blocks Used Available Use% Mounted on
	input := `Filesystem     1K-blocks      Used Available Use% Mounted on
/dev/sda1      102400000  51200000  51200000  50% /
tmpfs            8000000   100000   7900000   2% /tmp
`
	parser := fileops.NewDFParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.DFOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.DFOutput, got %T", result.Data)
	}

	if len(output.Filesystems) != 2 {
		t.Fatalf("Filesystems len = %d, want 2", len(output.Filesystems))
	}

	// Check first filesystem
	if output.Filesystems[0].Filesystem != "/dev/sda1" {
		t.Errorf("Filesystems[0].Filesystem = %q, want %q", output.Filesystems[0].Filesystem, "/dev/sda1")
	}
	if output.Filesystems[0].Size != 102400000*1024 {
		t.Errorf("Filesystems[0].Size = %d, want %d", output.Filesystems[0].Size, 102400000*1024)
	}
	if output.Filesystems[0].Used != 51200000*1024 {
		t.Errorf("Filesystems[0].Used = %d, want %d", output.Filesystems[0].Used, 51200000*1024)
	}
	if output.Filesystems[0].Available != 51200000*1024 {
		t.Errorf("Filesystems[0].Available = %d, want %d", output.Filesystems[0].Available, 51200000*1024)
	}
	if output.Filesystems[0].UsePercent != 50 {
		t.Errorf("Filesystems[0].UsePercent = %f, want 50", output.Filesystems[0].UsePercent)
	}
	if output.Filesystems[0].MountedOn != "/" {
		t.Errorf("Filesystems[0].MountedOn = %q, want %q", output.Filesystems[0].MountedOn, "/")
	}
}

func TestDFParser_HumanReadable(t *testing.T) {
	// df -h output
	input := `Filesystem      Size  Used Avail Use% Mounted on
/dev/sda1        98G   49G   49G  50% /
tmpfs           7.7G  100M  7.6G   2% /tmp
`
	parser := fileops.NewDFParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.DFOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.DFOutput, got %T", result.Data)
	}

	if len(output.Filesystems) != 2 {
		t.Fatalf("Filesystems len = %d, want 2", len(output.Filesystems))
	}

	// Human-readable should preserve the display strings
	if output.Filesystems[0].SizeHuman != "98G" {
		t.Errorf("Filesystems[0].SizeHuman = %q, want %q", output.Filesystems[0].SizeHuman, "98G")
	}
	if output.Filesystems[0].UsedHuman != "49G" {
		t.Errorf("Filesystems[0].UsedHuman = %q, want %q", output.Filesystems[0].UsedHuman, "49G")
	}
	if output.Filesystems[0].AvailableHuman != "49G" {
		t.Errorf("Filesystems[0].AvailableHuman = %q, want %q", output.Filesystems[0].AvailableHuman, "49G")
	}
}

func TestDFParser_WithFilesystemType(t *testing.T) {
	// df -T output includes filesystem type
	input := `Filesystem     Type     1K-blocks      Used Available Use% Mounted on
/dev/sda1      ext4     102400000  51200000  51200000  50% /
tmpfs          tmpfs      8000000   100000   7900000   2% /tmp
`
	parser := fileops.NewDFParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.DFOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.DFOutput, got %T", result.Data)
	}

	if len(output.Filesystems) != 2 {
		t.Fatalf("Filesystems len = %d, want 2", len(output.Filesystems))
	}

	if output.Filesystems[0].Type != "ext4" {
		t.Errorf("Filesystems[0].Type = %q, want %q", output.Filesystems[0].Type, "ext4")
	}
	if output.Filesystems[1].Type != "tmpfs" {
		t.Errorf("Filesystems[1].Type = %q, want %q", output.Filesystems[1].Type, "tmpfs")
	}
}

func TestDFParser_EmptyInput(t *testing.T) {
	input := ``
	parser := fileops.NewDFParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.DFOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.DFOutput, got %T", result.Data)
	}

	if len(output.Filesystems) != 0 {
		t.Errorf("Filesystems len = %d, want 0", len(output.Filesystems))
	}
}

func TestDFParser_Schema(t *testing.T) {
	parser := fileops.NewDFParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema().ID is empty")
	}
	if schema.Title == "" {
		t.Error("Schema().Title is empty")
	}
}

func TestDFParser_Matches(t *testing.T) {
	parser := fileops.NewDFParser()

	tests := []struct {
		cmd         string
		subcommands []string
		want        bool
	}{
		{"df", []string{}, true},
		{"df", []string{"-h"}, true},
		{"df", []string{"-T"}, true},
		{"df", []string{"-hT", "/"}, true},
		{"du", []string{}, false},
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

func TestDFParser_HeaderOnly(t *testing.T) {
	input := `Filesystem     1K-blocks      Used Available Use% Mounted on
`
	parser := fileops.NewDFParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.DFOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.DFOutput, got %T", result.Data)
	}

	if len(output.Filesystems) != 0 {
		t.Errorf("Filesystems len = %d, want 0", len(output.Filesystems))
	}
}

func TestDFParser_MountWithSpaces(t *testing.T) {
	input := `Filesystem     1K-blocks      Used Available Use% Mounted on
/dev/sda1      102400000  51200000  51200000  50% /mnt/My Drive
`
	parser := fileops.NewDFParser()
	result, err := parser.Parse(strings.NewReader(input))

	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	output, ok := result.Data.(*fileops.DFOutput)
	if !ok {
		t.Fatalf("result.Data is not *fileops.DFOutput, got %T", result.Data)
	}

	if len(output.Filesystems) != 1 {
		t.Fatalf("Filesystems len = %d, want 1", len(output.Filesystems))
	}

	if output.Filesystems[0].MountedOn != "/mnt/My Drive" {
		t.Errorf("Filesystems[0].MountedOn = %q, want %q", output.Filesystems[0].MountedOn, "/mnt/My Drive")
	}
}
