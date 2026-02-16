package ports

import (
	"io"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// OutputWriter formats and writes parse results to an output stream.
// Different implementations handle different output modes:
// - JSONWriter: outputs structured JSON
// - PassthroughWriter: outputs raw command output unchanged
//
// The schema parameter enables format-specific behavior (e.g., validation).
type OutputWriter interface {
	// Write formats the result and writes it to the given writer.
	// The schema provides context about the expected output structure.
	// Returns an error if writing fails.
	Write(w io.Writer, result domain.ParseResult, schema domain.Schema) error
}
