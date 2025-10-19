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

// TypeScriptParser implements Parser for Benchmark.js output from TypeScript
// Uses identical format to Node.js parser since TypeScript compiles to JavaScript
type TypeScriptParser struct{}

// NewTypeScriptParser creates a new TypeScript benchmark parser
func NewTypeScriptParser() *TypeScriptParser {
	return &TypeScriptParser{}
}

// Language returns the language this parser supports
func (p *TypeScriptParser) Language() string {
	return "typescript"
}

// Parse parses Benchmark.js text output from TypeScript benchmarks
// Expected format: test_name x ops/sec ±percentage% (runs sampled)
// Example: StringComparison x 1,234,567 ops/sec ±1.23% (90 runs sampled)
func (p *TypeScriptParser) Parse(output []byte) (*BenchmarkSuite, error) {
	suite := &BenchmarkSuite{
		Language:  "typescript",
		Timestamp: time.Now(),
		Results:   make([]*BenchmarkResult, 0),
		Metadata:  make(map[string]string),
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	lineNum := 0

	// Regex for Benchmark.js format: name x ops/sec ±percentage% (runs sampled)
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
		// For approximation: StdDev ≈ (RME * mean) / 1.96
		stdDevNs := (marginOfError / 100.0) * timePerOpNs / 1.96

		// Create benchmark result
		result := &BenchmarkResult{
			Name:       name,
			Language:   "typescript",
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
