package aggregator

import (
	"time"

	"github.com/jpequegn/benchflow/internal/parser"
)

// AggregatedResult represents aggregated statistics for a single benchmark
type AggregatedResult struct {
	Name       string        `json:"name"`
	Language   string        `json:"language"`
	Mean       time.Duration `json:"mean"`
	Median     time.Duration `json:"median"`
	Min        time.Duration `json:"min"`
	Max        time.Duration `json:"max"`
	StdDev     time.Duration `json:"stddev"`
	Iterations int64         `json:"iterations"`
	Timestamp  time.Time     `json:"timestamp"`
}

// AggregatedSuite represents a collection of aggregated benchmark results
type AggregatedSuite struct {
	Results   []*AggregatedResult `json:"results"`
	Metadata  map[string]string   `json:"metadata"`
	Timestamp time.Time           `json:"timestamp"`
	Duration  time.Duration       `json:"duration"`
	Stats     *SuiteStats         `json:"stats"`
}

// SuiteStats contains overall statistics for a suite
type SuiteStats struct {
	TotalBenchmarks int           `json:"total_benchmarks"`
	TotalDuration   time.Duration `json:"total_duration"`
	FastestBench    string        `json:"fastest_bench"`
	SlowestBench    string        `json:"slowest_bench"`
	FastestTime     time.Duration `json:"fastest_time"`
	SlowestTime     time.Duration `json:"slowest_time"`
}

// Comparison represents a comparison between two benchmark runs
type Comparison struct {
	Name         string            `json:"name"`
	Baseline     *AggregatedResult `json:"baseline"`
	Current      *AggregatedResult `json:"current"`
	Delta        time.Duration     `json:"delta"`
	DeltaPercent float64           `json:"delta_percent"`
	Regression   bool              `json:"regression"`
	Improvement  bool              `json:"improvement"`
}

// ComparisonSuite represents a collection of benchmark comparisons
type ComparisonSuite struct {
	Comparisons      []*Comparison     `json:"comparisons"`
	Threshold        float64           `json:"threshold"`
	RegressionCount  int               `json:"regression_count"`
	ImprovementCount int               `json:"improvement_count"`
	UnchangedCount   int               `json:"unchanged_count"`
	Timestamp        time.Time         `json:"timestamp"`
	Metadata         map[string]string `json:"metadata"`
}

// ExportFormat represents supported export formats
type ExportFormat string

const (
	FormatJSON ExportFormat = "json"
	FormatCSV  ExportFormat = "csv"
)

// Aggregator defines the interface for result aggregation
type Aggregator interface {
	// Aggregate aggregates a benchmark suite into statistics
	Aggregate(suite *parser.BenchmarkSuite) (*AggregatedSuite, error)

	// Compare compares two aggregated suites
	Compare(baseline, current *AggregatedSuite, threshold float64) (*ComparisonSuite, error)

	// Export exports aggregated results to the specified format
	Export(suite *AggregatedSuite, format ExportFormat) ([]byte, error)
}
