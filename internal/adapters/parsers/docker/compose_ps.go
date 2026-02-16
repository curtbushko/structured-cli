package docker

import (
	"bufio"
	"encoding/json"
	"io"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// composePSJSONEntry represents a service in JSON format from docker compose ps --format json.
type composePSJSONEntry struct {
	Name    string `json:"Name"`
	Image   string `json:"Image"`
	Service string `json:"Service"`
	Status  string `json:"Status"`
	State   string `json:"State"`
	Ports   string `json:"Ports"`
	Health  string `json:"Health,omitempty"`
}

// ComposePSParser parses the output of 'docker compose ps'.
type ComposePSParser struct {
	schema domain.Schema
}

// NewComposePSParser creates a new ComposePSParser with the docker-compose-ps schema.
func NewComposePSParser() *ComposePSParser {
	return &ComposePSParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/docker-compose-ps.json",
			"Docker Compose PS Output",
			"object",
			map[string]domain.PropertySchema{
				"success":      {Type: "boolean", Description: "Whether the command completed successfully"},
				"project_name": {Type: "string", Description: "Compose project name"},
				"services":     {Type: "array", Description: "Services and their status"},
			},
			[]string{"success", "services"},
		),
	}
}

// Parse reads docker compose ps output and returns structured data.
func (p *ComposePSParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &ComposePSResult{
		Success:  true,
		Services: []ComposeService{},
	}

	// Try JSON format first
	if strings.HasPrefix(strings.TrimSpace(raw), "[") {
		parseJSONComposePSOutput(raw, result)
	} else {
		parseTableComposePSOutput(raw, result)
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// parseJSONComposePSOutput parses JSON format output from docker compose ps --format json.
func parseJSONComposePSOutput(output string, result *ComposePSResult) {
	var entries []composePSJSONEntry
	if err := json.Unmarshal([]byte(output), &entries); err != nil {
		result.Success = false
		return
	}

	for _, entry := range entries {
		service := ComposeService{
			Name:   entry.Service,
			Image:  entry.Image,
			Status: entry.Status,
			State:  entry.State,
			Ports:  entry.Ports,
			Health: entry.Health,
		}
		// Extract container ID from Name if present
		if entry.Name != "" {
			service.ContainerID = entry.Name
		}
		result.Services = append(result.Services, service)
	}
}

// parseTableComposePSOutput parses table format output from docker compose ps.
func parseTableComposePSOutput(output string, result *ComposePSResult) {
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

		service := parseComposePSLine(line)
		if service.Name != "" {
			result.Services = append(result.Services, service)
		}
	}
}

// parseComposePSLine parses a single line from docker compose ps table output.
func parseComposePSLine(line string) ComposeService {
	// Docker compose ps output columns:
	// NAME    IMAGE    COMMAND    SERVICE    CREATED    STATUS    PORTS

	parts := splitBySpaces(line)
	if len(parts) < 6 {
		return ComposeService{}
	}

	service := ComposeService{
		ContainerID: parts[0],
		Image:       parts[1],
		Ports:       parts[len(parts)-1],
	}

	// Find status and set state
	statusIdx := findStatusIndex(parts)
	if statusIdx > 0 {
		service.Status = strings.Join(parts[statusIdx:len(parts)-1], " ")
	}
	service.State = determineState(service.Status)

	// Find SERVICE name
	service.Name = findServiceName(parts, statusIdx)
	if service.Name == "" {
		service.Name = extractServiceFromContainerName(parts[0])
	}

	return service
}

// findStatusIndex finds the index where status starts in parts.
func findStatusIndex(parts []string) int {
	for i := len(parts) - 2; i >= 2; i-- {
		if isStatusKeyword(parts[i]) {
			return i
		}
	}
	return -1
}

// isStatusKeyword checks if a string is a status keyword.
func isStatusKeyword(s string) bool {
	lower := strings.ToLower(s)
	return lower == statusUp || lower == statusExited || lower == statusCreated ||
		lower == statusPaused || lower == statusRestarting
}

// findServiceName looks for the service name in parts before statusIdx.
func findServiceName(parts []string, statusIdx int) string {
	for i := statusIdx - 1; i >= 2; i-- {
		part := parts[i]
		if shouldSkipPart(part) {
			continue
		}
		if isValidServiceName(part) {
			return part
		}
	}
	return ""
}

// shouldSkipPart checks if a part should be skipped when looking for service name.
func shouldSkipPart(part string) bool {
	return isTimeUnit(part) ||
		isNumeric(part) ||
		strings.Contains(part, "\"") ||
		strings.Contains(part, "'") ||
		strings.Contains(part, "...") ||
		strings.Contains(part, "\u2026")
}

// isTimeUnit checks if a string is a time unit word.
func isTimeUnit(s string) bool {
	lower := strings.ToLower(s)
	return lower == "ago" || lower == "minutes" || lower == "hours" || lower == "days" ||
		lower == "seconds" || lower == "weeks" || lower == "month" || lower == "months"
}

// extractServiceFromContainerName extracts service name from container name.
func extractServiceFromContainerName(containerName string) string {
	if !strings.Contains(containerName, "-") {
		return ""
	}
	nameParts := strings.Split(containerName, "-")
	if len(nameParts) >= 2 {
		return nameParts[len(nameParts)-2]
	}
	return ""
}

// isNumeric checks if a string is a number.
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

// isValidServiceName checks if a string is a valid service name.
func isValidServiceName(s string) bool {
	if len(s) == 0 || len(s) > 50 {
		return false
	}
	// Service names are alphanumeric with possible underscores/hyphens
	// but shouldn't end with -N (replica number)
	if endsWithReplicaNumber(s) {
		return false
	}
	// Check for valid characters using isAlphanumericOrDash
	for _, c := range s {
		if !isAlphanumericOrDash(c) {
			return false
		}
	}
	return true
}

// endsWithReplicaNumber checks if string ends with -N pattern (replica number).
func endsWithReplicaNumber(s string) bool {
	if len(s) < 2 {
		return false
	}
	lastChar := s[len(s)-1]
	return lastChar >= '0' && lastChar <= '9' && s[len(s)-2] == '-'
}

// isAlphanumericOrDash checks if a character is alphanumeric, underscore, or dash.
func isAlphanumericOrDash(c rune) bool {
	return (c >= 'a' && c <= 'z') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= '0' && c <= '9') ||
		c == '_' ||
		c == '-'
}

// Schema returns the JSON Schema for docker compose ps output.
func (p *ComposePSParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *ComposePSParser) Matches(cmd string, subcommands []string) bool {
	// docker compose ps
	if cmd == dockerCommand {
		if len(subcommands) >= 2 && subcommands[0] == subCompose && subcommands[1] == subPs {
			return true
		}
	}

	// docker-compose ps
	if cmd == dockerComposeCommand {
		if len(subcommands) >= 1 && subcommands[0] == subPs {
			return true
		}
	}

	return false
}
