package tracking_test

import (
	"context"
	"testing"
	"time"

	"github.com/curtbushko/structured-cli/internal/adapters/tracking"
	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"
)

func TestSQLiteTracker_ImplementsInterface(t *testing.T) {
	var _ ports.Tracker = (*tracking.SQLiteTracker)(nil)
}

func TestSQLiteTracker_Record(t *testing.T) {
	tracker, err := tracking.NewSQLiteTracker(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteTracker() error = %v", err)
	}
	deferClose(t, tracker)

	ctx := context.Background()
	record := domain.NewCommandRecord("git", []string{"status"}, 100, 50, time.Second, "/project")

	err = tracker.Record(ctx, record)

	if err != nil {
		t.Errorf("Record() error = %v, want nil", err)
	}

	// Verify record was stored by checking history
	history, err := tracker.History(ctx, 10)
	if err != nil {
		t.Fatalf("History() error = %v", err)
	}
	if len(history) != 1 {
		t.Errorf("History() returned %d records, want 1", len(history))
	}
	if history[0].Command != "git" {
		t.Errorf("History()[0].Command = %q, want %q", history[0].Command, "git")
	}
}

func TestSQLiteTracker_RecordFailure(t *testing.T) {
	tracker, err := tracking.NewSQLiteTracker(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteTracker() error = %v", err)
	}
	deferClose(t, tracker)

	ctx := context.Background()
	failure := domain.NewParseFailure("unknown-cmd", "no parser found", true)

	err = tracker.RecordFailure(ctx, failure)

	if err != nil {
		t.Errorf("RecordFailure() error = %v, want nil", err)
	}
}

func TestSQLiteTracker_Stats(t *testing.T) {
	tracker, err := tracking.NewSQLiteTracker(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteTracker() error = %v", err)
	}
	deferClose(t, tracker)

	ctx := context.Background()

	// Insert 5 records
	for i := 0; i < 5; i++ {
		record := domain.NewCommandRecord("git", []string{"status"}, 100, 50, time.Second, "/project")
		recordErr := tracker.Record(ctx, record)
		if recordErr != nil {
			t.Fatalf("Record() error = %v", recordErr)
		}
	}

	stats, err := tracker.Stats(ctx, ports.StatsOptions{})

	if err != nil {
		t.Errorf("Stats() error = %v, want nil", err)
	}
	if stats.TotalCommands != 5 {
		t.Errorf("Stats().TotalCommands = %d, want 5", stats.TotalCommands)
	}
	// Each record saves 50 tokens (100 - 50)
	if stats.TotalTokensSaved != 250 {
		t.Errorf("Stats().TotalTokensSaved = %d, want 250", stats.TotalTokensSaved)
	}
}

func TestSQLiteTracker_Stats_Project(t *testing.T) {
	tracker, err := tracking.NewSQLiteTracker(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteTracker() error = %v", err)
	}
	deferClose(t, tracker)

	ctx := context.Background()

	// Insert records for different projects
	record1 := domain.NewCommandRecord("git", []string{"status"}, 100, 50, time.Second, "/projectA")
	record2 := domain.NewCommandRecord("git", []string{"diff"}, 200, 100, time.Second, "/projectA")
	record3 := domain.NewCommandRecord("git", []string{"log"}, 300, 150, time.Second, "/projectB")

	mustRecord(t, tracker, ctx, record1)
	mustRecord(t, tracker, ctx, record2)
	mustRecord(t, tracker, ctx, record3)

	// Filter by projectA
	stats, err := tracker.Stats(ctx, ports.StatsOptions{Project: "/projectA"})

	if err != nil {
		t.Errorf("Stats() error = %v, want nil", err)
	}
	if stats.TotalCommands != 2 {
		t.Errorf("Stats().TotalCommands = %d, want 2", stats.TotalCommands)
	}
	// 50 + 100 = 150
	if stats.TotalTokensSaved != 150 {
		t.Errorf("Stats().TotalTokensSaved = %d, want 150", stats.TotalTokensSaved)
	}
}

