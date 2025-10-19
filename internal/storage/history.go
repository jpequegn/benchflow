package storage

import (
	"fmt"
	"time"

	"github.com/jpequegn/benchflow/internal/analyzer"
	"github.com/jpequegn/benchflow/internal/comparator"
)

// StoredComparison represents a stored comparison result
type StoredComparison struct {
	ID               int64
	BaselineSuiteID  int64
	CurrentSuiteID   int64
	BenchmarkName    string
	Language         string
	BaselineTimeNs   int64
	CurrentTimeNs    int64
	TimeDeltaPercent float64
	IsRegression     bool
	CommitHash       string
	BranchName       string
	Author           string
	CreatedAt        time.Time
}

// HistoryStorage extends Storage with comparison history capabilities
type HistoryStorage interface {
	// SaveComparison saves a comparison result to history
	SaveComparison(baselineSuiteID, currentSuiteID int64, result *comparator.ComparisonResult, metadata map[string]string) error

	// GetComparisonHistory retrieves comparison history for a benchmark
	GetComparisonHistory(benchmarkName, language string, limit int) ([]*analyzer.HistoricalComparison, error)

	// GetComparisonHistoryRange retrieves comparisons within a time range
	GetComparisonHistoryRange(benchmarkName, language string, start, end time.Time) ([]*analyzer.HistoricalComparison, error)

	// PruneComparisonHistory removes old comparison records
	PruneComparisonHistory(retentionDays int) error
}

// SaveComparison saves comparison results to storage
func (s *SQLiteStorage) SaveComparison(baselineSuiteID, currentSuiteID int64, result *comparator.ComparisonResult, metadata map[string]string) error {
	if result == nil || len(result.Benchmarks) == 0 {
		return fmt.Errorf("comparison result cannot be empty")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	for _, comp := range result.Benchmarks {
		query := `
		INSERT INTO comparison_history
			(baseline_suite_id, current_suite_id, benchmark_name, language,
			 baseline_time_ns, current_time_ns, time_delta_percent, is_regression,
			 commit_hash, branch_name, author, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`

		commitHash := ""
		branchName := ""
		author := ""

		if metadata != nil {
			if v, ok := metadata["commit_hash"]; ok {
				commitHash = v
			}
			if v, ok := metadata["branch_name"]; ok {
				branchName = v
			}
			if v, ok := metadata["author"]; ok {
				author = v
			}
		}

		_, err := tx.Exec(query,
			baselineSuiteID,
			currentSuiteID,
			comp.Name,
			comp.Language,
			comp.Baseline.Time.Nanoseconds(),
			comp.Current.Time.Nanoseconds(),
			comp.TimeDelta,
			comp.IsRegression,
			commitHash,
			branchName,
			author,
			time.Now(),
		)

		if err != nil {
			return fmt.Errorf("failed to insert comparison: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetComparisonHistory retrieves comparison history for a benchmark
func (s *SQLiteStorage) GetComparisonHistory(benchmarkName, language string, limit int) ([]*analyzer.HistoricalComparison, error) {
	query := `
	SELECT id, benchmark_name, language, baseline_time_ns, current_time_ns,
	       time_delta_percent, is_regression, commit_hash, branch_name, author, created_at
	FROM comparison_history
	WHERE benchmark_name = ? AND language = ?
	ORDER BY created_at DESC
	LIMIT ?
	`

	rows, err := s.db.Query(query, benchmarkName, language, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query comparison history: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var history []*analyzer.HistoricalComparison
	for rows.Next() {
		comp := &analyzer.HistoricalComparison{}
		err := rows.Scan(
			&comp.ID,
			&comp.BenchmarkName,
			&comp.Language,
			&comp.BaselineTimeNs,
			&comp.CurrentTimeNs,
			&comp.TimeDeltaPercent,
			&comp.IsRegression,
			&comp.CommitHash,
			&comp.BranchName,
			&comp.Author,
			&comp.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		history = append(history, comp)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Reverse to get oldest first (since we queried DESC)
	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}

	return history, nil
}

// GetComparisonHistoryRange retrieves comparisons within a time range
func (s *SQLiteStorage) GetComparisonHistoryRange(benchmarkName, language string, start, end time.Time) ([]*analyzer.HistoricalComparison, error) {
	query := `
	SELECT id, benchmark_name, language, baseline_time_ns, current_time_ns,
	       time_delta_percent, is_regression, commit_hash, branch_name, author, created_at
	FROM comparison_history
	WHERE benchmark_name = ? AND language = ? AND created_at BETWEEN ? AND ?
	ORDER BY created_at ASC
	`

	rows, err := s.db.Query(query, benchmarkName, language, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query comparison history range: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var history []*analyzer.HistoricalComparison
	for rows.Next() {
		comp := &analyzer.HistoricalComparison{}
		err := rows.Scan(
			&comp.ID,
			&comp.BenchmarkName,
			&comp.Language,
			&comp.BaselineTimeNs,
			&comp.CurrentTimeNs,
			&comp.TimeDeltaPercent,
			&comp.IsRegression,
			&comp.CommitHash,
			&comp.BranchName,
			&comp.Author,
			&comp.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		history = append(history, comp)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return history, nil
}

// PruneComparisonHistory removes old comparison records
func (s *SQLiteStorage) PruneComparisonHistory(retentionDays int) error {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	query := `DELETE FROM comparison_history WHERE created_at < ?`
	result, err := s.db.Exec(query, cutoff)
	if err != nil {
		return fmt.Errorf("failed to prune comparison history: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected > 0 {
		fmt.Printf("Pruned %d old comparison records\n", rowsAffected)
	}

	return nil
}

// InitComparisonHistory initializes comparison history table
func (s *SQLiteStorage) InitComparisonHistory() error {
	schema := `
	CREATE TABLE IF NOT EXISTS comparison_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		baseline_suite_id INTEGER,
		current_suite_id INTEGER,
		benchmark_name TEXT NOT NULL,
		language TEXT NOT NULL,
		baseline_time_ns INTEGER NOT NULL,
		current_time_ns INTEGER NOT NULL,
		time_delta_percent REAL NOT NULL,
		is_regression BOOLEAN NOT NULL,
		commit_hash TEXT,
		branch_name TEXT,
		author TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (baseline_suite_id) REFERENCES suites(id) ON DELETE CASCADE,
		FOREIGN KEY (current_suite_id) REFERENCES suites(id) ON DELETE CASCADE
	);

	CREATE INDEX IF NOT EXISTS idx_comparison_history_benchmark_language
		ON comparison_history(benchmark_name, language);

	CREATE INDEX IF NOT EXISTS idx_comparison_history_created_at
		ON comparison_history(created_at);

	CREATE INDEX IF NOT EXISTS idx_comparison_history_regression
		ON comparison_history(is_regression, created_at);
	`

	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create comparison history schema: %w", err)
	}

	return nil
}
