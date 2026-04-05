// Package ports defines the interfaces (contracts) for the structured-cli.
// This layer only imports from domain - never from adapters or services.
// Adapters implement these interfaces; services depend on them.
package ports

import (
	"github.com/curtbushko/structured-cli/internal/domain"
)

// Deduplicator performs deduplication on parsed command output.
// It operates on the Data field of a ParseResult after parsing.
//
// Implementations should:
//   - For raw text: collapse identical lines
//   - For JSON arrays: collapse identical objects at same level
type Deduplicator interface {
	// Dedupe takes any data structure (typically the parsed result)
	// and returns the deduplicated data along with statistics.
	//
	// Parameters:
	//   - data: the data to deduplicate (typically from ParseResult.Data)
	//
	// Returns:
	//   - the deduplicated data
	//   - statistics about the deduplication (original count, deduped count, reduction)
	Dedupe(data any) (any, domain.DedupResult)
}
