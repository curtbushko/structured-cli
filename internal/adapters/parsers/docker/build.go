package docker

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Common subcommands for docker build.
const (
	subcommandBuild  = "build"
	subcommandImage  = "image"
	subcommandBuildx = "buildx"
)

// Regex patterns for parsing docker build output.
var (
	// stepPattern matches build step lines like "#1 [internal] load build definition".
	stepPattern = regexp.MustCompile(`^#(\d+)\s+(.+)$`)

	// cachedPattern matches "CACHED" in build output.
	cachedPattern = regexp.MustCompile(`(?i)CACHED`)

	// imageIDPattern matches "writing image sha256:xxx".
	imageIDPattern = regexp.MustCompile(`writing image (sha256:[a-f0-9]+)`)

	// namingPattern matches "naming to xxx".
	namingPattern = regexp.MustCompile(`naming to (.+?)\s+done`)

	// buildErrorPattern matches ERROR in build output.
	buildErrorPattern = regexp.MustCompile(`(?i)ERROR:|error:`)
)

// BuildParser parses the output of 'docker build'.
type BuildParser struct {
	schema domain.Schema
}

// NewBuildParser creates a new BuildParser with the docker-build schema.
func NewBuildParser() *BuildParser {
	return &BuildParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/docker-build.json",
			"Docker Build Output",
			"object",
			map[string]domain.PropertySchema{
				"success":     {Type: "boolean", Description: "Whether the build completed successfully"},
				"image_id":    {Type: "string", Description: "ID of the built image"},
				"tags":        {Type: "array", Description: "Tags applied to the image"},
				"steps":       {Type: "array", Description: "Build steps executed"},
				"total_steps": {Type: "integer", Description: "Total number of build steps"},
				"cached":      {Type: "integer", Description: "Number of cached steps"},
				"errors":      {Type: "array", Description: "Build errors"},
				"warnings":    {Type: "array", Description: "Build warnings"},
			},
			[]string{"success", "image_id", "steps"},
		),
	}
}

// Parse reads docker build output and returns structured data.
func (p *BuildParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &BuildResult{
		Success:  true,
		Tags:     []string{},
		Steps:    []BuildStep{},
		Errors:   []string{},
		Warnings: []string{},
	}

	parseBuildOutput(raw, result)

	return domain.NewParseResult(result, raw, 0), nil
}

// parseBuildOutput extracts build information from the output.
func parseBuildOutput(output string, result *BuildResult) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	seenSteps := make(map[int]bool)

	for scanner.Scan() {
		line := scanner.Text()

		// Parse build steps
		if matches := stepPattern.FindStringSubmatch(line); matches != nil {
			stepNum, _ := strconv.Atoi(matches[1])
			instruction := strings.TrimSpace(matches[2])

			// Skip duplicate step entries (steps appear multiple times with different status)
			if !seenSteps[stepNum] {
				seenSteps[stepNum] = true
				step := BuildStep{
					Number:      stepNum,
					Instruction: instruction,
					Cached:      false,
				}
				result.Steps = append(result.Steps, step)
				result.TotalSteps = stepNum
			}
		}

		// Check for cached steps
		if cachedPattern.MatchString(line) {
			// Find which step this CACHED belongs to
			if matches := stepPattern.FindStringSubmatch(line); matches != nil {
				stepNum, _ := strconv.Atoi(matches[1])
				for i := range result.Steps {
					if result.Steps[i].Number == stepNum {
						result.Steps[i].Cached = true
						result.Cached++
						break
					}
				}
			}
		}

		// Extract image ID
		if matches := imageIDPattern.FindStringSubmatch(line); matches != nil {
			result.ImageID = matches[1]
		}

		// Extract tags
		if matches := namingPattern.FindStringSubmatch(line); matches != nil {
			result.Tags = append(result.Tags, matches[1])
		}

		// Check for errors
		if buildErrorPattern.MatchString(line) {
			result.Success = false
			result.Errors = append(result.Errors, line)
		}
	}
}

// Schema returns the JSON Schema for docker build output.
func (p *BuildParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *BuildParser) Matches(cmd string, subcommands []string) bool {
	if cmd != dockerCommand {
		return false
	}

	if len(subcommands) == 0 {
		return false
	}

	// docker build
	if subcommands[0] == subcommandBuild {
		return true
	}

	// docker image build
	if len(subcommands) >= 2 && subcommands[0] == subcommandImage && subcommands[1] == subcommandBuild {
		return true
	}

	// docker buildx build
	if len(subcommands) >= 2 && subcommands[0] == subcommandBuildx && subcommands[1] == subcommandBuild {
		return true
	}

	return false
}
