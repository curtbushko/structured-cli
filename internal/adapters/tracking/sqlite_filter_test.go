package tracking_test

import (
	"context"
	"testing"
	"time"

	"github.com/curtbushko/structured-cli/internal/adapters/tracking"
	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

func TestSQLiteTracker_RecordWithFilters(t *testing.T) {
	tracker, err := tracking.NewSQLiteTracker(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteTracker() error = %v", err)
	}
	deferClose(t, tracker)

	ctx := context.Background()
	filters := []string{"small", "success"}
	record := domain.NewCommandRecordWithFilters(
		"git", []string{"status"}, 100, 50, time.Second, "/project", filters,
	)

	err = tracker.Record(ctx, record)
	if err != nil {
		t.Errorf("Record() error = %v, want nil", err)
	}

	// Verify filters were stored
	history, err := tracker.History(ctx, 1)
	if err != nil {
		t.Fatalf("History() error = %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("History() returned %d records, want 1", len(history))
	}
	if len(history[0].FiltersApplied) != 2 {
		t.Errorf("FiltersApplied length = %d, want 2", len(history[0].FiltersApplied))
	}
}

func TestSQLiteTracker_StatsByFilter(t *testing.T) {
	tracker, err := tracking.NewSQLiteTracker(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteTracker() error = %v", err)
	}
	deferClose(t, tracker)

	ctx := context.Background()

	// Insert records with different filters
	smallFilter := domain.NewCommandRecordWithFilters(
		"git", []string{"status"}, 100, 20, time.Second, "/project", []string{"small"},
	)
	successFilter := domain.NewCommandRecordWithFilters(
		"go", []string{"test"}, 200, 50, time.Second, "/project", []string{"success"},
	)
	noFilter := domain.NewCommandRecord("npm", []string{"install"}, 150, 100, time.Second, "/project")

	// 3 small filter, 2 success filter, 1 no filter
	for i := 0; i < 3; i++ {
		mustRecord(t, tracker, ctx, smallFilter)
	}
	for i := 0; i < 2; i++ {
		mustRecord(t, tracker, ctx, successFilter)
	}
	mustRecord(t, tracker, ctx, noFilter)

	filterStats, err := tracker.StatsByFilter(ctx)
	if err != nil {
		t.Errorf("StatsByFilter() error = %v, want nil", err)
	}
	if len(filterStats) < 2 {
		t.Errorf("StatsByFilter() returned %d stats, want at least 2", len(filterStats))
	}

	// Check small filter stats
	var smallStats domain.FilterStats
	for _, s := range filterStats {
		if s.FilterName == "small" {
			smallStats = s
			break
		}
	}
	if smallStats.ActivationCount != 3 {
		t.Errorf("small filter ActivationCount = %d, want 3", smallStats.ActivationCount)
	}
}

func TestSQLiteTracker_Stats_ExcludesNegativeSavingsFromFiltered(t *testing.T) {
	tracker, err := tracking.NewSQLiteTracker(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteTracker() error = %v", err)
	}
	deferClose(t, tracker)

	ctx := context.Background()

	// Record without filter - positive savings
	positiveSavings := domain.NewCommandRecord("git", []string{"status"}, 100, 50, time.Second, "/project")
	mustRecord(t, tracker, ctx, positiveSavings)

	// Record with filter - would be negative without filter, but filter made it positive
	// rawTokens=20, parsedTokens=10 after filter (savings=10)
	filteredRecord := domain.NewCommandRecordWithFilters(
		"git", []string{"status"}, 20, 10, time.Second, "/project", []string{"small"},
	)
	mustRecord(t, tracker, ctx, filteredRecord)

	// Get stats - should show filtered count
	stats, err := tracker.Stats(ctx, ports.StatsOptions{})
	if err != nil {
		t.Errorf("Stats() error = %v, want nil", err)
	}

	if stats.FilteredCount != 1 {
		t.Errorf("Stats().FilteredCount = %d, want 1", stats.FilteredCount)
	}
	if stats.TotalCommands != 2 {
		t.Errorf("Stats().TotalCommands = %d, want 2", stats.TotalCommands)
	}
}
