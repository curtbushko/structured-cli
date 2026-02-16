package docker

import (
	"strings"
	"testing"
)

func TestExecParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData ExecResult
	}{
		{
			name:  "empty output indicates success",
			input: "",
			wantData: ExecResult{
				Success: true,
				Output:  "",
				Errors:  []string{},
			},
		},
		{
			name: "exec with output",
			input: `root
total 64
drwxr-xr-x   1 root root 4096 Jan 15 10:30 app
drwxr-xr-x   2 root root 4096 Jan 10 00:00 bin`,
			wantData: ExecResult{
				Success: true,
				Output:  "root\ntotal 64\ndrwxr-xr-x   1 root root 4096 Jan 15 10:30 app\ndrwxr-xr-x   2 root root 4096 Jan 10 00:00 bin",
				Errors:  []string{},
			},
		},
		{
			name:  "exec command output",
			input: "Hello from container\n",
			wantData: ExecResult{
				Success: true,
				Output:  "Hello from container",
				Errors:  []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewExecParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*ExecResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *ExecResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("ExecResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if got.Output != tt.wantData.Output {
				t.Errorf("ExecResult.Output = %q, want %q", got.Output, tt.wantData.Output)
			}
		})
	}
}

func TestExecParser_WithErrors(t *testing.T) {
	input := `Error response from daemon: Container abc123def456 is not running`

	parser := NewExecParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*ExecResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ExecResult", result.Data)
	}

	if got.Success {
		t.Error("ExecResult.Success = true, want false for error output")
	}

	if len(got.Errors) == 0 {
		t.Error("ExecResult.Errors should not be empty")
	}
}

func TestExecParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches docker exec",
			cmd:         "docker",
			subcommands: []string{"exec"},
			want:        true,
		},
		{
			name:        "matches docker exec with container and command",
			cmd:         "docker",
			subcommands: []string{"exec", "mycontainer", "ls", "-la"},
			want:        true,
		},
		{
			name:        "matches docker exec with flags",
			cmd:         "docker",
			subcommands: []string{"exec", "-it", "mycontainer", "/bin/bash"},
			want:        true,
		},
		{
			name:        "matches docker container exec",
			cmd:         "docker",
			subcommands: []string{"container", "exec"},
			want:        true,
		},
		{
			name:        "does not match docker run",
			cmd:         "docker",
			subcommands: []string{"run"},
			want:        false,
		},
		{
			name:        "does not match podman",
			cmd:         "podman",
			subcommands: []string{"exec"},
			want:        false,
		},
	}

	parser := NewExecParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestExecParser_Schema(t *testing.T) {
	parser := NewExecParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema.ID should not be empty")
	}

	if schema.Title == "" {
		t.Error("Schema.Title should not be empty")
	}

	if schema.Type != "object" {
		t.Errorf("Schema.Type = %q, want %q", schema.Type, "object")
	}

	requiredProps := []string{"success", "output"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}