func TestSQLiteTracker_History(t *testing.T) {
	tracker, err := tracking.NewSQLiteTracker(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteTracker() error = %v", err)
	}
	deferClose(t, tracker)

	ctx := context.Background()

	// Insert 10 records
	for i := 0; i < 10; i++ {
		record := domain.NewCommandRecord("git", []string{"status"}, 100, 50, time.Second, "/project")
		recordErr := tracker.Record(ctx, record)
		if recordErr != nil {
			t.Fatalf("Record() error = %v", recordErr)
		}
	}

	// Request only 5
	history, err := tracker.History(ctx, 5)

	if err != nil {
		t.Errorf("History() error = %v, want nil", err)
	}
	if len(history) != 5 {
		t.Errorf("History() returned %d records, want 5", len(history))
	}
}

func TestSQLiteTracker_History_DefaultLimit(t *testing.T) {
	tracker, err := tracking.NewSQLiteTracker(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteTracker() error = %v", err)
	}
	deferClose(t, tracker)

	ctx := context.Background()

	// Insert some records
	for i := 0; i < 5; i++ {
		record := domain.NewCommandRecord("git", []string{"status"}, 100, 50, time.Second, "/project")
		recordErr := tracker.Record(ctx, record)
		if recordErr != nil {
			t.Fatalf("Record() error = %v", recordErr)
		}
	}

	// Request with limit=0 should use default
	history, err := tracker.History(ctx, 0)

	if err != nil {
		t.Errorf("History() error = %v, want nil", err)
	}
	if len(history) != 5 {
		t.Errorf("History() returned %d records, want 5", len(history))
	}
}

func TestSQLiteTracker_StatsByParser(t *testing.T) {
	tracker, err := tracking.NewSQLiteTracker(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteTracker() error = %v", err)
	}
	deferClose(t, tracker)

	ctx := context.Background()

	// Insert records with different commands
	gitStatus := domain.NewCommandRecord("git", []string{"status"}, 100, 50, time.Second, "/project")
	gitDiff := domain.NewCommandRecord("git", []string{"diff"}, 200, 100, time.Second, "/project")
	kubectl := domain.NewCommandRecord("kubectl", []string{"get"}, 300, 150, time.Second, "/project")

	// 3 git-status, 2 git-diff, 1 kubectl-get
	for i := 0; i < 3; i++ {
		mustRecord(t, tracker, ctx, gitStatus)
	}
	for i := 0; i < 2; i++ {
		mustRecord(t, tracker, ctx, gitDiff)
	}
	mustRecord(t, tracker, ctx, kubectl)

	parserStats, err := tracker.StatsByParser(ctx)

	if err != nil {
		t.Errorf("StatsByParser() error = %v, want nil", err)
	}
	if len(parserStats) != 3 {
		t.Errorf("StatsByParser() returned %d stats, want 3", len(parserStats))
	}

	// Should be ordered by invocation count descending
	if parserStats[0].ParserName != "git-status" {
		t.Errorf("StatsByParser()[0].ParserName = %q, want %q", parserStats[0].ParserName, "git-status")
	}
	if parserStats[0].InvocationCount != 3 {
		t.Errorf("StatsByParser()[0].InvocationCount = %d, want 3", parserStats[0].InvocationCount)
	}
}

func TestSQLiteTracker_Cleanup(t *testing.T) {
	tracker, err := tracking.NewSQLiteTracker(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteTracker() error = %v", err)
	}
	deferClose(t, tracker)

	ctx := context.Background()

	// Insert a record with old timestamp by directly manipulating
	// For this test, we need to insert directly with raw SQL
	oldRecord := domain.NewCommandRecord("git", []string{"status"}, 100, 50, time.Second, "/project")
	mustRecord(t, tracker, ctx, oldRecord)

	// Update the timestamp to be old (91 days ago)
	mustUpdateTimestamp(t, tracker, ctx, 91*24*time.Hour)

	// Add a new record
	newRecord := domain.NewCommandRecord("git", []string{"diff"}, 200, 100, time.Second, "/project")
	mustRecord(t, tracker, ctx, newRecord)

	// Cleanup old records (90 days retention)
	cleanupErr := tracker.Cleanup(ctx, 90*24*time.Hour)
	if cleanupErr != nil {
		t.Fatalf("Cleanup() error = %v", cleanupErr)
	}

	// Verify only new record remains
	history, err := tracker.History(ctx, 10)
	if err != nil {
		t.Fatalf("History() error = %v", err)
	}
	if len(history) != 1 {
		t.Errorf("History() returned %d records after cleanup, want 1", len(history))
	}
}

