package test

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// mochaPassingPattern matches "N passing (Xms)" or "N passing (Xs)" or "N passing (Xm)"
var mochaPassingPattern = regexp.MustCompile(`(\d+)\s+passing\s+\((\d+)(ms|s|m)\)`)

// mochaFailingPattern matches "N failing"
var mochaFailingPattern = regexp.MustCompile(`(\d+)\s+failing`)

// mochaPendingPattern matches "N pending"
var mochaPendingPattern = regexp.MustCompile(`(\d+)\s+pending`)

// MochaParser parses the output of 'mocha'.
// Mocha outputs test results in a standard format.
type MochaParser struct {
	schema domain.Schema
}

// NewMochaParser creates a new MochaParser with the mocha schema.
func NewMochaParser() *MochaParser {
	return &MochaParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/mocha.json",
			"Mocha Output",
			"object",
			map[string]domain.PropertySchema{
				"passed":   {Type: "integer", Description: "Total number of tests that passed"},
				"failed":   {Type: "integer", Description: "Total number of tests that failed"},
				"pending":  {Type: "integer", Description: "Total number of pending tests"},
				"duration": {Type: "number", Description: "Total time in milliseconds for the test run"},
				"suites":   {Type: "array", Description: "Test suite results"},
			},
			[]string{"passed", "failed", "pending", "suites"},
		),
	}
}

// Parse reads mocha output and returns structured data.
func (p *MochaParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &MochaResult{
		Passed:   0,
		Failed:   0,
		Pending:  0,
		Duration: 0,
		Suites:   []MochaSuite{},
	}

	if len(data) == 0 {
		return domain.NewParseResult(result, raw, 0), nil
	}

	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		line := scanner.Text()

		// Parse passing line (includes duration)
		if matches := mochaPassingPattern.FindStringSubmatch(line); matches != nil {
			result.Passed = parseIntOrZero(matches[1])
			duration, _ := strconv.ParseFloat(matches[2], 64)
			unit := matches[3]
			switch unit {
			case "s":
				duration = duration * 1000
			case "m":
				duration = duration * 60000
			}
			result.Duration = duration
		}

		// Parse failing line
		if matches := mochaFailingPattern.FindStringSubmatch(line); matches != nil {
			result.Failed = parseIntOrZero(matches[1])
		}

		// Parse pending line
		if matches := mochaPendingPattern.FindStringSubmatch(line); matches != nil {
			result.Pending = parseIntOrZero(matches[1])
		}
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// Schema returns the JSON Schema for mocha output.
func (p *MochaParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *MochaParser) Matches(cmd string, subcommands []string) bool {
	if cmd == "mocha" {
		return true
	}

	// Handle npx mocha or yarn mocha
	if (cmd == cmdNpx || cmd == cmdYarn || cmd == cmdPnpm) && len(subcommands) >= 1 {
		return subcommands[0] == "mocha"
	}

	return false
}
