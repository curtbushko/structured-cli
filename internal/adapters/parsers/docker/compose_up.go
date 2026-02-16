package docker

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// dockerComposeCommand is the docker-compose command name constant.
const dockerComposeCommand = "docker-compose"

// Regex patterns for parsing docker compose up output.
var (
	// composeNetworkPattern matches "Network xxx Created".
	composeNetworkPattern = regexp.MustCompile(`Network\s+"?([^"\s]+)"?\s+(Created|Removed)`)

	// composeVolumePattern matches "Volume xxx Created".
	composeVolumePattern = regexp.MustCompile(`Volume\s+"?([^"\s]+)"?\s+(Created|Removed)`)

	// composeContainerPattern matches "Container xxx-service-1 Started".
	composeContainerPattern = regexp.MustCompile(`Container\s+([^\s]+)-([^\s]+)-(\d+)\s+(Started|Created|Stopped|Removed|Running|Healthy)`)

	// composeServiceErrorPattern matches "service Error".
	composeServiceErrorPattern = regexp.MustCompile(`(\w+)\s+Error`)

	// composeErrorPattern matches error messages.
	composeErrorPattern = regexp.MustCompile(`(?i)Error response from daemon|error:`)
)

// ComposeUpParser parses the output of 'docker compose up'.
type ComposeUpParser struct {
	schema domain.Schema
}

// NewComposeUpParser creates a new ComposeUpParser with the docker-compose-up schema.
func NewComposeUpParser() *ComposeUpParser {
	return &ComposeUpParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/docker-compose-up.json",
			"Docker Compose Up Output",
			"object",
			map[string]domain.PropertySchema{
				"success":  {Type: "boolean", Description: "Whether all services started successfully"},
				"services": {Type: "array", Description: "Services and their status"},
				"networks": {Type: "array", Description: "Created networks"},
				"volumes":  {Type: "array", Description: "Created volumes"},
				"errors":   {Type: "array", Description: "Error messages"},
				"warnings": {Type: "array", Description: "Warning messages"},
			},
			[]string{"success", "services"},
		),
	}
}

// Parse reads docker compose up output and returns structured data.
func (p *ComposeUpParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &ComposeUpResult{
		Success:  true,
		Services: []ComposeService{},
		Networks: []string{},
		Volumes:  []string{},
		Errors:   []string{},
		Warnings: []string{},
	}

	parseComposeUpOutput(raw, result)

	return domain.NewParseResult(result, raw, 0), nil
}

// parseComposeUpOutput extracts compose up information from the output.
func parseComposeUpOutput(output string, result *ComposeUpResult) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	seenServices := make(map[string]bool)

	for scanner.Scan() {
		line := scanner.Text()

		// Check for errors
		if composeErrorPattern.MatchString(line) {
			result.Success = false
			result.Errors = append(result.Errors, strings.TrimSpace(line))
			continue
		}

		// Check for service errors
		if matches := composeServiceErrorPattern.FindStringSubmatch(line); matches != nil {
			result.Success = false
			continue
		}

		// Parse network creation
		if matches := composeNetworkPattern.FindStringSubmatch(line); matches != nil {
			if matches[2] == "Created" {
				result.Networks = append(result.Networks, matches[1])
			}
			continue
		}

		// Parse volume creation
		if matches := composeVolumePattern.FindStringSubmatch(line); matches != nil {
			if matches[2] == "Created" {
				result.Volumes = append(result.Volumes, matches[1])
			}
			continue
		}

		// Parse container status
		if matches := composeContainerPattern.FindStringSubmatch(line); matches != nil {
			serviceName := matches[2]
			status := matches[4]

			// Only add each service once
			if !seenServices[serviceName] {
				seenServices[serviceName] = true
				service := ComposeService{
					Name:   serviceName,
					Status: status,
				}
				result.Services = append(result.Services, service)
			}
		}
	}
}

// Schema returns the JSON Schema for docker compose up output.
func (p *ComposeUpParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *ComposeUpParser) Matches(cmd string, subcommands []string) bool {
	// docker compose up
	if cmd == dockerCommand {
		if len(subcommands) >= 2 && subcommands[0] == subCompose && subcommands[1] == subUp {
			return true
		}
	}

	// docker-compose up
	if cmd == dockerComposeCommand {
		if len(subcommands) >= 1 && subcommands[0] == subUp {
			return true
		}
	}

	return false
}
