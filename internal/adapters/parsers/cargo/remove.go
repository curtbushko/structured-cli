package cargo

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// removingPattern matches "Removing <name> from <type>-dependencies"
var removingPattern = regexp.MustCompile(`Removing\s+(\S+)\s+from\s+`)

// RemoveParser parses the output of 'cargo remove'.
type RemoveParser struct {
	schema domain.Schema
}

// NewRemoveParser creates a new RemoveParser with the cargo remove schema.
func NewRemoveParser() *RemoveParser {
	return &RemoveParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/cargo-remove.json",
			"Cargo Remove Output",
			"object",
			map[string]domain.PropertySchema{
				"success": {Type: "boolean", Description: "Whether the remove operation succeeded"},
				"removed": {Type: "array", Description: "Names of removed dependencies"},
				"errors":  {Type: "array", Description: "Error messages"},
			},
			[]string{"success", "removed", "errors"},
		),
	}
}

// Parse reads cargo remove output and returns structured data.
func (p *RemoveParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &RemoveResult{
		Success: true,
		Removed: []string{},
		Errors:  []string{},
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		// Check for errors
		if strings.HasPrefix(trimmed, "error:") {
			result.Success = false
			result.Errors = append(result.Errors, strings.TrimPrefix(trimmed, "error: "))
			continue
		}

		// Check for removing pattern
		if matches := removingPattern.FindStringSubmatch(line); matches != nil {
			result.Removed = append(result.Removed, matches[1])
		}
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// Schema returns the JSON Schema for cargo remove output.
func (p *RemoveParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *RemoveParser) Matches(cmd string, subcommands []string) bool {
	if cmd != cmdCargo {
		return false
	}
	if len(subcommands) == 0 {
		return false
	}
	// Match "remove" or "rm" (the short alias)
	return subcommands[0] == "remove" || subcommands[0] == "rm"
}
