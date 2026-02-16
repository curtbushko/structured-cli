package docker

import (
	"io"
	"regexp"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Regex patterns for parsing docker run output.
var (
	// containerIDPattern matches a full 64-character container ID.
	containerIDPattern = regexp.MustCompile(`^[a-f0-9]{64}$`)

	// shortContainerIDPattern matches a short 12-character container ID.
	shortContainerIDPattern = regexp.MustCompile(`^[a-f0-9]{12}$`)

	// dockerErrorPattern matches Docker error messages.
	dockerErrorPattern = regexp.MustCompile(`(?i)docker:\s*Error|Error response from daemon`)
)

// RunParser parses the output of 'docker run'.
type RunParser struct {
	schema domain.Schema
}

// NewRunParser creates a new RunParser with the docker-run schema.
func NewRunParser() *RunParser {
	return &RunParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/docker-run.json",
			"Docker Run Output",
			"object",
			map[string]domain.PropertySchema{
				"success":      {Type: "boolean", Description: "Whether the container ran successfully"},
				"container_id": {Type: "string", Description: "ID of the created/running container"},
				"output":       {Type: "string", Description: "Container output (if not detached)"},
				"exit_code":    {Type: "integer", Description: "Container exit code"},
				"detached":     {Type: "boolean", Description: "Whether container is running in background"},
				"errors":       {Type: "array", Description: "Error messages"},
			},
			[]string{"success", "container_id"},
		),
	}
}

// Parse reads docker run output and returns structured data.
func (p *RunParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &RunResult{
		Success: true,
		Errors:  []string{},
	}

	parseRunOutput(raw, result)

	return domain.NewParseResult(result, raw, 0), nil
}

// parseRunOutput extracts run information from the output.
func parseRunOutput(output string, result *RunResult) {
	trimmed := strings.TrimSpace(output)

	// Check for errors first
	if dockerErrorPattern.MatchString(output) {
		result.Success = false
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				result.Errors = append(result.Errors, strings.TrimSpace(line))
			}
		}
		return
	}

	// Check if output is a container ID (detached mode)
	if containerIDPattern.MatchString(trimmed) || shortContainerIDPattern.MatchString(trimmed) {
		result.ContainerID = trimmed
		result.Detached = true
		return
	}

	// Otherwise, it's regular output from the container
	result.Output = trimmed
	result.Detached = false
}

// Schema returns the JSON Schema for docker run output.
func (p *RunParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *RunParser) Matches(cmd string, subcommands []string) bool {
	if cmd != dockerCommand {
		return false
	}

	if len(subcommands) == 0 {
		return false
	}

	// docker run
	if subcommands[0] == subRun {
		return true
	}

	// docker container run
	if len(subcommands) >= 2 && subcommands[0] == subContainer && subcommands[1] == subRun {
		return true
	}

	return false
}
