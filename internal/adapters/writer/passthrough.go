package writer

import (
	"io"

	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

// Compile-time check that PassthroughWriter implements ports.OutputWriter.
var _ ports.OutputWriter = (*PassthroughWriter)(nil)

// PassthroughWriter outputs the raw command output unchanged.
// It implements the ports.OutputWriter interface and writes the ParseResult.Raw
// field directly to the output, ignoring any structured Data.
// This is used for passthrough mode where no JSON transformation is desired.
type PassthroughWriter struct{}

// NewPassthroughWriter creates a new PassthroughWriter.
func NewPassthroughWriter() *PassthroughWriter {
	return &PassthroughWriter{}
}

// Write outputs the raw command output unchanged.
// The Data and schema fields are ignored; only result.Raw is written.
// Returns an error if writing to w fails.
func (pw *PassthroughWriter) Write(w io.Writer, result domain.ParseResult, schema domain.Schema) error {
	_, err := io.WriteString(w, result.Raw)
	return err
}
