package docker

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// dockerImagesJSONEntry represents an image in JSON format from docker images --format json.
type dockerImagesJSONEntry struct {
	ID         string `json:"ID"`
	Repository string `json:"Repository"`
	Tag        string `json:"Tag"`
	CreatedAt  string `json:"CreatedAt"`
	Size       string `json:"Size"`
	Digest     string `json:"Digest,omitempty"`
}

// ImagesParser parses the output of 'docker images'.
type ImagesParser struct {
	schema domain.Schema
}

// NewImagesParser creates a new ImagesParser with the docker-images schema.
func NewImagesParser() *ImagesParser {
	return &ImagesParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/docker-images.json",
			"Docker Images Output",
			"object",
			map[string]domain.PropertySchema{
				"success": {Type: "boolean", Description: "Whether the command completed successfully"},
				"images":  {Type: "array", Description: "List of images"},
			},
			[]string{"success", "images"},
		),
	}
}

// Parse reads docker images output and returns structured data.
func (p *ImagesParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &ImagesResult{
		Success: true,
		Images:  []Image{},
	}

	// Try JSON format first
	if strings.HasPrefix(strings.TrimSpace(raw), "[") {
		parseJSONImagesOutput(raw, result)
	} else {
		parseTableImagesOutput(raw, result)
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// parseJSONImagesOutput parses JSON format output from docker images --format json.
func parseJSONImagesOutput(output string, result *ImagesResult) {
	var entries []dockerImagesJSONEntry
	if err := json.Unmarshal([]byte(output), &entries); err != nil {
		result.Success = false
		return
	}

	for _, entry := range entries {
		image := Image{
			ID:         entry.ID,
			Repository: entry.Repository,
			Tag:        entry.Tag,
			Created:    entry.CreatedAt,
			Size:       entry.Size,
			Digest:     entry.Digest,
		}
		result.Images = append(result.Images, image)
	}
}

// parseTableImagesOutput parses table format output from docker images.
func parseTableImagesOutput(output string, result *ImagesResult) {
	scanner := bufio.NewScanner(strings.NewReader(output))

	// Skip header line
	if !scanner.Scan() {
		return
	}

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		image := parseImageLine(line)
		if image.ID != "" {
			result.Images = append(result.Images, image)
		}
	}
}

// parseImageLine parses a single line from docker images table output.
func parseImageLine(line string) Image {
	// Docker images output is column-based
	// REPOSITORY   TAG       IMAGE ID       CREATED        SIZE

	parts := splitBySpaces(line)
	if len(parts) < 5 {
		return Image{}
	}

	image := Image{
		Repository: parts[0],
		Tag:        parts[1],
		ID:         parts[2],
	}

	// Created and Size are at the end - need to handle "X time ago" format
	// Work backwards from the end
	image.Size = parts[len(parts)-1]

	// Created is everything between ID and Size
	if len(parts) >= 5 {
		createdParts := parts[3 : len(parts)-1]
		image.Created = strings.Join(createdParts, " ")
	}

	return image
}

// Schema returns the JSON Schema for docker images output.
func (p *ImagesParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *ImagesParser) Matches(cmd string, subcommands []string) bool {
	if cmd != dockerCommand {
		return false
	}

	if len(subcommands) == 0 {
		return false
	}

	// docker images
	if subcommands[0] == subImages {
		return true
	}

	// docker image ls/list
	if len(subcommands) >= 2 && subcommands[0] == subcommandImage {
		switch subcommands[1] {
		case subLs, subList:
			return true
		}
	}

	return false
}
