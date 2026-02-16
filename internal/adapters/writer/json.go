// Package writer provides implementations of the OutputWriter port.
// This package is in the adapters layer and implements the ports.OutputWriter
// interface for formatting and writing parse results.
package writer

import (
	"encoding/json"
	"io"

	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

// Compile-time check that JSONWriter implements ports.OutputWriter.
var _ ports.OutputWriter = (*JSONWriter)(nil)

// JSONWriter outputs parse results as JSON.
// It implements the ports.OutputWriter interface and formats the ParseResult.Data
// field as JSON, optionally with indentation for human readability.
type JSONWriter struct {
	// indent controls whether the output JSON should be pretty-printed
	// with indentation and newlines.
	indent bool
}

// NewJSONWriter creates a new JSONWriter.
// If indent is true, the output will be pretty-printed with indentation.
func NewJSONWriter(indent bool) *JSONWriter {
	return &JSONWriter{
		indent: indent,
	}
}

// Write formats the parse result's Data field as JSON and writes it to w.
// The schema parameter is currently unused but available for future validation.
// Returns an error if JSON encoding fails.
func (jw *JSONWriter) Write(w io.Writer, result domain.ParseResult, schema domain.Schema) error {
	enc := json.NewEncoder(w)
	if jw.indent {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(result.Data)
}
