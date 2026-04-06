package helm

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Status field prefixes for helm status output.
const (
	fieldName         = "NAME:"
	fieldLastDeployed = "LAST DEPLOYED:"
	fieldNamespace    = "NAMESPACE:"
	fieldStatus       = "STATUS:"
	fieldRevision     = "REVISION:"
	fieldDescription  = "DESCRIPTION:"
	fieldNotes        = "NOTES:"
	fieldResources    = "RESOURCES:"
)

// resourceKindPattern matches resource kind lines like "==> v1/Service".
var resourceKindPattern = regexp.MustCompile(`^==>\s+(.+)$`)

// StatusParser parses the output of 'helm status'.
type StatusParser struct {
	schema domain.Schema
}

// NewStatusParser creates a new StatusParser with the helm-status schema.
func NewStatusParser() *StatusParser {
	return &StatusParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/helm-status.json",
			"Helm Status Output",
			"object",
			map[string]domain.PropertySchema{
				"name":          {Type: "string", Description: "Release name"},
				"namespace":     {Type: "string", Description: "Release namespace"},
				"status":        {Type: "string", Description: "Release status"},
				"revision":      {Type: "integer", Description: "Current revision number"},
				"last_deployed": {Type: "string", Description: "Timestamp of the last deployment"},
				"description":   {Type: "string", Description: "Status description"},
				"notes":         {Type: "string", Description: "NOTES.txt output from the chart"},
				"resources":     {Type: "array", Description: "Kubernetes resources in the release"},
			},
			[]string{"name", "status"},
		),
	}
}

// Parse reads helm status output and returns structured data.
func (p *StatusParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return emptyResultWithError(err, ""), nil
	}

	raw := string(data)
	trimmed := strings.TrimSpace(raw)

	if trimmed == "" {
		return emptyResultOK(&StatusResult{}, raw), nil
	}

	result := parseStatusOutput(trimmed)
	return domain.NewParseResult(result, raw, 0), nil
}

// Schema returns the JSON Schema for helm status output.
func (p *StatusParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *StatusParser) Matches(cmd string, subcommands []string) bool {
	if cmd != cmdHelm || len(subcommands) < 1 {
		return false
	}
	return subcommands[0] == cmdStatus
}

// parseStatusOutput parses the helm status output into a StatusResult.
func parseStatusOutput(input string) *StatusResult {
	result := &StatusResult{}
	scanner := bufio.NewScanner(strings.NewReader(input))

	var section string
	var notesBuilder strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		// Check for section starts
		if strings.HasPrefix(line, fieldNotes) {
			section = fieldNotes
			continue
		}
		if strings.HasPrefix(line, fieldResources) {
			section = fieldResources
			continue
		}

		// Parse based on current section
		switch section {
		case fieldNotes:
			if notesBuilder.Len() > 0 {
				notesBuilder.WriteString("\n")
			}
			notesBuilder.WriteString(line)
		case fieldResources:
			parseResourceLine(line, result)
		default:
			parseHeaderField(line, result)
		}
	}

	// Set notes from builder
	notes := strings.TrimSpace(notesBuilder.String())
	if notes != "" {
		result.Notes = notes
	}

	return result
}

// parseHeaderField parses a key-value header field.
func parseHeaderField(line string, result *StatusResult) {
	switch {
	case strings.HasPrefix(line, fieldName):
		result.Name = extractValue(line, fieldName)
	case strings.HasPrefix(line, fieldLastDeployed):
		result.LastDeployed = extractValue(line, fieldLastDeployed)
	case strings.HasPrefix(line, fieldNamespace):
		result.Namespace = extractValue(line, fieldNamespace)
	case strings.HasPrefix(line, fieldStatus):
		result.Status = extractValue(line, fieldStatus)
	case strings.HasPrefix(line, fieldRevision):
		result.Revision = parseRevision(extractValue(line, fieldRevision))
	case strings.HasPrefix(line, fieldDescription):
		result.Description = extractValue(line, fieldDescription)
	}
}

// parseResourceLine parses a resource line in the RESOURCES section.
func parseResourceLine(line string, result *StatusResult) {
	// Check for resource kind header like "==> v1/Service"
	matches := resourceKindPattern.FindStringSubmatch(line)
	if len(matches) >= 2 {
		// Store the kind for upcoming resource entries
		result.currentResourceKind = matches[1]
		return
	}

	// Skip header lines and empty lines
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "NAME") {
		return
	}

	// Parse resource entry - first field is the name
	if result.currentResourceKind != "" {
		fields := strings.Fields(trimmed)
		if len(fields) >= 1 {
			resource := ReleaseResource{
				Kind: result.currentResourceKind,
				Name: fields[0],
			}
			result.Resources = append(result.Resources, resource)
		}
	}
}

// extractValue extracts the value after a field prefix.
func extractValue(line, prefix string) string {
	return strings.TrimSpace(strings.TrimPrefix(line, prefix))
}

// parseRevision parses a revision string to int.
func parseRevision(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
