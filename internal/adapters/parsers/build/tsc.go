package build

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// tscErrorPattern matches "file(line,column): error TSxxxx: message"
// The file path can be Unix or Windows style (including drive letters like C:\)
var tscErrorPattern = regexp.MustCompile(`^(.+)\((\d+),(\d+)\):\s+error\s+(TS\d+):\s+(.+)$`)

// TSCParser parses the output of 'tsc' (TypeScript compiler).
// TypeScript outputs nothing on success (exit code 0).
type TSCParser struct {
	schema domain.Schema
}

// NewTSCParser creates a new TSCParser with the tsc schema.
func NewTSCParser() *TSCParser {
	return &TSCParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/tsc.json",
			"TypeScript Compiler Output",
			"object",
			map[string]domain.PropertySchema{
				"success": {Type: "boolean", Description: "Whether the compilation succeeded"},
				"errors":  {Type: "array", Description: "TypeScript compilation errors"},
			},
			[]string{"success", "errors"},
		),
	}
}

// Parse reads tsc output and returns structured data.
// For successful builds, tsc produces no output.
func (p *TSCParser) Parse(r io.Reader) (domain.ParseResult, error) {
	// Read all input
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	tscResult := &TSCResult{
		Success: true,
		Errors:  []TSCError{},
	}

	// Parse error lines from output
	tscResult.Errors = parseTSCErrorLines(raw)

	// Success is false when errors are present
	if len(tscResult.Errors) > 0 {
		tscResult.Success = false
	}

	return domain.NewParseResult(tscResult, raw, 0), nil
}

// parseTSCErrorLines scans tsc output and extracts error information.
func parseTSCErrorLines(output string) []TSCError {
	var errors []TSCError

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		if tscErr := parseTSCErrorLine(line); tscErr != nil {
			errors = append(errors, *tscErr)
		}
	}

	return errors
}

// parseTSCErrorLine attempts to parse a single line as a tsc error.
// Returns nil if the line is not a recognized error format.
func parseTSCErrorLine(line string) *TSCError {
	// Match format: file(line,column): error TSxxxx: message
	if matches := tscErrorPattern.FindStringSubmatch(line); matches != nil {
		lineNum, _ := strconv.Atoi(matches[2])
		colNum, _ := strconv.Atoi(matches[3])
		return &TSCError{
			File:    matches[1],
			Line:    lineNum,
			Column:  colNum,
			Code:    matches[4],
			Message: matches[5],
		}
	}

	return nil
}

// Schema returns the JSON Schema for tsc output.
func (p *TSCParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
// The tsc parser matches the "tsc" command with any subcommands/flags.
func (p *TSCParser) Matches(cmd string, subcommands []string) bool {
	return cmd == "tsc"
}
