package comparator

import (
	"testing"
	"time"

	"github.com/jpequegn/benchflow/internal/parser"
)

// BenchmarkComparison_Uncached benchmarks comparison without caching
func BenchmarkComparison_Uncached(b *testing.B) {
	bc := NewBasicComparator()

	baseline := createLargeBenchmarkSuite(1000)
	current := createLargeBenchmarkSuite(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		bc.Compare(baseline, current)
	}
}

// BenchmarkComparison_Cached benchmarks comparison with caching (cache hits)
func BenchmarkComparison_Cached(b *testing.B) {
	bc := NewBasicComparator()
	cached := NewCachedComparator(bc, 100)

	baseline := createLargeBenchmarkSuite(1000)
	current := createLargeBenchmarkSuite(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cached.Compare(baseline, current)
	}
}

// BenchmarkComparison_CachedMiss benchmarks comparison with cache misses
func BenchmarkComparison_CachedMiss(b *testing.B) {
	bc := NewBasicComparator()
	cached := NewCachedComparator(bc, 10)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		baseline := createLargeBenchmarkSuite(100)
		current := createLargeBenchmarkSuite(100)
		cached.Compare(baseline, current)
	}
}

// BenchmarkLRUCache_Set benchmarks LRU cache Set operation
func BenchmarkLRUCache_Set(b *testing.B) {
	lru := NewLRUCache(1000)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := string(rune(i % 1000))
		lru.Set(key, &ComparisonResult{})
	}
}

// BenchmarkLRUCache_Get benchmarks LRU cache Get operation
func BenchmarkLRUCache_Get(b *testing.B) {
	lru := NewLRUCache(1000)

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		lru.Set(string(rune(i)), &ComparisonResult{})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := string(rune(i % 1000))
		lru.Get(key)
	}
}

// createLargeBenchmarkSuite creates a suite with many benchmarks
func createLargeBenchmarkSuite(count int) *parser.BenchmarkSuite {
	results := make([]*parser.BenchmarkResult, count)
	for i := 0; i < count; i++ {
		results[i] = &parser.BenchmarkResult{
			Name:     string(rune(i % 100)),
			Language: "go",
			Time:     time.Duration(1000+i%100) * time.Nanosecond,
			StdDev:   time.Duration(50) * time.Nanosecond,
		}
	}
	return &parser.BenchmarkSuite{Results: results}
}
