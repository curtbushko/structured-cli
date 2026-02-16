package docker

import (
	"strings"
	"testing"
)

func TestRunParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData RunResult
	}{
		{
			name:  "detached run returns container ID",
			input: "abc123def4567890123456789012345678901234567890123456789012345678\n",
			wantData: RunResult{
				Success:     true,
				ContainerID: "abc123def4567890123456789012345678901234567890123456789012345678",
				Detached:    true,
				Errors:      []string{},
			},
		},
		{
			name: "interactive run with output",
			input: `Hello, World!
This is some output.
Process completed.`,
			wantData: RunResult{
				Success:  true,
				Detached: false,
				Output:   "Hello, World!\nThis is some output.\nProcess completed.",
				Errors:   []string{},
			},
		},
		{
			name:  "empty output indicates success",
			input: "",
			wantData: RunResult{
				Success:  true,
				Detached: false,
				Errors:   []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewRunParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*RunResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *RunResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("RunResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if got.Detached != tt.wantData.Detached {
				t.Errorf("RunResult.Detached = %v, want %v", got.Detached, tt.wantData.Detached)
			}

			if tt.wantData.ContainerID != "" && got.ContainerID != tt.wantData.ContainerID {
				t.Errorf("RunResult.ContainerID = %q, want %q", got.ContainerID, tt.wantData.ContainerID)
			}
		})
	}
}

func TestRunParser_WithErrors(t *testing.T) {
	input := `docker: Error response from daemon: pull access denied for nonexistent, repository does not exist or may require 'docker login'.
See 'docker run --help'.`

	parser := NewRunParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*RunResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *RunResult", result.Data)
	}

	if got.Success {
		t.Error("RunResult.Success = true, want false for error output")
	}

	if len(got.Errors) == 0 {
		t.Error("RunResult.Errors should not be empty")
	}
}

func TestRunParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches docker run",
			cmd:         "docker",
			subcommands: []string{"run"},
			want:        true,
		},
		{
			name:        "matches docker run with image",
			cmd:         "docker",
			subcommands: []string{"run", "nginx"},
			want:        true,
		},
		{
			name:        "matches docker run with flags",
			cmd:         "docker",
			subcommands: []string{"run", "-d", "-p", "80:80", "nginx"},
			want:        true,
		},
		{
			name:        "matches docker container run",
			cmd:         "docker",
			subcommands: []string{"container", "run"},
			want:        true,
		},
		{
			name:        "does not match docker ps",
			cmd:         "docker",
			subcommands: []string{"ps"},
			want:        false,
		},
		{
			name:        "does not match podman",
			cmd:         "podman",
			subcommands: []string{"run"},
			want:        false,
		},
	}

	parser := NewRunParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestRunParser_Schema(t *testing.T) {
	parser := NewRunParser()
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

	requiredProps := []string{"success", "container_id"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}
