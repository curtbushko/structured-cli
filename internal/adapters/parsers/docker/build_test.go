package docker

import (
	"strings"
	"testing"
)

func TestBuildParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData BuildResult
	}{
		{
			name:  "empty output indicates build issue",
			input: "",
			wantData: BuildResult{
				Success:  true,
				Tags:     []string{},
				Steps:    []BuildStep{},
				Errors:   []string{},
				Warnings: []string{},
			},
		},
		{
			name: "successful build with steps",
			input: `#1 [internal] load build definition from Dockerfile
#1 DONE 0.0s

#2 [internal] load .dockerignore
#2 DONE 0.0s

#3 [1/3] FROM docker.io/library/golang:1.21
#3 CACHED

#4 [2/3] COPY go.mod go.sum ./
#4 DONE 0.1s

#5 [3/3] RUN go build -o app
#5 DONE 5.2s

#6 exporting to image
#6 exporting layers done
#6 writing image sha256:abc123def456789 done
#6 naming to docker.io/library/myapp:latest done
#6 DONE 0.1s`,
			wantData: BuildResult{
				Success:    true,
				ImageID:    "sha256:abc123def456789",
				Tags:       []string{"docker.io/library/myapp:latest"},
				TotalSteps: 6,
				Cached:     1,
				Steps: []BuildStep{
					{Number: 1, Instruction: "[internal] load build definition from Dockerfile", Cached: false},
					{Number: 2, Instruction: "[internal] load .dockerignore", Cached: false},
					{Number: 3, Instruction: "[1/3] FROM docker.io/library/golang:1.21", Cached: true},
					{Number: 4, Instruction: "[2/3] COPY go.mod go.sum ./", Cached: false},
					{Number: 5, Instruction: "[3/3] RUN go build -o app", Cached: false},
					{Number: 6, Instruction: "exporting to image", Cached: false},
				},
				Errors:   []string{},
				Warnings: []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewBuildParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*BuildResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *BuildResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("BuildResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if got.ImageID != tt.wantData.ImageID {
				t.Errorf("BuildResult.ImageID = %q, want %q", got.ImageID, tt.wantData.ImageID)
			}

			if got.TotalSteps != tt.wantData.TotalSteps {
				t.Errorf("BuildResult.TotalSteps = %d, want %d", got.TotalSteps, tt.wantData.TotalSteps)
			}

			if got.Cached != tt.wantData.Cached {
				t.Errorf("BuildResult.Cached = %d, want %d", got.Cached, tt.wantData.Cached)
			}

			if len(got.Tags) != len(tt.wantData.Tags) {
				t.Errorf("BuildResult.Tags length = %d, want %d", len(got.Tags), len(tt.wantData.Tags))
			}
		})
	}
}

func TestBuildParser_WithErrors(t *testing.T) {
	input := `#1 [internal] load build definition from Dockerfile
#1 DONE 0.0s

#2 [1/2] FROM docker.io/library/golang:1.21
#2 CACHED

#3 [2/2] RUN go build -o app
#3 ERROR: process "go build -o app" did not complete successfully: exit code: 1

------
 > [2/2] RUN go build -o app:
------
Dockerfile:5
--------------------
   3 |     WORKDIR /app
   4 |     COPY . .
   5 | >>> RUN go build -o app
--------------------
error: failed to build: exit code: 1`

	parser := NewBuildParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*BuildResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *BuildResult", result.Data)
	}

	if got.Success {
		t.Error("BuildResult.Success = true, want false for build with errors")
	}

	if len(got.Errors) == 0 {
		t.Error("BuildResult.Errors should not be empty")
	}
}

func TestBuildParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches docker build",
			cmd:         "docker",
			subcommands: []string{"build"},
			want:        true,
		},
		{
			name:        "matches docker build with path",
			cmd:         "docker",
			subcommands: []string{"build", "."},
			want:        true,
		},
		{
			name:        "matches docker build with tag",
			cmd:         "docker",
			subcommands: []string{"build", "-t", "myapp:latest", "."},
			want:        true,
		},
		{
			name:        "matches docker image build",
			cmd:         "docker",
			subcommands: []string{"image", "build"},
			want:        true,
		},
		{
			name:        "matches docker buildx build",
			cmd:         "docker",
			subcommands: []string{"buildx", "build"},
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
			subcommands: []string{"build"},
			want:        false,
		},
	}

	parser := NewBuildParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestBuildParser_Schema(t *testing.T) {
	parser := NewBuildParser()
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

	requiredProps := []string{"success", "image_id", "steps"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}
