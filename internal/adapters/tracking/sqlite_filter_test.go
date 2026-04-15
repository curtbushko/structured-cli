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
		"git", []string{"status"}, 500, 100, time.Second, "/project", filters,
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
		"git", []string{"status"}, 500, 100, time.Second, "/project", []string{"small"},
	)
	successFilter := domain.NewCommandRecordWithFilters(
		"go", []string{"test"}, 600, 100, time.Second, "/project", []string{"success"},
	)
	noFilter := domain.NewCommandRecord("npm", []string{"install"}, 550, 100, time.Second, "/project")

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

	// Record without filter
	positiveSavings := domain.NewCommandRecord("git", []string{"status"}, 500, 100, time.Second, "/project")
	mustRecord(t, tracker, ctx, positiveSavings)

	// Record with filter
	filteredRecord := domain.NewCommandRecordWithFilters(
		"git", []string{"status"}, 600, 100, time.Second, "/project", []string{"small"},
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

func TestSQLiteTracker_Stats_TokenSavingsFilter(t *testing.T) {
	tracker, err := tracking.NewSQLiteTracker(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteTracker() error = %v", err)
	}
	deferClose(t, tracker)

	ctx := context.Background()

	// Records with tokens_saved between -100 and 100 (should be EXCLUDED from stats)
	// TokensSaved = RawTokens - ParsedTokens
	smallSavings := []struct {
		rawTokens    int
		parsedTokens int
		tokensSaved  int
	}{
		{100, 100, 0},    // zero savings
		{150, 100, 50},   // small positive
		{100, 150, -50},  // small negative
		{200, 100, 100},  // exactly +100 (excluded)
		{100, 200, -100}, // exactly -100 (excluded)
	}

	for _, tc := range smallSavings {
		record := domain.NewCommandRecord("git", []string{"status"}, tc.rawTokens, tc.parsedTokens, time.Second, "/project")
		mustRecord(t, tracker, ctx, record)
	}

	// Records with tokens_saved > 100 or < -100 (should be INCLUDED in stats)
	largeSavings := []struct {
		rawTokens    int
		parsedTokens int
		tokensSaved  int
	}{
		{201, 100, 101},  // just over +100
		{100, 201, -101}, // just under -100
		{500, 100, 400},  // large positive
	}

	for _, tc := range largeSavings {
		record := domain.NewCommandRecord("npm", []string{"install"}, tc.rawTokens, tc.parsedTokens, time.Second, "/project")
		mustRecord(t, tracker, ctx, record)
	}

	// Get stats - should only count records with |tokens_saved| > 100
	stats, err := tracker.Stats(ctx, ports.StatsOptions{})
	if err != nil {
		t.Fatalf("Stats() error = %v, want nil", err)
	}

	// Should only count the 3 large savings records
	if stats.TotalCommands != 3 {
		t.Errorf("Stats().TotalCommands = %d, want 3 (only records with |tokens_saved| > 100)", stats.TotalCommands)
	}

	// Total tokens saved should be 101 + (-101) + 400 = 400
	expectedTokensSaved := 101 + (-101) + 400
	if stats.TotalTokensSaved != expectedTokensSaved {
		t.Errorf("Stats().TotalTokensSaved = %d, want %d", stats.TotalTokensSaved, expectedTokensSaved)
	}
}

func TestSQLiteTracker_Stats_TokenSavingsFilter_WithProjectFilter(t *testing.T) {
	tracker, err := tracking.NewSQLiteTracker(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteTracker() error = %v", err)
	}
	deferClose(t, tracker)

	ctx := context.Background()

	// Project A: one large savings, one small savings
	largeA := domain.NewCommandRecord("git", []string{"status"}, 500, 100, time.Second, "/projectA")
	smallA := domain.NewCommandRecord("git", []string{"status"}, 150, 100, time.Second, "/projectA")
	mustRecord(t, tracker, ctx, largeA)
	mustRecord(t, tracker, ctx, smallA)

	// Project B: one large savings
	largeB := domain.NewCommandRecord("git", []string{"status"}, 600, 100, time.Second, "/projectB")
	mustRecord(t, tracker, ctx, largeB)

	// Get stats for projectA only - should only count the one large savings record
	stats, err := tracker.Stats(ctx, ports.StatsOptions{Project: "/projectA"})
	if err != nil {
		t.Fatalf("Stats() error = %v, want nil", err)
	}

	if stats.TotalCommands != 1 {
		t.Errorf("Stats().TotalCommands = %d, want 1 (only projectA records with |tokens_saved| > 100)", stats.TotalCommands)
	}
	if stats.TotalTokensSaved != 400 {
		t.Errorf("Stats().TotalTokensSaved = %d, want 400", stats.TotalTokensSaved)
	}
}

func TestSQLiteTracker_Stats_TokenSavingsFilter_WithSinceFilter(t *testing.T) {
	tracker, err := tracking.NewSQLiteTracker(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteTracker() error = %v", err)
	}
	deferClose(t, tracker)

	ctx := context.Background()

	// Insert old record with large savings
	oldRecord := domain.NewCommandRecord("git", []string{"status"}, 500, 100, time.Second, "/project")
	mustRecord(t, tracker, ctx, oldRecord)

	// Update timestamp to be 10 days ago
	mustUpdateTimestamp(t, tracker, ctx, 10*24*time.Hour)

	// Insert new record with large savings
	newRecord := domain.NewCommandRecord("git", []string{"diff"}, 600, 100, time.Second, "/project")
	mustRecord(t, tracker, ctx, newRecord)

	// Insert new record with small savings (should be excluded)
	smallRecord := domain.NewCommandRecord("git", []string{"log"}, 150, 100, time.Second, "/project")
	mustRecord(t, tracker, ctx, smallRecord)

	// Filter since 5 days ago - should only get the new large savings record
	since := time.Now().Add(-5 * 24 * time.Hour)
	stats, err := tracker.Stats(ctx, ports.StatsOptions{Since: since})
	if err != nil {
		t.Fatalf("Stats() error = %v, want nil", err)
	}

	if stats.TotalCommands != 1 {
		t.Errorf("Stats().TotalCommands = %d, want 1", stats.TotalCommands)
	}
	if stats.TotalTokensSaved != 500 {
		t.Errorf("Stats().TotalTokensSaved = %d, want 500", stats.TotalTokensSaved)
	}
}

func TestSQLiteTracker_StatsByParser_TokenSavingsFilter(t *testing.T) {
	tracker, err := tracking.NewSQLiteTracker(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteTracker() error = %v", err)
	}
	deferClose(t, tracker)

	ctx := context.Background()

	// Records with tokens_saved between -100 and 100 (should be EXCLUDED from StatsByParser)
	// TokensSaved = RawTokens - ParsedTokens
	smallSavings := []struct {
		rawTokens    int
		parsedTokens int
	}{
		{100, 100}, // zero savings
		{150, 100}, // 50 positive (excluded)
		{100, 150}, // -50 negative (excluded)
		{200, 100}, // exactly +100 (excluded)
		{100, 200}, // exactly -100 (excluded)
	}

	for _, tc := range smallSavings {
		record := domain.NewCommandRecord("git", []string{"status"}, tc.rawTokens, tc.parsedTokens, time.Second, "/project")
		mustRecord(t, tracker, ctx, record)
	}

	// Records with tokens_saved > 100 or < -100 (should be INCLUDED in StatsByParser)
	largeSavings := []struct {
		rawTokens    int
		parsedTokens int
	}{
		{301, 100}, // 201 positive (included)
		{100, 301}, // -201 negative (included)
		{500, 100}, // 400 positive (included)
	}

	for _, tc := range largeSavings {
		record := domain.NewCommandRecord("npm", []string{"install"}, tc.rawTokens, tc.parsedTokens, time.Second, "/project")
		mustRecord(t, tracker, ctx, record)
	}

	// Get per-parser stats - should only count records with |tokens_saved| > 100
	parserStats, err := tracker.StatsByParser(ctx)
	if err != nil {
		t.Fatalf("StatsByParser() error = %v, want nil", err)
	}

	// Should only have npm-install parser (git-status records all have small savings)
	if len(parserStats) != 1 {
		t.Errorf("StatsByParser() returned %d parsers, want 1 (only npm-install)", len(parserStats))
	}

	if len(parserStats) > 0 {
		if parserStats[0].ParserName != "npm-install" {
			t.Errorf("StatsByParser()[0].ParserName = %q, want %q", parserStats[0].ParserName, "npm-install")
		}
		if parserStats[0].InvocationCount != 3 {
			t.Errorf("StatsByParser()[0].InvocationCount = %d, want 3", parserStats[0].InvocationCount)
		}
		// Total tokens saved for npm-install: 201 + (-201) + 400 = 400
		expectedTokensSaved := 201 + (-201) + 400
		if parserStats[0].TotalTokensSaved != expectedTokensSaved {
			t.Errorf("StatsByParser()[0].TotalTokensSaved = %d, want %d", parserStats[0].TotalTokensSaved, expectedTokensSaved)
		}
	}
}

func TestSQLiteTracker_StatsByFilter_TokenSavingsFilter(t *testing.T) {
	tracker, err := tracking.NewSQLiteTracker(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteTracker() error = %v", err)
	}
	deferClose(t, tracker)

	ctx := context.Background()

	// Records with tokens_saved between -100 and 100 (should be EXCLUDED from StatsByFilter)
	// TokensSaved = RawTokens - ParsedTokens
	smallSavings := []struct {
		rawTokens    int
		parsedTokens int
	}{
		{100, 100}, // zero savings
		{150, 100}, // 50 positive (excluded)
		{100, 150}, // -50 negative (excluded)
		{200, 100}, // exactly +100 (excluded)
		{100, 200}, // exactly -100 (excluded)
	}

	// Add records with "small" filter but small token savings (should be excluded)
	for _, tc := range smallSavings {
		record := domain.NewCommandRecordWithFilters("git", []string{"status"}, tc.rawTokens, tc.parsedTokens, time.Second, "/project", []string{"small"})
		mustRecord(t, tracker, ctx, record)
	}

	// Records with tokens_saved > 100 or < -100 (should be INCLUDED in StatsByFilter)
	largeSavings := []struct {
		rawTokens    int
		parsedTokens int
		tokensSaved  int
	}{
		{301, 100, 201},  // 201 positive (included)
		{100, 301, -201}, // -201 negative (included)
		{500, 100, 400},  // 400 positive (included)
	}

	// Add records with "large" filter and large token savings (should be included)
	for _, tc := range largeSavings {
		record := domain.NewCommandRecordWithFilters("npm", []string{"install"}, tc.rawTokens, tc.parsedTokens, time.Second, "/project", []string{"large"})
		mustRecord(t, tracker, ctx, record)
	}

	// Get per-filter stats - should only count records with |tokens_saved| > 100
	filterStats, err := tracker.StatsByFilter(ctx)
	if err != nil {
		t.Fatalf("StatsByFilter() error = %v, want nil", err)
	}

	// Should only have "large" filter (small filter records all have small savings)
	if len(filterStats) != 1 {
		t.Errorf("StatsByFilter() returned %d filters, want 1 (only large)", len(filterStats))
	}

	if len(filterStats) > 0 {
		if filterStats[0].FilterName != "large" {
			t.Errorf("StatsByFilter()[0].FilterName = %q, want %q", filterStats[0].FilterName, "large")
		}
		if filterStats[0].ActivationCount != 3 {
			t.Errorf("StatsByFilter()[0].ActivationCount = %d, want 3", filterStats[0].ActivationCount)
		}
		// Total tokens saved for large filter: 201 + (-201) + 400 = 400
		expectedTokensSaved := 201 + (-201) + 400
		if filterStats[0].TotalTokensSaved != expectedTokensSaved {
			t.Errorf("StatsByFilter()[0].TotalTokensSaved = %d, want %d", filterStats[0].TotalTokensSaved, expectedTokensSaved)
		}
	}
}

func TestSQLiteTracker_History_TokenSavingsFilter(t *testing.T) {
	tracker, err := tracking.NewSQLiteTracker(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteTracker() error = %v", err)
	}
	deferClose(t, tracker)

	ctx := context.Background()

	// Records with tokens_saved between -100 and 100 (should be EXCLUDED from History)
	// TokensSaved = RawTokens - ParsedTokens
	smallSavings := []struct {
		name         string
		rawTokens    int
		parsedTokens int
	}{
		{"zero savings", 100, 100},
		{"50 positive", 150, 100},
		{"-50 negative", 100, 150},
		{"exactly +100", 200, 100},
		{"exactly -100", 100, 200},
	}

	for _, tc := range smallSavings {
		record := domain.NewCommandRecord("git", []string{"status"}, tc.rawTokens, tc.parsedTokens, time.Second, "/project")
		mustRecord(t, tracker, ctx, record)
	}

	// Records with tokens_saved > 100 or < -100 (should be INCLUDED in History)
	largeSavings := []struct {
		name         string
		rawTokens    int
		parsedTokens int
	}{
		{"201 positive", 301, 100},
		{"-201 negative", 100, 301},
		{"400 positive", 500, 100},
	}

	for _, tc := range largeSavings {
		record := domain.NewCommandRecord("npm", []string{"install"}, tc.rawTokens, tc.parsedTokens, time.Second, "/project")
		mustRecord(t, tracker, ctx, record)
	}

	// Get history - should only include records with |tokens_saved| > 100
	history, err := tracker.History(ctx, 10)
	if err != nil {
		t.Fatalf("History() error = %v", err)
	}

	// Should only have the 3 large savings records
	if len(history) != 3 {
		t.Errorf("History() returned %d records, want 3 (only records with |tokens_saved| > 100)", len(history))
	}

	// All returned records should be npm-install (those with large savings)
	for _, h := range history {
		if h.Command != "npm" {
			t.Errorf("History() record.Command = %q, want %q (only npm commands have large savings)", h.Command, "npm")
		}
	}
}

func TestSQLiteTracker_History_TokenSavingsFilter_LimitStillWorks(t *testing.T) {
	tracker, err := tracking.NewSQLiteTracker(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteTracker() error = %v", err)
	}
	deferClose(t, tracker)

	ctx := context.Background()

	// Insert 5 records with large savings
	for i := 0; i < 5; i++ {
		record := domain.NewCommandRecord("git", []string{"status"}, 500, 100, time.Second, "/project")
		mustRecord(t, tracker, ctx, record)
	}

	// Request only 3 - limit should still work with filter
	history, err := tracker.History(ctx, 3)
	if err != nil {
		t.Fatalf("History() error = %v", err)
	}

	if len(history) != 3 {
		t.Errorf("History() returned %d records, want 3 (limit should still work)", len(history))
	}
}

func TestSQLiteTracker_History_TokenSavingsFilter_BoundaryValues(t *testing.T) {
	tracker, err := tracking.NewSQLiteTracker(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteTracker() error = %v", err)
	}
	deferClose(t, tracker)

	ctx := context.Background()

	// Boundary test cases
	testCases := []struct {
		name         string
		rawTokens    int
		parsedTokens int
		shouldShow   bool
	}{
		{"exactly +100 (excluded)", 200, 100, false},
		{"exactly -100 (excluded)", 100, 200, false},
		{"exactly +101 (included)", 201, 100, true},
		{"exactly -101 (included)", 100, 201, true},
	}

	for _, tc := range testCases {
		record := domain.NewCommandRecord("test", []string{tc.name}, tc.rawTokens, tc.parsedTokens, time.Second, "/project")
		mustRecord(t, tracker, ctx, record)
	}

	// Get history
	history, err := tracker.History(ctx, 10)
	if err != nil {
		t.Fatalf("History() error = %v", err)
	}

	// Should only have 2 records (the +101 and -101 cases)
	if len(history) != 2 {
		t.Errorf("History() returned %d records, want 2 (only +101 and -101 boundary values)", len(history))
	}
}

// TestSQLiteTracker_FilterSmallSavings verifies that commands with small savings
// ARE stored in the database but filtered out during retrieval via History() and Stats().
func TestSQLiteTracker_FilterSmallSavings(t *testing.T) {
	tracker, err := tracking.NewSQLiteTracker(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteTracker() error = %v", err)
	}
	deferClose(t, tracker)

	ctx := context.Background()

	// Test cases with small token savings (|tokens_saved| <= 100)
	// These should be STORED but FILTERED during retrieval
	testCases := []struct {
		name         string
		rawTokens    int
		parsedTokens int
		tokensSaved  int
	}{
		{"zero savings", 100, 100, 0},
		{"positive 50", 150, 100, 50},
		{"positive 100", 200, 100, 100},
		{"negative 50", 100, 150, -50},
		{"negative 100", 100, 200, -100},
	}

	for _, tc := range testCases {
		record := domain.NewCommandRecord("git", []string{"status"}, tc.rawTokens, tc.parsedTokens, time.Second, "/project")
		err = tracker.Record(ctx, record)
		if err != nil {
			t.Fatalf("Record() error for %s = %v", tc.name, err)
		}
	}

	// Step 1: Verify ALL records ARE stored by querying database directly
	totalCount, err := tracker.CountAllRecordsForTest(ctx)
	if err != nil {
		t.Fatalf("CountAllRecordsForTest() error = %v", err)
	}
	if totalCount != len(testCases) {
		t.Errorf("Database contains %d records, want %d (all small savings should be stored)", totalCount, len(testCases))
	}

	// Step 2: Verify History() returns 0 records (due to filter)
	history, err := tracker.History(ctx, 10)
	if err != nil {
		t.Fatalf("History() error = %v", err)
	}
	if len(history) != 0 {
		t.Errorf("History() returned %d records, want 0 (all records have |tokens_saved| <= 100)", len(history))
	}

	// Step 3: Verify Stats() shows 0 commands (due to filter)
	stats, err := tracker.Stats(ctx, ports.StatsOptions{})
	if err != nil {
		t.Fatalf("Stats() error = %v", err)
	}
	if stats.TotalCommands != 0 {
		t.Errorf("Stats().TotalCommands = %d, want 0 (all records have |tokens_saved| <= 100)", stats.TotalCommands)
	}
	if stats.TotalTokensSaved != 0 {
		t.Errorf("Stats().TotalTokensSaved = %d, want 0 (all records have |tokens_saved| <= 100)", stats.TotalTokensSaved)
	}
}
