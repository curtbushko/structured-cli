// Package fileops provides parsers for file operation commands (ls, find, grep, etc.).
//
// # Compact Format Design
//
// The grep parser uses a compact array-based JSON format optimized for token efficiency
// when consumed by LLMs. Traditional object-based formats repeat keys for every match:
//
//	{"file": "foo.go", "line": 42, "content": "..."}  // 45+ tokens per match
//
// The compact format uses positional arrays instead:
//
//	["foo.go", 3, [[42, "..."], [57, "..."]]]  // ~15 tokens per file group
//
// This reduces token usage by 60-80% for typical grep results with many matches,
// significantly reducing API costs and context window consumption.
//
// See types.go for MatchTuple and FileMatchGroup which implement the array marshaling.
package fileops

import (
	"bufio"
	"io"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// GrepParser parses the output of 'grep' command into a compact JSON format.
// It supports grep, egrep, and fgrep commands with various output formats.
type GrepParser struct {
	schema domain.Schema
}

// NewGrepParser creates a new GrepParser with the grep schema.
func NewGrepParser() *GrepParser {
	return &GrepParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/grep.json",
			"Grep Output",
			"object",
			map[string]domain.PropertySchema{
				"total":     {Type: "integer", Description: "Total number of matches across all files"},
				"files":     {Type: "integer", Description: "Number of files with matches"},
				"results":   {Type: "array", Description: "List of file match groups"},
				"truncated": {Type: "boolean", Description: "Whether output was truncated due to size limits"},
			},
			[]string{"total", "files", "results", "truncated"},
		),
	}
}

// Truncation limits for compact format.
//
// These limits prevent unbounded output growth while preserving useful results.
// The limits are designed to balance completeness with token efficiency:
//   - maxMatchesPerFile: Limits matches per file to show representative samples
//     rather than overwhelming output for files with many matches.
//   - maxTotalMatches: Hard cap on total matches to prevent runaway token usage
//     for queries matching thousands of lines (e.g., grep "import" in large codebases).
//
// When truncation occurs, the Truncated field is set to true so consumers know
// the results are incomplete and can refine their search if needed.
const (
	maxMatchesPerFile = 10  // Maximum matches to include per file (excess truncated)
	maxTotalMatches   = 200 // Maximum total matches across all files (hard limit)
)

// isSkippableLine returns true if the line should be skipped (not a real match).
func isSkippableLine(line string) bool {
	// Skip empty lines
	if line == "" {
		return true
	}
	// Skip "Binary file X matches" lines
	if strings.HasPrefix(line, "Binary file ") && strings.HasSuffix(line, " matches") {
		return true
	}
	// Skip "grep: X: Is a directory" warnings
	if strings.HasPrefix(line, "grep: ") && strings.HasSuffix(line, ": Is a directory") {
		return true
	}
	// Skip "grep: X: Permission denied" errors
	if strings.HasPrefix(line, "grep: ") && strings.HasSuffix(line, ": Permission denied") {
		return true
	}
	// Skip "grep: X: No such file or directory" errors
	if strings.HasPrefix(line, "grep: ") && strings.HasSuffix(line, ": No such file or directory") {
		return true
	}
	return false
}

// truncateResults applies truncation limits to file match groups.
func truncateResults(fileOrder []string, fileGroups map[string][]MatchTuple) ([]FileMatchGroup, bool) {
	results := make([]FileMatchGroup, 0, len(fileOrder))
	truncated := false
	globalMatchCount := 0
	filesProcessed := 0

	for _, filename := range fileOrder {
		matches := fileGroups[filename]
		fileMatchCount := len(matches)
		filesProcessed++

		truncatedMatches := applyTruncation(matches, &globalMatchCount, &truncated)
		if len(truncatedMatches) > 0 {
			results = append(results, FileMatchGroup{
				Filename: filename,
				Count:    fileMatchCount,
				Matches:  truncatedMatches,
			})
		}

		if globalMatchCount >= maxTotalMatches {
			break
		}
	}

	// Check if we stopped before processing all files (global truncation)
	if filesProcessed < len(fileOrder) {
		truncated = true
	}

	return results, truncated
}

