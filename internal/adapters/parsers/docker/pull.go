package docker

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Regex patterns for parsing docker pull output.
var (
	// layerPattern matches layer status lines like "a2abf6c4d29d: Pull complete".
	layerPattern = regexp.MustCompile(`^([a-f0-9]+):\s+(.+)$`)

	// digestPattern matches "Digest: sha256:xxx".
	digestPattern = regexp.MustCompile(`^Digest:\s+(sha256:[a-f0-9]+)`)

	// statusPattern matches "Status: xxx".
	statusPattern = regexp.MustCompile(`^Status:\s+(.+)$`)

	// pullErrorPattern matches pull error messages.
	pullErrorPattern = regexp.MustCompile(`(?i)Error response from daemon|error pulling image`)
)

// PullParser parses the output of 'docker pull'.
type PullParser struct {
	schema domain.Schema
}

// NewPullParser creates a new PullParser with the docker-pull schema.
func NewPullParser() *PullParser {
	return &PullParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/docker-pull.json",
			"Docker Pull Output",
			"object",
			map[string]domain.PropertySchema{
				"success": {Type: "boolean", Description: "Whether the pull completed successfully"},
				"image":   {Type: "string", Description: "Pulled image reference"},
				"digest":  {Type: "string", Description: "Image digest"},
				"status":  {Type: "string", Description: "Pull status"},
				"layers":  {Type: "array", Description: "Layer statuses"},
				"errors":  {Type: "array", Description: "Error messages"},
			},
			[]string{"success", "image", "status"},
		),
	}
}

// Parse reads docker pull output and returns structured data.
func (p *PullParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &PullResult{
		Success: true,
		Layers:  []LayerStatus{},
		Errors:  []string{},
	}

	parsePullOutput(raw, result)

	return domain.NewParseResult(result, raw, 0), nil
}

// parsePullOutput extracts pull information from the output.
func parsePullOutput(output string, result *PullResult) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	var lastLine string

	for scanner.Scan() {
		line := scanner.Text()
		lastLine = line

		// Check for errors
		if pullErrorPattern.MatchString(line) {
			result.Success = false
			result.Errors = append(result.Errors, line)
			continue
		}

		// Parse layer status
		if matches := layerPattern.FindStringSubmatch(line); matches != nil {
			layer := LayerStatus{
				ID:     matches[1],
				Status: matches[2],
			}
			result.Layers = append(result.Layers, layer)
			continue
		}

		// Parse digest
		if matches := digestPattern.FindStringSubmatch(line); matches != nil {
			result.Digest = matches[1]
			continue
		}

		// Parse status
		if matches := statusPattern.FindStringSubmatch(line); matches != nil {
			result.Status = matches[1]
			continue
		}
	}

	// The last line is typically the full image reference
	if result.Image == "" && lastLine != "" && !strings.HasPrefix(lastLine, "Status:") {
		// Check if it looks like an image reference
		if strings.Contains(lastLine, "/") || strings.Contains(lastLine, ":") {
			result.Image = strings.TrimSpace(lastLine)
		}
	}
}

// Schema returns the JSON Schema for docker pull output.
func (p *PullParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *PullParser) Matches(cmd string, subcommands []string) bool {
	if cmd != dockerCommand {
		return false
	}

	if len(subcommands) == 0 {
		return false
	}

	// docker pull
	if subcommands[0] == subPull {
		return true
	}

	// docker image pull
	if len(subcommands) >= 2 && subcommands[0] == subcommandImage && subcommands[1] == subPull {
		return true
	}

	return false
}
