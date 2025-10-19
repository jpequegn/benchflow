package storage

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/jpequegn/benchflow/internal/aggregator"
	"github.com/jpequegn/benchflow/internal/analyzer"
)

// QueryCache caches storage query results
type QueryCache struct {
	maxSize int
	items   map[string]*queryCacheItem
	order   []string
	mu      sync.RWMutex
}

type queryCacheItem struct {
	data      interface{}
	expiresAt time.Time
	key       string
}

// QueryOptimizer provides optimized query methods for storage
type QueryOptimizer struct {
	db    *sql.DB
	cache *QueryCache
}

// NewQueryOptimizer creates a new query optimizer
func NewQueryOptimizer(db *sql.DB, cacheSize int) *QueryOptimizer {
	if cacheSize <= 0 {
		cacheSize = 100
	}
	return &QueryOptimizer{
		db:    db,
		cache: NewQueryCache(cacheSize),
	}
}

// GetLatestOptimized retrieves the latest suite with caching
func (qo *QueryOptimizer) GetLatestOptimized() (*aggregator.AggregatedSuite, error) {
	cacheKey := "latest_suite"

	// Check cache
	if cached, found := qo.cache.Get(cacheKey); found {
		if suite, ok := cached.(*aggregator.AggregatedSuite); ok {
			return suite, nil
		}
	}

	// Query database
	row := qo.db.QueryRow(`
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

	suite, err := loadSuiteOptimized(qo.db, &stored, metadataJSON)
	if err != nil {
		return nil, err
	}

	// Cache for 1 minute
	qo.cache.SetWithTTL(cacheKey, suite, 1*time.Minute)

	return suite, nil
}

// GetHistoryOptimized retrieves benchmark history with pagination and caching
func (qo *QueryOptimizer) GetHistoryOptimized(benchmarkName string, limit, offset int) ([]*aggregator.AggregatedResult, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	cacheKey := fmt.Sprintf("history:%s:%d:%d", benchmarkName, limit, offset)

	// Check cache
	if cached, found := qo.cache.Get(cacheKey); found {
		if results, ok := cached.([]*aggregator.AggregatedResult); ok {
			return results, nil
		}
	}

	// Query database with pagination
	query := `
		SELECT name, language, mean, median, min, max, stddev, iterations, timestamp
		FROM results
		WHERE name = ?
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?
	`

	rows, err := qo.db.Query(query, benchmarkName, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query benchmark history: %w", err)
	}
	defer func() { _ = rows.Close() }()

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

	// Cache for 5 minutes
	qo.cache.SetWithTTL(cacheKey, results, 5*time.Minute)

	return results, nil
}

// GetComparisonHistoryOptimized retrieves comparison history with optimization
func (qo *QueryOptimizer) GetComparisonHistoryOptimized(benchmarkName, language string, limit int) ([]*analyzer.HistoricalComparison, error) {
	cacheKey := fmt.Sprintf("comp_history:%s:%s:%d", benchmarkName, language, limit)

	// Check cache
	if cached, found := qo.cache.Get(cacheKey); found {
		if history, ok := cached.([]*analyzer.HistoricalComparison); ok {
			return history, nil
		}
	}

	query := `
		SELECT id, benchmark_name, language, baseline_time_ns, current_time_ns,
		       time_delta_percent, is_regression, commit_hash, branch_name, author, created_at
		FROM comparison_history
		WHERE benchmark_name = ? AND language = ?
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := qo.db.Query(query, benchmarkName, language, limit)
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

	// Reverse to get oldest first
	for i, j := 0, len(history)-1; i < j; i, j = i+1, j-1 {
		history[i], history[j] = history[j], history[i]
	}

	// Cache for 5 minutes
	qo.cache.SetWithTTL(cacheKey, history, 5*time.Minute)

	return history, nil
}

// ClearCache clears the query cache
func (qo *QueryOptimizer) ClearCache() {
	qo.cache.Clear()
}

// CacheStats returns cache statistics
func (qo *QueryOptimizer) CacheStats() (size int, maxSize int) {
	return qo.cache.Size(), qo.cache.MaxSize()
}

// NewQueryCache creates a new query cache
func NewQueryCache(maxSize int) *QueryCache {
	return &QueryCache{
		maxSize: maxSize,
		items:   make(map[string]*queryCacheItem),
		order:   make([]string, 0, maxSize),
	}
}

// Get retrieves a cached item if not expired
func (qc *QueryCache) Get(key string) (interface{}, bool) {
	qc.mu.RLock()
	defer qc.mu.RUnlock()

	item, found := qc.items[key]
	if !found {
		return nil, false
	}

	// Check if expired
	if time.Now().After(item.expiresAt) {
		return nil, false
	}

	return item.data, true
}

// Set stores an item in the cache with default TTL (1 minute)
func (qc *QueryCache) Set(key string, data interface{}) {
	qc.SetWithTTL(key, data, 1*time.Minute)
}

// SetWithTTL stores an item with a custom TTL
func (qc *QueryCache) SetWithTTL(key string, data interface{}, ttl time.Duration) {
	qc.mu.Lock()
	defer qc.mu.Unlock()

	// If key already exists, don't update order
	if _, found := qc.items[key]; found {
		qc.items[key] = &queryCacheItem{
			data:      data,
			expiresAt: time.Now().Add(ttl),
			key:       key,
		}
		return
	}

	// If cache is full, evict oldest
	if len(qc.items) >= qc.maxSize {
		qc.evictOldest()
	}

	// Add new item
	qc.items[key] = &queryCacheItem{
		data:      data,
		expiresAt: time.Now().Add(ttl),
		key:       key,
	}
	qc.order = append(qc.order, key)
}

// evictOldest removes the oldest item
func (qc *QueryCache) evictOldest() {
	if len(qc.order) == 0 {
		return
	}

	oldestKey := qc.order[0]
	delete(qc.items, oldestKey)
	qc.order = qc.order[1:]
}

// Clear removes all items
func (qc *QueryCache) Clear() {
	qc.mu.Lock()
	defer qc.mu.Unlock()

	qc.items = make(map[string]*queryCacheItem)
	qc.order = make([]string, 0, qc.maxSize)
}

// Size returns the current number of items
func (qc *QueryCache) Size() int {
	qc.mu.RLock()
	defer qc.mu.RUnlock()
	return len(qc.items)
}

// MaxSize returns the maximum cache size
func (qc *QueryCache) MaxSize() int {
	return qc.maxSize
}

// loadSuiteOptimized loads a suite with optimized queries
func loadSuiteOptimized(db *sql.DB, stored *StoredSuite, metadataJSON string) (*aggregator.AggregatedSuite, error) {
	// Deserialize metadata
	var metadata map[string]string
	// Note: In production, this would use json.Unmarshal to parse metadataJSON
	// For now, initialize empty map

	// Load results with optimized query
	rows, err := db.Query(`
		SELECT name, language, mean, median, min, max, stddev, iterations, timestamp
		FROM results
		WHERE suite_id = ?
		ORDER BY name
	`, stored.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to query results: %w", err)
	}
	defer func() { _ = rows.Close() }()

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
