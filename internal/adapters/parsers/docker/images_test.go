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

	if schema.Type != schemaTypeObject {
		t.Errorf("Schema.Type = %q, want %q", schema.Type, schemaTypeObject)
	}

	requiredProps := []string{"success", "images"}
	for _, prop := range requiredProps {
		if _, ok := schema.Properties[prop]; !ok {
			t.Errorf("Schema.Properties missing %q", prop)
		}
	}
}

func TestImagesParser_VaryingColumnWidths(t *testing.T) {
	// Real docker images output with varying repository name lengths
	// Column positions are determined by the header
	input := `REPOSITORY                                      TAG       IMAGE ID       CREATED         SIZE
nginx                                           latest    abc123def456   2 weeks ago     142MB
gcr.io/my-project/my-very-long-image-name       v1.2.3    def789abc012   3 months ago    256MB
redis                                           alpine    ghi456def789   About an hour ago   32.4MB`

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

	// Must parse all 3 images - this is the bug: images may be missing
	if len(got.Images) != 3 {
		t.Fatalf("ImagesResult.Images length = %d, want 3 - images are missing!", len(got.Images))
	}

	// Verify each image is correctly parsed
	expected := []Image{
		{Repository: "nginx", Tag: "latest", ID: "abc123def456", Created: "2 weeks ago", Size: "142MB"},
		{Repository: "gcr.io/my-project/my-very-long-image-name", Tag: "v1.2.3", ID: "def789abc012", Created: "3 months ago", Size: "256MB"},
		{Repository: "redis", Tag: "alpine", ID: "ghi456def789", Created: "About an hour ago", Size: "32.4MB"},
	}

	for i, want := range expected {
		if got.Images[i].Repository != want.Repository {
			t.Errorf("Image[%d].Repository = %q, want %q", i, got.Images[i].Repository, want.Repository)
		}
		if got.Images[i].Tag != want.Tag {
			t.Errorf("Image[%d].Tag = %q, want %q", i, got.Images[i].Tag, want.Tag)
		}
		if got.Images[i].ID != want.ID {
			t.Errorf("Image[%d].ID = %q, want %q", i, got.Images[i].ID, want.ID)
		}
		if got.Images[i].Created != want.Created {
			t.Errorf("Image[%d].Created = %q, want %q", i, got.Images[i].Created, want.Created)
		}
		if got.Images[i].Size != want.Size {
			t.Errorf("Image[%d].Size = %q, want %q", i, got.Images[i].Size, want.Size)
		}
	}
}

func TestImagesParser_ShortIDs(t *testing.T) {
	// Docker can show short image IDs (12 chars) or full IDs
	input := `REPOSITORY   TAG       IMAGE ID       CREATED        SIZE
nginx        latest    abc123def456   2 weeks ago    142MB
redis        alpine    def789       3 weeks ago    32.4MB`

	parser := NewImagesParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*ImagesResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ImagesResult", result.Data)
	}

	if len(got.Images) != 2 {
		t.Fatalf("ImagesResult.Images length = %d, want 2", len(got.Images))
	}

	// Second image has a short ID
	if got.Images[1].ID != "def789" {
		t.Errorf("Image[1].ID = %q, want %q", got.Images[1].ID, "def789")
	}
}

func TestImagesParser_LongCreatedDates(t *testing.T) {
	// Test multi-word Created dates like "About an hour ago", "Less than a second ago"
	input := `REPOSITORY   TAG       IMAGE ID       CREATED                   SIZE
nginx        latest    abc123def456   About an hour ago         142MB
redis        alpine    def789abc012   Less than a second ago    32.4MB
postgres     14        ghi456def789   3 months ago              379MB`

	parser := NewImagesParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*ImagesResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ImagesResult", result.Data)
	}

	if len(got.Images) != 3 {
		t.Fatalf("ImagesResult.Images length = %d, want 3", len(got.Images))
	}

	// Verify Created field captures all date components
	expectedCreated := []string{
		"About an hour ago",
		"Less than a second ago",
		"3 months ago",
	}

	for i, want := range expectedCreated {
		if got.Images[i].Created != want {
			t.Errorf("Image[%d].Created = %q, want %q", i, got.Images[i].Created, want)
		}
	}
}

