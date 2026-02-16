package docker

import (
	"strings"
	"testing"
)

func TestComposeDownParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData ComposeDownResult
	}{
		{
			name:  "empty output indicates success",
			input: "",
			wantData: ComposeDownResult{
				Success:           true,
				StoppedContainers: []string{},
				RemovedContainers: []string{},
				RemovedNetworks:   []string{},
				RemovedVolumes:    []string{},
				Errors:            []string{},
			},
		},
		{
			name: "successful compose down",
			input: `[+] Running 3/3
 ✔ Container myapp-web-1    Stopped
 ✔ Container myapp-db-1     Stopped
 ✔ Network myapp_default    Removed`,
			wantData: ComposeDownResult{
				Success:           true,
				StoppedContainers: []string{"myapp-web-1", "myapp-db-1"},
				RemovedContainers: []string{},
				RemovedNetworks:   []string{"myapp_default"},
				RemovedVolumes:    []string{},
				Errors:            []string{},
			},
		},
		{
			name: "compose down with volumes",
			input: `[+] Running 4/4
 ✔ Container myapp-web-1     Removed
 ✔ Container myapp-db-1      Removed
 ✔ Volume "myapp_data"       Removed
 ✔ Network myapp_default     Removed`,
			wantData: ComposeDownResult{
				Success:           true,
				StoppedContainers: []string{},
				RemovedContainers: []string{"myapp-web-1", "myapp-db-1"},
				RemovedNetworks:   []string{"myapp_default"},
				RemovedVolumes:    []string{"myapp_data"},
				Errors:            []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewComposeDownParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*ComposeDownResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *ComposeDownResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("ComposeDownResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if len(got.StoppedContainers) != len(tt.wantData.StoppedContainers) {
				t.Errorf("ComposeDownResult.StoppedContainers length = %d, want %d", len(got.StoppedContainers), len(tt.wantData.StoppedContainers))
			}

			if len(got.RemovedContainers) != len(tt.wantData.RemovedContainers) {
				t.Errorf("ComposeDownResult.RemovedContainers length = %d, want %d", len(got.RemovedContainers), len(tt.wantData.RemovedContainers))
			}

			if len(got.RemovedNetworks) != len(tt.wantData.RemovedNetworks) {
				t.Errorf("ComposeDownResult.RemovedNetworks length = %d, want %d", len(got.RemovedNetworks), len(tt.wantData.RemovedNetworks))
			}

			if len(got.RemovedVolumes) != len(tt.wantData.RemovedVolumes) {
				t.Errorf("ComposeDownResult.RemovedVolumes length = %d, want %d", len(got.RemovedVolumes), len(tt.wantData.RemovedVolumes))
			}
		})
	}
}

func TestComposeDownParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches docker compose down",
			cmd:         "docker",
			subcommands: []string{"compose", "down"},
			want:        true,
		},
		{
			name:        "matches docker compose down with flags",
			cmd:         "docker",
			subcommands: []string{"compose", "down", "-v"},
			want:        true,
		},
		{
			name:        "matches docker-compose down",
			cmd:         "docker-compose",
			subcommands: []string{"down"},
			want:        true,
		},
		{
			name:        "does not match docker compose up",
			cmd:         "docker",
			subcommands: []string{"compose", "up"},
			want:        false,
		},
		{
			name:        "does not match docker run",
			cmd:         "docker",
			subcommands: []string{"run"},
			want:        false,
		},
	}

	parser := NewComposeDownParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestComposeDownParser_Schema(t *testing.T) {
	parser := NewComposeDownParser()
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

	requiredProps := []string{"success", "stopped_containers", "removed_containers"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}
