package lint

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// prettierFilePattern matches "[warn] filepath" format
var prettierFilePattern = regexp.MustCompile(`^\[warn\]\s+(.+)$`)

// PrettierParser parses the output of 'prettier --check'.
// Prettier outputs nothing on success, or lists files that need formatting.
type PrettierParser struct {
	schema domain.Schema
}

// NewPrettierParser creates a new PrettierParser with the prettier schema.
func NewPrettierParser() *PrettierParser {
	return &PrettierParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/prettier.json",
			"Prettier Check Output",
			"object",
			map[string]domain.PropertySchema{
				"success":     {Type: "boolean", Description: "Whether all files are properly formatted"},
				"unformatted": {Type: "array", Description: "List of files that need formatting"},
			},
			[]string{"success", "unformatted"},
		),
	}
}

// Parse reads prettier --check output and returns structured data.
func (p *PrettierParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &PrettierResult{
		Success:     true,
		Unformatted: []string{},
	}

	result.Unformatted = parsePrettierOutput(raw)

	if len(result.Unformatted) > 0 {
		result.Success = false
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// parsePrettierOutput extracts unformatted file paths from prettier output.
func parsePrettierOutput(output string) []string {
	var files []string

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Match [warn] filepath pattern
		if matches := prettierFilePattern.FindStringSubmatch(line); matches != nil {
			filePath := strings.TrimSpace(matches[1])
			// Skip summary messages that contain "Code style" or other non-file patterns
			if !strings.Contains(filePath, "Code style") && !strings.Contains(filePath, "Run Prettier") {
				files = append(files, filePath)
			}
		}
	}

	return files
}

// Schema returns the JSON Schema for prettier output.
func (p *PrettierParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *PrettierParser) Matches(cmd string, subcommands []string) bool {
	return cmd == "prettier"
}
