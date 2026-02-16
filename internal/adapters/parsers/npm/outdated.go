package npm

import (
	"bufio"
	"io"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// OutdatedParser parses the output of 'npm outdated'.
type OutdatedParser struct {
	schema domain.Schema
}

// NewOutdatedParser creates a new OutdatedParser with the npm-outdated schema.
func NewOutdatedParser() *OutdatedParser {
	return &OutdatedParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/npm-outdated.json",
			"NPM Outdated Output",
			"object",
			map[string]domain.PropertySchema{
				"success":  {Type: "boolean", Description: "Whether all packages are up to date"},
				"packages": {Type: "array", Description: "List of outdated packages"},
			},
			[]string{"success", "packages"},
		),
	}
}

// Parse reads npm outdated output and returns structured data.
func (p *OutdatedParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &OutdatedResult{
		Success:  true,
		Packages: []OutdatedPackage{},
	}

	parseOutdatedOutput(raw, result)

	// If there are outdated packages, mark as not successful
	if len(result.Packages) > 0 {
		result.Success = false
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// parseOutdatedOutput extracts outdated package information from the output.
func parseOutdatedOutput(output string, result *OutdatedResult) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	headerSeen := false

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Skip header line
		if strings.HasPrefix(line, "Package") {
			headerSeen = true
			continue
		}

		if !headerSeen {
			continue
		}

		// Parse package line - columns are variable width
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}

		pkg := OutdatedPackage{
			Name:     fields[0],
			Current:  fields[1],
			Wanted:   fields[2],
			Latest:   fields[3],
			Location: fields[4],
		}

		// Type is optional (6th field if present)
		if len(fields) >= 6 {
			pkg.Type = fields[5]
		}

		result.Packages = append(result.Packages, pkg)
	}
}

// Schema returns the JSON Schema for npm outdated output.
func (p *OutdatedParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *OutdatedParser) Matches(cmd string, subcommands []string) bool {
	if cmd != npmCommand {
		return false
	}

	if len(subcommands) == 0 {
		return false
	}

	return subcommands[0] == "outdated"
}
