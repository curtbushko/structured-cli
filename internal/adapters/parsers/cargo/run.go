package cargo

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"strings"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// RunParser parses the output of 'cargo run'.
// It handles both JSON build output and program stdout.
type RunParser struct {
	schema domain.Schema
}

// NewRunParser creates a new RunParser with the cargo run schema.
func NewRunParser() *RunParser {
	return &RunParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/cargo-run.json",
			"Cargo Run Output",
			"object",
			map[string]domain.PropertySchema{
				"success":       {Type: "boolean", Description: "Whether the build and run succeeded"},
				"build_success": {Type: "boolean", Description: "Whether the build phase succeeded"},
				"executable":    {Type: "string", Description: "Path to the executed binary"},
				"errors":        {Type: "array", Description: "Build errors"},
				"output":        {Type: "string", Description: "Program stdout output"},
			},
			[]string{"success", "build_success", "errors"},
		),
	}
}

// runMessage represents a single JSON message from cargo run output.
type runMessage struct {
	Reason     string        `json:"reason"`
	Success    *bool         `json:"success,omitempty"`
	Message    *runDiagMsg   `json:"message,omitempty"`
	Executable *string       `json:"executable,omitempty"`
	Target     *runTargetMsg `json:"target,omitempty"`
}

// runTargetMsg represents target information.
type runTargetMsg struct {
	Kind       []string `json:"kind"`
	CrateTypes []string `json:"crate_types"`
	Name       string   `json:"name"`
	SrcPath    string   `json:"src_path"`
	Edition    string   `json:"edition"`
}

// runDiagMsg represents a diagnostic message.
type runDiagMsg struct {
	Message  string       `json:"message"`
	Code     *runCodeMsg  `json:"code,omitempty"`
	Level    string       `json:"level"`
	Spans    []runSpanMsg `json:"spans"`
	Rendered string       `json:"rendered,omitempty"`
}

// runCodeMsg represents the diagnostic code.
type runCodeMsg struct {
	Code        string `json:"code"`
	Explanation string `json:"explanation,omitempty"`
}

// runSpanMsg represents a source code span.
type runSpanMsg struct {
	FileName    string `json:"file_name"`
	LineStart   int    `json:"line_start"`
	ColumnStart int    `json:"column_start"`
	IsPrimary   bool   `json:"is_primary"`
}

// Parse reads cargo run output and returns structured data.
func (p *RunParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &RunResult{
		Success:      true,
		BuildSuccess: true,
		Executable:   "",
		Errors:       []RunError{},
		Output:       "",
	}

	var outputLines []string
	buildFinished := false

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var msg runMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			// Non-JSON line - this is program output
			if buildFinished {
				outputLines = append(outputLines, line)
			}
			continue
		}

		switch msg.Reason {
		case reasonBuildFinished:
			buildFinished = true
			if msg.Success != nil {
				result.BuildSuccess = *msg.Success
				result.Success = *msg.Success
			}

		case reasonCompilerMessage:
			if msg.Message != nil && msg.Message.Level == levelError {
				processRunDiagnostic(msg.Message, result)
			}

		case "compiler-artifact":
			if msg.Executable != nil {
				result.Executable = *msg.Executable
			}
		}
	}

	result.Output = strings.Join(outputLines, "\n")

	return domain.NewParseResult(result, raw, 0), nil
}

// processRunDiagnostic extracts error information from a diagnostic.
func processRunDiagnostic(diag *runDiagMsg, result *RunResult) {
	var file string
	var line, column int
	for _, span := range diag.Spans {
		if span.IsPrimary {
			file = span.FileName
			line = span.LineStart
			column = span.ColumnStart
			break
		}
	}

	var code string
	if diag.Code != nil {
		code = diag.Code.Code
	}

	result.Errors = append(result.Errors, RunError{
		Message: diag.Message,
		Code:    code,
		File:    file,
		Line:    line,
		Column:  column,
	})
}

// Schema returns the JSON Schema for cargo run output.
func (p *RunParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *RunParser) Matches(cmd string, subcommands []string) bool {
	if cmd != cmdCargo {
		return false
	}
	if len(subcommands) == 0 {
		return false
	}
	// Match "run" or "r" (the short alias)
	return subcommands[0] == "run" || subcommands[0] == "r"
}
