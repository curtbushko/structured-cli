package test

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// jestTestSummaryPattern matches "Tests: X failed, Y passed, Z total"
var jestTestSummaryPattern = regexp.MustCompile(`Tests:\s+(?:(\d+)\s+failed,\s*)?(?:(\d+)\s+skipped,\s*)?(?:(\d+)\s+passed,\s*)?(\d+)\s+total`)

// jestTimePattern matches "Time: X.XXX s" or "Time: XXX ms"
var jestTimePattern = regexp.MustCompile(`Time:\s+([\d.]+)\s*(s|ms)`)

// JestParser parses the output of 'jest'.
// Jest outputs test results in a standard format.
type JestParser struct {
	schema domain.Schema
}

// NewJestParser creates a new JestParser with the jest schema.
func NewJestParser() *JestParser {
	return &JestParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/jest.json",
			"Jest Output",
			"object",
			map[string]domain.PropertySchema{
				"passed":   {Type: "integer", Description: "Total number of tests that passed"},
				"failed":   {Type: "integer", Description: "Total number of tests that failed"},
				"skipped":  {Type: "integer", Description: "Total number of tests that were skipped"},
				"total":    {Type: "integer", Description: "Total number of tests"},
				"duration": {Type: "number", Description: "Total time in seconds for the test run"},
				"suites":   {Type: "array", Description: "Test suite results"},
			},
			[]string{"passed", "failed", "skipped", "total", "suites"},
		),
	}
}

// Parse reads jest output and returns structured data.
func (p *JestParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &JestResult{
		Passed:   0,
		Failed:   0,
		Skipped:  0,
		Total:    0,
		Duration: 0,
		Suites:   []JestSuite{},
	}

	if len(data) == 0 {
		return domain.NewParseResult(result, raw, 0), nil
	}

	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		line := scanner.Text()

		// Parse test summary line
		if matches := jestTestSummaryPattern.FindStringSubmatch(line); matches != nil {
			result.Failed = parseIntOrZero(matches[1])
			result.Skipped = parseIntOrZero(matches[2])
			result.Passed = parseIntOrZero(matches[3])
			result.Total = parseIntOrZero(matches[4])
		}

		// Parse time line
		if matches := jestTimePattern.FindStringSubmatch(line); matches != nil {
			duration, _ := strconv.ParseFloat(matches[1], 64)
			if matches[2] == "ms" {
				duration = duration / 1000.0
			}
			result.Duration = duration
		}
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// parseIntOrZero parses a string to int, returning 0 on failure.
func parseIntOrZero(s string) int {
	if s == "" {
		return 0
	}
	n, _ := strconv.Atoi(s)
	return n
}

// Schema returns the JSON Schema for jest output.
func (p *JestParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *JestParser) Matches(cmd string, subcommands []string) bool {
	if cmd == "jest" {
		return true
	}

	// Handle npx jest or yarn jest
	if (cmd == cmdNpx || cmd == cmdYarn) && len(subcommands) >= 1 {
		return subcommands[0] == "jest"
	}

	return false
}
