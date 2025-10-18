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

// GoParser implements Parser for Go testing.B output
type GoParser struct{}

// NewGoParser creates a new Go benchmark parser
func NewGoParser() *GoParser {
	return &GoParser{}
}

// Language returns the language this parser supports
func (p *GoParser) Language() string {
	return "go"
}

// Parse parses Go testing.B output
// Expected format: BenchmarkName-N  iterations  ns/op  [B/op  allocs/op]
// Example: BenchmarkSort-8  1000000  1234 ns/op  512 B/op  10 allocs/op
func (p *GoParser) Parse(output []byte) (*BenchmarkSuite, error) {
	suite := &BenchmarkSuite{
		Language:  "go",
		Timestamp: time.Now(),
		Results:   make([]*BenchmarkResult, 0),
		Metadata:  make(map[string]string),
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	lineNum := 0

	// Regex for benchmark line: BenchmarkName-N  iterations  ns/op  [B/op  allocs/op]
	// Pattern explanation:
	// - ^Benchmark(\S+): starts with "Benchmark" followed by name/suffix (no space)
	// - \s+: whitespace separator
	// - (\d+): iterations
	// - \s+: whitespace
	// - (\d+(?:\.\d+)?): time value (integer or float)
	// - \s+ns/op: literal "ns/op"
	// - (?:\s+(\d+)\s+B/op)?: optional bytes per op
	// - (?:\s+(\d+)\s+allocs/op)?: optional allocs per op
	benchRegex := regexp.MustCompile(
		`^Benchmark(\S+)\s+(\d+)\s+(\d+(?:\.\d+)?)\s+ns/op(?:\s+(\d+)\s+B/op)?(?:\s+(\d+)\s+allocs/op)?`,
	)

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and non-benchmark lines
		if line == "" || !strings.HasPrefix(line, "Benchmark") {
			continue
		}

		// Skip lines with FAIL, PASS, --- (debug output), ok, goos, goarch, pkg, cpu
		if strings.Contains(line, "FAIL") || strings.Contains(line, "PASS") ||
			strings.HasPrefix(line, "---") || strings.HasPrefix(line, "ok ") ||
			strings.HasPrefix(line, "goos:") || strings.HasPrefix(line, "goarch:") ||
			strings.HasPrefix(line, "pkg:") || strings.HasPrefix(line, "cpu:") {
			continue
		}

		// Match benchmark line
		matches := benchRegex.FindStringSubmatch(line)
		if matches == nil {
			// Line starts with "Benchmark" but doesn't match format - might be error
			continue
		}

		// Extract fields (group 0 is full match, 1+ are capture groups)
		// Group 1: name (e.g., "Sort-8")
		// Group 2: iterations
		// Group 3: time
		// Group 4: bytes per op (optional)
		// Group 5: allocs per op (optional)
		nameStr := matches[1]
		iterationsStr := matches[2]
		timeStr := matches[3]
		bytesOpStr := matches[4] // Optional
		allocsOpStr := matches[5] // Optional

		// Reconstruct full name with "Benchmark" prefix
		name := "Benchmark" + nameStr

		// Parse iterations
		iterations, err := strconv.ParseInt(iterationsStr, 10, 64)
		if err != nil {
			return nil, &ParseError{
				Line:    lineNum,
				Message: fmt.Sprintf("failed to parse iterations: %v", err),
				Input:   line,
			}
		}

		// Parse time (can be float like 10.5)
		timeFloat, err := strconv.ParseFloat(timeStr, 64)
		if err != nil {
			return nil, &ParseError{
				Line:    lineNum,
				Message: fmt.Sprintf("failed to parse time: %v", err),
				Input:   line,
			}
		}

		// Convert from nanoseconds to time.Duration
		timeNs := int64(timeFloat)
		if timeNs < 0 {
			return nil, &ParseError{
				Line:    lineNum,
				Message: fmt.Sprintf("invalid time value: %f", timeFloat),
				Input:   line,
			}
		}

		// Create benchmark result
		result := &BenchmarkResult{
			Name:       name,
			Language:   "go",
			Time:       time.Duration(timeNs) * time.Nanosecond,
			Iterations: iterations,
			StdDev:     0, // Go testing.B doesn't report stddev
			Metadata:   make(map[string]string),
		}

		// Parse optional B/op field
		if bytesOpStr != "" {
			bytesOp, err := strconv.ParseInt(bytesOpStr, 10, 64)
			if err == nil && bytesOp > 0 {
				result.Metadata["bytes_per_op"] = fmt.Sprintf("%d", bytesOp)
			}
		}

		// Parse optional allocs/op field
		if allocsOpStr != "" {
			allocsOp, err := strconv.ParseInt(allocsOpStr, 10, 64)
			if err == nil && allocsOp > 0 {
				result.Metadata["allocs_per_op"] = fmt.Sprintf("%d", allocsOp)
			}
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
