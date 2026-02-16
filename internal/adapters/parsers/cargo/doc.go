package cargo

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// DocParser parses the JSON output of 'cargo doc --message-format=json'.
type DocParser struct {
	schema domain.Schema
}

// NewDocParser creates a new DocParser with the cargo doc schema.
func NewDocParser() *DocParser {
	return &DocParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/cargo-doc.json",
			"Cargo Doc Output",
			"object",
			map[string]domain.PropertySchema{
				"success":        {Type: "boolean", Description: "Whether doc generation succeeded"},
				"warnings":       {Type: "array", Description: "Documentation warnings"},
				"errors":         {Type: "array", Description: "Documentation errors"},
				"generated_docs": {Type: "array", Description: "Paths to generated documentation"},
			},
			[]string{"success", "warnings", "errors"},
		),
	}
}

// docMessage represents a single JSON message from cargo doc output.
type docMessage struct {
	Reason  string        `json:"reason"`
	Success *bool         `json:"success,omitempty"`
	Message *docDiagMsg   `json:"message,omitempty"`
	Target  *docTargetMsg `json:"target,omitempty"`
}

// docTargetMsg represents target information.
type docTargetMsg struct {
	Kind       []string `json:"kind"`
	CrateTypes []string `json:"crate_types"`
	Name       string   `json:"name"`
	SrcPath    string   `json:"src_path"`
	Edition    string   `json:"edition"`
}

// docDiagMsg represents a diagnostic message.
type docDiagMsg struct {
	Message  string       `json:"message"`
	Code     *docCodeMsg  `json:"code,omitempty"`
	Level    string       `json:"level"`
	Spans    []docSpanMsg `json:"spans"`
	Rendered string       `json:"rendered,omitempty"`
}

// docCodeMsg represents the diagnostic code.
type docCodeMsg struct {
	Code        string `json:"code"`
	Explanation string `json:"explanation,omitempty"`
}

// docSpanMsg represents a source code span.
type docSpanMsg struct {
	FileName    string `json:"file_name"`
	LineStart   int    `json:"line_start"`
	ColumnStart int    `json:"column_start"`
	IsPrimary   bool   `json:"is_primary"`
}

// Parse reads cargo doc JSON output and returns structured data.
func (p *DocParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &DocResult{
		Success:       true,
		Warnings:      []DocWarning{},
		Errors:        []DocError{},
		GeneratedDocs: []string{},
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var msg docMessage
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
				processDocDiagnostic(msg.Message, result)
			}
		}
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// processDocDiagnostic extracts diagnostic information from a doc message.
func processDocDiagnostic(diag *docDiagMsg, result *DocResult) {
	var file string
	var line int
	for _, span := range diag.Spans {
		if span.IsPrimary {
			file = span.FileName
			line = span.LineStart
			break
		}
	}

	var code string
	if diag.Code != nil {
		code = diag.Code.Code
	}

	switch diag.Level {
	case levelError, levelErrorICE:
		result.Errors = append(result.Errors, DocError{
			Message: diag.Message,
			Code:    code,
			File:    file,
			Line:    line,
		})
	case levelWarning:
		result.Warnings = append(result.Warnings, DocWarning{
			Message: diag.Message,
			File:    file,
			Line:    line,
		})
	}
}

// Schema returns the JSON Schema for cargo doc output.
func (p *DocParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
func (p *DocParser) Matches(cmd string, subcommands []string) bool {
	if cmd != cmdCargo {
		return false
	}
	if len(subcommands) == 0 {
		return false
	}
	// Match "doc" or "d" (the short alias)
	return subcommands[0] == "doc" || subcommands[0] == "d"
}