func TestSQLiteTracker_AutoCleanup(t *testing.T) {
	tracker, err := tracking.NewSQLiteTracker(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteTracker() error = %v", err)
	}
	deferClose(t, tracker)

	ctx := context.Background()

	// Insert a record
	oldRecord := domain.NewCommandRecord("git", []string{"status"}, 100, 50, time.Second, "/project")
	mustRecord(t, tracker, ctx, oldRecord)

	// Update the timestamp to be old (91 days ago)
	mustUpdateTimestamp(t, tracker, ctx, 91*24*time.Hour)

	// Insert a new record - this should trigger auto-cleanup
	newRecord := domain.NewCommandRecord("git", []string{"diff"}, 200, 100, time.Second, "/project")
	mustRecord(t, tracker, ctx, newRecord)

	// Verify only new record remains (old one cleaned up)
	history, err := tracker.History(ctx, 10)
	if err != nil {
		t.Fatalf("History() error = %v", err)
	}
	if len(history) != 1 {
		t.Errorf("History() returned %d records after auto-cleanup, want 1", len(history))
	}
}

func TestSQLiteTracker_CreatesDatabaseDir(t *testing.T) {
	// Use a temp directory that doesn't exist
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/subdir/nested/tracking.db"

	tracker, err := tracking.NewSQLiteTracker(dbPath)
	if err != nil {
		t.Fatalf("NewSQLiteTracker() error = %v", err)
	}
	deferClose(t, tracker)

	// Verify tracker works
	ctx := context.Background()
	record := domain.NewCommandRecord("git", []string{"status"}, 100, 50, time.Second, "/project")
	recordErr := tracker.Record(ctx, record)
	if recordErr != nil {
		t.Errorf("Record() error = %v", recordErr)
	}
}

func TestSQLiteTracker_Close(t *testing.T) {
	tracker, err := tracking.NewSQLiteTracker(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteTracker() error = %v", err)
	}

	err = tracker.Close()

	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}

func TestSQLiteTracker_Stats_Since(t *testing.T) {
	tracker, err := tracking.NewSQLiteTracker(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteTracker() error = %v", err)
	}
	deferClose(t, tracker)

	ctx := context.Background()

	// Insert a record
	oldRecord := domain.NewCommandRecord("git", []string{"status"}, 100, 50, time.Second, "/project")
	mustRecord(t, tracker, ctx, oldRecord)

	// Update timestamp to be 10 days ago
	mustUpdateTimestamp(t, tracker, ctx, 10*24*time.Hour)

	// Insert another record (now)
	newRecord := domain.NewCommandRecord("git", []string{"diff"}, 200, 100, time.Second, "/project")
	mustRecord(t, tracker, ctx, newRecord)

	// Filter since 5 days ago - should only get the new record
	since := time.Now().Add(-5 * 24 * time.Hour)
	stats, err := tracker.Stats(ctx, ports.StatsOptions{Since: since})

	if err != nil {
		t.Errorf("Stats() error = %v, want nil", err)
	}
	if stats.TotalCommands != 1 {
		t.Errorf("Stats().TotalCommands = %d, want 1", stats.TotalCommands)
	}
}

// mustRecord is a test helper that records a command or fails the test.
func mustRecord(t *testing.T, tracker *tracking.SQLiteTracker, ctx context.Context, record domain.CommandRecord) {
	t.Helper()
	if err := tracker.Record(ctx, record); err != nil {
		t.Fatalf("Record() error = %v", err)
	}
}

// mustUpdateTimestamp is a test helper that updates timestamps or fails the test.
func mustUpdateTimestamp(t *testing.T, tracker *tracking.SQLiteTracker, ctx context.Context, ago time.Duration) {
	t.Helper()
	if err := tracker.UpdateTimestampForTest(ctx, ago); err != nil {
		t.Fatalf("UpdateTimestampForTest() error = %v", err)
	}
}

// deferClose returns a cleanup function for use with t.Cleanup.
func deferClose(t *testing.T, tracker *tracking.SQLiteTracker) {
	t.Helper()
	t.Cleanup(func() {
		if err := tracker.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
}
