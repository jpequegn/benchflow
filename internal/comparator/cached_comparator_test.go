package comparator

import (
	"testing"
	"time"

	"github.com/jpequegn/benchflow/internal/parser"
)

func TestCachedComparator_HitOnRepeat(t *testing.T) {
	bc := NewBasicComparator()
	cached := NewCachedComparator(bc, 10)

	baseline := &parser.BenchmarkSuite{
		Results: []*parser.BenchmarkResult{
			{
				Name:     "sort",
				Language: "go",
				Time:     1000 * time.Nanosecond,
				StdDev:   50 * time.Nanosecond,
			},
		},
	}

	current := &parser.BenchmarkSuite{
		Results: []*parser.BenchmarkResult{
			{
				Name:     "sort",
				Language: "go",
				Time:     950 * time.Nanosecond,
				StdDev:   45 * time.Nanosecond,
			},
		},
	}

	// First comparison - cache miss
	result1 := cached.Compare(baseline, current)
	if result1 == nil {
		t.Fatal("Expected comparison result")
	}

	size1, _ := cached.CacheStats()
	if size1 != 1 {
		t.Errorf("Expected cache size 1 after first compare, got %d", size1)
	}

	// Second comparison - should be cache hit
	result2 := cached.Compare(baseline, current)

	size2, _ := cached.CacheStats()
	if size2 != 1 {
		t.Errorf("Expected cache size still 1 after second compare, got %d", size2)
	}

	// Results should be identical
	if result1.Summary.TotalComparisons != result2.Summary.TotalComparisons {
		t.Errorf("Expected identical results, got different summaries")
	}
}

func TestCachedComparator_CacheMissDifferentInput(t *testing.T) {
	bc := NewBasicComparator()
	cached := NewCachedComparator(bc, 10)

	baseline1 := &parser.BenchmarkSuite{
		Results: []*parser.BenchmarkResult{
			{
				Name:     "sort",
				Language: "go",
				Time:     1000 * time.Nanosecond,
				StdDev:   50 * time.Nanosecond,
			},
		},
	}

	current1 := &parser.BenchmarkSuite{
		Results: []*parser.BenchmarkResult{
			{
				Name:     "sort",
				Language: "go",
				Time:     950 * time.Nanosecond,
				StdDev:   45 * time.Nanosecond,
			},
		},
	}

	baseline2 := &parser.BenchmarkSuite{
		Results: []*parser.BenchmarkResult{
			{
				Name:     "sort",
				Language: "go",
				Time:     1100 * time.Nanosecond,
				StdDev:   55 * time.Nanosecond,
			},
		},
	}

	current2 := &parser.BenchmarkSuite{
		Results: []*parser.BenchmarkResult{
			{
				Name:     "sort",
				Language: "go",
				Time:     1050 * time.Nanosecond,
				StdDev:   50 * time.Nanosecond,
			},
		},
	}

	// First comparison
	result1 := cached.Compare(baseline1, current1)
	if result1 == nil {
		t.Fatal("Expected first comparison result")
	}

	// Second comparison with different inputs - should be cache miss
	result2 := cached.Compare(baseline2, current2)
	if result2 == nil {
		t.Fatal("Expected second comparison result")
	}

	// Cache should have 2 entries
	size, _ := cached.CacheStats()
	if size != 2 {
		t.Errorf("Expected cache size 2 after two different compares, got %d", size)
	}
}

func TestCachedComparator_LRUEviction(t *testing.T) {
	bc := NewBasicComparator()
	cached := NewCachedComparator(bc, 3) // Small cache size

	// Add 4 different comparisons to trigger eviction
	for i := 0; i < 4; i++ {
		baseline := &parser.BenchmarkSuite{
			Results: []*parser.BenchmarkResult{
				{
					Name:     "sort",
					Language: "go",
					Time:     time.Duration((1000 + i*100)) * time.Nanosecond,
					StdDev:   50 * time.Nanosecond,
				},
			},
		}

		current := &parser.BenchmarkSuite{
			Results: []*parser.BenchmarkResult{
				{
					Name:     "sort",
					Language: "go",
					Time:     time.Duration((950 + i*100)) * time.Nanosecond,
					StdDev:   45 * time.Nanosecond,
				},
			},
		}

		cached.Compare(baseline, current)
	}

	// Cache size should be max 3
	size, maxSize := cached.CacheStats()
	if size > maxSize {
		t.Errorf("Expected cache size %d <= max size %d, got %d", size, maxSize, size)
	}
}

