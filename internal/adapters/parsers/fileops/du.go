package fileops

import (
	"bufio"
	"io"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// DUParser parses the output of 'du' command.
type DUParser struct {
	schema domain.Schema
}

// NewDUParser creates a new DUParser with the du schema.
func NewDUParser() *DUParser {
	return &DUParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/du.json",
			"DU Output",
			"object",
			map[string]domain.PropertySchema{
				"entries":    {Type: "array", Description: "Disk usage entries"},
				"total":      {Type: "integer", Description: "Total disk usage in bytes"},
				"totalHuman": {Type: "string", Description: "Human-readable total size"},
			},
			[]string{"entries"},
		),
	}
}

// Parse reads du output and returns structured data.
func (p *DUParser) Parse(r io.Reader) (domain.ParseResult, error) {
	output := &DUOutput{
		Entries: []DUEntry{},
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

		entry := p.parseLine(line)

		// Check if this is the "total" line
		if entry.Path == "total" {
			output.Total = entry.Size
			output.TotalHuman = entry.SizeHuman
		} else {
			output.Entries = append(output.Entries, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return domain.NewParseResultWithError(err, rawBuilder.String(), 0), nil
	}

	return domain.NewParseResult(output, rawBuilder.String(), 0), nil
}

// Schema returns the JSON Schema for du output.
func (p *DUParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *DUParser) Matches(cmd string, _ []string) bool {
	return cmd == "du"
}

// parseLine parses a single line of du output.
// Format: size<tab>path
func (p *DUParser) parseLine(line string) DUEntry {
	entry := DUEntry{}

	// Split by tab
	parts := strings.SplitN(line, "\t", 2)
	if len(parts) < 2 {
		// Try splitting by spaces
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			parts = []string{fields[0], strings.Join(fields[1:], " ")}
		} else {
			return entry
		}
	}

	sizeStr := strings.TrimSpace(parts[0])
	entry.Path = strings.TrimSpace(parts[1])

	// Parse size
	entry.Size, entry.SizeHuman = parseSize(sizeStr)

	return entry
}

// parseSize parses a size string which may be numeric or human-readable.
// Returns the size in bytes and the human-readable string (if applicable).
func parseSize(s string) (int64, string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, ""
	}

	// Try direct numeric parse first
	if size, err := strconv.ParseInt(s, 10, 64); err == nil {
		return size, ""
	}

	// Human-readable format (e.g., "4.0K", "1.5M", "2.3G")
	humanStr := s
	var multiplier float64 = 1

	// Get last character for suffix
	lastChar := s[len(s)-1]
	numStr := s

	switch lastChar {
	case 'K':
		multiplier = 1024
		numStr = s[:len(s)-1]
	case 'M':
		multiplier = 1024 * 1024
		numStr = s[:len(s)-1]
	case 'G':
		multiplier = 1024 * 1024 * 1024
		numStr = s[:len(s)-1]
	case 'T':
		multiplier = 1024 * 1024 * 1024 * 1024
		numStr = s[:len(s)-1]
	case 'P':
		multiplier = 1024 * 1024 * 1024 * 1024 * 1024
		numStr = s[:len(s)-1]
	}

	// Parse the numeric part
	if num, err := strconv.ParseFloat(numStr, 64); err == nil {
		return int64(num * multiplier), humanStr
	}

	return 0, humanStr
}