// applyTruncation applies per-file and global truncation limits.
func applyTruncation(matches []MatchTuple, globalCount *int, truncated *bool) []MatchTuple {
	result := matches

	// Apply per-file truncation
	if len(matches) > maxMatchesPerFile {
		result = matches[:maxMatchesPerFile]
		*truncated = true
	}

	// Apply global truncation
	if *globalCount+len(result) > maxTotalMatches {
		remaining := maxTotalMatches - *globalCount
		if remaining > 0 {
			result = result[:remaining]
		} else {
			result = nil
		}
		*truncated = true
	}

	*globalCount += len(result)
	return result
}

// Parse reads grep output and returns structured data in compact format.
//
// The parsing strategy groups matches by filename while preserving file order
// (first occurrence order). This grouping enables the compact format where
// the filename appears once per group rather than repeated for each match.
//
// Supported input formats:
//   - file:line:content (grep -n with multiple files)
//   - file:content (grep without -n, multiple files)
//   - line:content (grep -n with single file)
//   - content (grep without -n, single file)
//
// The output is truncated according to maxMatchesPerFile and maxTotalMatches
// to prevent excessive token usage in LLM contexts.
func (p *GrepParser) Parse(r io.Reader) (domain.ParseResult, error) {
	// fileGroups maps filename to matches, fileOrder preserves first-seen order
	fileGroups := make(map[string][]MatchTuple)
	fileOrder := make([]string, 0)

	scanner := bufio.NewScanner(r)
	var rawBuilder strings.Builder
	totalMatches := 0

	for scanner.Scan() {
		line := scanner.Text()
		rawBuilder.WriteString(line)
		rawBuilder.WriteString("\n")

		if isSkippableLine(line) {
			continue
		}

		match := p.parseLine(line)
		totalMatches++

		filename := match.File
		if _, exists := fileGroups[filename]; !exists {
			fileOrder = append(fileOrder, filename)
			fileGroups[filename] = make([]MatchTuple, 0)
		}

		fileGroups[filename] = append(fileGroups[filename], MatchTuple{
			Line:    match.Line,
			Content: match.Content,
		})
	}

	if err := scanner.Err(); err != nil {
		return domain.NewParseResultWithError(err, rawBuilder.String(), 0), nil
	}

	results, truncated := truncateResults(fileOrder, fileGroups)

	output := &GrepOutputCompact{
		Total:     totalMatches,
		Files:     len(fileOrder),
		Results:   results,
		Truncated: truncated,
	}

	return domain.NewParseResult(output, rawBuilder.String(), 0), nil
}

// Schema returns the JSON Schema for grep output.
func (p *GrepParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *GrepParser) Matches(cmd string, _ []string) bool {
	return cmd == "grep" || cmd == "egrep" || cmd == "fgrep"
}

// parseLine parses a single line of grep output.
// Formats:
//   - file:line:content (with -n)
//   - file:content (without -n)
//   - line:content (single file with -n)
//   - content (single file without -n) - not distinguishable, treated as content
func (p *GrepParser) parseLine(line string) GrepMatch {
	match := GrepMatch{}

	// Try to parse as file:line:content
	colonIdx := strings.Index(line, ":")
	if colonIdx == -1 {
		// No colons - just content
		match.Content = line
		return match
	}

	firstPart := line[:colonIdx]
	rest := line[colonIdx+1:]

	// Check if first part is a number (line number for single file)
	if lineNum, err := strconv.Atoi(firstPart); err == nil {
		match.Line = lineNum
		match.Content = rest
		return match
	}

	// First part is a filename
	match.File = firstPart

	// Check for line number in rest
	colonIdx2 := strings.Index(rest, ":")
	if colonIdx2 != -1 {
		possibleLineNum := rest[:colonIdx2]
		if lineNum, err := strconv.Atoi(possibleLineNum); err == nil {
			match.Line = lineNum
			match.Content = rest[colonIdx2+1:]
			return match
		}
	}

	// No line number, rest is content
	match.Content = rest
	return match
}
