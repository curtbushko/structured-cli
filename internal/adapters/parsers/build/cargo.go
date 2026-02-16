package build

import (
	"bufio"
	"encoding/json"
	"io"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// CargoParser parses the JSON output of 'cargo build --message-format=json'.
type CargoParser struct {
	schema domain.Schema
}

// NewCargoParser creates a new CargoParser with the cargo build schema.
func NewCargoParser() *CargoParser {
	return &CargoParser{
		schema: domain.NewSchema(
			"https://structured-cli.dev/schemas/cargo-build.json",
			"Cargo Build Output",
			"object",
			map[string]domain.PropertySchema{
				"success":       {Type: "boolean", Description: "Whether the build succeeded"},
				"errors":        {Type: "array", Description: "Compilation errors"},
				"warnings":      {Type: "array", Description: "Compilation warnings"},
				"artifacts":     {Type: "array", Description: "Compiled artifacts"},
				"build_scripts": {Type: "array", Description: "Build script outputs"},
			},
			[]string{"success", "errors", "warnings", "artifacts"},
		),
	}
}

// cargoMessage represents a single JSON message from cargo build output.
// The Reason field determines the message type.
type cargoMessage struct {
	Reason string `json:"reason"`

	// For build-finished messages
	Success *bool `json:"success,omitempty"`

	// For compiler-message, compiler-artifact
	PackageID    string           `json:"package_id,omitempty"`
	ManifestPath string           `json:"manifest_path,omitempty"`
	Target       *cargoTargetMsg  `json:"target,omitempty"`
	Message      *cargoDiagnostic `json:"message,omitempty"`

	// For compiler-artifact
	Profile    *cargoProfileMsg `json:"profile,omitempty"`
	Features   []string         `json:"features,omitempty"`
	Filenames  []string         `json:"filenames,omitempty"`
	Executable *string          `json:"executable,omitempty"`
	Fresh      bool             `json:"fresh,omitempty"`

	// For build-script-executed
	LinkedLibs  []string   `json:"linked_libs,omitempty"`
	LinkedPaths []string   `json:"linked_paths,omitempty"`
	Cfgs        []string   `json:"cfgs,omitempty"`
	Env         [][]string `json:"env,omitempty"`
	OutDir      string     `json:"out_dir,omitempty"`
}

// cargoTargetMsg represents target information in cargo JSON output.
type cargoTargetMsg struct {
	Kind       []string `json:"kind"`
	CrateTypes []string `json:"crate_types"`
	Name       string   `json:"name"`
	SrcPath    string   `json:"src_path"`
	Edition    string   `json:"edition"`
}

// cargoProfileMsg represents profile information in cargo JSON output.
type cargoProfileMsg struct {
	OptLevel        string `json:"opt_level"`
	Debuginfo       int    `json:"debuginfo"`
	DebugAssertions bool   `json:"debug_assertions"`
	OverflowChecks  bool   `json:"overflow_checks"`
	Test            bool   `json:"test"`
}

// cargoDiagnostic represents a rustc diagnostic message.
type cargoDiagnostic struct {
	Message  string            `json:"message"`
	Code     *cargoCodeMsg     `json:"code,omitempty"`
	Level    string            `json:"level"`
	Spans    []cargoSpanMsg    `json:"spans"`
	Children []cargoDiagnostic `json:"children,omitempty"`
	Rendered string            `json:"rendered,omitempty"`
}

// cargoCodeMsg represents error/warning code information.
type cargoCodeMsg struct {
	Code        string `json:"code"`
	Explanation string `json:"explanation,omitempty"`
}

// cargoSpanMsg represents a source code span in a diagnostic.
type cargoSpanMsg struct {
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

// Parse reads cargo build JSON output and returns structured data.
func (p *CargoParser) Parse(r io.Reader) (domain.ParseResult, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return domain.NewParseResultWithError(err, "", 0), nil
	}

	raw := string(data)

	result := &CargoResult{
		Success:      true, // Assume success unless build-finished says otherwise
		Errors:       []CargoError{},
		Warnings:     []CargoWarning{},
		Artifacts:    []CargoArtifact{},
		BuildScripts: []CargoBuildScript{},
	}

	// Parse each line as a JSON message
	scanner := bufio.NewScanner(reader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var msg cargoMessage
		if err := json.Unmarshal([]byte(line), &msg); err != nil {
			// Skip invalid JSON lines (could be non-JSON output mixed in)
			continue
		}

		switch msg.Reason {
		case "build-finished":
			if msg.Success != nil {
				result.Success = *msg.Success
			}

		case "compiler-message":
			if msg.Message != nil {
				processDiagnostic(msg.Message, result)
			}

		case "compiler-artifact":
			artifact := CargoArtifact{
				PackageID: msg.PackageID,
				Features:  msg.Features,
				Filenames: msg.Filenames,
				Fresh:     msg.Fresh,
			}
			if msg.Executable != nil {
				artifact.Executable = *msg.Executable
			}
			if msg.Target != nil {
				artifact.Target = CargoTarget{
					Kind:       msg.Target.Kind,
					CrateTypes: msg.Target.CrateTypes,
					Name:       msg.Target.Name,
					SrcPath:    msg.Target.SrcPath,
					Edition:    msg.Target.Edition,
				}
			}
			if msg.Profile != nil {
				artifact.Profile = CargoProfile{
					OptLevel:        msg.Profile.OptLevel,
					Debuginfo:       msg.Profile.Debuginfo,
					DebugAssertions: msg.Profile.DebugAssertions,
					OverflowChecks:  msg.Profile.OverflowChecks,
					Test:            msg.Profile.Test,
				}
			}
			result.Artifacts = append(result.Artifacts, artifact)

		case "build-script-executed":
			buildScript := CargoBuildScript{
				PackageID:   msg.PackageID,
				LinkedLibs:  msg.LinkedLibs,
				LinkedPaths: msg.LinkedPaths,
				Cfgs:        msg.Cfgs,
				Env:         msg.Env,
				OutDir:      msg.OutDir,
			}
			result.BuildScripts = append(result.BuildScripts, buildScript)
		}
	}

	return domain.NewParseResult(result, raw, 0), nil
}

// reader creates a bufio.Scanner-compatible reader from byte data.
func reader(data []byte) io.Reader {
	return &byteReader{data: data, pos: 0}
}

type byteReader struct {
	data []byte
	pos  int
}

func (r *byteReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

// processDiagnostic extracts error/warning information from a rustc diagnostic.
func processDiagnostic(diag *cargoDiagnostic, result *CargoResult) {
	// Find the primary span for file/line/column info
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

	// Extract error code
	var code string
	if diag.Code != nil {
		code = diag.Code.Code
	}

	switch diag.Level {
	case "error", "error: internal compiler error":
		result.Errors = append(result.Errors, CargoError{
			Message:  diag.Message,
			Code:     code,
			File:     file,
			Line:     line,
			Column:   column,
			Rendered: diag.Rendered,
		})
	case "warning":
		result.Warnings = append(result.Warnings, CargoWarning{
			Message:  diag.Message,
			Code:     code,
			File:     file,
			Line:     line,
			Column:   column,
			Rendered: diag.Rendered,
		})
	}
	// Note: "note", "help", "failure-note" levels are typically children
	// and are part of the rendered output; we skip them at the top level
}

// Schema returns the JSON Schema for cargo build output.
func (p *CargoParser) Schema() domain.Schema {
	return p.schema
}

// Matches returns true if this parser handles the given command.
// The cargo parser matches "cargo build" and "cargo b" (the alias).
func (p *CargoParser) Matches(cmd string, subcommands []string) bool {
	if cmd != "cargo" {
		return false
	}
	if len(subcommands) == 0 {
		return false
	}
	// Match "build" or "b" (the short alias)
	return subcommands[0] == "build" || subcommands[0] == "b"
}
