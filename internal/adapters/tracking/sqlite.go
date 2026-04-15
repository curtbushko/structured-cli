package tracking

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/curtbushko/structured-cli/internal/domain"
	"github.com/curtbushko/structured-cli/internal/ports"

	_ "modernc.org/sqlite" // SQLite driver for database/sql
)

// Default retention period for auto-cleanup.
const defaultRetention = 90 * 24 * time.Hour

// defaultHistoryLimit is the default number of records to return in History.
const defaultHistoryLimit = 100

// Token savings threshold for display filtering:
// The SQL queries in this file use a threshold of |tokens_saved| > 100 to filter
// records for display. Records with negligible savings (|tokens_saved| <= 100) are
// excluded from Stats, StatsByParser, StatsByFilter, and History results.
// This threshold is applied only on retrieval - all commands are still recorded
// to preserve complete data. The value 100 tokens represents roughly 75 words of text.

// SQL statements for schema creation.
const createCommandsTableSQL = `
CREATE TABLE IF NOT EXISTS commands (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	command TEXT NOT NULL,
	subcommands TEXT NOT NULL,
	raw_tokens INTEGER NOT NULL,
	parsed_tokens INTEGER NOT NULL,
	tokens_saved INTEGER NOT NULL,
	savings_percent REAL NOT NULL,
	execution_time_ns INTEGER NOT NULL,
	timestamp DATETIME NOT NULL,
	project TEXT NOT NULL,
	filters_applied TEXT NOT NULL DEFAULT ''
)`

// Migration to add filters_applied column to existing databases.
const addFiltersColumnSQL = `
ALTER TABLE commands ADD COLUMN filters_applied TEXT NOT NULL DEFAULT ''`

const createParseFailuresTableSQL = `
CREATE TABLE IF NOT EXISTS parse_failures (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	command TEXT NOT NULL,
	error_message TEXT NOT NULL,
	fallback_success INTEGER NOT NULL,
	timestamp DATETIME NOT NULL
)`

// SQL statements for CRUD operations.
const insertCommandSQL = `
INSERT INTO commands (
	command, subcommands, raw_tokens, parsed_tokens, tokens_saved,
	savings_percent, execution_time_ns, timestamp, project, filters_applied
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

const insertParseFailureSQL = `
INSERT INTO parse_failures (command, error_message, fallback_success, timestamp)
VALUES (?, ?, ?, ?)`

const selectStatsSQL = `
SELECT
	COUNT(*) as total_commands,
	COALESCE(SUM(tokens_saved), 0) as total_tokens_saved,
	COALESCE(AVG(savings_percent), 0) as avg_savings_percent,
	COALESCE(SUM(execution_time_ns), 0) as total_execution_time_ns,
	COALESCE(SUM(CASE WHEN filters_applied != '' THEN 1 ELSE 0 END), 0) as filtered_count
FROM commands`

const selectHistorySQL = `
SELECT id, command, subcommands, raw_tokens, parsed_tokens, tokens_saved,
       savings_percent, execution_time_ns, timestamp, project, filters_applied
FROM commands
WHERE tokens_saved > 100 OR tokens_saved < -100
ORDER BY timestamp DESC
LIMIT ?`

const selectStatsByParserSQL = `
SELECT
	command || '-' || subcommands as parser_name,
	COUNT(*) as invocation_count,
	COALESCE(SUM(tokens_saved), 0) as total_tokens_saved,
	COALESCE(AVG(execution_time_ns), 0) as avg_execution_time_ns
FROM commands
WHERE tokens_saved > 100 OR tokens_saved < -100
GROUP BY command, subcommands
ORDER BY invocation_count DESC`

const cleanupCommandsSQL = `DELETE FROM commands WHERE timestamp < ?`

const cleanupParseFailuresSQL = `DELETE FROM parse_failures WHERE timestamp < ?`

// selectStatsByFilterSQL returns per-filter statistics.
// It uses a WITH clause to split comma-separated filter names and aggregate.
// Only includes records where tokens_saved > 100 OR tokens_saved < -100.
const selectStatsByFilterSQL = `
WITH RECURSIVE split_filters AS (
	SELECT
		id,
		tokens_saved,
		TRIM(SUBSTR(filters_applied || ',', 1, INSTR(filters_applied || ',', ',') - 1)) AS filter_name,
		SUBSTR(filters_applied || ',', INSTR(filters_applied || ',', ',') + 1) AS remaining
	FROM commands
	WHERE filters_applied != ''
		AND (tokens_saved > 100 OR tokens_saved < -100)
	UNION ALL
	SELECT
		id,
		tokens_saved,
		TRIM(SUBSTR(remaining, 1, INSTR(remaining, ',') - 1)) AS filter_name,
		SUBSTR(remaining, INSTR(remaining, ',') + 1) AS remaining
	FROM split_filters
	WHERE remaining != ''
)
SELECT
	filter_name,
	COUNT(*) as activation_count,
	COALESCE(SUM(tokens_saved), 0) as total_tokens_saved
