// Package aggregator provides benchmark result aggregation with statistical analysis.
//
// # Overview
//
// The aggregator package processes raw benchmark results from parsers and generates
// aggregated statistics including mean, median, standard deviation, and more. It also
// provides comparison capabilities to detect performance regressions between benchmark runs.
//
// # Features
//
//   - Statistical aggregation (mean, median, min, max, std dev)
//   - Suite-level statistics (fastest/slowest benchmarks, totals)
//   - Comparison between baseline and current results
//   - Regression detection with configurable thresholds
//   - Export to JSON and CSV formats
//
// # Usage
//
// Basic aggregation:
//
//	agg := aggregator.NewAggregator()
//
//	// Aggregate parser results
//	suite, err := agg.Aggregate(parserSuite)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Export to JSON
//	data, err := agg.Export(suite, aggregator.FormatJSON)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Comparison and regression detection:
//
//	// Compare baseline vs current
//	comparison, err := agg.Compare(baseline, current, 5.0) // 5% threshold
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Check for regressions
//	if comparison.RegressionCount > 0 {
//	    for _, comp := range comparison.Comparisons {
//	        if comp.Regression {
//	            log.Printf("%s regressed by %.2f%%\n",
//	                comp.Name, comp.DeltaPercent)
//	        }
//	    }
//	}
//
// # Statistical Calculations
//
// The aggregator calculates the following statistics:
//
//   - **Mean**: Average time across iterations
//   - **Median**: Middle value when sorted (more resistant to outliers)
//   - **Min**: Fastest observed time
//   - **Max**: Slowest observed time
//   - **StdDev**: Standard deviation (measure of variance)
//
// For suites, it also calculates:
//
//   - Total benchmarks count
//   - Total duration (sum of all means)
//   - Fastest and slowest benchmarks
//
// # Comparison Logic
//
// When comparing two benchmark runs:
//
//   - **Delta**: Absolute time difference (current - baseline)
//   - **DeltaPercent**: Percentage change ((delta / baseline) × 100)
//   - **Regression**: DeltaPercent > threshold AND positive (slower)
//   - **Improvement**: DeltaPercent > threshold AND negative (faster)
//   - **Unchanged**: |DeltaPercent| ≤ threshold
//
// Example: If baseline is 100ns and current is 120ns with 5% threshold:
//   - Delta = 20ns
//   - DeltaPercent = 20%
//   - Regression = true (20% > 5% and positive)
//
// # Export Formats
//
// ## JSON Format
//
// Exports complete aggregated results as JSON with full type information:
//
//	{
//	  "results": [
//	    {
//	      "name": "bench_sort",
//	      "language": "rust",
//	      "mean": 1234000,
//	      "median": 1200000,
//	      "stddev": 56000,
//	      ...
//	    }
//	  ],
//	  "stats": { ... },
//	  "timestamp": "2025-01-15T10:30:00Z"
//	}
//
// ## CSV Format
//
// Exports results as comma-separated values for spreadsheet analysis:
//
//	Name,Language,Mean (ns),Median (ns),Min (ns),Max (ns),StdDev (ns),Iterations
//	bench_sort,rust,1234,1200,1100,1300,56,1000
//
// # Thread Safety
//
// The DefaultAggregator is stateless and safe for concurrent use. Multiple
// goroutines can call methods on the same aggregator instance.
//
// # Performance
//
// Aggregation is O(n) where n is the number of benchmark results. Statistical
// calculations are performed in-memory with minimal overhead:
//
//   - 1000 results: ~1ms
//   - 10000 results: ~10ms
//   - Memory: ~100 bytes per result
//
// # Integration
//
// The aggregator sits between the parser and storage/reporter layers:
//
//	Parser → Aggregator → Storage/Reporter
//
// It consumes parser.BenchmarkSuite and produces AggregatedSuite which
// can be stored in SQLite or exported to various formats.
package aggregator
