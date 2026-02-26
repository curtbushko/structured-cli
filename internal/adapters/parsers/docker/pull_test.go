package docker

import (
	"strings"
	"testing"
)

func TestPullParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData PullResult
	}{
		{
			name: "successful pull",
			input: `Using default tag: latest
latest: Pulling from library/nginx
a2abf6c4d29d: Pull complete
a9edb18cadd1: Pull complete
589b7251471a: Pull complete
186b1aaa4aa6: Pull complete
b4df32aa5a72: Pull complete
a0bcbecc962e: Pull complete
Digest: sha256:0d17b565c37bcbd895e9d92315a05c1c3c9a29f762b011a10c54a66cd53c9b31
Status: Downloaded newer image for nginx:latest
docker.io/library/nginx:latest`,
			wantData: PullResult{
				Success: true,
				Image:   "docker.io/library/nginx:latest",
				Digest:  "sha256:0d17b565c37bcbd895e9d92315a05c1c3c9a29f762b011a10c54a66cd53c9b31",
				Status:  "Downloaded newer image for nginx:latest",
				Layers: []LayerStatus{
					{ID: "a2abf6c4d29d", Status: "Pull complete"},
					{ID: "a9edb18cadd1", Status: "Pull complete"},
					{ID: "589b7251471a", Status: "Pull complete"},
					{ID: "186b1aaa4aa6", Status: "Pull complete"},
					{ID: "b4df32aa5a72", Status: "Pull complete"},
					{ID: "a0bcbecc962e", Status: "Pull complete"},
				},
				Errors: []string{},
			},
		},
		{
			name: "image already up to date",
			input: `Using default tag: latest
latest: Pulling from library/nginx
Digest: sha256:0d17b565c37bcbd895e9d92315a05c1c3c9a29f762b011a10c54a66cd53c9b31
Status: Image is up to date for nginx:latest
docker.io/library/nginx:latest`,
			wantData: PullResult{
				Success: true,
				Image:   "docker.io/library/nginx:latest",
				Digest:  "sha256:0d17b565c37bcbd895e9d92315a05c1c3c9a29f762b011a10c54a66cd53c9b31",
				Status:  "Image is up to date for nginx:latest",
				Layers:  []LayerStatus{},
				Errors:  []string{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewPullParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*PullResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *PullResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("PullResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if got.Digest != tt.wantData.Digest {
				t.Errorf("PullResult.Digest = %q, want %q", got.Digest, tt.wantData.Digest)
			}

			if got.Status != tt.wantData.Status {
				t.Errorf("PullResult.Status = %q, want %q", got.Status, tt.wantData.Status)
			}

			if len(got.Layers) != len(tt.wantData.Layers) {
				t.Errorf("PullResult.Layers length = %d, want %d", len(got.Layers), len(tt.wantData.Layers))
			}
		})
	}
}

func TestPullParser_WithErrors(t *testing.T) {
	input := `Using default tag: latest
Error response from daemon: pull access denied for nonexistent, repository does not exist or may require 'docker login': denied: requested access to the resource is denied`

	parser := NewPullParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*PullResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *PullResult", result.Data)
	}

	if got.Success {
		t.Error("PullResult.Success = true, want false for pull failure")
	}

	if len(got.Errors) == 0 {
		t.Error("PullResult.Errors should not be empty")
	}
}

func TestPullParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches docker pull",
			cmd:         "docker",
			subcommands: []string{"pull"},
			want:        true,
		},
		{
			name:        "matches docker pull with image",
			cmd:         "docker",
			subcommands: []string{"pull", "nginx:latest"},
			want:        true,
		},
		{
			name:        "matches docker image pull",
			cmd:         "docker",
			subcommands: []string{"image", "pull"},
			want:        true,
		},
		{
			name:        "does not match docker push",
			cmd:         "docker",
			subcommands: []string{"push"},
			want:        false,
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
			subcommands: []string{"pull"},
			want:        false,
		},
	}

	parser := NewPullParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestPullParser_Schema(t *testing.T) {
	parser := NewPullParser()
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

	requiredProps := []string{"success", "image", "status"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}
