package golang

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// vetIssueWithColumnPattern matches "file.go:line:column: message"
var vetIssueWithColumnPattern = regexp.MustCompile(`^(.+):(\d+):(\d+):\s*(.+)$`)

// vetIssueNoColumnPattern matches "file.go:line: message"
var vetIssueNoColumnPattern = regexp.MustCompile(`^(.+):(\d+):\s*(.+)$`)

// VetParser parses the output of 'go vet'.
// Go vet outputs issues in format: "file.go:line:column: message"
type VetParser struct {
	schema domain.Schema
}

// NewVetParser creates a new VetParser with the go-vet schema.
func NewVetParser() *VetParser {
	return &VetParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/go-vet.json",
			"Go Vet Output",
			"object",
			map[string]domain.PropertySchema{
				"issues": {Type: "array", Description: "List of vet issues found"},
			},
			[]string{"issues"},
		),
	}
}

// Parse reads go vet output and returns structured data.
// For clean code, go vet produces no output.
func (p *VetParser) Parse(r io.Reader) (domain.ParseResult, error) {
	// Read all input
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	vet := &VetResult{
		Issues: []VetIssue{},
	}

	// Parse issue lines from output
	vet.Issues = parseVetIssueLines(raw)

	return domain.NewParseResult(vet, raw, 0), nil
}

// parseVetIssueLines scans vet output and extracts issue information.
func parseVetIssueLines(output string) []VetIssue {
	var issues []VetIssue

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		if issue := parseVetIssueLine(line); issue != nil {
			issues = append(issues, *issue)
		}
	}

	return issues
}

// parseVetIssueLine attempts to parse a single line as a vet issue.
// Returns nil if the line is not a recognized issue format.
func parseVetIssueLine(line string) *VetIssue {
	// Try file:line:column: message format first
	if matches := vetIssueWithColumnPattern.FindStringSubmatch(line); matches != nil {
		lineNum, _ := strconv.Atoi(matches[2])
		colNum, _ := strconv.Atoi(matches[3])
		return &VetIssue{
			File:    matches[1],
			Line:    lineNum,
			Column:  colNum,
			Message: matches[4],
		}
	}

	// Try file:line: message format (no column)
	if matches := vetIssueNoColumnPattern.FindStringSubmatch(line); matches != nil {
		lineNum, _ := strconv.Atoi(matches[2])
		return &VetIssue{
			File:    matches[1],
			Line:    lineNum,
			Column:  0,
			Message: matches[3],
		}
	}

	return nil
}

// Schema returns the JSON Schema for go vet output.
func (p *VetParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *VetParser) Matches(cmd string, subcommands []string) bool {
	if cmd != "go" {
		return false
	}
	if len(subcommands) == 0 {
		return false
	}
	return subcommands[0] == "vet"
}
