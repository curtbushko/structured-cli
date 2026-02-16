package cargo

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// CheckParser parses the JSON output of 'cargo check --message-format=json'.
type CheckParser struct {
	schema domain.Schema
}

// NewCheckParser creates a new CheckParser with the cargo check schema.
func NewCheckParser() *CheckParser {
	return &CheckParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/cargo-check.json",
			"Cargo Check Output",
			"object",
			map[string]domain.PropertySchema{
				"success":  {Type: "boolean", Description: "Whether the check succeeded"},
				"errors":   {Type: "array", Description: "Compilation errors"},
				"warnings": {Type: "array", Description: "Compilation warnings"},
			},
			[]string{"success", "errors", "warnings"},
		),
	}
}

// checkMessage represents a single JSON message from cargo check output.
type checkMessage struct {
	Reason  string          `json:"reason"`
	Success *bool           `json:"success,omitempty"`
	Message *checkDiagMsg   `json:"message,omitempty"`
	Target  *checkTargetMsg `json:"target,omitempty"`
}

// checkTargetMsg represents target information.
type checkTargetMsg struct {
	Kind       []string `json:"kind"`
	CrateTypes []string `json:"crate_types"`
	Name       string   `json:"name"`
	SrcPath    string   `json:"src_path"`
	Edition    string   `json:"edition"`
}

// checkDiagMsg represents a diagnostic message.
type checkDiagMsg struct {
	Message  string         `json:"message"`
	Code     *checkCodeMsg  `json:"code,omitempty"`
	Level    string         `json:"level"`
	Spans    []checkSpanMsg `json:"spans"`
	Rendered string         `json:"rendered,omitempty"`
}

// checkCodeMsg represents the diagnostic code.
type checkCodeMsg struct {
	Code        string `json:"code"`
	Explanation string `json:"explanation,omitempty"`
}

// checkSpanMsg represents a source code span.
type checkSpanMsg struct {
	FileName    string `json:"file_name"`
	LineStart   int    `json:"line_start"`
	ColumnStart int    `json:"column_start"`
	IsPrimary   bool   `json:"is_primary"`
}

// Parse reads cargo check JSON output and returns structured data.
//
//nolint:dupl // check and clippy parsers have similar structure but different result types
func (p *CheckParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &CheckResult{
		Success:  true,
		Errors:   []CheckError{},
		Warnings: []CheckWarning{},
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var msg checkMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			continue
		}

		switch msg.Reason {
		case reasonBuildFinished:
			if msg.Success != nil {
				result.Success = *msg.Success
			}

		case reasonCompilerMessage:
			if msg.Message != nil {
				processCheckDiagnostic(msg.Message, result)
			}
		}
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// processCheckDiagnostic extracts diagnostic information from a check message.
func processCheckDiagnostic(diag *checkDiagMsg, result *CheckResult) {
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

	switch diag.Level {
	case levelError, levelErrorICE:
		result.Errors = append(result.Errors, CheckError{
			Message:  diag.Message,
			Code:     code,
			File:     file,
			Line:     line,
			Column:   column,
			Rendered: diag.Rendered,
		})
	case levelWarning:
		result.Warnings = append(result.Warnings, CheckWarning{
			Message:  diag.Message,
			Code:     code,
			File:     file,
			Line:     line,
			Column:   column,
			Rendered: diag.Rendered,
		})
	}
}

// Schema returns the JSON Schema for cargo check output.
func (p *CheckParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *CheckParser) Matches(cmd string, subcommands []string) bool {
	if cmd != cmdCargo {
		return false
	}
	if len(subcommands) == 0 {
		return false
	}
	// Match "check" or "c" (the short alias)
	return subcommands[0] == "check" || subcommands[0] == "c"
}
