package golang

import (
	"bufio"
	"io"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// FmtParser parses the output of 'gofmt -l' or 'go fmt'.
// These commands list files that need formatting (or were formatted).
type FmtParser struct {
	schema domain.Schema
}

// NewFmtParser creates a new FmtParser with the gofmt schema.
func NewFmtParser() *FmtParser {
	return &FmtParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/gofmt.json",
			"Go Fmt Output",
			"object",
			map[string]domain.PropertySchema{
				"unformatted": {Type: "array", Description: "List of files that need formatting"},
			},
			[]string{"unformatted"},
		),
	}
}

// Parse reads gofmt output and returns structured data.
// For properly formatted code, gofmt -l produces no output.
func (p *FmtParser) Parse(r io.Reader) (domain.ParseResult, error) {
	// Read all input
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	fmt := &FmtResult{
		Unformatted: []string{},
	}

	// Parse file paths from output
	fmt.Unformatted = parseFmtLines(raw)

	return domain.NewParseResult(fmt, raw, 0), nil
}

// parseFmtLines scans gofmt output and extracts file paths.
func parseFmtLines(output string) []string {
	var files []string

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		files = append(files, line)
	}

	return files
}

// Schema returns the JSON Schema for gofmt output.
func (p *FmtParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
// Matches both 'gofmt' and 'go fmt' commands.
func (p *FmtParser) Matches(cmd string, subcommands []string) bool {
	// Match 'gofmt' command
	if cmd == "gofmt" {
		return true
	}

	// Match 'go fmt' command
	if cmd == "go" && len(subcommands) > 0 && subcommands[0] == "fmt" {
		return true
	}

	return false
}
