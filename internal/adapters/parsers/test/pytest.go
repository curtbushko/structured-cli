package test

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// summaryPattern matches pytest summary lines like "3 passed, 1 failed in 0.15s"
var summaryPattern = regexp.MustCompile(`=+\s*(\d+\s+passed)?[,\s]*(\d+\s+failed)?[,\s]*(\d+\s+skipped)?[,\s]*(\d+\s+error)?[,\s]*.*?in\s+([\d.]+)s\s*=+`)

// verboseTestPattern matches verbose test output like "test_example.py::test_one PASSED"
var verboseTestPattern = regexp.MustCompile(`^(.+?)::(\w+)\s+(PASSED|FAILED|SKIPPED|ERROR)\s*`)

// PytestParser parses the output of 'pytest'.
// Pytest outputs test results in various formats.
type PytestParser struct {
	schema domain.Schema
}

// NewPytestParser creates a new PytestParser with the pytest schema.
func NewPytestParser() *PytestParser {
	return &PytestParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/pytest.json",
			"Pytest Output",
			"object",
			map[string]domain.PropertySchema{
				"passed":   {Type: "integer", Description: "Total number of tests that passed"},
				"failed":   {Type: "integer", Description: "Total number of tests that failed"},
				"skipped":  {Type: "integer", Description: "Total number of tests that were skipped"},
				"errors":   {Type: "integer", Description: "Total number of tests that had errors"},
				"duration": {Type: "number", Description: "Total time in seconds for the test run"},
				"tests":    {Type: "array", Description: "Individual test case results"},
			},
			[]string{"passed", "failed", "skipped", "tests"},
		),
	}
}

// Parse reads pytest output and returns structured data.
func (p *PytestParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &PytestResult{
		Passed:   0,
		Failed:   0,
		Skipped:  0,
		Errors:   0,
		Duration: 0,
		Tests:    []PytestCase{},
	}

	if len(data) == 0 {
		return domain.NewParseResult(result, raw, 0), nil
	}

	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		line := scanner.Text()

		// Try to parse verbose test output
		if tc := parseVerboseTestLine(line); tc != nil {
			result.Tests = append(result.Tests, *tc)
		}

		// Try to parse summary line
		if matches := summaryPattern.FindStringSubmatch(line); matches != nil {
			result.Passed = extractCount(matches[1])
			result.Failed = extractCount(matches[2])
			result.Skipped = extractCount(matches[3])
			result.Errors = extractCount(matches[4])
			if dur, err := strconv.ParseFloat(matches[5], 64); err == nil {
				result.Duration = dur
			}
		}
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// extractCount extracts the number from strings like "3 passed".
func extractCount(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	parts := strings.Fields(s)
	if len(parts) > 0 {
		if n, err := strconv.Atoi(parts[0]); err == nil {
			return n
		}
	}
	return 0
}

// parseVerboseTestLine parses a verbose test output line.
func parseVerboseTestLine(line string) *PytestCase {
	if matches := verboseTestPattern.FindStringSubmatch(line); matches != nil {
		outcome := strings.ToLower(matches[3])
		return &PytestCase{
			Name:    matches[2],
			File:    matches[1],
			Outcome: outcome,
		}
	}
	return nil
}

// Schema returns the JSON Schema for pytest output.
func (p *PytestParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *PytestParser) Matches(cmd string, subcommands []string) bool {
	if cmd == "pytest" {
		return true
	}

	// Handle python -m pytest or python3 -m pytest
	if (cmd == "python" || cmd == "python3") && len(subcommands) >= 2 {
		return subcommands[0] == "-m" && subcommands[1] == "pytest"
	}

	return false
}
