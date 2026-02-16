package npm

import (
	"io"
	"regexp"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Regex patterns for parsing npm run output
var (
	// scriptNamePattern matches "> package@version scriptname"
	scriptNamePattern = regexp.MustCompile(`^> \S+@\S+ (\S+)$`)

	// npmErrCodePattern matches "npm ERR! code ELIFECYCLE"
	npmErrCodePattern = regexp.MustCompile(`^npm ERR! code`)
)

// RunParser parses the output of 'npm run'.
type RunParser struct {
	schema domain.Schema
}

// NewRunParser creates a new RunParser with the npm-run schema.
func NewRunParser() *RunParser {
	return &RunParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/npm-run.json",
			"NPM Run Output",
			"object",
			map[string]domain.PropertySchema{
				"success":   {Type: "boolean", Description: "Whether the script ran successfully"},
				"script":    {Type: "string", Description: "Script name that was run"},
				"output":    {Type: "string", Description: "Script output"},
				"exit_code": {Type: "integer", Description: "Exit code from the script"},
			},
			[]string{"success", "script", "output"},
		),
	}
}

// Parse reads npm run output and returns structured data.
func (p *RunParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &RunResult{
		Success:  true,
		Output:   raw,
		ExitCode: 0,
	}

	parseRunOutput(raw, result)

	return domain.NewParseResult(result, raw, 0), nil
}

// parseRunOutput extracts run information from the output.
func parseRunOutput(output string, result *RunResult) {
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		// Extract script name
		if matches := scriptNamePattern.FindStringSubmatch(line); matches != nil {
			result.Script = matches[1]
		}

		// Check for npm error
		if npmErrCodePattern.MatchString(line) {
			result.Success = false
		}
	}
}

// Schema returns the JSON Schema for npm run output.
func (p *RunParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *RunParser) Matches(cmd string, subcommands []string) bool {
	if cmd != npmCommand {
		return false
	}

	if len(subcommands) == 0 {
		return false
	}

	// npm run, npm run-script
	switch subcommands[0] {
	case "run", "run-script":
		return true
	default:
		return false
	}
}
