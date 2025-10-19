package storage

import (
	"os"
	"testing"
	"time"

	"github.com/jpequegn/benchflow/internal/aggregator"
)

func TestSQLiteStorage_Init(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	// Init should create tables
	if err := storage.Init(); err != nil {
		t.Fatalf("failed to initialize storage: %v", err)
	}

	// Verify tables exist
	var count int
	err := storage.db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('suites', 'results')").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query tables: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 tables, got %d", count)
	}
}

func TestSQLiteStorage_SaveAndGetLatest(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	suite := &aggregator.AggregatedSuite{
		Results: []*aggregator.AggregatedResult{
			{
				Name:       "bench_test",
				Language:   "rust",
				Mean:       100 * time.Nanosecond,
				Median:     100 * time.Nanosecond,
				Min:        90 * time.Nanosecond,
				Max:        110 * time.Nanosecond,
				StdDev:     10 * time.Nanosecond,
				Iterations: 1000,
				Timestamp:  time.Now(),
			},
		},
		Metadata: map[string]string{
			"version": "1.0.0",
		},
		Timestamp: time.Now(),
		Duration:  5 * time.Second,
	}

	// Save suite
	if err := storage.Save(suite); err != nil {
		t.Fatalf("failed to save suite: %v", err)
	}

	// Get latest
	latest, err := storage.GetLatest()
	if err != nil {
		t.Fatalf("failed to get latest: %v", err)
	}

	if latest == nil {
		t.Fatal("expected suite, got nil")
	}

	if len(latest.Results) != 1 {
		t.Errorf("expected 1 result, got %d", len(latest.Results))
	}

	if latest.Results[0].Name != "bench_test" {
		t.Errorf("expected name bench_test, got %s", latest.Results[0].Name)
	}

	if latest.Metadata["version"] != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", latest.Metadata["version"])
	}
}

func TestSQLiteStorage_Save_NilSuite(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	err := storage.Save(nil)
	if err == nil {
		t.Fatal("expected error for nil suite")
	}
}

func TestSQLiteStorage_GetLatest_Empty(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	latest, err := storage.GetLatest()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if latest != nil {
		t.Error("expected nil for empty database")
	}
}

func TestSQLiteStorage_GetByTimestamp(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	timestamp := time.Now().Truncate(time.Second)

	suite := &aggregator.AggregatedSuite{
		Results: []*aggregator.AggregatedResult{
			{
				Name:      "bench_test",
				Language:  "rust",
				Mean:      100 * time.Nanosecond,
				Timestamp: timestamp,
			},
		},
		Timestamp: timestamp,
		Duration:  1 * time.Second,
	}

	if err := storage.Save(suite); err != nil {
		t.Fatalf("failed to save suite: %v", err)
	}

	// Get by timestamp
	retrieved, err := storage.GetByTimestamp(timestamp)
	if err != nil {
		t.Fatalf("failed to get by timestamp: %v", err)
	}

	if retrieved == nil {
		t.Fatal("expected suite, got nil")
	}

	if !retrieved.Timestamp.Equal(timestamp) {
		t.Errorf("expected timestamp %v, got %v", timestamp, retrieved.Timestamp)
	}
}

