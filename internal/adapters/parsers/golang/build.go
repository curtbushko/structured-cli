package golang

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// errorWithColumnPattern matches "file.go:line:column: message"
var errorWithColumnPattern = regexp.MustCompile(`^([^:]+):(\d+):(\d+):\s*(.+)$`)

// errorNoColumnPattern matches "file.go:line: message"
var errorNoColumnPattern = regexp.MustCompile(`^([^:]+):(\d+):\s*(.+)$`)

// BuildParser parses the output of 'go build'.
// Go build outputs nothing on success (exit code 0).
type BuildParser struct {
	schema domain.Schema
}

// NewBuildParser creates a new BuildParser with the go-build schema.
func NewBuildParser() *BuildParser {
	return &BuildParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/go-build.json",
			"Go Build Output",
			"object",
			map[string]domain.PropertySchema{
				"success":  {Type: "boolean", Description: "Whether the build succeeded"},
				"packages": {Type: "array", Description: "Packages that were built"},
				"errors":   {Type: "array", Description: "Compilation errors if build failed"},
			},
			[]string{"success", "packages", "errors"},
		),
	}
}

// Parse reads go build output and returns structured data.
// For successful builds, go build produces no output.
func (p *BuildParser) Parse(r io.Reader) (domain.ParseResult, error) {
	// Read all input
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	build := &Build{
		Success:  true,
		Packages: []string{},
		Errors:   []BuildError{},
	}

	// Parse error lines from output
	build.Errors = parseErrorLines(raw)

	// Success is false when errors are present
	if len(build.Errors) > 0 {
		build.Success = false
	}

	return domain.NewParseResult(build, raw, 0), nil
}

// parseErrorLines scans build output and extracts error information.
func parseErrorLines(output string) []BuildError {
	var errors []BuildError

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		if buildErr := parseErrorLine(line); buildErr != nil {
			errors = append(errors, *buildErr)
		}
	}

	return errors
}

// parseErrorLine attempts to parse a single line as a build error.
// Returns nil if the line is not a recognized error format.
func parseErrorLine(line string) *BuildError {
	// Try file:line:column: message format first
	if matches := errorWithColumnPattern.FindStringSubmatch(line); matches != nil {
		lineNum, _ := strconv.Atoi(matches[2])
		colNum, _ := strconv.Atoi(matches[3])
		return &BuildError{
			File:    matches[1],
			Line:    lineNum,
			Column:  colNum,
			Message: matches[4],
		}
	}

	// Try file:line: message format (no column)
	if matches := errorNoColumnPattern.FindStringSubmatch(line); matches != nil {
		lineNum, _ := strconv.Atoi(matches[2])
		return &BuildError{
			File:    matches[1],
			Line:    lineNum,
			Column:  0,
			Message: matches[3],
		}
	}

	// Check for package-level errors (start with "package ")
	if strings.HasPrefix(line, "package ") {
		return &BuildError{
			File:    "",
			Line:    0,
			Column:  0,
			Message: line,
		}
	}

	return nil
}

// Schema returns the JSON Schema for go build output.
func (p *BuildParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *BuildParser) Matches(cmd string, subcommands []string) bool {
	if cmd != "go" {
		return false
	}
	if len(subcommands) == 0 {
		return false
	}
	return subcommands[0] == "build"
}
