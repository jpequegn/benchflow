// Package parser provides benchmark output parsers for multiple languages.
//
// The parser package defines a common interface for parsing benchmark results
// from different programming languages and testing frameworks. Each parser
// implementation converts language-specific output into a unified BenchmarkSuite
// format for aggregation and reporting.
//
// # Supported Formats
//
// Currently supported benchmark formats:
//
//   - Rust: cargo bench bencher format
//   - Python: pytest-benchmark JSON
//
// Planned support:
//
//   - Rust: criterion format
//   - Go: testing.B output
//
// # Usage
//
// Basic usage example:
//
//	parser := parser.NewRustParser()
//	output := []byte(`test bench_sort ... bench:   1,234 ns/iter (+/- 56)`)
//	suite, err := parser.Parse(output)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	for _, result := range suite.Results {
//	    fmt.Printf("%s: %v ± %v\n", result.Name, result.Time, result.StdDev)
//	}
//
// # Parser Interface
//
// All parsers implement the Parser interface:
//
//	type Parser interface {
//	    Parse(output []byte) (*BenchmarkSuite, error)
//	    Language() string
//	}
//
// This allows for polymorphic handling of different benchmark formats:
//
//	var parser parser.Parser
//	switch lang {
//	case "rust":
//	    parser = parser.NewRustParser()
//	case "python":
//	    parser = parser.NewPythonParser()
//	default:
//	    return fmt.Errorf("unsupported language: %s", lang)
//	}
//
//	suite, err := parser.Parse(benchmarkOutput)
//
// # Data Structures
//
// BenchmarkResult represents a single benchmark with:
//   - Name: benchmark identifier
//   - Language: source language
//   - Time: average execution time per iteration
//   - Iterations: number of iterations run
//   - StdDev: standard deviation of measurements
//   - Throughput: optional throughput metrics (bytes/sec, ops/sec)
//   - Metadata: additional key-value data
//
// BenchmarkSuite represents a collection of benchmark results with:
//   - Results: slice of BenchmarkResult
//   - Language: source language for all results
//   - Timestamp: when benchmarks were parsed
//   - Metadata: suite-level metadata
//
// # Error Handling
//
// Parsers return ParseError for recoverable parsing errors with context:
//
//	if err != nil {
//	    if parseErr, ok := err.(*parser.ParseError); ok {
//	        fmt.Printf("Parse error at line %d: %s\n", parseErr.Line, parseErr.Message)
//	    }
//	    return err
//	}
//
// # Rust Parser Specifics
//
// The Rust parser supports cargo bench bencher format output:
//
// Expected format:
//
//	test bench_name ... bench:   1,234 ns/iter (+/- 56)
//
// Features:
//   - Handles comma-separated numbers (1,234)
//   - Extracts benchmark name, time, and standard deviation
//   - Skips failed and ignored tests
//   - Tolerates compiler warnings and other output
//   - Parses zero nanosecond results
//   - Handles large numbers (microseconds, milliseconds)
//
// Edge cases handled:
//   - Zero variance: bench:   100 ns/iter (+/- 0)
//   - Large numbers: bench:  12,345,678 ns/iter (+/- 987,654)
//   - Failed tests: test bench_failed ... FAILED (skipped)
//   - Ignored tests: test bench_ignored ... ignored (skipped)
//
// # Python Parser Specifics
//
// The Python parser supports pytest-benchmark JSON output format:
//
// Expected format (from pytest-benchmark --json-report):
//
//	{
//	  "benchmarks": [
//	    {
//	      "name": "test_sort",
//	      "fullname": "tests/test_perf.py::test_sort",
//	      "stats": {
//	        "min": 0.0001234,
//	        "max": 0.0005678,
//	        "mean": 0.0002456,
//	        "stddev": 0.0000123,
//	        "rounds": 100,
//	        "median": 0.0002400,
//	        "ops": 4071.66
//	      }
//	    }
//	  ],
//	  "datetime": "2025-10-18T14:30:00",
//	  "version": "4.0.1"
//	}
//
// Features:
//   - Parses JSON format with automatic deserialization
//   - Converts times from seconds to nanoseconds
//   - Extracts benchmark name, mean time, standard deviation, iterations (rounds)
//   - Captures throughput metrics (ops per second)
//   - Stores quartile data and IQR in metadata
//   - Handles suite-level metadata (datetime, version)
//
// Edge cases handled:
//   - Zero-time benchmarks: mean: 0.0
//   - Large time values: mean in seconds converted to nanoseconds
//   - Missing stats field: skipped gracefully
//   - Partial stats: skipped if key metrics missing
//   - Zero throughput: skipped if ops not present
//
// # Future Extensions
//
// Planned additions:
//   - Criterion format parser with histogram data
//   - Go testing.B output parser
//   - Custom format support via configuration
package parser