func TestImagesParser_ColumnBasedParsing(t *testing.T) {
	// Test with exact docker images output format using column positions
	// The header determines column positions - some outputs use fixed-width columns
	input := `REPOSITORY                                              TAG       IMAGE ID       CREATED        SIZE
bender                                                  latest    f6036cbdc4ad   2 weeks ago    18.1MB
ghcr.io/curtbushko/minecraft-servers/dj-server          latest    0aea42cd4c12   3 weeks ago    813MB
ghcr.io/curtbushko/minecraft-servers/homestead          latest    a0650662d3d2   4 weeks ago    774MB`

	parser := NewImagesParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*ImagesResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ImagesResult", result.Data)
	}

	// All 3 images must be parsed
	if len(got.Images) != 3 {
		t.Fatalf("ImagesResult.Images length = %d, want 3 - images are missing!", len(got.Images))
	}

	// Verify the long repository names are parsed correctly
	if got.Images[1].Repository != "ghcr.io/curtbushko/minecraft-servers/dj-server" {
		t.Errorf("Image[1].Repository = %q, want %q", got.Images[1].Repository, "ghcr.io/curtbushko/minecraft-servers/dj-server")
	}
	if got.Images[2].Repository != "ghcr.io/curtbushko/minecraft-servers/homestead" {
		t.Errorf("Image[2].Repository = %q, want %q", got.Images[2].Repository, "ghcr.io/curtbushko/minecraft-servers/homestead")
	}
}

func TestImagesParser_MinimalParts(t *testing.T) {
	// Test case where parsing might fail due to insufficient parts
	// This tests when an image line has fewer than expected parts
	input := `REPOSITORY   TAG       IMAGE ID       CREATED        SIZE
nginx        latest    abc123def456   now            142MB`

	parser := NewImagesParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*ImagesResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ImagesResult", result.Data)
	}

	// Image with single-word Created should still be parsed
	if len(got.Images) != 1 {
		t.Fatalf("ImagesResult.Images length = %d, want 1", len(got.Images))
	}

	if got.Images[0].Created != "now" {
		t.Errorf("Image[0].Created = %q, want %q", got.Images[0].Created, "now")
	}
}

func TestImagesParser_FourPartLine(t *testing.T) {
	// Test case where line splits into exactly 4 parts
	// This should NOT result in a missing image (bug case)
	// Parts: [nginx, latest, abc123, 142MB] - missing Created entirely
	// The current code requires 5+ parts, so this would be skipped
	input := `REPOSITORY   TAG       IMAGE ID   SIZE
nginx        latest    abc123     142MB`

	parser := NewImagesParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*ImagesResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ImagesResult", result.Data)
	}

	// Image should still be parsed even without Created column
	if len(got.Images) != 1 {
		t.Fatalf("ImagesResult.Images length = %d, want 1 - image is missing!", len(got.Images))
	}

	if got.Images[0].Repository != "nginx" {
		t.Errorf("Image[0].Repository = %q, want %q", got.Images[0].Repository, "nginx")
	}
}

func TestImagesParser_NoTruncOutput(t *testing.T) {
	// Test docker images --no-trunc output with full image IDs
	input := `REPOSITORY   TAG       IMAGE ID                                                                  CREATED        SIZE
nginx        latest    sha256:abc123def456789012345678901234567890123456789012345678901234567890   2 weeks ago    142MB`

	parser := NewImagesParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*ImagesResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ImagesResult", result.Data)
	}

	if len(got.Images) != 1 {
		t.Fatalf("ImagesResult.Images length = %d, want 1", len(got.Images))
	}

	// Full image ID should be captured
	expectedID := "sha256:abc123def456789012345678901234567890123456789012345678901234567890"
	if got.Images[0].ID != expectedID {
		t.Errorf("Image[0].ID = %q, want %q", got.Images[0].ID, expectedID)
	}
}

func TestImagesParser_DigestOutput(t *testing.T) {
	// Test docker images --digests output which includes DIGEST column
	input := `REPOSITORY   TAG       DIGEST                                                                    IMAGE ID       CREATED        SIZE
nginx        latest    sha256:abc123def456789012345678901234567890123456789012345678901234567890   def456abc123   2 weeks ago    142MB`

	parser := NewImagesParser()
	result, err := parser.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	got, ok := result.Data.(*ImagesResult)
	if !ok {
		t.Fatalf("ParseResult.Data type = %T, want *ImagesResult", result.Data)
	}

	// Image should be parsed even with extra DIGEST column
	if len(got.Images) != 1 {
		t.Fatalf("ImagesResult.Images length = %d, want 1 - image is missing!", len(got.Images))
	}
}
