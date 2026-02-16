package golang

import (
	"bufio"
	"io"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// ModTidyParser parses the output of 'go mod tidy'.
// When run with -v flag, go mod tidy outputs added/removed modules.
// Without -v flag, it produces no output on success.
type ModTidyParser struct {
	schema domain.Schema
}

// NewModTidyParser creates a new ModTidyParser with the go-mod-tidy schema.
func NewModTidyParser() *ModTidyParser {
	return &ModTidyParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/go-mod-tidy.json",
			"Go Mod Tidy Output",
			"object",
			map[string]domain.PropertySchema{
				"added":   {Type: "array", Description: "Dependencies that were added"},
				"removed": {Type: "array", Description: "Dependencies that were removed"},
			},
			[]string{"added", "removed"},
		),
	}
}

// Parse reads go mod tidy output and returns structured data.
// For successful tidy with no changes, go mod tidy produces no output.
// With -v flag, it shows added/removed modules.
func (p *ModTidyParser) Parse(r io.Reader) (domain.ParseResult, error) {
	// Read all input
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &ModTidyResult{
		Added:   []string{},
		Removed: []string{},
	}

	// Parse lines for added/removed modules
	scanner := bufio.NewScanner(strings.NewReader(raw))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Check for "go: added module@version" format
		if strings.HasPrefix(line, "go: added ") {
			module := strings.TrimPrefix(line, "go: added ")
			result.Added = append(result.Added, module)
			continue
		}

		// Check for "go: removed module@version" format
		if strings.HasPrefix(line, "go: removed ") {
			module := strings.TrimPrefix(line, "go: removed ")
			result.Removed = append(result.Removed, module)
			continue
		}
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// Schema returns the JSON Schema for go mod tidy output.
func (p *ModTidyParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
// It matches "go" ["mod", "tidy", ...].
func (p *ModTidyParser) Matches(cmd string, subcommands []string) bool {
	if cmd != "go" {
		return false
	}
	if len(subcommands) < 2 {
		return false
	}
	return subcommands[0] == "mod" && subcommands[1] == "tidy"
}
