package docker

import (
	"strings"
	"testing"
)

func TestComposeUpParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData ComposeUpResult
	}{
		{
			name:  "empty output indicates success",
			input: "",
			wantData: ComposeUpResult{
				Success:  true,
				Services: []ComposeService{},
				Networks: []string{},
				Volumes:  []string{},
				Errors:   []string{},
				Warnings: []string{},
			},
		},
		{
			name: "successful compose up",
			input: `[+] Running 3/3
 ✔ Network myapp_default    Created
 ✔ Container myapp-db-1     Started
 ✔ Container myapp-web-1    Started`,
			wantData: ComposeUpResult{
				Success: true,
				Services: []ComposeService{
					{Name: "db", Status: "Started"},
					{Name: "web", Status: "Started"},
				},
				Networks: []string{"myapp_default"},
				Volumes:  []string{},
				Errors:   []string{},
				Warnings: []string{},
			},
		},
		{
			name: "compose up with volumes",
			input: `[+] Running 4/4
 ✔ Network myapp_default     Created
 ✔ Volume "myapp_data"       Created
 ✔ Container myapp-db-1      Started
 ✔ Container myapp-web-1     Started`,
			wantData: ComposeUpResult{
				Success: true,
				Services: []ComposeService{
					{Name: "db", Status: "Started"},
					{Name: "web", Status: "Started"},
				},
				Networks: []string{"myapp_default"},
				Volumes:  []string{"myapp_data"},
				Errors:   []string{},
				Warnings: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewComposeUpParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*ComposeUpResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *ComposeUpResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("ComposeUpResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if len(got.Services) != len(tt.wantData.Services) {
				t.Errorf("ComposeUpResult.Services length = %d, want %d", len(got.Services), len(tt.wantData.Services))
			}

			if len(got.Networks) != len(tt.wantData.Networks) {
				t.Errorf("ComposeUpResult.Networks length = %d, want %d", len(got.Networks), len(tt.wantData.Networks))
			}

			if len(got.Volumes) != len(tt.wantData.Volumes) {
				t.Errorf("ComposeUpResult.Volumes length = %d, want %d", len(got.Volumes), len(tt.wantData.Volumes))
			}
		})
	}
}

func TestComposeUpParser_WithErrors(t *testing.T) {
	input := `[+] Running 0/1
 ⠿ web Error
Error response from daemon: pull access denied for myapp, repository does not exist`

	parser := NewComposeUpParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*ComposeUpResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ComposeUpResult", result.Data)
	}

	if got.Success {
		t.Error("ComposeUpResult.Success = true, want false for compose error")
	}

	if len(got.Errors) == 0 {
		t.Error("ComposeUpResult.Errors should not be empty")
	}
}

func TestComposeUpParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches docker compose up",
			cmd:         "docker",
			subcommands: []string{"compose", "up"},
			want:        true,
		},
		{
			name:        "matches docker compose up with flags",
			cmd:         "docker",
			subcommands: []string{"compose", "up", "-d"},
			want:        true,
		},
		{
			name:        "matches docker-compose up",
			cmd:         "docker-compose",
			subcommands: []string{"up"},
			want:        true,
		},
		{
			name:        "does not match docker compose down",
			cmd:         "docker",
			subcommands: []string{"compose", "down"},
			want:        false,
		},
		{
			name:        "does not match docker run",
			cmd:         "docker",
			subcommands: []string{"run"},
			want:        false,
		},
	}

	parser := NewComposeUpParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestComposeUpParser_Schema(t *testing.T) {
	parser := NewComposeUpParser()
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
