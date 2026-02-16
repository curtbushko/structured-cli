package npm

import (
	"io"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// TestParser parses the output of 'npm test'.
type TestParser struct {
	schema domain.Schema
}

// NewTestParser creates a new TestParser with the npm-test schema.
func NewTestParser() *TestParser {
	return &TestParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/npm-test.json",
			"NPM Test Output",
			"object",
			map[string]domain.PropertySchema{
				"success":   {Type: "boolean", Description: "Whether tests passed"},
				"output":    {Type: "string", Description: "Test output"},
				"exit_code": {Type: "integer", Description: "Exit code from tests"},
			},
			[]string{"success", "output"},
		),
	}
}

// Parse reads npm test output and returns structured data.
func (p *TestParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &TestResult{
		Success:  true,
		Output:   raw,
		ExitCode: 0,
	}

	parseTestOutput(raw, result)

	return domain.NewParseResult(result, raw, 0), nil
}

// parseTestOutput extracts test information from the output.
func parseTestOutput(output string, result *TestResult) {
	// Check for npm error codes (test failure)
	if strings.Contains(output, "npm ERR!") {
		result.Success = false
	}

	// Check for common test failure indicators
	if strings.Contains(output, "FAIL ") {
		result.Success = false
	}

	// Check for "Tests: X failed" pattern (Jest)
	if strings.Contains(output, "failed,") && strings.Contains(output, "Tests:") {
		result.Success = false
	}

	// Check for "failing" (Mocha)
	if strings.Contains(output, "failing") {
		result.Success = false
	}
}

// Schema returns the JSON Schema for npm test output.
func (p *TestParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *TestParser) Matches(cmd string, subcommands []string) bool {
	if cmd != npmCommand {
		return false
	}

	if len(subcommands) == 0 {
		return false
	}

	// npm test, npm t, npm tst
	switch subcommands[0] {
	case "test", "t", "tst":
		return true
	default:
		return false
	}
}
