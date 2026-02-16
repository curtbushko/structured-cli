package fileops

import (
	"bufio"
	"io"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

const totalFileName = "total"

// WCParser parses the output of 'wc' command.
type WCParser struct {
	schema domain.Schema
}

// NewWCParser creates a new WCParser with the wc schema.
func NewWCParser() *WCParser {
	return &WCParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/wc.json",
			"WC Output",
			"object",
			map[string]domain.PropertySchema{
				"files": {Type: "array", Description: "File statistics"},
				"total": {Type: "object", Description: "Total statistics"},
			},
			[]string{"files"},
		),
	}
}

// Parse reads wc output and returns structured data.
func (p *WCParser) Parse(r io.Reader) (domain.ParseResult, error) {
	output := &WCOutput{
		Files: []WCStats{},
	}

	scanner := bufio.NewScanner(r)
	var rawBuilder strings.Builder

	for scanner.Scan() {
		line := scanner.Text()
		rawBuilder.WriteString(line)
		rawBuilder.WriteString("\n")

		if line == "" {
			continue
		}

		stats := p.parseLine(line)

		// Check if this is the "total" line
		if stats.File == totalFileName {
			totalStats := stats
			totalStats.File = totalFileName
			output.Total = &totalStats
		} else {
			output.Files = append(output.Files, stats)
		}
	}

	if err := scanner.Err(); err != nil {
		return domain.NewParseResultWithError(err, rawBuilder.String(), 0), nil
	}

	return domain.NewParseResult(output, rawBuilder.String(), 0), nil
}

// Schema returns the JSON Schema for wc output.
func (p *WCParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *WCParser) Matches(cmd string, _ []string) bool {
	return cmd == "wc"
}

// parseLine parses a single line of wc output.
// Formats:
//   - lines words bytes filename (full wc)
//   - lines words bytes (stdin, no filename)
//   - value filename (single stat like wc -l)
//   - value (single stat from stdin)
func (p *WCParser) parseLine(line string) WCStats {
	stats := WCStats{}
	fields := strings.Fields(line)

	if len(fields) == 0 {
		return stats
	}

	// Parse numeric values
	nums := []int{}
	var filename string

	for _, f := range fields {
		if n, err := strconv.Atoi(f); err == nil {
			nums = append(nums, n)
		} else {
			// Non-numeric field is the filename
			filename = f
		}
	}

	stats.File = filename

	// Assign values based on count
	switch len(nums) {
	case 1:
		// Single value (like wc -l)
		stats.Lines = nums[0]
	case 2:
		// Two values: lines, words or lines, bytes
		stats.Lines = nums[0]
		stats.Words = nums[1]
	case 3:
		// Full output: lines, words, bytes
		stats.Lines = nums[0]
		stats.Words = nums[1]
		stats.Bytes = nums[2]
	case 4:
		// Some versions include chars too: lines, words, chars, bytes
		stats.Lines = nums[0]
		stats.Words = nums[1]
		stats.Chars = nums[2]
		stats.Bytes = nums[3]
	}

	return stats
}
