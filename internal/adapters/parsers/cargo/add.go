package cargo

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// addingPattern matches "Adding <name> v<version> to <type>-dependencies"
var addingPattern = regexp.MustCompile(`Adding\s+(\S+)\s+v(\S+)\s+to\s+(\S+)`)

// AddParser parses the output of 'cargo add'.
type AddParser struct {
	schema domain.Schema
}

// NewAddParser creates a new AddParser with the cargo add schema.
func NewAddParser() *AddParser {
	return &AddParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/cargo-add.json",
			"Cargo Add Output",
			"object",
			map[string]domain.PropertySchema{
				"success":      {Type: "boolean", Description: "Whether the add operation succeeded"},
				"dependencies": {Type: "array", Description: "Added dependencies"},
				"errors":       {Type: "array", Description: "Error messages"},
			},
			[]string{"success", "dependencies", "errors"},
		),
	}
}

// Parse reads cargo add output and returns structured data.
func (p *AddParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &AddResult{
		Success:      true,
		Dependencies: []AddedDependency{},
		Errors:       []string{},
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

		// Check for adding pattern
		if matches := addingPattern.FindStringSubmatch(line); matches != nil {
			dep := AddedDependency{
				Name:     matches[1],
				Version:  matches[2],
				Features: []string{},
			}

			depType := matches[3]
			switch depType {
			case "dev-dependencies":
				dep.Dev = true
			case "build-dependencies":
				dep.Build = true
			}

			result.Dependencies = append(result.Dependencies, dep)
		}
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// Schema returns the JSON Schema for cargo add output.
func (p *AddParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *AddParser) Matches(cmd string, subcommands []string) bool {
	if cmd != cmdCargo {
		return false
	}
	if len(subcommands) == 0 {
		return false
	}
	return subcommands[0] == "add"
}
