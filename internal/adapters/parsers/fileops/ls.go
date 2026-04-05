package fileops

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// LSParser parses the output of 'ls' command.
type LSParser struct {
	schema domain.Schema
}

// NewLSParser creates a new LSParser with the ls schema.
func NewLSParser() *LSParser {
	return &LSParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/ls.json",
			"LS Output",
			"object",
			map[string]domain.PropertySchema{
				"entries": {Type: "array", Description: "List of directory entries"},
				"total":   {Type: "integer", Description: "Total block size"},
			},
			[]string{"entries"},
		),
	}
}

// Parse reads ls output and returns structured data.
func (p *LSParser) Parse(r io.Reader) (domain.ParseResult, error) {
	output := &LSOutput{
		Entries: []LSEntry{},
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

		// Check for "total N" line
		if strings.HasPrefix(line, "total ") {
			total, err := strconv.ParseInt(strings.TrimPrefix(line, "total "), 10, 64)
			if err == nil {
				output.Total = total
			}
			continue
		}

		entry := p.parseLine(line)
		output.Entries = append(output.Entries, entry)
	}

	if err := scanner.Err(); err != nil {
		return domain.NewParseResultWithError(err, rawBuilder.String(), 0), nil
	}

	return domain.NewParseResult(output, rawBuilder.String(), 0), nil
}

// Schema returns the JSON Schema for ls output.
func (p *LSParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *LSParser) Matches(cmd string, _ []string) bool {
	return cmd == "ls"
}

// parseLine parses a single line of ls output.
func (p *LSParser) parseLine(line string) LSEntry {
	// Check if it's a long format line (starts with permission string)
	if len(line) >= 10 && isPermissionChar(line[0]) {
		return p.parseLongFormat(line)
	}

	// Simple format - just the filename
	// Note: In JSON mode, -l is auto-added so this branch is rarely hit
	return LSEntry{
		Name: line,
		Type: TypeFile, // Default to file for simple listing (rare in JSON mode)
	}
}

// parseLongFormat parses a line from "ls -l" output.
// Format: drwxr-xr-x  5 user group  4096 2024-01-15 10:30 name
func (p *LSParser) parseLongFormat(line string) LSEntry {
	entry := LSEntry{}

	// Determine type from first character
	entry.Type = fileTypeFromChar(line[0])

	// Extract permissions (skip first char which is type)
	entry.Permissions = line[1:10]

	// Parse the rest of the fields using regex to handle variable spacing
	pattern := regexp.MustCompile(`^[dlbcps-][rwxstST-]{9}\s+(\d+)\s+(\S+)\s+(\S+)\s+(\S+)\s+(\d{4}-\d{2}-\d{2}\s+\d{2}:\d{2})\s+(.+)$`)
	matches := pattern.FindStringSubmatch(line)

	if matches == nil {
		// Fallback: just extract the name from the end
		return p.parseFallback(line, entry)
	}

	return p.parseMatchedEntry(matches, entry)
}

// parseMatchedEntry extracts entry data from regex matches.
func (p *LSParser) parseMatchedEntry(matches []string, entry LSEntry) LSEntry {
	entry.Links, _ = strconv.Atoi(matches[1])
	entry.Owner = matches[2]
	entry.Group = matches[3]
	sizeStr := matches[4]
	entry.Modified = matches[5]
	name := matches[6]

	// Parse size (handle block/char device format "8, 0")
	if !strings.Contains(sizeStr, ",") {
		entry.Size, _ = strconv.ParseInt(sizeStr, 10, 64)
	}

	// Handle symlinks: "name -> target"
	entry.Name, entry.Target = p.parseSymlinkName(name, entry.Type)

	return entry
}

// parseSymlinkName parses a name field that may contain symlink target.
func (p *LSParser) parseSymlinkName(name, fileType string) (string, string) {
	if fileType == TypeSymlink && strings.Contains(name, " -> ") {
		parts := strings.SplitN(name, " -> ", 2)
		if len(parts) > 1 {
			return parts[0], parts[1]
		}
		return parts[0], ""
	}
	return name, ""
}

// parseFallback extracts just the name from a line when regex doesn't match.
func (p *LSParser) parseFallback(line string, entry LSEntry) LSEntry {
	fields := strings.Fields(line)
	if len(fields) > 0 {
		entry.Name = fields[len(fields)-1]
	}
	return entry
}

// isPermissionChar checks if a character is a valid file type indicator.
func isPermissionChar(c byte) bool {
	return c == '-' || c == 'd' || c == 'l' || c == 'b' || c == 'c' || c == 'p' || c == 's'
}

// fileTypeFromChar converts the first character of ls -l output to a type.
func fileTypeFromChar(c byte) string {
	switch c {
	case 'd':
		return TypeDirectory
	case 'l':
		return TypeSymlink
	case 's':
		return TypeSocket
	case 'p':
		return TypeFIFO
	case 'b':
		return TypeBlock
	case 'c':
		return TypeChar
	case '-':
		return TypeFile
	default:
		return TypeUnknown
	}
}
