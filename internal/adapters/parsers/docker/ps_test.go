package docker

import (
	"strings"
	"testing"
)

const stateRunning = "running"

func TestPSParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData PSResult
	}{
		{
			name:  "empty output indicates no containers",
			input: "",
			wantData: PSResult{
				Success:    true,
				Containers: []Container{},
			},
		},
		{
			name: "single running container",
			input: `CONTAINER ID   IMAGE         COMMAND       CREATED         STATUS         PORTS     NAMES
abc123def456   nginx:latest  "nginx -g"    5 minutes ago   Up 5 minutes   80/tcp    web-server`,
			wantData: PSResult{
				Success: true,
				Containers: []Container{
					{
						ID:      "abc123def456",
						Image:   "nginx:latest",
						Command: "nginx -g",
						Created: "5 minutes ago",
						Status:  "Up 5 minutes",
						Ports:   "80/tcp",
						Names:   "web-server",
						State:   "running",
					},
				},
			},
		},
		{
			name: "multiple containers with various states",
			input: `CONTAINER ID   IMAGE           COMMAND          CREATED          STATUS                     PORTS                    NAMES
abc123def456   nginx:latest    "nginx -g"       5 minutes ago    Up 5 minutes               80/tcp                   web-server
def789abc012   redis:alpine    "redis-server"   10 minutes ago   Up 10 minutes              6379/tcp                 redis-cache
ghi456def789   postgres:14     "postgres"       1 hour ago       Exited (0) 30 minutes ago                           db-server`,
			wantData: PSResult{
				Success: true,
				Containers: []Container{
					{
						ID:      "abc123def456",
						Image:   "nginx:latest",
						Command: "nginx -g",
						Created: "5 minutes ago",
						Status:  "Up 5 minutes",
						Ports:   "80/tcp",
						Names:   "web-server",
						State:   "running",
					},
					{
						ID:      "def789abc012",
						Image:   "redis:alpine",
						Command: "redis-server",
						Created: "10 minutes ago",
						Status:  "Up 10 minutes",
						Ports:   "6379/tcp",
						Names:   "redis-cache",
						State:   "running",
					},
					{
						ID:      "ghi456def789",
						Image:   "postgres:14",
						Command: "postgres",
						Created: "1 hour ago",
						Status:  "Exited (0) 30 minutes ago",
						Ports:   "",
						Names:   "db-server",
						State:   "exited",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewPSParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*PSResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *PSResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("PSResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if len(got.Containers) != len(tt.wantData.Containers) {
				t.Errorf("PSResult.Containers length = %d, want %d", len(got.Containers), len(tt.wantData.Containers))
				return
			}

			for i, container := range got.Containers {
				want := tt.wantData.Containers[i]
				if container.ID != want.ID {
					t.Errorf("Container[%d].ID = %q, want %q", i, container.ID, want.ID)
				}
				if container.Image != want.Image {
					t.Errorf("Container[%d].Image = %q, want %q", i, container.Image, want.Image)
				}
				if container.Names != want.Names {
					t.Errorf("Container[%d].Names = %q, want %q", i, container.Names, want.Names)
				}
				if container.State != want.State {
					t.Errorf("Container[%d].State = %q, want %q", i, container.State, want.State)
				}
			}
		})
	}
}

func TestPSParser_JSONFormat(t *testing.T) {
	input := `[{"ID":"abc123def456","Image":"nginx:latest","Command":"nginx -g 'daemon off;'","CreatedAt":"2024-01-15 10:30:00 +0000 UTC","Status":"Up 5 minutes","Ports":"0.0.0.0:80->80/tcp","Names":"web-server","State":"running"}]`

	parser := NewPSParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*PSResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *PSResult", result.Data)
	}

	if !got.Success {
		t.Error("PSResult.Success = false, want true")
	}

	if len(got.Containers) != 1 {
		t.Fatalf("PSResult.Containers length = %d, want 1", len(got.Containers))
	}

	container := got.Containers[0]
	if container.ID != "abc123def456" {
		t.Errorf("Container.ID = %q, want %q", container.ID, "abc123def456")
	}
	if container.Image != "nginx:latest" {
		t.Errorf("Container.Image = %q, want %q", container.Image, "nginx:latest")
	}
	if container.State != stateRunning {
		t.Errorf("Container.State = %q, want %q", container.State, stateRunning)
	}
}

func TestPSParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches docker ps",
			cmd:         "docker",
			subcommands: []string{"ps"},
			want:        true,
		},
		{
			name:        "matches docker ps with flags",
			cmd:         "docker",
			subcommands: []string{"ps", "-a"},
			want:        true,
		},
		{
			name:        "matches docker container ls",
			cmd:         "docker",
			subcommands: []string{"container", "ls"},
			want:        true,
		},
		{
			name:        "matches docker container list",
			cmd:         "docker",
			subcommands: []string{"container", "list"},
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
			subcommands: []string{"ps"},
			want:        false,
		},
		{
			name:        "does not match empty",
			cmd:         "",
			subcommands: []string{},
			want:        false,
		},
	}

	parser := NewPSParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestPSParser_Schema(t *testing.T) {
	parser := NewPSParser()
	schema := parser.Schema()

	if schema.ID == "" {
		t.Error("Schema.ID should not be empty")
	}

	if schema.Title == "" {
		t.Error("Schema.Title should not be empty")
	}

	if schema.Type != schemaTypeObject {
		t.Errorf("Schema.Type = %q, want %q", schema.Type, schemaTypeObject)
	}

	requiredProps := []string{"success", "containers"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}
