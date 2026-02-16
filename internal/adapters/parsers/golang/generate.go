package golang

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// generateDirectivePattern matches "file.go:line: running command" format
// from go generate -v output.
var generateDirectivePattern = regexp.MustCompile(`^(.+\.go):(\d+):\s*running\s+(.+)$`)

// GenerateParser parses the output of 'go generate'.
// Go generate with -v flag outputs lines like: "file.go:line: running command"
type GenerateParser struct {
	schema domain.Schema
}

// NewGenerateParser creates a new GenerateParser with the go-generate schema.
func NewGenerateParser() *GenerateParser {
	return &GenerateParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/go-generate.json",
			"Go Generate Output",
			"object",
			map[string]domain.PropertySchema{
				"success":   {Type: "boolean", Description: "Whether generation completed without errors"},
				"generated": {Type: "array", Description: "List of files that ran generate directives"},
			},
			[]string{"success", "generated"},
		),
	}
}

// Parse reads go generate output and returns structured data.
// For successful generation, go generate -v outputs the files and directives run.
func (p *GenerateParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &GenerateResult{
		Success:   true,
		Generated: []string{},
	}

	// Parse directive lines from verbose output
	result.Generated = parseGenerateDirectives(raw)

	return domain.NewParseResult(result, raw, 0), nil
}

// parseGenerateDirectives scans go generate -v output and extracts file paths.
func parseGenerateDirectives(output string) []string {
	var files []string

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if file := parseGenerateDirectiveLine(line); file != "" {
			files = append(files, file)
		}
	}

	return files
}

// parseGenerateDirectiveLine attempts to parse a single line as a generate directive.
// Returns the file path if matched, empty string otherwise.
func parseGenerateDirectiveLine(line string) string {
	if matches := generateDirectivePattern.FindStringSubmatch(line); matches != nil {
		return matches[1]
	}
	return ""
}

// Schema returns the JSON Schema for go generate output.
func (p *GenerateParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *GenerateParser) Matches(cmd string, subcommands []string) bool {
	if cmd != "go" {
		return false
	}
	if len(subcommands) == 0 {
		return false
	}
	return subcommands[0] == "generate"
}
