package parser

import "time"

// BenchmarkResult represents a single benchmark result
type BenchmarkResult struct {
	Name       string            // Benchmark name (e.g., "bench_sort")
	Language   string            // Language (e.g., "rust", "python", "go")
	Time       time.Duration     // Average time per iteration
	Iterations int64             // Number of iterations
	StdDev     time.Duration     // Standard deviation
	Throughput *Throughput       // Optional throughput metrics
	Metadata   map[string]string // Additional metadata
}

// Throughput represents throughput metrics (bytes/sec, ops/sec, etc.)
type Throughput struct {
	Value float64
	Unit  string // "MB/s", "ops/s", etc.
}

// BenchmarkSuite represents a collection of benchmark results
type BenchmarkSuite struct {
	Results   []*BenchmarkResult
	Language  string
	Timestamp time.Time
	Metadata  map[string]string
}

// Parser defines the interface for benchmark parsers
type Parser interface {
	// Parse parses benchmark output and returns results
	Parse(output []byte) (*BenchmarkSuite, error)

	// Language returns the language this parser supports
	Language() string
}

// ParseError represents a parsing error with context
type ParseError struct {
	Line    int
	Message string
	Input   string
}

func (e *ParseError) Error() string {
	if e.Line > 0 {
		return e.Message + " at line " + string(rune(e.Line))
	}
	return e.Message
}
