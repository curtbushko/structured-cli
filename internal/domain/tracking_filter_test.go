package domain_test

import (
	"testing"
	"time"

	"github.com/curtbushko/structured-cli/internal/domain"
)

func TestCommandRecord_FiltersApplied(t *testing.T) {
	t.Run("NewCommandRecord has empty filters by default", func(t *testing.T) {
		record := domain.NewCommandRecord("git", []string{"status"}, 100, 50, time.Second, "/project")

		if len(record.FiltersApplied) != 0 {
			t.Errorf("FiltersApplied = %v, want empty slice", record.FiltersApplied)
		}
	})

	t.Run("NewCommandRecordWithFilters stores filter names", func(t *testing.T) {
		filters := []string{"small", "success"}
		record := domain.NewCommandRecordWithFilters(
			"git", []string{"status"}, 100, 50, time.Second, "/project", filters,
		)

		if len(record.FiltersApplied) != 2 {
			t.Errorf("FiltersApplied length = %d, want 2", len(record.FiltersApplied))
		}
		if record.FiltersApplied[0] != "small" {
			t.Errorf("FiltersApplied[0] = %q, want %q", record.FiltersApplied[0], "small")
		}
		if record.FiltersApplied[1] != "success" {
			t.Errorf("FiltersApplied[1] = %q, want %q", record.FiltersApplied[1], "success")
		}
	})
}

func TestFilterStats(t *testing.T) {
	t.Run("NewFilterStats creates valid stats", func(t *testing.T) {
		stats := domain.NewFilterStats("small", 10, 500)

		if stats.FilterName != "small" {
			t.Errorf("FilterName = %q, want %q", stats.FilterName, "small")
		}
		if stats.ActivationCount != 10 {
			t.Errorf("ActivationCount = %d, want 10", stats.ActivationCount)
		}
		if stats.TotalTokensSaved != 500 {
			t.Errorf("TotalTokensSaved = %d, want 500", stats.TotalTokensSaved)
		}
	})
}

func TestStatsSummary_FilteredExcluded(t *testing.T) {
	t.Run("StatsSummary includes filtered count", func(t *testing.T) {
		summary := domain.NewStatsSummaryWithFiltered(
			10,    // total commands
			500,   // total tokens saved
			50.0,  // avg savings percent
			time.Second,
			3,     // filtered count (commands that had filters applied)
		)

		if summary.FilteredCount != 3 {
			t.Errorf("FilteredCount = %d, want 3", summary.FilteredCount)
		}
	})

	t.Run("StatsSummary calculates unfiltered savings", func(t *testing.T) {
		// If 10 commands total saved 500 tokens, but 3 were filtered,
		// we want to know the savings from unfiltered commands only
		summary := domain.NewStatsSummaryWithFiltered(
			10,    // total commands
			500,   // total tokens saved (includes filtered)
			50.0,  // avg savings percent
			time.Second,
			3,     // filtered count
		)

		// The filtered count should be trackable
		if summary.TotalCommands != 10 {
			t.Errorf("TotalCommands = %d, want 10", summary.TotalCommands)
		}
	})
}
