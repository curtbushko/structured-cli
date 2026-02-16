package docker

import (
	"strings"
	"testing"
)

func TestComposePSParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData ComposePSResult
	}{
		{
			name:  "empty output indicates no services",
			input: "",
			wantData: ComposePSResult{
				Success:  true,
				Services: []ComposeService{},
			},
		},
		{
			name: "single running service",
			input: `NAME                IMAGE          COMMAND                  SERVICE     CREATED         STATUS          PORTS
myapp-web-1         nginx:latest   "nginx -g 'daemon of…"   web         5 minutes ago   Up 5 minutes    0.0.0.0:80->80/tcp`,
			wantData: ComposePSResult{
				Success: true,
				Services: []ComposeService{
					{
						Name:   "web",
						Image:  "nginx:latest",
						Status: "Up 5 minutes",
						State:  "running",
						Ports:  "0.0.0.0:80->80/tcp",
					},
				},
			},
		},
		{
			name: "multiple services",
			input: `NAME                IMAGE            COMMAND                  SERVICE     CREATED          STATUS          PORTS
myapp-web-1         nginx:latest     "nginx -g 'daemon of…"   web         10 minutes ago   Up 10 minutes   0.0.0.0:80->80/tcp
myapp-db-1          postgres:14      "docker-entrypoint.s…"   db          10 minutes ago   Up 10 minutes   5432/tcp
myapp-redis-1       redis:alpine     "docker-entrypoint.s…"   redis       10 minutes ago   Up 10 minutes   6379/tcp`,
			wantData: ComposePSResult{
				Success: true,
				Services: []ComposeService{
					{
						Name:   "web",
						Image:  "nginx:latest",
						Status: "Up 10 minutes",
						State:  "running",
						Ports:  "0.0.0.0:80->80/tcp",
					},
					{
						Name:   "db",
						Image:  "postgres:14",
						Status: "Up 10 minutes",
						State:  "running",
						Ports:  "5432/tcp",
					},
					{
						Name:   "redis",
						Image:  "redis:alpine",
						Status: "Up 10 minutes",
						State:  "running",
						Ports:  "6379/tcp",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewComposePSParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*ComposePSResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *ComposePSResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("ComposePSResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if len(got.Services) != len(tt.wantData.Services) {
				t.Errorf("ComposePSResult.Services length = %d, want %d", len(got.Services), len(tt.wantData.Services))
				return
			}

			for i, service := range got.Services {
				want := tt.wantData.Services[i]
				if service.Name != want.Name {
					t.Errorf("Service[%d].Name = %q, want %q", i, service.Name, want.Name)
				}
				if service.Image != want.Image {
					t.Errorf("Service[%d].Image = %q, want %q", i, service.Image, want.Image)
				}
				if service.State != want.State {
					t.Errorf("Service[%d].State = %q, want %q", i, service.State, want.State)
				}
			}
		})
	}
}

func TestComposePSParser_JSONFormat(t *testing.T) {
	input := `[{"Name":"myapp-web-1","Image":"nginx:latest","Service":"web","Status":"Up 5 minutes","State":"running","Ports":"0.0.0.0:80->80/tcp"}]`

	parser := NewComposePSParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*ComposePSResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ComposePSResult", result.Data)
	}

	if !got.Success {
		t.Error("ComposePSResult.Success = false, want true")
	}

	if len(got.Services) != 1 {
		t.Fatalf("ComposePSResult.Services length = %d, want 1", len(got.Services))
	}

	service := got.Services[0]
	if service.Name != "web" {
		t.Errorf("Service.Name = %q, want %q", service.Name, "web")
	}
	if service.Image != "nginx:latest" {
		t.Errorf("Service.Image = %q, want %q", service.Image, "nginx:latest")
	}
	if service.State != "running" {
		t.Errorf("Service.State = %q, want %q", service.State, "running")
	}
}

func TestComposePSParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches docker compose ps",
			cmd:         "docker",
			subcommands: []string{"compose", "ps"},
			want:        true,
		},
		{
			name:        "matches docker compose ps with flags",
			cmd:         "docker",
			subcommands: []string{"compose", "ps", "-a"},
			want:        true,
		},
		{
			name:        "matches docker-compose ps",
			cmd:         "docker-compose",
			subcommands: []string{"ps"},
			want:        true,
		},
		{
			name:        "does not match docker compose up",
			cmd:         "docker",
			subcommands: []string{"compose", "up"},
			want:        false,
		},
		{
			name:        "does not match docker ps",
			cmd:         "docker",
			subcommands: []string{"ps"},
			want:        false,
		},
	}

	parser := NewComposePSParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestComposePSParser_Schema(t *testing.T) {
	parser := NewComposePSParser()
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

	requiredProps := []string{"success", "services"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}
