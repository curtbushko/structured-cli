package makejust

import (
	"bufio"
	"io"
	"regexp"
	"strconv"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// Parser parses the output of 'make' build commands.
type Parser struct {
	schema domain.Schema
}

// NewMakeParser creates a new Parser for make commands.
func NewMakeParser() *Parser {
	return &Parser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/make.json",
			"Make Build Output",
			"object",
			map[string]domain.PropertySchema{
				"success":   {Type: "boolean", Description: "Whether the build succeeded"},
				"target":    {Type: "string", Description: "The target that was built"},
				"duration":  {Type: "integer", Description: "Build duration in milliseconds"},
				"error":     {Type: "string", Description: "Error message if build failed"},
				"exit_code": {Type: "integer", Description: "Exit code from make"},
				"targets":   {Type: "array", Description: "List of available targets"},
				"commands":  {Type: "array", Description: "Commands to be run (dry run)"},
			},
			[]string{"success", "exit_code"},
		),
	}
}

// Regular expressions for parsing make output.
var (
	// Matches make error lines: make: *** [target] Error N
	makeErrorRe = regexp.MustCompile(`^make(?:\[\d+\])?: \*\*\* \[([^\]]*)\] Error (\d+)`)
	// Matches make stop lines: make: *** message. Stop.
	makeStopRe = regexp.MustCompile(`^make(?:\[\d+\])?: \*\*\* (.+)\. Stop\.$`)
	// Matches "Nothing to be done" message
	nothingToBeDoneRe = regexp.MustCompile(`^make(?:\[\d+\])?: Nothing to be done for ['"]?([^'"]+)['"]?\.$`)
	// Matches target listing lines: "  name        description"
	targetListingRe = regexp.MustCompile(`^\s{2}(\S+)\s{2,}(.+)$`)
	// Matches "Available targets:" header
	availableTargetsRe = regexp.MustCompile(`(?i)^available targets:?\s*$`)
)

// Parse reads make output and returns structured data.
func (p *Parser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &Result{
		Success:  true,
		ExitCode: 0,
		Targets:  []Target{},
		Commands: []string{},
	}

	scanner := bufio.NewScanner(strings.NewReader(raw))
	inTargetListing := false
	var errorMessages []string

	for scanner.Scan() {
		line := scanner.Text()

		// Skip empty lines
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Check for error patterns
		if matches := makeErrorRe.FindStringSubmatch(line); matches != nil {
			result.Success = false
			exitCode, _ := strconv.Atoi(matches[2])
			result.ExitCode = exitCode
			errorMessages = append(errorMessages, line)
			continue
		}

		if matches := makeStopRe.FindStringSubmatch(line); matches != nil {
			result.Success = false
			if result.ExitCode == 0 {
				result.ExitCode = 2
			}
			errorMessages = append(errorMessages, matches[1])
			continue
		}

		// Check for "nothing to be done"
		if nothingToBeDoneRe.MatchString(line) {
			result.Success = true
			continue
		}

		// Check for target listing header
		if availableTargetsRe.MatchString(line) {
			inTargetListing = true
			continue
		}

		// Parse target listing entries
		if inTargetListing {
			if matches := targetListingRe.FindStringSubmatch(line); matches != nil {
				result.Targets = append(result.Targets, Target{
					Name:        matches[1],
					Description: strings.TrimSpace(matches[2]),
				})
				continue
			}
		}

		// For non-error output, treat as commands (for dry run or normal output)
		// Only add as commands if it looks like a shell command
		if !strings.HasPrefix(line, "make[") && !strings.HasPrefix(line, "make:") {
			result.Commands = append(result.Commands, line)
		}
	}

	// Aggregate error messages
	if len(errorMessages) > 0 {
		result.Error = strings.Join(errorMessages, "\n")
	}

	return domain.NewParseResult(result, raw, result.ExitCode), nil
}

// Schema returns the JSON Schema for make output.
func (p *Parser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
// The make parser matches "make" and "gmake" (GNU make).
func (p *Parser) Matches(cmd string, subcommands []string) bool {
	return cmd == "make" || cmd == "gmake"
}
