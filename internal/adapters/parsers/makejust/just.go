package makejust

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// JustParser parses the output of 'just' command runner.
type JustParser struct {
	schema domain.Schema
}

// NewJustParser creates a new JustParser with the just schema.
func NewJustParser() *JustParser {
	return &JustParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/just.json",
			"Just Command Runner Output",
			"object",
			map[string]domain.PropertySchema{
				"success":   {Type: "boolean", Description: "Whether the recipe succeeded"},
				"recipe":    {Type: "string", Description: "The recipe that was executed"},
				"duration":  {Type: "integer", Description: "Execution duration in milliseconds"},
				"error":     {Type: "string", Description: "Error message if execution failed"},
				"exit_code": {Type: "integer", Description: "Exit code from just"},
				"recipes":   {Type: "array", Description: "List of available recipes"},
				"commands":  {Type: "array", Description: "Commands to be run (dry run)"},
			},
			[]string{"success", "exit_code"},
		),
	}
}

// Regular expressions for parsing just output.
var (
	// Matches just error lines: error: message
	justErrorRe = regexp.MustCompile(`^error: (.+)$`)
	// Matches recipe failure: error: Recipe 'name' failed on line N with exit code M
	justRecipeFailRe = regexp.MustCompile(`^error: Recipe '([^']+)' failed.* exit code (\d+)`)
	// Matches "Available recipes:" header
	availableRecipesRe = regexp.MustCompile(`(?i)^Available recipes:?\s*$`)
	// Matches recipe listing: "    name param1 param2 # description"
	recipeListingRe = regexp.MustCompile(`^\s{4}(\S+)(.*)#\s*(.*)$`)
	// Matches recipe listing without description: "    name param1 param2"
	recipeListingNoDescRe = regexp.MustCompile(`^\s{4}(\S+)(.*)$`)
	// Matches parameter with default: name='value' or name="value"
	paramWithDefaultRe = regexp.MustCompile(`(\w+)=['"]([^'"]+)['"]`)
	// Matches simple parameter
	simpleParamRe = regexp.MustCompile(`\b(\w+)\b`)
)

// Parse reads just output and returns structured data.
func (p *JustParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &JustResult{
		Success:  true,
		ExitCode: 0,
		Recipes:  []JustRecipe{},
		Commands: []string{},
	}

	scanner := bufio.NewScanner(strings.NewReader(raw))
	inRecipeListing := false
	var errorMessages []string

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Check for recipe failure with exit code
		if matches := justRecipeFailRe.FindStringSubmatch(line); matches != nil {
			result.Success = false
			result.Recipe = matches[1]
			exitCode, _ := strconv.Atoi(matches[2])
			result.ExitCode = exitCode
			errorMessages = append(errorMessages, line)
			continue
		}

		// Check for general error patterns
		if matches := justErrorRe.FindStringSubmatch(line); matches != nil {
			result.Success = false
			if result.ExitCode == 0 {
				result.ExitCode = 1
			}
			errorMessages = append(errorMessages, matches[1])
			continue
		}

		// Check for recipe listing header
		if availableRecipesRe.MatchString(line) {
			inRecipeListing = true
			continue
		}

		// Parse recipe listing entries
		if inRecipeListing {
			if recipe := parseRecipeLine(line); recipe != nil {
				result.Recipes = append(result.Recipes, *recipe)
				continue
			}
		}

		// For non-error output, treat as commands (for dry run or normal output)
		if !strings.HasPrefix(line, "error:") && !strings.HasPrefix(line, "warning:") {
			result.Commands = append(result.Commands, line)
		}
	}

	// Aggregate error messages
	if len(errorMessages) > 0 {
		result.Error = strings.Join(errorMessages, "\n")
	}

	return domain.NewParseResult(result, raw, result.ExitCode), nil
}

// parseRecipeLine parses a single recipe listing line.
func parseRecipeLine(line string) *JustRecipe {
	// Try matching with description first
	if matches := recipeListingRe.FindStringSubmatch(line); matches != nil {
		recipe := &JustRecipe{
			Name:        matches[1],
			Description: strings.TrimSpace(matches[3]),
			Parameters:  parseParameters(matches[2]),
		}
		return recipe
	}

	// Try matching without description
	if matches := recipeListingNoDescRe.FindStringSubmatch(line); matches != nil {
		recipe := &JustRecipe{
			Name:       matches[1],
			Parameters: parseParameters(matches[2]),
		}
		return recipe
	}

	return nil
}

// parseParameters extracts parameters from the parameter string.
func parseParameters(paramStr string) []JustParameter {
	paramStr = strings.TrimSpace(paramStr)
	if paramStr == "" {
		return nil
	}

	var params []JustParameter

	// Find parameters with default values first
	defaultMatches := paramWithDefaultRe.FindAllStringSubmatch(paramStr, -1)
	defaultParams := make(map[string]string)
	for _, match := range defaultMatches {
		defaultParams[match[1]] = match[2]
	}

	// Remove the default value syntax for simpler parsing
	cleanedStr := paramWithDefaultRe.ReplaceAllString(paramStr, "$1")

	// Find all simple parameter names
	simpleMatches := simpleParamRe.FindAllStringSubmatch(cleanedStr, -1)
	for _, match := range simpleMatches {
		paramName := match[1]
		param := JustParameter{Name: paramName}
		if defaultVal, ok := defaultParams[paramName]; ok {
			param.Default = defaultVal
		}
		params = append(params, param)
	}

	return params
}

// Schema returns the JSON Schema for just output.
func (p *JustParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
// The just parser matches only "just".
func (p *JustParser) Matches(cmd string, subcommands []string) bool {
	return cmd == "just"
}
