package docker

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// dockerPSJSONEntry represents a container in JSON format from docker ps --format json.
type dockerPSJSONEntry struct {
	ID        string `json:"ID"`
	Image     string `json:"Image"`
	Command   string `json:"Command"`
	CreatedAt string `json:"CreatedAt"`
	Status    string `json:"Status"`
	Ports     string `json:"Ports"`
	Names     string `json:"Names"`
	State     string `json:"State"`
	Size      string `json:"Size,omitempty"`
}

// Common status strings used for state detection.
const (
	statusUp         = "up"
	statusExited     = "exited"
	statusCreated    = "created"
	statusPaused     = "paused"
	statusRestarting = "restarting"
)

// PSParser parses the output of 'docker ps'.
type PSParser struct {
	schema domain.Schema
}

// NewPSParser creates a new PSParser with the docker-ps schema.
func NewPSParser() *PSParser {
	return &PSParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/docker-ps.json",
			"Docker PS Output",
			"object",
			map[string]domain.PropertySchema{
				"success":    {Type: "boolean", Description: "Whether the command completed successfully"},
				"containers": {Type: "array", Description: "List of containers"},
			},
			[]string{"success", "containers"},
		),
	}
}

// Parse reads docker ps output and returns structured data.
func (p *PSParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &PSResult{
		Success:    true,
		Containers: []Container{},
	}

	// Try JSON format first
	if strings.HasPrefix(strings.TrimSpace(raw), "[") {
		parseJSONPSOutput(raw, result)
	} else {
		parseTablePSOutput(raw, result)
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// parseJSONPSOutput parses JSON format output from docker ps --format json.
func parseJSONPSOutput(output string, result *PSResult) {
	var entries []dockerPSJSONEntry
	if err := json.Unmarshal([]byte(output), &entries); err != nil {
		result.Success = false
		return
	}

	for _, entry := range entries {
		container := Container{
			ID:      entry.ID,
			Names:   entry.Names,
			Image:   entry.Image,
			Command: entry.Command,
			Created: entry.CreatedAt,
			Status:  entry.Status,
			Ports:   entry.Ports,
			Size:    entry.Size,
			State:   entry.State,
		}
		result.Containers = append(result.Containers, container)
	}
}

// parseTablePSOutput parses table format output from docker ps.
func parseTablePSOutput(output string, result *PSResult) {
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

		container := parseContainerLine(line)
		if container.ID != "" {
			result.Containers = append(result.Containers, container)
		}
	}
}

// parseContainerLine parses a single line from docker ps table output.
func parseContainerLine(line string) Container {
	// Docker ps output is column-based with variable-width columns
	// CONTAINER ID   IMAGE         COMMAND       CREATED         STATUS         PORTS     NAMES
	// The columns are separated by multiple spaces

	// Split by multiple spaces
	parts := splitBySpaces(line)
	if len(parts) < 6 {
		return Container{}
	}

	container := Container{
		ID:    parts[0],
		Image: parts[1],
	}

	// Find the command (it may be quoted)
	commandIdx := 2
	if commandIdx < len(parts) {
		container.Command = strings.Trim(parts[commandIdx], "\"")
	}

	// Parse remaining fields - this is complex because fields can have spaces
	// We need to work backwards from the end
	container.Names = parts[len(parts)-1]
	container.Ports = parts[len(parts)-2]

	// Status and Created are in the middle - combine parts
	// Status typically starts with "Up", "Exited", "Created", etc.
	statusStart := -1
	for i := 3; i < len(parts)-2; i++ {
		if isStatusStart(parts[i]) {
			statusStart = i
			break
		}
	}

	if statusStart >= 3 {
		container.Created = strings.Join(parts[3:statusStart], " ")
		container.Status = strings.Join(parts[statusStart:len(parts)-2], " ")
	} else {
		// If status not found, try to find it in the combined string
		remaining := strings.Join(parts[3:len(parts)-2], " ")
		container.Status = remaining
	}

	// Determine state from status
	container.State = determineState(container.Status)

	return container
}

// splitBySpaces splits a string by multiple consecutive spaces.
func splitBySpaces(s string) []string {
	var parts []string
	current := strings.Builder{}

	inSpaces := false
	for _, r := range s {
		if r == ' ' {
			if !inSpaces && current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			inSpaces = true
		} else {
			inSpaces = false
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}

// isStatusStart checks if a string starts a status field.
func isStatusStart(s string) bool {
	lower := strings.ToLower(s)
	return lower == statusUp ||
		lower == statusExited ||
		lower == statusCreated ||
		lower == statusPaused ||
		lower == statusRestarting
}

// determineState determines the container state from the status string.
func determineState(status string) string {
	lower := strings.ToLower(status)
	switch {
	case strings.HasPrefix(lower, statusUp):
		return "running"
	case strings.HasPrefix(lower, statusExited):
		return statusExited
	case strings.HasPrefix(lower, statusCreated):
		return statusCreated
	case strings.HasPrefix(lower, statusPaused):
		return statusPaused
	case strings.HasPrefix(lower, statusRestarting):
		return statusRestarting
	default:
		return "unknown"
	}
}

// Schema returns the JSON Schema for docker ps output.
func (p *PSParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *PSParser) Matches(cmd string, subcommands []string) bool {
	if cmd != dockerCommand {
		return false
	}

	if len(subcommands) == 0 {
		return false
	}

	// docker ps
	if subcommands[0] == subPs {
		return true
	}

	// docker container ls/list
	if len(subcommands) >= 2 && subcommands[0] == subContainer {
		switch subcommands[1] {
		case subLs, subList:
			return true
		}
	}

	return false
}
