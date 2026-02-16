package golang

import (
	"io"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// RunParser parses the output of 'go run'.
// It captures stdout from the executed program. Note that stderr handling
// requires runner adapter changes to capture both streams separately.
type RunParser struct {
	schema domain.Schema
}

// NewRunParser creates a new RunParser with the go-run schema.
func NewRunParser() *RunParser {
	return &RunParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/go-run.json",
			"Go Run Output",
			"object",
			map[string]domain.PropertySchema{
				"exitCode": {Type: "integer", Description: "Exit code of the executed program"},
				"stdout":   {Type: "string", Description: "Standard output of the program"},
				"stderr":   {Type: "string", Description: "Standard error output of the program"},
			},
			[]string{"exitCode", "stdout", "stderr"},
		),
	}
}

// Parse reads go run output and returns structured data.
// The reader contains the combined output from the executed program.
// Currently captures all input as stdout; stderr requires runner adapter changes.
func (p *RunParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	run := &RunResult{
		ExitCode: 0,
		Stdout:   raw,
		Stderr:   "",
	}

	return domain.NewParseResult(run, raw, 0), nil
}

// Schema returns the JSON Schema for go run output.
func (p *RunParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *RunParser) Matches(cmd string, subcommands []string) bool {
	if cmd != "go" {
		return false
	}
	if len(subcommands) == 0 {
		return false
	}
	return subcommands[0] == "run"
}
