package docker

import (
	"io"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// ExecParser parses the output of 'docker exec'.
type ExecParser struct {
	schema domain.Schema
}

// NewExecParser creates a new ExecParser with the docker-exec schema.
func NewExecParser() *ExecParser {
	return &ExecParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/docker-exec.json",
			"Docker Exec Output",
			"object",
			map[string]domain.PropertySchema{
				"success":      {Type: "boolean", Description: "Whether the command completed successfully"},
				"container_id": {Type: "string", Description: "Container ID"},
				"output":       {Type: "string", Description: "Command output"},
				"exit_code":    {Type: "integer", Description: "Command exit code"},
				"errors":       {Type: "array", Description: "Error messages"},
			},
			[]string{"success", "output"},
		),
	}
}

// Parse reads docker exec output and returns structured data.
func (p *ExecParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &ExecResult{
		Success: true,
		Errors:  []string{},
	}

	parseExecOutput(raw, result)

	return domain.NewParseResult(result, raw, 0), nil
}

// parseExecOutput extracts exec information from the output.
func parseExecOutput(output string, result *ExecResult) {
	trimmed := strings.TrimSpace(output)

	// Check for errors
	if strings.Contains(output, "Error response from daemon") ||
		strings.HasPrefix(output, "Error:") ||
		strings.Contains(output, "OCI runtime exec failed") {
		result.Success = false
		lines := strings.Split(output, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				result.Errors = append(result.Errors, strings.TrimSpace(line))
			}
		}
		return
	}

	result.Output = trimmed
}

// Schema returns the JSON Schema for docker exec output.
func (p *ExecParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *ExecParser) Matches(cmd string, subcommands []string) bool {
	if cmd != dockerCommand {
		return false
	}

	if len(subcommands) == 0 {
		return false
	}

	// docker exec
	if subcommands[0] == subExec {
		return true
	}

	// docker container exec
	if len(subcommands) >= 2 && subcommands[0] == subContainer && subcommands[1] == subExec {
		return true
	}

	return false
}
