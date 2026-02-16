package fileops

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// DFParser parses the output of 'df' command.
type DFParser struct {
	schema domain.Schema
}

// NewDFParser creates a new DFParser with the df schema.
func NewDFParser() *DFParser {
	return &DFParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/df.json",
			"DF Output",
			"object",
			map[string]domain.PropertySchema{
				"filesystems": {Type: "array", Description: "Filesystem entries"},
			},
			[]string{"filesystems"},
		),
	}
}

// Parse reads df output and returns structured data.
func (p *DFParser) Parse(r io.Reader) (domain.ParseResult, error) {
	output := &DFOutput{
		Filesystems: []DFEntry{},
	}

	scanner := bufio.NewScanner(r)
	var rawBuilder strings.Builder
	var headerParsed bool
	var hasType bool

	for scanner.Scan() {
		line := scanner.Text()
		rawBuilder.WriteString(line)
		rawBuilder.WriteString("\n")

		if line == "" {
			continue
		}

		// Skip and analyze header line
		if !headerParsed {
			headerParsed = true
			hasType = strings.Contains(line, "Type")
			continue
		}

		entry := p.parseLine(line, hasType)
		if entry.Filesystem != "" {
			output.Filesystems = append(output.Filesystems, entry)
		}
	}

	if err := scanner.Err(); err != nil {
		return domain.NewParseResultWithError(err, rawBuilder.String(), 0), nil
	}

	return domain.NewParseResult(output, rawBuilder.String(), 0), nil
}

// Schema returns the JSON Schema for df output.
func (p *DFParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *DFParser) Matches(cmd string, _ []string) bool {
	return cmd == "df"
}

// parseLine parses a single line of df output.
// Format without type: Filesystem 1K-blocks Used Available Use% Mounted on
// Format with type: Filesystem Type 1K-blocks Used Available Use% Mounted on
func (p *DFParser) parseLine(line string, hasType bool) DFEntry {
	entry := DFEntry{}

	// Use regex to parse the line - handle variable spacing
	// Capture groups: filesystem, (type)?, size, used, available, use%, mount
	var pattern *regexp.Regexp
	if hasType {
		// With type column
		pattern = regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\d+)%\s+(.+)$`)
	} else {
		// Without type column
		pattern = regexp.MustCompile(`^(\S+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\d+)%\s+(.+)$`)
	}

	matches := pattern.FindStringSubmatch(line)
	if matches == nil {
		return entry
	}

	if hasType {
		entry.Filesystem = matches[1]
		entry.Type = matches[2]
		entry.Size, entry.SizeHuman = parseDFSize(matches[3])
		entry.Used, entry.UsedHuman = parseDFSize(matches[4])
		entry.Available, entry.AvailableHuman = parseDFSize(matches[5])
		entry.UsePercent, _ = strconv.ParseFloat(matches[6], 64)
		entry.MountedOn = matches[7]
	} else {
		entry.Filesystem = matches[1]
		entry.Size, entry.SizeHuman = parseDFSize(matches[2])
		entry.Used, entry.UsedHuman = parseDFSize(matches[3])
		entry.Available, entry.AvailableHuman = parseDFSize(matches[4])
		entry.UsePercent, _ = strconv.ParseFloat(matches[5], 64)
		entry.MountedOn = matches[6]
	}

	return entry
}

// parseDFSize parses a df size value which may be in K-blocks or human-readable format.
// Returns the size in bytes and the human-readable string (if applicable).
func parseDFSize(s string) (int64, string) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, ""
	}

	// Try direct numeric parse first (K-blocks)
	if size, err := strconv.ParseInt(s, 10, 64); err == nil {
		// df outputs in K-blocks by default
		return size * 1024, ""
	}

	// Human-readable format (e.g., "98G", "7.7G", "100M")
	humanStr := s
	size, _ := parseSize(s)

	return size, humanStr
}
