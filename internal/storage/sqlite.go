package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jpequegn/benchflow/internal/aggregator"
	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStorage implements Storage using SQLite
type SQLiteStorage struct {
	db   *sql.DB
	path string
}

// NewSQLiteStorage creates a new SQLite storage instance
func NewSQLiteStorage(path string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	storage := &SQLiteStorage{
		db:   db,
		path: path,
	}

	return storage, nil
}

// Init initializes the database schema
func (s *SQLiteStorage) Init() error {
	schema := `
	CREATE TABLE IF NOT EXISTS suites (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME NOT NULL,
		duration INTEGER NOT NULL,
		metadata TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_suites_timestamp ON suites(timestamp);

	CREATE TABLE IF NOT EXISTS results (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		suite_id INTEGER NOT NULL,
		name TEXT NOT NULL,
		language TEXT NOT NULL,
		mean INTEGER NOT NULL,
		median INTEGER NOT NULL,
		min INTEGER NOT NULL,
		max INTEGER NOT NULL,
		stddev INTEGER NOT NULL,
		iterations INTEGER NOT NULL,
		timestamp DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (suite_id) REFERENCES suites(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_results_suite_id ON results(suite_id);
	CREATE INDEX IF NOT EXISTS idx_results_name ON results(name);
	CREATE INDEX IF NOT EXISTS idx_results_timestamp ON results(timestamp);
	`

	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	return nil
}

