package lint

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// mypyErrorWithCodePattern matches "file:line: severity: message  [error-code]"
var mypyErrorWithCodePattern = regexp.MustCompile(`^(.+):(\d+):\s*(error|warning|note):\s*(.+?)\s+\[([^\]]+)\]$`)

// mypyErrorNoCodePattern matches "file:line: severity: message"
var mypyErrorNoCodePattern = regexp.MustCompile(`^(.+):(\d+):\s*(error|warning|note):\s*(.+)$`)

// mypySummaryPattern matches "Found X errors" or "Success: no issues"
var mypySummaryPattern = regexp.MustCompile(`^(Found \d+ error|Success: no issues)`)

// MypyParser parses the output of 'mypy'.
// Mypy outputs errors in format: file:line: severity: message [error-code]
type MypyParser struct {
	schema domain.Schema
}

// NewMypyParser creates a new MypyParser with the mypy schema.
func NewMypyParser() *MypyParser {
	return &MypyParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/mypy.json",
			"Mypy Type Check Output",
			"object",
			map[string]domain.PropertySchema{
				"success": {Type: "boolean", Description: "Whether type checking completed without errors"},
				"errors":  {Type: "array", Description: "List of type errors found"},
				"summary": {Type: "string", Description: "Summary line from mypy output"},
			},
			[]string{"success", "errors", "summary"},
		),
	}
}

// Parse reads mypy output and returns structured data.
func (p *MypyParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &MypyResult{
		Success: true,
		Errors:  []MypyError{},
		Summary: "",
	}

	result.Errors, result.Summary = parseMypyOutput(raw)

	if len(result.Errors) > 0 {
		result.Success = false
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// parseMypyOutput extracts errors and summary from mypy output.
func parseMypyOutput(output string) ([]MypyError, string) {
	var errors []MypyError
	var summary string

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Check if this is a summary line
		if mypySummaryPattern.MatchString(line) {
			summary = line
			continue
		}

		if mypyErr := parseMypyErrorLine(line); mypyErr != nil {
			errors = append(errors, *mypyErr)
		}
	}

	return errors, summary
}

// parseMypyErrorLine attempts to parse a single line as a mypy error.
// Returns nil if the line is not a recognized error format.
func parseMypyErrorLine(line string) *MypyError {
	// Try file:line: severity: message [error-code] format first
	if matches := mypyErrorWithCodePattern.FindStringSubmatch(line); matches != nil {
		lineNum, _ := strconv.Atoi(matches[2])
		return &MypyError{
			File:     matches[1],
			Line:     lineNum,
			Severity: matches[3],
			Message:  strings.TrimSpace(matches[4]),
			Code:     matches[5],
		}
	}

	// Try file:line: severity: message format (no code)
	if matches := mypyErrorNoCodePattern.FindStringSubmatch(line); matches != nil {
		lineNum, _ := strconv.Atoi(matches[2])
		return &MypyError{
			File:     matches[1],
			Line:     lineNum,
			Severity: matches[3],
			Message:  strings.TrimSpace(matches[4]),
			Code:     "",
		}
	}

	return nil
}

// Schema returns the JSON Schema for mypy output.
func (p *MypyParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *MypyParser) Matches(cmd string, subcommands []string) bool {
	return cmd == "mypy"
}
