package tracking_test

import (
	"context"
	"testing"
	"time"

	"github.com/curtbushko/structured-cli/internal/adapters/tracking"
	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

func TestNoOpTracker_ImplementsInterface(t *testing.T) {
	var _ ports.Tracker = (*tracking.NoOpTracker)(nil)
}

func TestNoOpTracker_Record(t *testing.T) {
	tracker := tracking.NewNoOpTracker()
	ctx := context.Background()
	record := domain.NewCommandRecord("git", []string{"status"}, 100, 50, time.Second, "/project")

	err := tracker.Record(ctx, record)

	if err != nil {
		t.Errorf("Record() error = %v, want nil", err)
	}
}

func TestNoOpTracker_RecordFailure(t *testing.T) {
	tracker := tracking.NewNoOpTracker()
	ctx := context.Background()
	failure := domain.NewParseFailure("unknown-cmd", "no parser found", true)

	err := tracker.RecordFailure(ctx, failure)

	if err != nil {
		t.Errorf("RecordFailure() error = %v, want nil", err)
	}
}

func TestNoOpTracker_Stats(t *testing.T) {
	tracker := tracking.NewNoOpTracker()
	ctx := context.Background()

	stats, err := tracker.Stats(ctx, ports.StatsOptions{})

	if err != nil {
		t.Errorf("Stats() error = %v, want nil", err)
	}
	if stats.TotalCommands != 0 {
		t.Errorf("Stats().TotalCommands = %d, want 0", stats.TotalCommands)
	}
	if stats.TotalTokensSaved != 0 {
		t.Errorf("Stats().TotalTokensSaved = %d, want 0", stats.TotalTokensSaved)
	}
	if stats.AvgSavingsPercent != 0 {
		t.Errorf("Stats().AvgSavingsPercent = %f, want 0", stats.AvgSavingsPercent)
	}
	if stats.TotalExecutionTime != 0 {
		t.Errorf("Stats().TotalExecutionTime = %v, want 0", stats.TotalExecutionTime)
	}
}

func TestNoOpTracker_History(t *testing.T) {
	tracker := tracking.NewNoOpTracker()
	ctx := context.Background()

	history, err := tracker.History(ctx, 10)

	if err != nil {
		t.Errorf("History() error = %v, want nil", err)
	}
	if len(history) != 0 {
		t.Errorf("History() returned %d records, want 0", len(history))
	}
}

func TestNoOpTracker_StatsByParser(t *testing.T) {
	tracker := tracking.NewNoOpTracker()
	ctx := context.Background()

	parserStats, err := tracker.StatsByParser(ctx)

	if err != nil {
		t.Errorf("StatsByParser() error = %v, want nil", err)
	}
	if len(parserStats) != 0 {
		t.Errorf("StatsByParser() returned %d stats, want 0", len(parserStats))
	}
}

func TestNoOpTracker_Cleanup(t *testing.T) {
	tracker := tracking.NewNoOpTracker()
	ctx := context.Background()

	err := tracker.Cleanup(ctx, 90*24*time.Hour)

	if err != nil {
		t.Errorf("Cleanup() error = %v, want nil", err)
	}
}

func TestNoOpTracker_Close(t *testing.T) {
	tracker := tracking.NewNoOpTracker()

	err := tracker.Close()

	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}