// Close closes the database connection
func (s *SQLiteStorage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Save saves an aggregated suite to storage
func (s *SQLiteStorage) Save(suite *aggregator.AggregatedSuite) error {
	if suite == nil {
		return fmt.Errorf("suite cannot be nil")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	// Serialize metadata
	metadataJSON, err := json.Marshal(suite.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Insert suite
	result, err := tx.Exec(`
		INSERT INTO suites (timestamp, duration, metadata)
		VALUES (?, ?, ?)
	`, suite.Timestamp, suite.Duration.Nanoseconds(), string(metadataJSON))
	if err != nil {
		return fmt.Errorf("failed to insert suite: %w", err)
	}

	suiteID, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get suite ID: %w", err)
	}

	// Insert results
	stmt, err := tx.Prepare(`
		INSERT INTO results (suite_id, name, language, mean, median, min, max, stddev, iterations, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, r := range suite.Results {
		_, err := stmt.Exec(
			suiteID,
			r.Name,
			r.Language,
			r.Mean.Nanoseconds(),
			r.Median.Nanoseconds(),
			r.Min.Nanoseconds(),
			r.Max.Nanoseconds(),
			r.StdDev.Nanoseconds(),
			r.Iterations,
			r.Timestamp,
		)
		if err != nil {
			return fmt.Errorf("failed to insert result: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetLatest retrieves the most recent suite
func (s *SQLiteStorage) GetLatest() (*aggregator.AggregatedSuite, error) {
	row := s.db.QueryRow(`
		SELECT id, timestamp, duration, metadata
		FROM suites
		ORDER BY timestamp DESC
		LIMIT 1
	`)

	var stored StoredSuite
	var metadataJSON string

	err := row.Scan(&stored.ID, &stored.Timestamp, &stored.Duration, &metadataJSON)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query latest suite: %w", err)
	}

	return s.loadSuite(&stored, metadataJSON)
}

// GetByTimestamp retrieves a suite by timestamp
func (s *SQLiteStorage) GetByTimestamp(timestamp time.Time) (*aggregator.AggregatedSuite, error) {
	row := s.db.QueryRow(`
		SELECT id, timestamp, duration, metadata
		FROM suites
		WHERE timestamp = ?
		LIMIT 1
	`, timestamp)

	var stored StoredSuite
	var metadataJSON string

	err := row.Scan(&stored.ID, &stored.Timestamp, &stored.Duration, &metadataJSON)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query suite by timestamp: %w", err)
	}

	return s.loadSuite(&stored, metadataJSON)
}

// GetRange retrieves suites within a time range
func (s *SQLiteStorage) GetRange(start, end time.Time) ([]*aggregator.AggregatedSuite, error) {
	rows, err := s.db.Query(`
		SELECT id, timestamp, duration, metadata
		FROM suites
		WHERE timestamp BETWEEN ? AND ?
		ORDER BY timestamp ASC
	`, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query suite range: %w", err)
	}
	defer rows.Close()

	var suites []*aggregator.AggregatedSuite

	for rows.Next() {
		var stored StoredSuite
		var metadataJSON string

		if err := rows.Scan(&stored.ID, &stored.Timestamp, &stored.Duration, &metadataJSON); err != nil {
			return nil, fmt.Errorf("failed to scan suite: %w", err)
		}

		suite, err := s.loadSuite(&stored, metadataJSON)
		if err != nil {
			return nil, err
		}

		suites = append(suites, suite)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return suites, nil
}

// GetHistory retrieves all suites for a specific benchmark
func (s *SQLiteStorage) GetHistory(benchmarkName string, limit int) ([]*aggregator.AggregatedResult, error) {
	query := `
		SELECT name, language, mean, median, min, max, stddev, iterations, timestamp
		FROM results
		WHERE name = ?
		ORDER BY timestamp DESC
	`

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := s.db.Query(query, benchmarkName)
	if err != nil {
		return nil, fmt.Errorf("failed to query benchmark history: %w", err)
	}
	defer rows.Close()

	var results []*aggregator.AggregatedResult

	for rows.Next() {
		var r aggregator.AggregatedResult
		var mean, median, min, max, stddev, iterations int64

		err := rows.Scan(
			&r.Name,
			&r.Language,
			&mean,
			&median,
			&min,
			&max,
			&stddev,
			&iterations,
			&r.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan result: %w", err)
		}

		r.Mean = time.Duration(mean)
		r.Median = time.Duration(median)
		r.Min = time.Duration(min)
		r.Max = time.Duration(max)
		r.StdDev = time.Duration(stddev)
		r.Iterations = iterations

		results = append(results, &r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// Cleanup removes old records beyond retention period
func (s *SQLiteStorage) Cleanup(retentionDays int) error {
	if retentionDays <= 0 {
		return fmt.Errorf("retention days must be positive")
	}

	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	result, err := s.db.Exec(`
		DELETE FROM suites
		WHERE timestamp < ?
	`, cutoff)
	if err != nil {
		return fmt.Errorf("failed to cleanup old records: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	_ = rowsAffected // Optionally log this

	return nil
}

// loadSuite loads a complete suite with all results
func (s *SQLiteStorage) loadSuite(stored *StoredSuite, metadataJSON string) (*aggregator.AggregatedSuite, error) {
	// Deserialize metadata
	var metadata map[string]string
	if metadataJSON != "" {
		if err := json.Unmarshal([]byte(metadataJSON), &metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	// Load results
	rows, err := s.db.Query(`
		SELECT name, language, mean, median, min, max, stddev, iterations, timestamp
		FROM results
		WHERE suite_id = ?
		ORDER BY name
	`, stored.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query results: %w", err)
	}
	defer rows.Close()

	var results []*aggregator.AggregatedResult

	for rows.Next() {
		var r aggregator.AggregatedResult
		var mean, median, min, max, stddev, iterations int64

		err := rows.Scan(
			&r.Name,
			&r.Language,
			&mean,
			&median,
			&min,
			&max,
			&stddev,
			&iterations,
			&r.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan result: %w", err)
		}

		r.Mean = time.Duration(mean)
		r.Median = time.Duration(median)
		r.Min = time.Duration(min)
		r.Max = time.Duration(max)
		r.StdDev = time.Duration(stddev)
		r.Iterations = iterations

		results = append(results, &r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating results: %w", err)
	}

	suite := &aggregator.AggregatedSuite{
		Results:   results,
		Metadata:  metadata,
		Timestamp: stored.Timestamp,
		Duration:  time.Duration(stored.Duration),
	}

	// Calculate stats if results exist
	if len(results) > 0 {
		suite.Stats = calculateStats(results)
	}

	return suite, nil
}

// calculateStats calculates suite statistics from results
func calculateStats(results []*aggregator.AggregatedResult) *aggregator.SuiteStats {
	if len(results) == 0 {
		return &aggregator.SuiteStats{}
	}

	stats := &aggregator.SuiteStats{
		TotalBenchmarks: len(results),
	}

	fastest := results[0]
	slowest := results[0]

	for _, r := range results {
		stats.TotalDuration += r.Mean

		if r.Mean < fastest.Mean {
			fastest = r
		}
		if r.Mean > slowest.Mean {
			slowest = r
		}
	}

	stats.FastestBench = fastest.Name
	stats.FastestTime = fastest.Mean
	stats.SlowestBench = slowest.Name
	stats.SlowestTime = slowest.Mean

	return stats
}
