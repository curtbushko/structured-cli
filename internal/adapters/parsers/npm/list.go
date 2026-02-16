package npm

import (
	"bufio"
	"io"
	"regexp"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Regex patterns for parsing npm list output
var (
	// rootPackagePattern matches "name@version /path"
	rootPackagePattern = regexp.MustCompile(`^(\S+)@(\S+)\s+(/\S+)$`)

	// dependencyPattern matches tree lines like "+-- package@version"
	dependencyPattern = regexp.MustCompile("^[+|`]-- (?:UNMET DEPENDENCY )?(\\S+)@(\\S+)$")

	// unmetDependencyPattern matches "UNMET DEPENDENCY" lines
	unmetDependencyPattern = regexp.MustCompile(`UNMET DEPENDENCY`)

	// npmErrPattern matches npm ERR! lines
	npmErrPattern = regexp.MustCompile(`^npm ERR!`)
)

// ListParser parses the output of 'npm list'.
type ListParser struct {
	schema domain.Schema
}

// NewListParser creates a new ListParser with the npm-list schema.
func NewListParser() *ListParser {
	return &ListParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/npm-list.json",
			"NPM List Output",
			"object",
			map[string]domain.PropertySchema{
				"success":      {Type: "boolean", Description: "Whether list command succeeded without problems"},
				"name":         {Type: "string", Description: "Root package name"},
				"version":      {Type: "string", Description: "Root package version"},
				"dependencies": {Type: "array", Description: "List of direct dependencies"},
				"problems":     {Type: "array", Description: "List of problems found"},
			},
			[]string{"success", "name", "version", "dependencies"},
		),
	}
}

// Parse reads npm list output and returns structured data.
func (p *ListParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &ListResult{
		Success:      true,
		Dependencies: []ListDependency{},
		Problems:     []string{},
	}

	parseListOutput(raw, result)

	// If there are problems, mark as not successful
	if len(result.Problems) > 0 {
		result.Success = false
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// parseListOutput extracts list information from the output.
func parseListOutput(output string, result *ListResult) {
	scanner := bufio.NewScanner(strings.NewReader(output))

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		// Check for root package
		if matches := rootPackagePattern.FindStringSubmatch(line); matches != nil {
			result.Name = matches[1]
			result.Version = matches[2]
			continue
		}

		// Check for dependency
		if matches := dependencyPattern.FindStringSubmatch(line); matches != nil {
			dep := ListDependency{
				Name:    matches[1],
				Version: matches[2],
			}
			result.Dependencies = append(result.Dependencies, dep)

			// Check if this is an unmet dependency
			if unmetDependencyPattern.MatchString(line) {
				result.Problems = append(result.Problems, "UNMET DEPENDENCY: "+matches[1]+"@"+matches[2])
			}
			continue
		}

		// Check for npm errors
		if npmErrPattern.MatchString(line) {
			problem := strings.TrimPrefix(line, "npm ERR! ")
			result.Problems = append(result.Problems, problem)
		}
	}
}

// Schema returns the JSON Schema for npm list output.
func (p *ListParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *ListParser) Matches(cmd string, subcommands []string) bool {
	if cmd != npmCommand {
		return false
	}

	if len(subcommands) == 0 {
		return false
	}

	// npm list, npm ls, npm ll, npm la
	switch subcommands[0] {
	case "list", "ls", "ll", "la":
		return true
	default:
		return false
	}
}
