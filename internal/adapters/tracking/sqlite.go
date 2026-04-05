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
	project TEXT NOT NULL
)`

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
	savings_percent, execution_time_ns, timestamp, project
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

const insertParseFailureSQL = `
INSERT INTO parse_failures (command, error_message, fallback_success, timestamp)
VALUES (?, ?, ?, ?)`

const selectStatsSQL = `
SELECT
	COUNT(*) as total_commands,
	COALESCE(SUM(tokens_saved), 0) as total_tokens_saved,
	COALESCE(AVG(savings_percent), 0) as avg_savings_percent,
	COALESCE(SUM(execution_time_ns), 0) as total_execution_time_ns
FROM commands`

const selectHistorySQL = `
SELECT id, command, subcommands, raw_tokens, parsed_tokens, tokens_saved,
       savings_percent, execution_time_ns, timestamp, project
FROM commands
ORDER BY timestamp DESC
LIMIT ?`

const selectStatsByParserSQL = `
SELECT
	command || '-' || subcommands as parser_name,
	COUNT(*) as invocation_count,
	COALESCE(SUM(tokens_saved), 0) as total_tokens_saved,
	COALESCE(AVG(execution_time_ns), 0) as avg_execution_time_ns
FROM commands
GROUP BY command, subcommands
ORDER BY invocation_count DESC`

const cleanupCommandsSQL = `DELETE FROM commands WHERE timestamp < ?`

const cleanupParseFailuresSQL = `DELETE FROM parse_failures WHERE timestamp < ?`

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

	return &SQLiteTracker{db: db}, nil
}

// Record stores a command execution record and triggers auto-cleanup.
func (t *SQLiteTracker) Record(ctx context.Context, record domain.CommandRecord) error {
	// Auto-cleanup old records first - ignore errors as cleanup is best-effort
	_ = t.Cleanup(ctx, defaultRetention)

	subcommands := strings.Join(record.Subcommands, ",")
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
func (t *SQLiteTracker) Stats(ctx context.Context, opts ports.StatsOptions) (domain.StatsSummary, error) {
	query := selectStatsSQL
	var args []any

	var conditions []string
	if opts.Project != "" {
		conditions = append(conditions, "project = ?")
		args = append(args, opts.Project)
	}
	if !opts.Since.IsZero() {
		conditions = append(conditions, "timestamp >= ?")
		args = append(args, opts.Since)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	var totalCommands int
	var totalTokensSaved int
	var avgSavingsPercent float64
	var totalExecTimeNs int64

	err := t.db.QueryRowContext(ctx, query, args...).Scan(
		&totalCommands,
		&totalTokensSaved,
		&avgSavingsPercent,
		&totalExecTimeNs,
	)
	if err != nil {
		return domain.StatsSummary{}, err
	}

	return domain.NewStatsSummary(
		totalCommands,
		totalTokensSaved,
		avgSavingsPercent,
		time.Duration(totalExecTimeNs),
	), nil
}

// History returns the most recent command records.
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
		var command, subcommandsStr, project string
		var rawTokens, parsedTokens, tokensSaved int
		var savingsPercent float64
		var execTimeNs int64
		var timestamp time.Time

		if err := rows.Scan(
			&id, &command, &subcommandsStr, &rawTokens, &parsedTokens,
			&tokensSaved, &savingsPercent, &execTimeNs, &timestamp, &project,
		); err != nil {
			return nil, err
		}

		var subcommands []string
		if subcommandsStr != "" {
			subcommands = strings.Split(subcommandsStr, ",")
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
		})
	}

	return records, rows.Err()
}

// StatsByParser returns per-parser statistics.
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
