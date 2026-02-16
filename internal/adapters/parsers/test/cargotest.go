package test

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// cargoTestResultPattern matches "test result: ok. X passed; Y failed; Z ignored; W measured; V filtered out; finished in X.XXs"
var cargoTestResultPattern = regexp.MustCompile(`test result:.*?(\d+)\s+passed;\s*(\d+)\s+failed;\s*(\d+)\s+ignored;\s*(\d+)\s+measured;\s*(\d+)\s+filtered out.*?finished in\s+([\d.]+)s`)

// cargoTestLinePattern matches "test name::path ... ok" or "test name::path ... FAILED" or "test name::path ... ignored"
var cargoTestLinePattern = regexp.MustCompile(`^test\s+(\S+)\s+\.\.\.\s+(ok|FAILED|ignored)$`)

// CargoTestParser parses the output of 'cargo test'.
// Cargo test outputs test results in a standard format.
type CargoTestParser struct {
	schema domain.Schema
}

// NewCargoTestParser creates a new CargoTestParser with the cargo-test schema.
func NewCargoTestParser() *CargoTestParser {
	return &CargoTestParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/cargo-test.json",
			"Cargo Test Output",
			"object",
			map[string]domain.PropertySchema{
				"passed":   {Type: "integer", Description: "Total number of tests that passed"},
				"failed":   {Type: "integer", Description: "Total number of tests that failed"},
				"ignored":  {Type: "integer", Description: "Total number of tests that were ignored"},
				"measured": {Type: "integer", Description: "Number of benchmark tests measured"},
				"filtered": {Type: "integer", Description: "Number of tests filtered out"},
				"duration": {Type: "number", Description: "Total time in seconds for the test run"},
				"tests":    {Type: "array", Description: "Individual test case results"},
			},
			[]string{"passed", "failed", "ignored", "tests"},
		),
	}
}

// Parse reads cargo test output and returns structured data.
func (p *CargoTestParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &CargoTestResult{
		Passed:   0,
		Failed:   0,
		Ignored:  0,
		Measured: 0,
		Filtered: 0,
		Duration: 0,
		Tests:    []CargoTestCase{},
	}

	if len(data) == 0 {
		return domain.NewParseResult(result, raw, 0), nil
	}

	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		line := scanner.Text()

		// Parse individual test line
		if matches := cargoTestLinePattern.FindStringSubmatch(line); matches != nil {
			result.Tests = append(result.Tests, CargoTestCase{
				Name:   matches[1],
				Status: matches[2],
			})
		}

		// Parse summary line
		if matches := cargoTestResultPattern.FindStringSubmatch(line); matches != nil {
			result.Passed = parseIntOrZero(matches[1])
			result.Failed = parseIntOrZero(matches[2])
			result.Ignored = parseIntOrZero(matches[3])
			result.Measured = parseIntOrZero(matches[4])
			result.Filtered = parseIntOrZero(matches[5])
			if dur, err := strconv.ParseFloat(matches[6], 64); err == nil {
				result.Duration = dur
			}
		}
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// Schema returns the JSON Schema for cargo test output.
func (p *CargoTestParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *CargoTestParser) Matches(cmd string, subcommands []string) bool {
	if cmd != "cargo" {
		return false
	}
	if len(subcommands) == 0 {
		return false
	}
	return subcommands[0] == "test"
}
