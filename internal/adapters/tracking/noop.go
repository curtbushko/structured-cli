package tracking

import (
	"context"
	"time"

	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

// NoOpTracker is a tracker implementation that does nothing.
// It is useful for testing or when tracking is disabled.
type NoOpTracker struct{}

// NewNoOpTracker creates a new NoOpTracker instance.
func NewNoOpTracker() *NoOpTracker {
	return &NoOpTracker{}
}

// Record does nothing and returns nil.
func (t *NoOpTracker) Record(_ context.Context, _ domain.CommandRecord) error {
	return nil
}

// RecordFailure does nothing and returns nil.
func (t *NoOpTracker) RecordFailure(_ context.Context, _ domain.ParseFailure) error {
	return nil
}

// Stats returns an empty StatsSummary.
func (t *NoOpTracker) Stats(_ context.Context, _ ports.StatsOptions) (domain.StatsSummary, error) {
	return domain.StatsSummary{}, nil
}

// History returns an empty slice.
func (t *NoOpTracker) History(_ context.Context, _ int) ([]domain.CommandRecord, error) {
	return nil, nil
}

// StatsByParser returns an empty slice.
func (t *NoOpTracker) StatsByParser(_ context.Context) ([]domain.CommandStats, error) {
	return nil, nil
}

// StatsByFilter returns an empty slice.
func (t *NoOpTracker) StatsByFilter(_ context.Context) ([]domain.FilterStats, error) {
	return nil, nil
}

// Cleanup does nothing and returns nil.
func (t *NoOpTracker) Cleanup(_ context.Context, _ time.Duration) error {
	return nil
}

// Close does nothing and returns nil.
func (t *NoOpTracker) Close() error {
	return nil
}
