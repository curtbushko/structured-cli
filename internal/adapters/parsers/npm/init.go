package npm

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// InitParser parses the output of 'npm init'.
type InitParser struct {
	schema domain.Schema
}

// NewInitParser creates a new InitParser with the npm-init schema.
func NewInitParser() *InitParser {
	return &InitParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/npm-init.json",
			"NPM Init Output",
			"object",
			map[string]domain.PropertySchema{
				"success":      {Type: "boolean", Description: "Whether initialization succeeded"},
				"package_name": {Type: "string", Description: "Created package name"},
				"version":      {Type: "string", Description: "Initial version"},
				"description":  {Type: "string", Description: "Package description"},
				"entry_point":  {Type: "string", Description: "Main entry point"},
				"test_command": {Type: "string", Description: "Test command"},
				"repository":   {Type: "string", Description: "Git repository URL"},
				"keywords":     {Type: "array", Description: "Package keywords"},
				"author":       {Type: "string", Description: "Package author"},
				"license":      {Type: "string", Description: "Package license"},
			},
			[]string{"success", "package_name", "version"},
		),
	}
}

// Parse reads npm init output and returns structured data.
func (p *InitParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &InitResult{
		Success:  true,
		Keywords: []string{},
	}

	parseInitOutput(raw, result)

	return domain.NewParseResult(result, raw, 0), nil
}

// packageJSON represents the structure of package.json for parsing
type packageJSON struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Description string            `json:"description"`
	Main        string            `json:"main"`
	Scripts     map[string]string `json:"scripts"`
	Keywords    []string          `json:"keywords"`
	Author      string            `json:"author"`
	License     string            `json:"license"`
	Repository  interface{}       `json:"repository"`
}

// parseInitOutput extracts init information from the output.
func parseInitOutput(output string, result *InitResult) {
	// Check for npm errors
	if strings.Contains(output, "npm ERR!") {
		result.Success = false
		return
	}

	// Try to find and parse the JSON portion
	jsonStr := extractJSONFromOutput(output)
	if jsonStr == "" {
		return
	}

	var pkg packageJSON
	if err := json.Unmarshal([]byte(jsonStr), &pkg); err != nil {
		return
	}

	populateResultFromPackage(result, &pkg)
}

// extractJSONFromOutput finds and extracts JSON from npm init output.
func extractJSONFromOutput(output string) string {
	lines := strings.Split(output, "\n")
	var jsonLines []string
	inJSON := false
	braceCount := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if trimmed == "{" && !inJSON {
			inJSON = true
			braceCount = 1
			jsonLines = append(jsonLines, line)
			continue
		}

		if inJSON {
			jsonLines = append(jsonLines, line)
			braceCount += strings.Count(trimmed, "{") - strings.Count(trimmed, "}")
			if braceCount <= 0 {
				break
			}
		}
	}

	if len(jsonLines) == 0 {
		return ""
	}

	return strings.Join(jsonLines, "\n")
}

// populateResultFromPackage fills the InitResult from parsed package.json.
func populateResultFromPackage(result *InitResult, pkg *packageJSON) {
	result.PackageName = pkg.Name
	result.Version = pkg.Version
	result.Description = pkg.Description
	result.EntryPoint = pkg.Main
	result.Author = pkg.Author
	result.License = pkg.License

	if pkg.Keywords != nil {
		result.Keywords = pkg.Keywords
	}

	if testCmd, ok := pkg.Scripts["test"]; ok {
		result.TestCommand = testCmd
	}

	// Handle repository which can be string or object
	switch repo := pkg.Repository.(type) {
	case string:
		result.Repository = repo
	case map[string]interface{}:
		if url, ok := repo["url"].(string); ok {
			result.Repository = url
		}
	}
}

// Schema returns the JSON Schema for npm init output.
func (p *InitParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *InitParser) Matches(cmd string, subcommands []string) bool {
	if cmd != npmCommand {
		return false
	}

	if len(subcommands) == 0 {
		return false
	}

	// npm init, npm create, npm innit
	switch subcommands[0] {
	case "init", "create", "innit":
		return true
	default:
		return false
	}
}
