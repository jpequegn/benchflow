package comparator

import (
	"crypto/md5"
	"fmt"
	"sync"

	"github.com/jpequegn/benchflow/internal/parser"
)

// CachedComparator wraps a Comparator with LRU caching for improved performance
type CachedComparator struct {
	comparator Comparator
	cache      *LRUCache
	mu         sync.RWMutex
}

// LRUCache implements a simple LRU cache for comparison results
type LRUCache struct {
	maxSize int
	items   map[string]*cacheItem
	order   []string
	mu      sync.RWMutex
}

type cacheItem struct {
	result *ComparisonResult
	key    string
}

// NewCachedComparator creates a new cached comparator with the specified cache size
func NewCachedComparator(comparator Comparator, cacheSize int) *CachedComparator {
	if cacheSize <= 0 {
		cacheSize = 100 // Default size
	}
	return &CachedComparator{
		comparator: comparator,
		cache:      NewLRUCache(cacheSize),
	}
}

// Compare implements the Comparator interface with caching
func (cc *CachedComparator) Compare(baseline, current *parser.BenchmarkSuite) *ComparisonResult {
	key := cc.cacheKey(baseline, current)

	// Check cache
	if result, found := cc.cache.Get(key); found {
		return result
	}

	// Cache miss - perform comparison
	result := cc.comparator.Compare(baseline, current)

	// Store in cache
	cc.cache.Set(key, result)

	return result
}

// GetSignificance delegates to the wrapped comparator (no caching needed)
func (cc *CachedComparator) GetSignificance(baseline, current *parser.BenchmarkResult, confidenceLevel float64) (bool, float64) {
	return cc.comparator.GetSignificance(baseline, current, confidenceLevel)
}

// CalculateConfidenceInterval delegates to the wrapped comparator
func (cc *CachedComparator) CalculateConfidenceInterval(results []*parser.BenchmarkResult, confidenceLevel float64) (lower, upper float64) {
	return cc.comparator.CalculateConfidenceInterval(results, confidenceLevel)
}

// ClearCache clears all cached entries
func (cc *CachedComparator) ClearCache() {
	cc.cache.Clear()
}

// CacheStats returns cache statistics for monitoring
func (cc *CachedComparator) CacheStats() (size int, maxSize int) {
	return cc.cache.Size(), cc.cache.MaxSize()
}

// cacheKey generates a cache key from baseline and current suites
func (cc *CachedComparator) cacheKey(baseline, current *parser.BenchmarkSuite) string {
	// Use MD5 hash of suite contents for cache key
	h := md5.New()

	if baseline != nil {
		for _, r := range baseline.Results {
			fmt.Fprintf(h, "%s:%s:%d", r.Name, r.Language, r.Time)
		}
	}

	if current != nil {
		for _, r := range current.Results {
			fmt.Fprintf(h, "%s:%s:%d", r.Name, r.Language, r.Time)
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil))
}

// NewLRUCache creates a new LRU cache
func NewLRUCache(maxSize int) *LRUCache {
	return &LRUCache{
		maxSize: maxSize,
		items:   make(map[string]*cacheItem),
		order:   make([]string, 0, maxSize),
	}
}

// Get retrieves a value from the cache
func (lru *LRUCache) Get(key string) (*ComparisonResult, bool) {
	lru.mu.RLock()
	defer lru.mu.RUnlock()

	item, found := lru.items[key]
	if !found {
		return nil, false
	}

	return item.result, true
}

// Set stores a value in the cache
func (lru *LRUCache) Set(key string, result *ComparisonResult) {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	// If key already exists, don't update order
	if _, found := lru.items[key]; found {
		lru.items[key] = &cacheItem{result: result, key: key}
		return
	}

	// If cache is full, evict least recently used
	if len(lru.items) >= lru.maxSize {
		lru.evictOldest()
	}

	// Add new item
	lru.items[key] = &cacheItem{result: result, key: key}
	lru.order = append(lru.order, key)
}

// evictOldest removes the oldest item from the cache
func (lru *LRUCache) evictOldest() {
	if len(lru.order) == 0 {
		return
	}

	oldestKey := lru.order[0]
	delete(lru.items, oldestKey)
	lru.order = lru.order[1:]
}

// Clear removes all items from the cache
func (lru *LRUCache) Clear() {
	lru.mu.Lock()
	defer lru.mu.Unlock()

	lru.items = make(map[string]*cacheItem)
	lru.order = make([]string, 0, lru.maxSize)
}

// Size returns the current number of items in the cache
func (lru *LRUCache) Size() int {
	lru.mu.RLock()
	defer lru.mu.RUnlock()
	return len(lru.items)
}

// MaxSize returns the maximum cache size
func (lru *LRUCache) MaxSize() int {
	return lru.maxSize
}