FROM split_filters
WHERE filter_name != ''
GROUP BY filter_name
ORDER BY activation_count DESC`

// SQLiteTracker implements the Tracker interface using SQLite.
type SQLiteTracker struct {
	db *sql.DB
}

// NewSQLiteTracker creates a new SQLiteTracker with the given database path.
// If path is ":memory:", an in-memory database is used.
// The database directory is created if it doesn't exist.
func NewSQLiteTracker(path string) (*SQLiteTracker, error) {
	return NewSQLiteTrackerWithContext(context.Background(), path)
}

// NewSQLiteTrackerWithContext creates a new SQLiteTracker with context support.
func NewSQLiteTrackerWithContext(ctx context.Context, path string) (*SQLiteTracker, error) {
	// Create directory if not in-memory
	if path != ":memory:" {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	// Create tables using context
	if _, err := db.ExecContext(ctx, createCommandsTableSQL); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, closeErr
		}
		return nil, err
	}
	if _, err := db.ExecContext(ctx, createParseFailuresTableSQL); err != nil {
		if closeErr := db.Close(); closeErr != nil {
			return nil, closeErr
		}
		return nil, err
	}

	// Run migration to add filters_applied column (ignore error if column exists)
	_, _ = db.ExecContext(ctx, addFiltersColumnSQL)

	return &SQLiteTracker{db: db}, nil
}

// Record stores a command execution record and triggers auto-cleanup.
// All commands are recorded regardless of token savings - filtering is applied
// only when retrieving data for display (Stats, StatsByParser, History, etc.).
// This "record all, filter on display" approach preserves complete data for
// potential future analysis while keeping displayed statistics meaningful.
func (t *SQLiteTracker) Record(ctx context.Context, record domain.CommandRecord) error {
	// Auto-cleanup old records first - ignore errors as cleanup is best-effort
	_ = t.Cleanup(ctx, defaultRetention)

	subcommands := strings.Join(record.Subcommands, ",")
	filtersApplied := strings.Join(record.FiltersApplied, ",")
	_, err := t.db.ExecContext(ctx, insertCommandSQL,
		record.Command,
		subcommands,
		record.RawTokens,
		record.ParsedTokens,
		record.TokensSaved,
		record.SavingsPercent,
		record.ExecutionTime.Nanoseconds(),
		record.Timestamp,
		record.Project,
		filtersApplied,
	)
	return err
}

// RecordFailure stores a parse failure record.
func (t *SQLiteTracker) RecordFailure(ctx context.Context, failure domain.ParseFailure) error {
	fallbackSuccess := 0
	if failure.FallbackSuccess {
		fallbackSuccess = 1
	}
	_, err := t.db.ExecContext(ctx, insertParseFailureSQL,
		failure.Command,
		failure.ErrorMessage,
		fallbackSuccess,
		failure.Timestamp,
	)
	return err
}

// Stats returns aggregated usage statistics.
// Results are filtered to only include records where |tokens_saved| > 100,
// excluding commands with negligible impact from the displayed statistics.
// This keeps the stats meaningful by focusing on commands that actually
// provide value (positive savings) or identify issues (negative savings).
func (t *SQLiteTracker) Stats(ctx context.Context, opts ports.StatsOptions) (domain.StatsSummary, error) {
	query := selectStatsSQL
	var args []any

	// Always filter to records with meaningful token savings (> 100 or < -100)
	conditions := []string{"(tokens_saved > 100 OR tokens_saved < -100)"}

	if opts.Project != "" {
		conditions = append(conditions, "project = ?")
		args = append(args, opts.Project)
	}
	if !opts.Since.IsZero() {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, opts.Since)
	}

	query += " WHERE " + strings.Join(conditions, " AND ")

	var totalCommands int
	var totalTokensSaved int
	var avgSavingsPercent float64
	var totalExecTimeNs int64
	var filteredCount int

	err := t.db.QueryRowContext(ctx, query, args...).Scan(
		&totalCommands,
		&totalTokensSaved,
		&avgSavingsPercent,
		&totalExecTimeNs,
		&filteredCount,
	)
	if err != nil {
		return domain.StatsSummary{}, err
	}

	return domain.NewStatsSummaryWithFiltered(
		totalCommands,
		totalTokensSaved,
		avgSavingsPercent,
		time.Duration(totalExecTimeNs),
		filteredCount,
	), nil
}

// History returns the most recent command records.
// Results are filtered to only include records where |tokens_saved| > 100,
// excluding commands with negligible token savings from the history display.
func (t *SQLiteTracker) History(ctx context.Context, limit int) ([]domain.CommandRecord, error) {
	if limit <= 0 {
		limit = defaultHistoryLimit
	}

	rows, err := t.db.QueryContext(ctx, selectHistorySQL, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var records []domain.CommandRecord
	for rows.Next() {
		var id int64
		var command, subcommandsStr, project, filtersAppliedStr string
		var rawTokens, parsedTokens, tokensSaved int
		var savingsPercent float64
		var execTimeNs int64
		var timestamp time.Time

		if err := rows.Scan(
			&id, &command, &subcommandsStr, &rawTokens, &parsedTokens,
			&tokensSaved, &savingsPercent, &execTimeNs, &timestamp, &project,
			&filtersAppliedStr,
		); err != nil {
			return nil, err
		}

		var subcommands []string
		if subcommandsStr != "" {
			subcommands = strings.Split(subcommandsStr, ",")
		}

		var filtersApplied []string
		if filtersAppliedStr != "" {
			filtersApplied = strings.Split(filtersAppliedStr, ",")
		}

		records = append(records, domain.CommandRecord{
			ID:             id,
			Command:        command,
			Subcommands:    subcommands,
			RawTokens:      rawTokens,
			ParsedTokens:   parsedTokens,
			TokensSaved:    tokensSaved,
			SavingsPercent: savingsPercent,
			ExecutionTime:  time.Duration(execTimeNs),
			Timestamp:      timestamp,
			Project:        project,
			FiltersApplied: filtersApplied,
		})
	}

	return records, rows.Err()
}

// StatsByParser returns per-parser statistics grouped by command/subcommand.
// Results are filtered to only include records where |tokens_saved| > 100,
// excluding commands with negligible token savings from per-parser statistics.
func (t *SQLiteTracker) StatsByParser(ctx context.Context) ([]domain.CommandStats, error) {
	rows, err := t.db.QueryContext(ctx, selectStatsByParserSQL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var stats []domain.CommandStats
	for rows.Next() {
		var parserName string
		var invocationCount, totalTokensSaved int
		var avgExecTimeNs float64

		if err := rows.Scan(&parserName, &invocationCount, &totalTokensSaved, &avgExecTimeNs); err != nil {
			return nil, err
		}

		stats = append(stats, domain.NewCommandStats(
			parserName,
			invocationCount,
			totalTokensSaved,
			time.Duration(avgExecTimeNs),
		))
	}

	return stats, rows.Err()
}

// StatsByFilter returns per-filter statistics showing how often each filter
// is activated and its total token savings contribution.
// Results are filtered to only include records where |tokens_saved| > 100,
// excluding commands with negligible token savings from per-filter statistics.
func (t *SQLiteTracker) StatsByFilter(ctx context.Context) ([]domain.FilterStats, error) {
	rows, err := t.db.QueryContext(ctx, selectStatsByFilterSQL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var stats []domain.FilterStats
	for rows.Next() {
		var filterName string
		var activationCount, totalTokensSaved int

		if err := rows.Scan(&filterName, &activationCount, &totalTokensSaved); err != nil {
			return nil, err
		}

		stats = append(stats, domain.NewFilterStats(
			filterName,
			activationCount,
			totalTokensSaved,
		))
	}

	return stats, rows.Err()
}

// Cleanup removes records older than the retention period.
func (t *SQLiteTracker) Cleanup(ctx context.Context, retention time.Duration) error {
	cutoff := time.Now().Add(-retention)

	if _, err := t.db.ExecContext(ctx, cleanupCommandsSQL, cutoff); err != nil {
		return err
	}
	if _, err := t.db.ExecContext(ctx, cleanupParseFailuresSQL, cutoff); err != nil {
		return err
	}

	return nil
}

// Close closes the database connection.
func (t *SQLiteTracker) Close() error {
	return t.db.Close()
}

// UpdateTimestampForTest is a test helper that updates all command timestamps
// to be the specified duration ago. This is only exported for testing.
func (t *SQLiteTracker) UpdateTimestampForTest(ctx context.Context, ago time.Duration) error {
	oldTime := time.Now().Add(-ago)
	_, err := t.db.ExecContext(ctx, "UPDATE commands SET timestamp = ?", oldTime)
	return err
}

// CountAllRecordsForTest is a test helper that returns the total number of records
// in the database without any filtering. This is only exported for testing.
func (t *SQLiteTracker) CountAllRecordsForTest(ctx context.Context) (int, error) {
	var count int
	err := t.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM commands").Scan(&count)
	return count, err
}
