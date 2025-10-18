package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// NodeJSParser implements Parser for Benchmark.js output
type NodeJSParser struct{}

// NewNodeJSParser creates a new Node.js benchmark parser
func NewNodeJSParser() *NodeJSParser {
	return &NodeJSParser{}
}

// Language returns the language this parser supports
func (p *NodeJSParser) Language() string {
	return "nodejs"
}

// Parse parses Benchmark.js text output
// Expected format: test_name x ops/sec ±percentage% (runs sampled)
// Example: Array#forEach x 1,234,567 ops/sec ±1.23% (90 runs sampled)
func (p *NodeJSParser) Parse(output []byte) (*BenchmarkSuite, error) {
	suite := &BenchmarkSuite{
		Language:  "nodejs",
		Timestamp: time.Now(),
		Results:   make([]*BenchmarkResult, 0),
		Metadata:  make(map[string]string),
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	lineNum := 0

	// Regex for Benchmark.js format: name x ops/sec ±percentage% (runs sampled)
	// Pattern explanation:
	// - ^(.+?): benchmark name (non-greedy, captures everything before ' x')
	// - \s+x\s+: literal ' x ' separator
	// - ([\d,]+): operations per second (with optional commas)
	// - \s+ops/sec\s+: literal ' ops/sec '
	// - ±([\d.]+)%: margin of error percentage
	// - \s+\((\d+)\s+runs?\s+sampled\): runs sampled (singular or plural)
	benchRegex := regexp.MustCompile(
		`^(.+?)\s+x\s+([\d,]+)\s+ops/sec\s+±([\d.]+)%\s+\((\d+)\s+runs?\s+sampled\)`,
	)

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Skip non-benchmark lines (e.g., "Fastest is X")
		if !strings.Contains(line, "ops/sec") {
			continue
		}

		// Match benchmark line
		matches := benchRegex.FindStringSubmatch(line)
		if matches == nil {
			// Line contains "ops/sec" but doesn't match format - skip
			continue
		}

		// Extract fields
		// Group 1: benchmark name
		// Group 2: operations per second (with commas)
		// Group 3: margin of error percentage
		// Group 4: runs sampled
		name := strings.TrimSpace(matches[1])
		opsPerSecStr := strings.ReplaceAll(matches[2], ",", "")
		marginOfErrorStr := matches[3]
		runsStr := matches[4]

		// Parse operations per second
		opsPerSec, err := strconv.ParseFloat(opsPerSecStr, 64)
		if err != nil {
			return nil, &ParseError{
				Line:    lineNum,
				Message: fmt.Sprintf("failed to parse ops/sec: %v", err),
				Input:   line,
			}
		}

		if opsPerSec <= 0 {
			return nil, &ParseError{
				Line:    lineNum,
				Message: fmt.Sprintf("invalid ops/sec value: %f", opsPerSec),
				Input:   line,
			}
		}

		// Parse margin of error percentage
		marginOfError, err := strconv.ParseFloat(marginOfErrorStr, 64)
		if err != nil {
			return nil, &ParseError{
				Line:    lineNum,
				Message: fmt.Sprintf("failed to parse margin of error: %v", err),
				Input:   line,
			}
		}

		// Parse runs sampled
		runs, err := strconv.ParseInt(runsStr, 10, 64)
		if err != nil {
			return nil, &ParseError{
				Line:    lineNum,
				Message: fmt.Sprintf("failed to parse runs: %v", err),
				Input:   line,
			}
		}

		// Convert ops/sec to time per operation (nanoseconds)
		timePerSecSec := 1.0 / opsPerSec
		timePerOpNs := timePerSecSec * 1e9

		// Approximate standard deviation from margin of error
		// RME (Relative Margin of Error) ≈ 1.96 * StdErr / mean
		// For approximation: StdDev ≈ RME * mean / 1.96
		stdDevNs := (marginOfError / 100.0) * timePerOpNs / 1.96

		// Create benchmark result
		result := &BenchmarkResult{
			Name:       name,
			Language:   "nodejs",
			Time:       time.Duration(int64(timePerOpNs)) * time.Nanosecond,
			Iterations: runs,
			StdDev:     time.Duration(int64(stdDevNs)) * time.Nanosecond,
			Throughput: &Throughput{
				Value: opsPerSec,
				Unit:  "ops/s",
			},
			Metadata: map[string]string{
				"margin_of_error": fmt.Sprintf("%.2f%%", marginOfError),
			},
		}

		suite.Results = append(suite.Results, result)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading input: %w", err)
	}

	if len(suite.Results) == 0 {
		return nil, &ParseError{
			Message: "no benchmark results found in output",
		}
	}

	return suite, nil
}

// approximateStdDev approximates standard deviation from margin of error and mean
// This is a simplified approximation since Benchmark.js doesn't provide exact stddev
func approximateStdDev(marginOfError float64, mean float64) float64 {
	// RME (Relative Margin of Error) ≈ 1.96 * StdErr / mean
	// Solving for StdDev: StdDev ≈ (RME * mean) / 1.96
	if marginOfError == 0 {
		return 0
	}
	return (marginOfError / 100.0) * mean / 1.96
}

// validateBenchmarkResult validates a benchmark result for correctness
func validateBenchmarkResult(result *BenchmarkResult) error {
	if result.Name == "" {
		return fmt.Errorf("benchmark name cannot be empty")
	}

	if result.Time < 0 {
		return fmt.Errorf("time cannot be negative")
	}

	if result.StdDev < 0 {
		return fmt.Errorf("standard deviation cannot be negative")
	}

	if result.Iterations <= 0 {
		return fmt.Errorf("iterations must be positive")
	}

	if result.Throughput != nil && result.Throughput.Value <= 0 {
		return fmt.Errorf("throughput value must be positive")
	}

	return nil
}
