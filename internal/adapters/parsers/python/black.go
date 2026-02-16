package python

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Regex patterns for parsing black --check output.
var (
	// wouldReformatPattern matches "would reformat path/to/file.py"
	wouldReformatPattern = regexp.MustCompile(`^would reformat (.+)$`)

	// errorPattern matches "error: cannot format file.py: message"
	errorPattern = regexp.MustCompile(`^error: cannot format (.+?):(.+)$`)

	// summaryPattern matches the summary line
	// Examples:
	// "3 files would be left unchanged."
	// "2 files would be reformatted, 3 files would be left unchanged."
	// "1 file would be reformatted, 1 file would fail to reformat."
	summaryReformattedPattern = regexp.MustCompile(`(\d+) files? would be reformatted`)
	summaryUnchangedPattern   = regexp.MustCompile(`(\d+) files? would be left unchanged`)
)

// BlackParser parses the output of 'black --check'.
type BlackParser struct {
	schema domain.Schema
}

// NewBlackParser creates a new BlackParser with the black schema.
func NewBlackParser() *BlackParser {
	return &BlackParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/black.json",
			"Black Check Output",
			"object",
			map[string]domain.PropertySchema{
				"success":              {Type: "boolean", Description: "Whether all files are properly formatted"},
				"files_checked":        {Type: "integer", Description: "Number of files checked"},
				"files_would_reformat": {Type: "array", Description: "Files that would be reformatted"},
				"files_unchanged":      {Type: "integer", Description: "Number of files already formatted"},
				"errors":               {Type: "array", Description: "Errors encountered during checking"},
			},
			[]string{"success", "files_would_reformat"},
		),
	}
}

// Parse reads black --check output and returns structured data.
func (p *BlackParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &BlackResult{
		Success:            true,
		FilesWouldReformat: []string{},
		Errors:             []BlackError{},
	}

	parseBlackOutput(raw, result)

	// Mark as unsuccessful if any files need reformatting or errors occurred
	if len(result.FilesWouldReformat) > 0 || len(result.Errors) > 0 {
		result.Success = false
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// parseBlackOutput extracts black check information from the output.
func parseBlackOutput(output string, result *BlackResult) {
	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		// Parse files that would be reformatted
		if matches := wouldReformatPattern.FindStringSubmatch(line); matches != nil {
			result.FilesWouldReformat = append(result.FilesWouldReformat, matches[1])
			continue
		}

		// Parse errors
		if matches := errorPattern.FindStringSubmatch(line); matches != nil {
			result.Errors = append(result.Errors, BlackError{
				File:    matches[1],
				Message: strings.TrimSpace(matches[2]),
			})
			continue
		}

		// Parse summary lines
		if matches := summaryUnchangedPattern.FindStringSubmatch(line); matches != nil {
			result.FilesUnchanged, _ = strconv.Atoi(matches[1])
		}

		// Note: We can calculate files checked from files_would_reformat + files_unchanged
		// But we'll parse from summary if available
		if matches := summaryReformattedPattern.FindStringSubmatch(line); matches != nil {
			reformatted, _ := strconv.Atoi(matches[1])
			result.FilesChecked = reformatted + result.FilesUnchanged
		}
	}

	// Calculate files checked if not set from summary
	if result.FilesChecked == 0 {
		result.FilesChecked = len(result.FilesWouldReformat) + result.FilesUnchanged
	}
}

// Schema returns the JSON Schema for black check output.
func (p *BlackParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *BlackParser) Matches(cmd string, subcommands []string) bool {
	if cmd != blackCommand {
		return false
	}

	// Must have --check flag somewhere in subcommands
	for _, sub := range subcommands {
		if sub == "--check" {
			return true
		}
	}

	return false
}
