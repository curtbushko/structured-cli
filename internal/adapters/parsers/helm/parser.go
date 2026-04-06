package helm

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Common constants for helm output parsing.
const (
	// Common command and subcommand names.
	cmdHelm    = "helm"
	cmdList    = "list"
	cmdLs      = "ls"
	cmdStatus  = "status"
	cmdHistory = "history"
	cmdSearch  = "search"
	cmdShow    = "show"
	cmdValues  = "values"

	// Common column names.
	colName       = "NAME"
	colNamespace  = "NAMESPACE"
	colRevision   = "REVISION"
	colUpdated    = "UPDATED"
	colStatus     = "STATUS"
	colChart      = "CHART"
	colAppVersion = "APP VERSION"
)

// columnInfo holds the position and width of a column.
type columnInfo struct {
	name  string
	start int
	end   int // -1 means until end of line
}

// parseInput holds prepared input for parsing.
type parseInput struct {
	Raw     string
	Scanner *bufio.Scanner
	Columns []columnInfo
}

// parseResult contains the result of initial input processing.
type parseResult struct {
	Input    parseInput
	IsEmpty  bool
	Error    error
	ErrorMsg string
}

// readAndPrepare reads input and prepares for parsing.
func readAndPrepare(r io.Reader, colNames []string, requiredCols []string) parseResult {
	data, err := io.ReadAll(r)
	if err != nil {
		return parseResult{
			Error:    fmt.Errorf("read input: %w", err),
			ErrorMsg: "",
			IsEmpty:  true,
		}
	}

	raw := string(data)

	// Handle empty output (no releases)
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return parseResult{
			Input:   parseInput{Raw: raw},
			IsEmpty: true,
		}
	}

	scanner := bufio.NewScanner(strings.NewReader(raw))

	// Read header line to determine columns
	if !scanner.Scan() {
		return parseResult{
			Input:   parseInput{Raw: raw},
			IsEmpty: true,
		}
	}

	header := scanner.Text()
	columns := parseHeaderColumns(header, colNames)

	// Check if this looks like valid helm output
	if !hasRequiredColumns(columns, requiredCols) {
		return parseResult{
			Input:   parseInput{Raw: raw},
			IsEmpty: true,
		}
	}

	return parseResult{
		Input: parseInput{
			Raw:     raw,
			Scanner: scanner,
			Columns: columns,
		},
		IsEmpty: false,
	}
}

// lineParser is a function that parses a single line and returns an item.
type lineParser[T any] func(line string, columns []columnInfo) T

// parseLines iterates over remaining lines and parses each one.
func parseLines[T any](scanner *bufio.Scanner, columns []columnInfo, parser lineParser[T]) []T {
	var items []T
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		items = append(items, parser(line, columns))
	}
	return items
}

// parseHeaderColumns parses the header line to determine column positions.
func parseHeaderColumns(header string, colNames []string) []columnInfo {
	var columns []columnInfo

	for _, name := range colNames {
		idx := findColumnIndex(header, name)
		if idx >= 0 {
			columns = append(columns, columnInfo{
				name:  name,
				start: idx,
				end:   -1,
			})
		}
	}

	// Sort columns by start position
	sortColumnsByPosition(columns)

	// Set end positions based on next column's start
	for i := 0; i < len(columns)-1; i++ {
		columns[i].end = columns[i+1].start
	}

	return columns
}

// sortColumnsByPosition sorts columns by their start position.
func sortColumnsByPosition(columns []columnInfo) {
	for i := 0; i < len(columns)-1; i++ {
		for j := i + 1; j < len(columns); j++ {
			if columns[i].start > columns[j].start {
				columns[i], columns[j] = columns[j], columns[i]
			}
		}
	}
}

// findColumnIndex finds the index of a column name, ensuring it's a complete word match.
func findColumnIndex(header, name string) int {
	idx := 0
	for {
		pos := strings.Index(header[idx:], name)
		if pos < 0 {
			return -1
		}
		absPos := idx + pos

		// Check word boundaries
		if isWordBoundary(header, absPos, len(name)) {
			return absPos
		}

		idx = absPos + 1
		if idx >= len(header) {
			return -1
		}
	}
}

// isWordBoundary checks if the position represents a complete word match.
func isWordBoundary(header string, pos, nameLen int) bool {
	// Check start boundary (start of string or preceded by space/tab)
	startOK := pos == 0 || header[pos-1] == ' ' || header[pos-1] == '\t'
	if !startOK {
		return false
	}

	// Check end boundary (end of string or followed by space/tab)
	endPos := pos + nameLen
	return endPos >= len(header) || header[endPos] == ' ' || header[endPos] == '\t'
}

// hasRequiredColumns checks if the header has all required columns.
func hasRequiredColumns(columns []columnInfo, required []string) bool {
	found := make(map[string]bool)
	for _, col := range columns {
		found[col.name] = true
	}
	for _, req := range required {
		if !found[req] {
			return false
		}
	}
	return true
}

// extractColumnValue extracts a column value from a line.
func extractColumnValue(line string, col columnInfo) string {
	if col.start >= len(line) {
		return ""
	}

	if col.end == -1 {
		return strings.TrimSpace(line[col.start:])
	}

	end := col.end
	if end > len(line) {
		end = len(line)
	}
	return strings.TrimSpace(line[col.start:end])
}

// parseInt parses a string to int, returning 0 on error.
func parseInt(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}

// emptyResultWithError creates an error parse result.
func emptyResultWithError(err error, raw string) domain.ParseResult {
	return domain.NewParseResultWithError(err, raw, 0)
}

// emptyResultOK creates a successful empty parse result.
func emptyResultOK(data any, raw string) domain.ParseResult {
	return domain.NewParseResult(data, raw, 0)
}
