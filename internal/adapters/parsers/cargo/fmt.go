package cargo

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// diffInPattern matches "Diff in <path>:<line>:"
var diffInPattern = regexp.MustCompile(`^Diff in (.+):(\d+):`)

// FmtParser parses the output of 'cargo fmt --check'.
type FmtParser struct {
	schema domain.Schema
}

// NewFmtParser creates a new FmtParser with the cargo fmt schema.
func NewFmtParser() *FmtParser {
	return &FmtParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/cargo-fmt.json",
			"Cargo Fmt Output",
			"object",
			map[string]domain.PropertySchema{
				"success": {Type: "boolean", Description: "Whether all files are correctly formatted"},
				"files":   {Type: "array", Description: "Files with formatting issues"},
			},
			[]string{"success", "files"},
		),
	}
}

// Parse reads cargo fmt output and returns structured data.
func (p *FmtParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &FmtResult{
		Success: true,
		Files:   []FmtFile{},
	}

	// Track files we've already seen to avoid duplicates
	seenFiles := make(map[string]bool)

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if trimmed == "" {
			continue
		}

		// Check for "Diff in" pattern
		if matches := diffInPattern.FindStringSubmatch(line); matches != nil {
			path := matches[1]
			if !seenFiles[path] {
				seenFiles[path] = true
				result.Files = append(result.Files, FmtFile{
					Path: path,
				})
				result.Success = false
			}
			continue
		}

		// Check for plain file paths (cargo fmt --check --list output)
		if strings.HasSuffix(trimmed, ".rs") && !strings.HasPrefix(trimmed, "-") && !strings.HasPrefix(trimmed, "+") {
			if !seenFiles[trimmed] {
				seenFiles[trimmed] = true
				result.Files = append(result.Files, FmtFile{
					Path: trimmed,
				})
				result.Success = false
			}
		}
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// Schema returns the JSON Schema for cargo fmt output.
func (p *FmtParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *FmtParser) Matches(cmd string, subcommands []string) bool {
	if cmd != cmdCargo {
		return false
	}
	if len(subcommands) == 0 {
		return false
	}
	return subcommands[0] == "fmt"
}
