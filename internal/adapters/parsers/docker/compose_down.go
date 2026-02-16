package docker

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Regex patterns for parsing docker compose down output.
var (
	// composeContainerStoppedPattern matches "Container xxx Stopped".
	composeContainerStoppedPattern = regexp.MustCompile(`Container\s+([^\s]+)\s+Stopped`)

	// composeContainerRemovedPattern matches "Container xxx Removed".
	composeContainerRemovedPattern = regexp.MustCompile(`Container\s+([^\s]+)\s+Removed`)

	// composeNetworkRemovedPattern matches "Network xxx Removed".
	composeNetworkRemovedPattern = regexp.MustCompile(`Network\s+"?([^"\s]+)"?\s+Removed`)

	// composeVolumeRemovedPattern matches "Volume xxx Removed".
	composeVolumeRemovedPattern = regexp.MustCompile(`Volume\s+"?([^"\s]+)"?\s+Removed`)
)

// ComposeDownParser parses the output of 'docker compose down'.
type ComposeDownParser struct {
	schema domain.Schema
}

// NewComposeDownParser creates a new ComposeDownParser with the docker-compose-down schema.
func NewComposeDownParser() *ComposeDownParser {
	return &ComposeDownParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/docker-compose-down.json",
			"Docker Compose Down Output",
			"object",
			map[string]domain.PropertySchema{
				"success":            {Type: "boolean", Description: "Whether all services stopped successfully"},
				"stopped_containers": {Type: "array", Description: "Stopped containers"},
				"removed_containers": {Type: "array", Description: "Removed containers"},
				"removed_networks":   {Type: "array", Description: "Removed networks"},
				"removed_volumes":    {Type: "array", Description: "Removed volumes"},
				"errors":             {Type: "array", Description: "Error messages"},
			},
			[]string{"success", "stopped_containers", "removed_containers"},
		),
	}
}

// Parse reads docker compose down output and returns structured data.
func (p *ComposeDownParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &ComposeDownResult{
		Success:           true,
		StoppedContainers: []string{},
		RemovedContainers: []string{},
		RemovedNetworks:   []string{},
		RemovedVolumes:    []string{},
		Errors:            []string{},
	}

	parseComposeDownOutput(raw, result)

	return domain.NewParseResult(result, raw, 0), nil
}

// parseComposeDownOutput extracts compose down information from the output.
func parseComposeDownOutput(output string, result *ComposeDownResult) {
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()

		// Check for errors
		if composeErrorPattern.MatchString(line) {
			result.Success = false
			result.Errors = append(result.Errors, strings.TrimSpace(line))
			continue
		}

		// Parse stopped containers
		if matches := composeContainerStoppedPattern.FindStringSubmatch(line); matches != nil {
			result.StoppedContainers = append(result.StoppedContainers, matches[1])
			continue
		}

		// Parse removed containers
		if matches := composeContainerRemovedPattern.FindStringSubmatch(line); matches != nil {
			result.RemovedContainers = append(result.RemovedContainers, matches[1])
			continue
		}

		// Parse removed networks
		if matches := composeNetworkRemovedPattern.FindStringSubmatch(line); matches != nil {
			result.RemovedNetworks = append(result.RemovedNetworks, matches[1])
			continue
		}

		// Parse removed volumes
		if matches := composeVolumeRemovedPattern.FindStringSubmatch(line); matches != nil {
			result.RemovedVolumes = append(result.RemovedVolumes, matches[1])
			continue
		}
	}
}

// Schema returns the JSON Schema for docker compose down output.
func (p *ComposeDownParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *ComposeDownParser) Matches(cmd string, subcommands []string) bool {
	// docker compose down
	if cmd == dockerCommand {
		if len(subcommands) >= 2 && subcommands[0] == subCompose && subcommands[1] == subDown {
			return true
		}
	}

	// docker-compose down
	if cmd == dockerComposeCommand {
		if len(subcommands) >= 1 && subcommands[0] == subDown {
			return true
		}
	}

	return false
}
