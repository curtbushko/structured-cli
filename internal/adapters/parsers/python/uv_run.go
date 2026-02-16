package python

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Regex patterns for parsing uv run output.
var (
	// uvRunInstalledPattern matches "+ package==version" lines from uv run
	uvRunInstalledPattern = regexp.MustCompile(`^\s*\+\s+(\S+)==(\S+)`)

	// uvRunMetaPattern matches uv metadata lines (Resolved, Prepared, Installed)
	uvRunMetaPattern = regexp.MustCompile(`^(Resolved|Prepared|Installed)\s+\d+\s+packages?`)
)

// UVRunParser parses the output of 'uv run'.
type UVRunParser struct {
	schema domain.Schema
}

// NewUVRunParser creates a new UVRunParser with the uv-run schema.
func NewUVRunParser() *UVRunParser {
	return &UVRunParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/uv-run.json",
			"UV Run Output",
			"object",
			map[string]domain.PropertySchema{
				"success":            {Type: "boolean", Description: "Whether the command ran successfully"},
				"script":             {Type: "string", Description: "The script or command that was run"},
				"output":             {Type: "string", Description: "The command output"},
				"installed_packages": {Type: "array", Description: "Packages installed before running"},
				"exit_code":          {Type: "integer", Description: "Exit code from the command"},
			},
			[]string{"success", "output"},
		),
	}
}

// Parse reads uv run output and returns structured data.
func (p *UVRunParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &UVRunResult{
		Success:           true,
		InstalledPackages: []InstalledPackage{},
	}

	parseUVRunOutput(raw, result)

	return domain.NewParseResult(result, raw, 0), nil
}

// parseUVRunOutput extracts run information from uv output.
func parseUVRunOutput(output string, result *UVRunResult) {
	var outputLines []string
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()

		// Parse installed packages
		if matches := uvRunInstalledPattern.FindStringSubmatch(line); matches != nil {
			result.InstalledPackages = append(result.InstalledPackages, InstalledPackage{
				Name:    matches[1],
				Version: matches[2],
			})
			continue
		}

		// Skip uv metadata lines
		if uvRunMetaPattern.MatchString(line) {
			continue
		}

		// Collect all other output as the script output
		outputLines = append(outputLines, line)
	}

	result.Output = strings.Join(outputLines, "\n")
}

// Schema returns the JSON Schema for uv run output.
func (p *UVRunParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *UVRunParser) Matches(cmd string, subcommands []string) bool {
	if cmd != uvCommand {
		return false
	}

	if len(subcommands) == 0 {
		return false
	}

	return subcommands[0] == "run"
}