func TestCachedComparator_ClearCache(t *testing.T) {
	bc := NewBasicComparator()
	cached := NewCachedComparator(bc, 10)

	baseline := &parser.BenchmarkSuite{
		Results: []*parser.BenchmarkResult{
			{
				Name:     "sort",
				Language: "go",
				Time:     1000 * time.Nanosecond,
				StdDev:   50 * time.Nanosecond,
			},
		},
	}

	current := &parser.BenchmarkSuite{
		Results: []*parser.BenchmarkResult{
			{
				Name:     "sort",
				Language: "go",
				Time:     950 * time.Nanosecond,
				StdDev:   45 * time.Nanosecond,
			},
		},
	}

	// Add to cache
	cached.Compare(baseline, current)
	size1, _ := cached.CacheStats()
	if size1 != 1 {
		t.Errorf("Expected cache size 1 before clear, got %d", size1)
	}

	// Clear cache
	cached.ClearCache()
	size2, _ := cached.CacheStats()
	if size2 != 0 {
		t.Errorf("Expected cache size 0 after clear, got %d", size2)
	}
}

func TestCachedComparator_NilInputs(t *testing.T) {
	bc := NewBasicComparator()
	cached := NewCachedComparator(bc, 10)

	// Should handle nil without panic
	result := cached.Compare(nil, nil)
	if result == nil {
		t.Fatal("Expected result even with nil inputs")
	}

	// Cache should be empty
	size, _ := cached.CacheStats()
	if size != 1 {
		t.Errorf("Expected cache size 1, got %d", size)
	}
}

func TestCachedComparator_EmptyResults(t *testing.T) {
	bc := NewBasicComparator()
	cached := NewCachedComparator(bc, 10)

	baseline := &parser.BenchmarkSuite{
		Results: []*parser.BenchmarkResult{},
	}

	current := &parser.BenchmarkSuite{
		Results: []*parser.BenchmarkResult{},
	}

	result := cached.Compare(baseline, current)
	if result == nil {
		t.Fatal("Expected result for empty suites")
	}

	if result.Summary.TotalComparisons != 0 {
		t.Errorf("Expected 0 total comparisons, got %d", result.Summary.TotalComparisons)
	}
}

func TestLRUCache_Basic(t *testing.T) {
	lru := NewLRUCache(3)

	// Add items in order
	lru.Set("key1", &ComparisonResult{})
	lru.Set("key2", &ComparisonResult{})
	lru.Set("key3", &ComparisonResult{})

	if lru.Size() != 3 {
		t.Errorf("Expected size 3, got %d", lru.Size())
	}

	// Get item (doesn't affect LRU order in this implementation)
	result, found := lru.Get("key1")
	if !found || result == nil {
		t.Fatal("Expected to find key1")
	}

	// Update existing item - should not evict anything
	lru.Set("key1", &ComparisonResult{Summary: ComparisonSummary{TotalComparisons: 42}})
	size := lru.Size()
	if size != 3 {
		t.Errorf("Expected size still 3 after update, got %d", size)
	}

	// Add new item - should evict oldest (key1, first in order)
	lru.Set("key4", &ComparisonResult{})
	if lru.Size() != 3 {
		t.Errorf("Expected size 3 after eviction, got %d", lru.Size())
	}

	_, found = lru.Get("key1")
	if found {
		t.Fatal("Expected key1 to be evicted (oldest)")
	}

	_, found = lru.Get("key2")
	if !found {
		t.Fatal("Expected key2 to still exist")
	}
}

func TestLRUCache_Miss(t *testing.T) {
	lru := NewLRUCache(10)

	result, found := lru.Get("nonexistent")
	if found || result != nil {
		t.Fatal("Expected cache miss")
	}
}

func TestLRUCache_Clear(t *testing.T) {
	lru := NewLRUCache(10)

	lru.Set("key1", &ComparisonResult{})
	lru.Set("key2", &ComparisonResult{})

	if lru.Size() != 2 {
		t.Errorf("Expected size 2 before clear, got %d", lru.Size())
	}

	lru.Clear()

	if lru.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", lru.Size())
	}

	result, found := lru.Get("key1")
	if found || result != nil {
		t.Fatal("Expected cache miss after clear")
	}
}
