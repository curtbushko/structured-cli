package test

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// vitestTestSummaryPattern matches "Tests  X failed | Y skipped | Z passed (N)"
var vitestTestSummaryPattern = regexp.MustCompile(`Tests\s+(?:(\d+)\s+failed\s*\|?\s*)?(?:(\d+)\s+skipped\s*\|?\s*)?(?:(\d+)\s+passed)?\s*\((\d+)\)`)

// vitestDurationPattern matches "Duration  X.XXs" or "Duration  XXXms"
var vitestDurationPattern = regexp.MustCompile(`Duration\s+([\d.]+)(s|ms)`)

// VitestParser parses the output of 'vitest'.
// Vitest outputs test results in a standard format.
type VitestParser struct {
	schema domain.Schema
}

// NewVitestParser creates a new VitestParser with the vitest schema.
func NewVitestParser() *VitestParser {
	return &VitestParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/vitest.json",
			"Vitest Output",
			"object",
			map[string]domain.PropertySchema{
				"passed":   {Type: "integer", Description: "Total number of tests that passed"},
				"failed":   {Type: "integer", Description: "Total number of tests that failed"},
				"skipped":  {Type: "integer", Description: "Total number of tests that were skipped"},
				"duration": {Type: "number", Description: "Total time in seconds for the test run"},
				"files":    {Type: "array", Description: "Test file results"},
			},
			[]string{"passed", "failed", "skipped", "files"},
		),
	}
}

// Parse reads vitest output and returns structured data.
func (p *VitestParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &VitestResult{
		Passed:   0,
		Failed:   0,
		Skipped:  0,
		Duration: 0,
		Files:    []VitestFile{},
	}

	if len(data) == 0 {
		return domain.NewParseResult(result, raw, 0), nil
	}

	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		line := scanner.Text()

		// Parse test summary line
		if matches := vitestTestSummaryPattern.FindStringSubmatch(line); matches != nil {
			result.Failed = parseIntOrZero(matches[1])
			result.Skipped = parseIntOrZero(matches[2])
			result.Passed = parseIntOrZero(matches[3])
		}

		// Parse duration line
		if matches := vitestDurationPattern.FindStringSubmatch(line); matches != nil {
			duration, _ := strconv.ParseFloat(matches[1], 64)
			if matches[2] == "ms" {
				duration = duration / 1000.0
			}
			result.Duration = duration
		}
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// Schema returns the JSON Schema for vitest output.
func (p *VitestParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *VitestParser) Matches(cmd string, subcommands []string) bool {
	if cmd == "vitest" {
		return true
	}

	// Handle npx vitest, yarn vitest, pnpm vitest
	if (cmd == cmdNpx || cmd == cmdYarn || cmd == cmdPnpm) && len(subcommands) >= 1 {
		return subcommands[0] == "vitest"
	}

	return false
}