func TestSQLiteStorage_GetByTimestamp_NotFound(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	timestamp := time.Now()

	retrieved, err := storage.GetByTimestamp(timestamp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if retrieved != nil {
		t.Error("expected nil for non-existent timestamp")
	}
}

func TestSQLiteStorage_GetRange(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	// Save multiple suites with different timestamps
	now := time.Now().Truncate(time.Second)

	for i := 0; i < 5; i++ {
		suite := &aggregator.AggregatedSuite{
			Results: []*aggregator.AggregatedResult{
				{
					Name:      "bench_test",
					Language:  "rust",
					Mean:      time.Duration(i) * time.Nanosecond,
					Timestamp: now.Add(time.Duration(i) * time.Hour),
				},
			},
			Timestamp: now.Add(time.Duration(i) * time.Hour),
			Duration:  1 * time.Second,
		}

		if err := storage.Save(suite); err != nil {
			t.Fatalf("failed to save suite %d: %v", i, err)
		}
	}

	// Get range
	start := now.Add(1 * time.Hour)
	end := now.Add(3 * time.Hour)

	suites, err := storage.GetRange(start, end)
	if err != nil {
		t.Fatalf("failed to get range: %v", err)
	}

	if len(suites) != 3 {
		t.Errorf("expected 3 suites, got %d", len(suites))
	}

	// Verify order (should be ascending)
	for i := 0; i < len(suites)-1; i++ {
		if suites[i].Timestamp.After(suites[i+1].Timestamp) {
			t.Error("suites not in ascending order")
		}
	}
}

func TestSQLiteStorage_GetRange_Empty(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	start := time.Now()
	end := start.Add(1 * time.Hour)

	suites, err := storage.GetRange(start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(suites) != 0 {
		t.Errorf("expected 0 suites, got %d", len(suites))
	}
}

func TestSQLiteStorage_GetHistory(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	// Save multiple suites with same benchmark name
	now := time.Now().Truncate(time.Second)

	for i := 0; i < 5; i++ {
		suite := &aggregator.AggregatedSuite{
			Results: []*aggregator.AggregatedResult{
				{
					Name:      "bench_target",
					Language:  "rust",
					Mean:      time.Duration(i*100) * time.Nanosecond,
					Timestamp: now.Add(time.Duration(i) * time.Hour),
				},
				{
					Name:      "bench_other",
					Language:  "rust",
					Mean:      200 * time.Nanosecond,
					Timestamp: now.Add(time.Duration(i) * time.Hour),
				},
			},
			Timestamp: now.Add(time.Duration(i) * time.Hour),
			Duration:  1 * time.Second,
		}

		if err := storage.Save(suite); err != nil {
			t.Fatalf("failed to save suite %d: %v", i, err)
		}
	}

	// Get history for bench_target
	history, err := storage.GetHistory("bench_target", 0)
	if err != nil {
		t.Fatalf("failed to get history: %v", err)
	}

	if len(history) != 5 {
		t.Errorf("expected 5 results, got %d", len(history))
	}

	// Verify order (should be descending by timestamp)
	for i := 0; i < len(history)-1; i++ {
		if history[i].Timestamp.Before(history[i+1].Timestamp) {
			t.Error("history not in descending order")
		}
	}

	// Verify only bench_target results
	for _, result := range history {
		if result.Name != "bench_target" {
			t.Errorf("expected bench_target, got %s", result.Name)
		}
	}
}

func TestSQLiteStorage_GetHistory_WithLimit(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	now := time.Now().Truncate(time.Second)

	// Save 10 suites
	for i := 0; i < 10; i++ {
		suite := &aggregator.AggregatedSuite{
			Results: []*aggregator.AggregatedResult{
				{
					Name:      "bench_test",
					Language:  "rust",
					Mean:      100 * time.Nanosecond,
					Timestamp: now.Add(time.Duration(i) * time.Hour),
				},
			},
			Timestamp: now.Add(time.Duration(i) * time.Hour),
			Duration:  1 * time.Second,
		}

		if err := storage.Save(suite); err != nil {
			t.Fatalf("failed to save suite %d: %v", i, err)
		}
	}

	// Get history with limit
	history, err := storage.GetHistory("bench_test", 5)
	if err != nil {
		t.Fatalf("failed to get history: %v", err)
	}

	if len(history) != 5 {
		t.Errorf("expected 5 results, got %d", len(history))
	}
}

func TestSQLiteStorage_Cleanup(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	now := time.Now()

	// Save old and new suites
	oldSuite := &aggregator.AggregatedSuite{
		Results: []*aggregator.AggregatedResult{
			{
				Name:      "bench_old",
				Language:  "rust",
				Mean:      100 * time.Nanosecond,
				Timestamp: now.AddDate(0, 0, -100), // 100 days ago
			},
		},
		Timestamp: now.AddDate(0, 0, -100),
		Duration:  1 * time.Second,
	}

	newSuite := &aggregator.AggregatedSuite{
		Results: []*aggregator.AggregatedResult{
			{
				Name:      "bench_new",
				Language:  "rust",
				Mean:      100 * time.Nanosecond,
				Timestamp: now,
			},
		},
		Timestamp: now,
		Duration:  1 * time.Second,
	}

	if err := storage.Save(oldSuite); err != nil {
		t.Fatalf("failed to save old suite: %v", err)
	}

	if err := storage.Save(newSuite); err != nil {
		t.Fatalf("failed to save new suite: %v", err)
	}

	// Cleanup old records (90 days retention)
	if err := storage.Cleanup(90); err != nil {
		t.Fatalf("failed to cleanup: %v", err)
	}

	// Verify old suite was deleted
	oldRetrieved, err := storage.GetByTimestamp(oldSuite.Timestamp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if oldRetrieved != nil {
		t.Error("expected old suite to be deleted")
	}

	// Verify new suite still exists
	newRetrieved, err := storage.GetByTimestamp(newSuite.Timestamp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newRetrieved == nil {
		t.Error("expected new suite to still exist")
	}
}

func TestSQLiteStorage_Cleanup_InvalidRetention(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	err := storage.Cleanup(0)
	if err == nil {
		t.Fatal("expected error for zero retention days")
	}

	err = storage.Cleanup(-1)
	if err == nil {
		t.Fatal("expected error for negative retention days")
	}
}

func TestSQLiteStorage_Close(t *testing.T) {
	storage, cleanup := setupTestStorage(t)
	defer cleanup()

	if err := storage.Close(); err != nil {
		t.Fatalf("failed to close storage: %v", err)
	}

	// Verify database is closed (operations should fail)
	err := storage.Save(&aggregator.AggregatedSuite{})
	if err == nil {
		t.Error("expected error after closing database")
	}
}

// setupTestStorage creates a test storage instance with a temporary database
func setupTestStorage(t *testing.T) (*SQLiteStorage, func()) {
	t.Helper()

	// Create temporary database file
	tmpFile, err := os.CreateTemp("", "benchflow_test_*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	_ = tmpFile.Close()

	path := tmpFile.Name()

	storage, err := NewSQLiteStorage(path)
	if err != nil {
		_ = os.Remove(path)
		t.Fatalf("failed to create storage: %v", err)
	}

	if err := storage.Init(); err != nil {
		_ = storage.Close()
		_ = os.Remove(path)
		t.Fatalf("failed to initialize storage: %v", err)
	}

	cleanup := func() {
		_ = storage.Close()
		_ = os.Remove(path)
	}

	return storage, cleanup
}
