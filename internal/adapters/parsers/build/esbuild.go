package build

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// esbuildErrorPattern matches "> file:line:column: error|warning: message"
// The file path can be Unix or Windows style (including drive letters like C:\)
// Format: > file:line:column: error: message
// Format: > file:line:column: warning: message
var esbuildErrorPattern = regexp.MustCompile(`^>\s+(.+):(\d+):(\d+):\s+(error|warning):\s+(.+)$`)

// esbuildDurationPattern matches "Done in Xms" or "Done in X.Xs"
var esbuildDurationPattern = regexp.MustCompile(`Done in (\d+(?:\.\d+)?)(ms|s)`)

// ESBuildParser parses the output of 'esbuild' bundler.
type ESBuildParser struct {
	schema domain.Schema
}

// NewESBuildParser creates a new ESBuildParser with the esbuild schema.
func NewESBuildParser() *ESBuildParser {
	return &ESBuildParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/esbuild.json",
			"esbuild Bundler Output",
			"object",
			map[string]domain.PropertySchema{
				"success":  {Type: "boolean", Description: "Whether the build succeeded"},
				"errors":   {Type: "array", Description: "Build errors"},
				"warnings": {Type: "array", Description: "Build warnings"},
				"outputs":  {Type: "array", Description: "Generated output files"},
				"duration": {Type: "number", Description: "Build duration in milliseconds"},
			},
			[]string{"success", "errors", "warnings", "outputs"},
		),
	}
}

// Parse reads esbuild output and returns structured data.
func (p *ESBuildParser) Parse(r io.Reader) (domain.ParseResult, error) {
	// Read all input
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &ESBuildResult{
		Success:  true,
		Errors:   []ESBuildError{},
		Warnings: []ESBuildWarning{},
		Outputs:  []ESBuildOutput{},
		Duration: 0,
	}

	// Parse error/warning lines and duration from output
	parseESBuildOutput(raw, result)

	// Success is false when errors are present
	if len(result.Errors) > 0 {
		result.Success = false
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// parseESBuildOutput scans esbuild output and extracts errors, warnings, and duration.
func parseESBuildOutput(output string, result *ESBuildResult) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Try to parse as error or warning
		if matches := esbuildErrorPattern.FindStringSubmatch(line); matches != nil {
			file := matches[1]
			lineNum, _ := strconv.Atoi(matches[2])
			colNum, _ := strconv.Atoi(matches[3])
			msgType := matches[4]
			message := matches[5]

			switch msgType {
			case "error":
				result.Errors = append(result.Errors, ESBuildError{
					File:    file,
					Line:    lineNum,
					Column:  colNum,
					Message: message,
				})
			case "warning":
				result.Warnings = append(result.Warnings, ESBuildWarning{
					File:    file,
					Line:    lineNum,
					Column:  colNum,
					Message: message,
				})
			}
			continue
		}

		// Try to parse duration
		if matches := esbuildDurationPattern.FindStringSubmatch(line); matches != nil {
			value, _ := strconv.ParseFloat(matches[1], 64)
			unit := matches[2]
			if unit == "s" {
				value *= 1000 // Convert seconds to milliseconds
			}
			result.Duration = value
		}
	}
}

// Schema returns the JSON Schema for esbuild output.
func (p *ESBuildParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
// The esbuild parser matches the "esbuild" command with any subcommands/flags.
func (p *ESBuildParser) Matches(cmd string, subcommands []string) bool {
	return cmd == "esbuild"
}
