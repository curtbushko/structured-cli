package cargo

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// ClippyParser parses the JSON output of 'cargo clippy --message-format=json'.
type ClippyParser struct {
	schema domain.Schema
}

// NewClippyParser creates a new ClippyParser with the cargo clippy schema.
func NewClippyParser() *ClippyParser {
	return &ClippyParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/cargo-clippy.json",
			"Cargo Clippy Output",
			"object",
			map[string]domain.PropertySchema{
				"success":  {Type: "boolean", Description: "Whether clippy completed without errors"},
				"warnings": {Type: "array", Description: "Clippy lint warnings"},
				"errors":   {Type: "array", Description: "Clippy lint errors"},
			},
			[]string{"success", "warnings", "errors"},
		),
	}
}

// clippyMessage represents a single JSON message from clippy output.
type clippyMessage struct {
	Reason  string           `json:"reason"`
	Success *bool            `json:"success,omitempty"`
	Message *clippyDiagMsg   `json:"message,omitempty"`
	Target  *clippyTargetMsg `json:"target,omitempty"`
}

// clippyTargetMsg represents target information in clippy JSON output.
type clippyTargetMsg struct {
	Kind       []string `json:"kind"`
	CrateTypes []string `json:"crate_types"`
	Name       string   `json:"name"`
	SrcPath    string   `json:"src_path"`
	Edition    string   `json:"edition"`
}

// clippyDiagMsg represents a diagnostic message from clippy.
type clippyDiagMsg struct {
	Message  string          `json:"message"`
	Code     *clippyCodeMsg  `json:"code,omitempty"`
	Level    string          `json:"level"`
	Spans    []clippySpanMsg `json:"spans"`
	Children []clippyDiagMsg `json:"children,omitempty"`
	Rendered string          `json:"rendered,omitempty"`
}

// clippyCodeMsg represents the diagnostic code.
type clippyCodeMsg struct {
	Code        string `json:"code"`
	Explanation string `json:"explanation,omitempty"`
}

// clippySpanMsg represents a source code span in a diagnostic.
type clippySpanMsg struct {
	FileName    string `json:"file_name"`
	ByteStart   int    `json:"byte_start"`
	ByteEnd     int    `json:"byte_end"`
	LineStart   int    `json:"line_start"`
	LineEnd     int    `json:"line_end"`
	ColumnStart int    `json:"column_start"`
	ColumnEnd   int    `json:"column_end"`
	IsPrimary   bool   `json:"is_primary"`
	Label       string `json:"label,omitempty"`
}

// Parse reads cargo clippy JSON output and returns structured data.
//
//nolint:dupl // clippy and check parsers have similar structure but different result types
func (p *ClippyParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &ClippyResult{
		Success:  true,
		Warnings: []ClippyDiagnostic{},
		Errors:   []ClippyDiagnostic{},
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var msg clippyMessage
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
				processClippyDiagnostic(msg.Message, result)
			}
		}
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// processClippyDiagnostic extracts diagnostic information from a clippy message.
func processClippyDiagnostic(diag *clippyDiagMsg, result *ClippyResult) {
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

	diagnostic := ClippyDiagnostic{
		Message:  diag.Message,
		Code:     code,
		Level:    diag.Level,
		File:     file,
		Line:     line,
		Column:   column,
		Rendered: diag.Rendered,
	}

	switch diag.Level {
	case levelError, levelErrorICE:
		result.Errors = append(result.Errors, diagnostic)
	case levelWarning:
		result.Warnings = append(result.Warnings, diagnostic)
	}
}

// Schema returns the JSON Schema for cargo clippy output.
func (p *ClippyParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *ClippyParser) Matches(cmd string, subcommands []string) bool {
	if cmd != cmdCargo {
		return false
	}
	if len(subcommands) == 0 {
		return false
	}
	return subcommands[0] == "clippy"
}
