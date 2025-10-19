package storage

import (
	"os"
	"testing"
	"time"

	"github.com/jpequegn/benchflow/internal/aggregator"
)

func TestQueryOptimizer_GetLatestOptimizedWithCache(t *testing.T) {
	// Create temporary database
	tmpFile, err := os.CreateTemp("", "benchflow_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	storage, err := NewSQLiteStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	if err := storage.Init(); err != nil {
		t.Fatalf("Failed to init storage: %v", err)
	}

	// Create and save a suite
	suite := &aggregator.AggregatedSuite{
		Results: []*aggregator.AggregatedResult{
			{
				Name:       "sort",
				Language:   "go",
				Mean:       1000 * time.Nanosecond,
				Median:     950 * time.Nanosecond,
				Min:        900 * time.Nanosecond,
				Max:        1100 * time.Nanosecond,
				StdDev:     50 * time.Nanosecond,
				Iterations: 1000,
				Timestamp:  time.Now(),
			},
		},
		Metadata:  map[string]string{"version": "1.0"},
		Timestamp: time.Now(),
		Duration:  5 * time.Second,
	}

	if err := storage.Save(suite); err != nil {
		t.Fatalf("Failed to save suite: %v", err)
	}

	// Create optimizer
	optimizer := NewQueryOptimizer(storage.db, 10)

	// First query - cache miss
	result1, err := optimizer.GetLatestOptimized()
	if err != nil {
		t.Fatalf("Failed to get latest: %v", err)
	}

	if result1 == nil {
		t.Fatal("Expected result")
	}

	size1, _ := optimizer.CacheStats()
	if size1 != 1 {
		t.Errorf("Expected cache size 1 after first query, got %d", size1)
	}

	// Second query - cache hit
	result2, err := optimizer.GetLatestOptimized()
	if err != nil {
		t.Fatalf("Failed to get latest (cached): %v", err)
	}

	size2, _ := optimizer.CacheStats()
	if size2 != 1 {
		t.Errorf("Expected cache size still 1, got %d", size2)
	}

	if result1.Results[0].Name != result2.Results[0].Name {
		t.Errorf("Expected identical results")
	}
}

func TestQueryOptimizer_GetHistoryOptimizedWithPagination(t *testing.T) {
	// Create temporary database
	tmpFile, err := os.CreateTemp("", "benchflow_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	storage, err := NewSQLiteStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	if err := storage.Init(); err != nil {
		t.Fatalf("Failed to init storage: %v", err)
	}

	// Create and save multiple suites
	for i := 0; i < 5; i++ {
		suite := &aggregator.AggregatedSuite{
			Results: []*aggregator.AggregatedResult{
				{
					Name:       "sort",
					Language:   "go",
					Mean:       time.Duration(1000+i*100) * time.Nanosecond,
					Median:     950 * time.Nanosecond,
					Min:        900 * time.Nanosecond,
					Max:        1100 * time.Nanosecond,
					StdDev:     50 * time.Nanosecond,
					Iterations: 1000,
					Timestamp:  time.Now(),
				},
			},
			Metadata:  map[string]string{},
			Timestamp: time.Now(),
			Duration:  5 * time.Second,
		}

		if err := storage.Save(suite); err != nil {
			t.Fatalf("Failed to save suite: %v", err)
		}
	}

	optimizer := NewQueryOptimizer(storage.db, 10)

	// Query with limit
	results, err := optimizer.GetHistoryOptimized("sort", 2, 0)
	if err != nil {
		t.Fatalf("Failed to get history: %v", err)
	}

	if len(results) > 2 {
		t.Errorf("Expected at most 2 results, got %d", len(results))
	}

	// Query with offset
	results2, err := optimizer.GetHistoryOptimized("sort", 2, 2)
	if err != nil {
		t.Fatalf("Failed to get history with offset: %v", err)
	}

	if len(results2) > 2 {
		t.Errorf("Expected at most 2 results, got %d", len(results2))
	}
}

func TestQueryCache_Expiration(t *testing.T) {
	cache := NewQueryCache(10)

	// Add item with short TTL
	cache.SetWithTTL("key1", "value1", 50*time.Millisecond)

	// Should be available immediately
	value, found := cache.Get("key1")
	if !found || value.(string) != "value1" {
		t.Fatal("Expected to find key1")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired
	_, found = cache.Get("key1")
	if found {
		t.Fatal("Expected key1 to be expired")
	}
}

func TestQueryCache_EvictionOnFullCache(t *testing.T) {
	cache := NewQueryCache(3)

	// Fill cache
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	if cache.Size() != 3 {
		t.Errorf("Expected size 3, got %d", cache.Size())
	}

	// Add new item - should evict oldest
	cache.Set("key4", "value4")

	if cache.Size() != 3 {
		t.Errorf("Expected size 3 after eviction, got %d", cache.Size())
	}

	_, found := cache.Get("key1")
	if found {
		t.Fatal("Expected key1 to be evicted")
	}

	_, found = cache.Get("key4")
	if !found {
		t.Fatal("Expected key4 to exist")
	}
}

func TestQueryCache_Clear(t *testing.T) {
	cache := NewQueryCache(10)

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")

	if cache.Size() != 2 {
		t.Errorf("Expected size 2, got %d", cache.Size())
	}

	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", cache.Size())
	}
}

func BenchmarkQueryOptimizer_GetLatestUncached(b *testing.B) {
	tmpFile, err := os.CreateTemp("", "benchflow_bench_*.db")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	storage, err := NewSQLiteStorage(tmpFile.Name())
	if err != nil {
		b.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	if err := storage.Init(); err != nil {
		b.Fatalf("Failed to init storage: %v", err)
	}

	// Add data
	for i := 0; i < 100; i++ {
		suite := &aggregator.AggregatedSuite{
			Results: []*aggregator.AggregatedResult{
				{
					Name:       "benchmark",
					Language:   "go",
					Mean:       1000 * time.Nanosecond,
					Iterations: 1000,
					Timestamp:  time.Now(),
				},
			},
			Timestamp: time.Now(),
			Duration:  5 * time.Second,
		}
		storage.Save(suite)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.GetLatest()
	}
}

func BenchmarkQueryOptimizer_GetLatestCached(b *testing.B) {
	tmpFile, err := os.CreateTemp("", "benchflow_bench_*.db")
	if err != nil {
		b.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	storage, err := NewSQLiteStorage(tmpFile.Name())
	if err != nil {
		b.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	if err := storage.Init(); err != nil {
		b.Fatalf("Failed to init storage: %v", err)
	}

	// Add data
	for i := 0; i < 100; i++ {
		suite := &aggregator.AggregatedSuite{
			Results: []*aggregator.AggregatedResult{
				{
					Name:       "benchmark",
					Language:   "go",
					Mean:       1000 * time.Nanosecond,
					Iterations: 1000,
					Timestamp:  time.Now(),
				},
			},
			Timestamp: time.Now(),
			Duration:  5 * time.Second,
		}
		storage.Save(suite)
	}

	optimizer := NewQueryOptimizer(storage.db, 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		optimizer.GetLatestOptimized()
	}
}
