package storage

import (
	"time"

	"github.com/jpequegn/benchflow/internal/aggregator"
)

// Storage defines the interface for benchmark result storage
type Storage interface {
	// Init initializes the storage (creates tables, etc.)
	Init() error

	// Close closes the storage connection
	Close() error

	// Save saves an aggregated suite to storage
	Save(suite *aggregator.AggregatedSuite) error

	// GetLatest retrieves the most recent suite
	GetLatest() (*aggregator.AggregatedSuite, error)

	// GetByTimestamp retrieves a suite by timestamp
	GetByTimestamp(timestamp time.Time) (*aggregator.AggregatedSuite, error)

	// GetRange retrieves suites within a time range
	GetRange(start, end time.Time) ([]*aggregator.AggregatedSuite, error)

	// GetHistory retrieves all suites for a specific benchmark
	GetHistory(benchmarkName string, limit int) ([]*aggregator.AggregatedResult, error)

	// Cleanup removes old records beyond retention period
	Cleanup(retentionDays int) error
}

// StoredSuite represents a suite stored in the database
type StoredSuite struct {
	ID        int64
	Timestamp time.Time
	Duration  int64  // Duration in nanoseconds
	Metadata  string // JSON-encoded metadata
	CreatedAt time.Time
}

// StoredResult represents a benchmark result stored in the database
type StoredResult struct {
	ID         int64
	SuiteID    int64
	Name       string
	Language   string
	Mean       int64 // Duration in nanoseconds
	Median     int64
	Min        int64
	Max        int64
	StdDev     int64
	Iterations int64
	Timestamp  time.Time
	CreatedAt  time.Time
}
