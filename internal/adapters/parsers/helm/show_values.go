package helm

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Regex patterns for YAML parsing.
var (
	// topLevelKeyPattern matches top-level YAML keys (no leading whitespace).
	topLevelKeyPattern = regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_-]*):\s*(.*)$`)

	// paramCommentPattern matches @param comments like "## @param key Description".
	paramCommentPattern = regexp.MustCompile(`^##\s*@param\s+(\S+)\s+(.+)$`)

	// errorPattern matches helm error messages.
	errorPattern = regexp.MustCompile(`^Error:`)
)

// ShowValuesParser parses the output of 'helm show values'.
type ShowValuesParser struct {
	schema domain.Schema
}

// NewShowValuesParser creates a new ShowValuesParser with the helm-show-values schema.
func NewShowValuesParser() *ShowValuesParser {
	return &ShowValuesParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/helm-show-values.json",
			"Helm Show Values Output",
			"object",
			map[string]domain.PropertySchema{
				"values": {Type: "array", Description: "Extracted chart values with top-level keys"},
				"raw":    {Type: "string", Description: "Raw YAML output from helm show values"},
			},
			[]string{"raw"},
		),
	}
}

// Parse reads helm show values output and returns structured data.
func (p *ShowValuesParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return emptyResultWithError(err, ""), nil
	}

	raw := string(data)
	trimmed := strings.TrimSpace(raw)

	if trimmed == "" {
		return emptyResultOK(&ShowValuesResult{}, raw), nil
	}

	// Check for error output
	if errorPattern.MatchString(trimmed) {
		return emptyResultOK(&ShowValuesResult{Raw: raw}, raw), nil
	}

	result := parseValuesOutput(trimmed)
	result.Raw = raw

	return domain.NewParseResult(result, raw, 0), nil
}

// Schema returns the JSON Schema for helm show values output.
func (p *ShowValuesParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *ShowValuesParser) Matches(cmd string, subcommands []string) bool {
	if cmd != cmdHelm || len(subcommands) < 2 {
		return false
	}
	return subcommands[0] == cmdShow && subcommands[1] == cmdValues
}

// parseValuesOutput parses the helm show values YAML output.
func parseValuesOutput(input string) *ShowValuesResult {
	result := &ShowValuesResult{}
	scanner := bufio.NewScanner(strings.NewReader(input))

	// Track description comments for @param annotations
	descriptions := make(map[string]string)
	var currentComments []string

	for scanner.Scan() {
		line := scanner.Text()

		// Check for @param comments
		if matches := paramCommentPattern.FindStringSubmatch(line); len(matches) >= 3 {
			paramName := extractParamBaseName(matches[1])
			descriptions[paramName] = matches[2]
			currentComments = append(currentComments, matches[2])
			continue
		}

		// Skip other comments
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}

		// Parse top-level key
		if matches := topLevelKeyPattern.FindStringSubmatch(line); len(matches) >= 3 {
			key := matches[1]
			valueStr := strings.TrimSpace(matches[2])

			value := &ChartValue{Key: key}

			// Set description from collected comments
			if desc, ok := descriptions[key]; ok {
				value.Description = desc
			} else if len(currentComments) > 0 {
				value.Description = currentComments[len(currentComments)-1]
			}

			// Parse the value
			value.Value = parseYAMLValue(valueStr)

			result.Values = append(result.Values, *value)
			currentComments = nil
		}
	}

	return result
}

// extractParamBaseName extracts the base name from a dotted path (e.g., "image.registry" -> "image").
func extractParamBaseName(path string) string {
	if idx := strings.Index(path, "."); idx > 0 {
		return path[:idx]
	}
	return path
}

// parseYAMLValue attempts to parse a simple YAML value string.
func parseYAMLValue(s string) any {
	// Empty or nested value
	if isEmptyYAMLValue(s) {
		return nil
	}

	// Handle special literals
	if val, ok := parseSpecialLiteral(s); ok {
		return val
	}

	// Handle booleans and null
	if val, ok := parseBoolOrNull(s); ok {
		return val
	}

	// Handle numbers
	if val, ok := parseNumber(s); ok {
		return val
	}

	// Handle quoted strings
	if val, ok := parseQuotedString(s); ok {
		return val
	}

	// Return as string
	return s
}

// isEmptyYAMLValue checks if the value represents empty/nested content.
func isEmptyYAMLValue(s string) bool {
	return s == "" || s == "|" || s == ">"
}

// parseSpecialLiteral handles empty arrays and objects.
func parseSpecialLiteral(s string) (any, bool) {
	switch s {
	case "[]":
		return []any{}, true
	case "{}":
		return map[string]any{}, true
	default:
		return nil, false
	}
}

// parseBoolOrNull handles boolean and null values.
func parseBoolOrNull(s string) (any, bool) {
	lower := strings.ToLower(s)
	switch lower {
	case "true":
		return true, true
	case "false":
		return false, true
	case "null", "~":
		return nil, true
	default:
		return nil, false
	}
}

// parseNumber handles integer and float values.
func parseNumber(s string) (any, bool) {
	if i, err := strconv.Atoi(s); err == nil {
		return i, true
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f, true
	}
	return nil, false
}

// parseQuotedString handles quoted strings by removing quotes.
func parseQuotedString(s string) (string, bool) {
	if len(s) < 2 {
		return "", false
	}
	if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
		return s[1 : len(s)-1], true
	}
	return "", false
}
