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

// RustParser implements Parser for Rust cargo bench output
type RustParser struct{}

// NewRustParser creates a new Rust benchmark parser
func NewRustParser() *RustParser {
	return &RustParser{}
}

// Language returns the language this parser supports
func (p *RustParser) Language() string {
	return "rust"
}

// Parse parses Rust cargo bench bencher format output
// Expected format: test bench_name ... bench:   1,234 ns/iter (+/- 56)
func (p *RustParser) Parse(output []byte) (*BenchmarkSuite, error) {
	suite := &BenchmarkSuite{
		Language:  "rust",
		Timestamp: time.Now(),
		Results:   make([]*BenchmarkResult, 0),
		Metadata:  make(map[string]string),
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	lineNum := 0

	// Regex for bencher format: test bench_name ... bench:   1,234 ns/iter (+/- 56)
	benchRegex := regexp.MustCompile(`^test\s+(\S+)\s+\.\.\.\s+bench:\s+([\d,]+)\s+ns/iter\s+\(\+/-\s+([\d,]+)\)`)

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and non-benchmark lines
		if line == "" || !strings.Contains(line, "bench:") {
			continue
		}

		// Match benchmark line
		matches := benchRegex.FindStringSubmatch(line)
		if matches == nil {
			// Line contains "bench:" but doesn't match format - might be error
			if strings.Contains(line, "FAILED") || strings.Contains(line, "ignored") {
				continue // Skip failed/ignored tests
			}
			continue
		}

		// Extract benchmark name, time, and std dev
		name := matches[1]
		timeStr := strings.ReplaceAll(matches[2], ",", "")
		stdDevStr := strings.ReplaceAll(matches[3], ",", "")

		// Parse time in nanoseconds
		timeNs, err := strconv.ParseInt(timeStr, 10, 64)
		if err != nil {
			return nil, &ParseError{
				Line:    lineNum,
				Message: fmt.Sprintf("failed to parse time: %v", err),
				Input:   line,
			}
		}

		// Parse std dev in nanoseconds
		stdDevNs, err := strconv.ParseInt(stdDevStr, 10, 64)
		if err != nil {
			return nil, &ParseError{
				Line:    lineNum,
				Message: fmt.Sprintf("failed to parse std dev: %v", err),
				Input:   line,
			}
		}

		// Create benchmark result
		result := &BenchmarkResult{
			Name:       name,
			Language:   "rust",
			Time:       time.Duration(timeNs) * time.Nanosecond,
			Iterations: 1, // Bencher doesn't report iterations, averaged internally
			StdDev:     time.Duration(stdDevNs) * time.Nanosecond,
			Metadata:   make(map[string]string),
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
