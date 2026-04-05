// Package ports defines the interfaces (contracts) for the structured-cli.
// This layer only imports from domain - never from adapters or application.
// Adapters implement these interfaces; application depends on them.
package ports

import (
	"context"
	"time"

	"github.com/curtbushko/structured-cli/internal/domain"
)

// StatsOptions provides filtering parameters for retrieving usage statistics.
// All fields are optional; zero values indicate no filtering.
type StatsOptions struct {
	// Project filters statistics to a specific project/directory context.
	// Empty string means all projects.
	Project string

	// Since filters statistics to records created after this time.
	// Zero time means no time filtering.
	Since time.Time
}

// TrackingRecorder records command executions and parse failures.
// This is the write-side of the tracking system.
type TrackingRecorder interface {
	// Record stores a command execution record.
	// Returns an error if storage fails.
	Record(ctx context.Context, record domain.CommandRecord) error

	// RecordFailure stores a parse failure record for debugging and analytics.
	// Returns an error if storage fails.
	RecordFailure(ctx context.Context, failure domain.ParseFailure) error
}

// TrackingReader provides read access to tracking data and statistics.
// This is the query-side of the tracking system.
type TrackingReader interface {
	// Stats returns aggregated usage statistics, optionally filtered by the given options.
	// Returns a StatsSummary containing totals and averages across all matching records.
	Stats(ctx context.Context, opts StatsOptions) (domain.StatsSummary, error)

	// History returns the most recent command records, limited to the specified count.
	// Records are returned in reverse chronological order (most recent first).
	// If limit is 0, a reasonable default is used by the implementation.
	History(ctx context.Context, limit int) ([]domain.CommandRecord, error)

	// StatsByParser returns per-parser statistics showing which parsers are most used.
	// Results are typically ordered by invocation count descending.
	StatsByParser(ctx context.Context) ([]domain.CommandStats, error)
}

// TrackingMaintainer provides maintenance operations for the tracking system.
type TrackingMaintainer interface {
	// Cleanup removes records older than the specified retention duration.
	// This is typically called automatically on insert but can be invoked manually.
	// Returns an error if cleanup fails.
	Cleanup(ctx context.Context, retention time.Duration) error

	// Close releases any resources held by the tracker (database connections, etc.).
	// After Close is called, other methods may return errors.
	Close() error
}

// Tracker provides usage tracking and analytics for command executions.
// It composes TrackingRecorder, TrackingReader, and TrackingMaintainer
// for implementations that need the full tracking capability.
//
// Implementations handle storage (SQLite, in-memory, etc.) and retrieval
// of command records and parse failures.
//
// The interface supports:
// - Recording successful command executions with token metrics
// - Recording failed parse attempts for debugging
// - Querying aggregated statistics
// - Retrieving command history
// - Cleanup of old records based on retention policies
//
// Implementations must be safe for concurrent use.
//
// Consumers should depend on the smallest interface they need:
// - TrackingRecorder for write-only access
// - TrackingReader for read-only access
// - TrackingMaintainer for maintenance operations
// - Tracker for full access
type Tracker interface {
	TrackingRecorder
	TrackingReader
	TrackingMaintainer
}
