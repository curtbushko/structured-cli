package docker

import (
	"strings"
	"testing"
)

func TestImagesParser_Success(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantData ImagesResult
	}{
		{
			name:  "empty output indicates no images",
			input: "",
			wantData: ImagesResult{
				Success: true,
				Images:  []Image{},
			},
		},
		{
			name: "single image",
			input: `REPOSITORY   TAG       IMAGE ID       CREATED        SIZE
nginx        latest    abc123def456   2 weeks ago    142MB`,
			wantData: ImagesResult{
				Success: true,
				Images: []Image{
					{
						Repository: "nginx",
						Tag:        "latest",
						ID:         "abc123def456",
						Created:    "2 weeks ago",
						Size:       "142MB",
					},
				},
			},
		},
		{
			name: "multiple images",
			input: `REPOSITORY   TAG       IMAGE ID       CREATED        SIZE
nginx        latest    abc123def456   2 weeks ago    142MB
redis        alpine    def789abc012   3 weeks ago    32.4MB
postgres     14        ghi456def789   1 month ago    379MB
<none>       <none>    xyz789abc123   2 months ago   1.2GB`,
			wantData: ImagesResult{
				Success: true,
				Images: []Image{
					{
						Repository: "nginx",
						Tag:        "latest",
						ID:         "abc123def456",
						Created:    "2 weeks ago",
						Size:       "142MB",
					},
					{
						Repository: "redis",
						Tag:        "alpine",
						ID:         "def789abc012",
						Created:    "3 weeks ago",
						Size:       "32.4MB",
					},
					{
						Repository: "postgres",
						Tag:        "14",
						ID:         "ghi456def789",
						Created:    "1 month ago",
						Size:       "379MB",
					},
					{
						Repository: "<none>",
						Tag:        "<none>",
						ID:         "xyz789abc123",
						Created:    "2 months ago",
						Size:       "1.2GB",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewImagesParser()
			result, err := parser.Parse(strings.NewReader(tt.input))
			if err != nil {
				t.Fatalf("Parse() returned error: %v", err)
			}

			if result.Error != nil {
				t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
			}

			got, ok := result.Data.(*ImagesResult)
			if !ok {
				t.Fatalf("ParseResult.Data type = %T, want *ImagesResult", result.Data)
			}

			if got.Success != tt.wantData.Success {
				t.Errorf("ImagesResult.Success = %v, want %v", got.Success, tt.wantData.Success)
			}

			if len(got.Images) != len(tt.wantData.Images) {
				t.Errorf("ImagesResult.Images length = %d, want %d", len(got.Images), len(tt.wantData.Images))
				return
			}

			for i, image := range got.Images {
				want := tt.wantData.Images[i]
				if image.Repository != want.Repository {
					t.Errorf("Image[%d].Repository = %q, want %q", i, image.Repository, want.Repository)
				}
				if image.Tag != want.Tag {
					t.Errorf("Image[%d].Tag = %q, want %q", i, image.Tag, want.Tag)
				}
				if image.ID != want.ID {
					t.Errorf("Image[%d].ID = %q, want %q", i, image.ID, want.ID)
				}
				if image.Size != want.Size {
					t.Errorf("Image[%d].Size = %q, want %q", i, image.Size, want.Size)
				}
			}
		})
	}
}

func TestImagesParser_JSONFormat(t *testing.T) {
	input := `[{"ID":"abc123def456","Repository":"nginx","Tag":"latest","CreatedAt":"2024-01-01 10:30:00 +0000 UTC","Size":"142MB"}]`

	parser := NewImagesParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if result.Error != nil {
		t.Fatalf("ParseResult.Error = %v, want nil", result.Error)
	}

	got, ok := result.Data.(*ImagesResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ImagesResult", result.Data)
	}

	if !got.Success {
		t.Error("ImagesResult.Success = false, want true")
	}

	if len(got.Images) != 1 {
		t.Fatalf("ImagesResult.Images length = %d, want 1", len(got.Images))
	}

	image := got.Images[0]
	if image.Repository != "nginx" {
		t.Errorf("Image.Repository = %q, want %q", image.Repository, "nginx")
	}
	if image.Tag != "latest" {
		t.Errorf("Image.Tag = %q, want %q", image.Tag, "latest")
	}
}

func TestImagesParser_Matches(t *testing.T) {
	tests := []struct {
		name        string
		cmd         string
		subcommands []string
		want        bool
	}{
		{
			name:        "matches docker images",
			cmd:         "docker",
			subcommands: []string{"images"},
			want:        true,
		},
		{
			name:        "matches docker images with flags",
			cmd:         "docker",
			subcommands: []string{"images", "-a"},
			want:        true,
		},
		{
			name:        "matches docker image ls",
			cmd:         "docker",
			subcommands: []string{"image", "ls"},
			want:        true,
		},
		{
			name:        "matches docker image list",
			cmd:         "docker",
			subcommands: []string{"image", "list"},
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
			subcommands: []string{"images"},
			want:        false,
		},
	}

	parser := NewImagesParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parser.Matches(tt.cmd, tt.subcommands)
			if got != tt.want {
				t.Errorf("Matches(%q, %v) = %v, want %v", tt.cmd, tt.subcommands, got, tt.want)
			}
		})
	}
}

func TestImagesParser_Schema(t *testing.T) {
	parser := NewImagesParser()
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

	requiredProps := []string{"success", "images"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}
