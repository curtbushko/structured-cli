package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/curtbushko/structured-cli/internal/adapters/cli"
	"github.com/curtbushko/structured-cli/internal/adapters/tracking"
	"github.com/curtbushko/structured-cli/internal/domain"
)

func TestStatsCommand_ByFilter(t *testing.T) {
	// Setup tracker with filter data
	tracker, err := tracking.NewSQLiteTracker(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteTracker() error = %v", err)
	}
	t.Cleanup(func() { _ = tracker.Close() })

	ctx := context.Background()

	// Insert records with different filters
	// Use token values with savings > 100 to avoid small savings filter
	smallFilter := domain.NewCommandRecordWithFilters(
		"git", []string{"status"}, 500, 100, time.Second, "/project", []string{"small"},
	)
	successFilter := domain.NewCommandRecordWithFilters(
		"go", []string{"test"}, 600, 100, time.Second, "/project", []string{"success"},
	)

	for i := 0; i < 3; i++ {
		if err := tracker.Record(ctx, smallFilter); err != nil {
			t.Fatalf("Record() error = %v", err)
		}
	}
	for i := 0; i < 2; i++ {
		if err := tracker.Record(ctx, successFilter); err != nil {
			t.Fatalf("Record() error = %v", err)
		}
	}

	// Create handler with tracker
	handler := cli.NewHandlerWithTracker(nil, nil, tracker)
	cmd := handler.RootCommand()

	t.Run("text output shows filter breakdown", func(t *testing.T) {
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"stats", "--by-filter"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		if !strings.Contains(output, "small") {
			t.Errorf("output should contain 'small' filter, got: %s", output)
		}
		if !strings.Contains(output, "success") {
			t.Errorf("output should contain 'success' filter, got: %s", output)
		}
	})

	t.Run("JSON output includes filter stats", func(t *testing.T) {
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"stats", "--by-filter", "--json"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		var result []struct {
			FilterName       string `json:"filter_name"`
			ActivationCount  int    `json:"activation_count"`
			TotalTokensSaved int    `json:"total_tokens_saved"`
		}
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("JSON unmarshal error = %v, output: %s", err, buf.String())
		}

		if len(result) < 2 {
			t.Errorf("expected at least 2 filter stats, got %d", len(result))
		}
	})
}

func TestStatsCommand_SummaryShowsFilteredCount(t *testing.T) {
	tracker, err := tracking.NewSQLiteTracker(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteTracker() error = %v", err)
	}
	t.Cleanup(func() { _ = tracker.Close() })

	ctx := context.Background()

	// Insert one unfiltered and one filtered record
	// Use token values with savings > 100 to avoid small savings filter
	unfiltered := domain.NewCommandRecord("git", []string{"status"}, 500, 100, time.Second, "/project")
	filtered := domain.NewCommandRecordWithFilters(
		"git", []string{"status"}, 600, 100, time.Second, "/project", []string{"small"},
	)

	if err := tracker.Record(ctx, unfiltered); err != nil {
		t.Fatalf("Record() error = %v", err)
	}
	if err := tracker.Record(ctx, filtered); err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	handler := cli.NewHandlerWithTracker(nil, nil, tracker)
	cmd := handler.RootCommand()

	t.Run("text output shows filtered count", func(t *testing.T) {
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"stats"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		output := buf.String()
		// Should mention filtered commands
		if !strings.Contains(output, "Filtered") && !strings.Contains(output, "filtered") {
			t.Errorf("output should mention filtered commands, got: %s", output)
		}
	})

	t.Run("JSON output includes filtered_count", func(t *testing.T) {
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetArgs([]string{"stats", "--json"})

		err := cmd.Execute()
		if err != nil {
			t.Fatalf("Execute() error = %v", err)
		}

		var result struct {
			TotalCommands int `json:"total_commands"`
			FilteredCount int `json:"filtered_count"`
		}
		if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
			t.Fatalf("JSON unmarshal error = %v, output: %s", err, buf.String())
		}

		if result.FilteredCount != 1 {
			t.Errorf("filtered_count = %d, want 1", result.FilteredCount)
		}
	})
}
