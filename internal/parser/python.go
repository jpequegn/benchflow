package parser

import (
	"encoding/json"
	"fmt"
	"time"
)

// PythonParser implements Parser for pytest-benchmark JSON output
type PythonParser struct{}

// NewPythonParser creates a new Python benchmark parser
func NewPythonParser() *PythonParser {
	return &PythonParser{}
}

// Language returns the language this parser supports
func (p *PythonParser) Language() string {
	return "python"
}

// pythonBenchmarkJSON represents the structure of pytest-benchmark JSON output
type pythonBenchmarkJSON struct {
	Benchmarks  []pythonBenchmark      `json:"benchmarks"`
	Datetime    string                 `json:"datetime"`
	Version     string                 `json:"version"`
	MachineInfo map[string]interface{} `json:"machine_info"`
}

// pythonBenchmark represents a single benchmark entry in pytest-benchmark JSON
type pythonBenchmark struct {
	Name      string                 `json:"name"`
	FullName  string                 `json:"fullname"`
	Params    interface{}            `json:"params"`
	Group     *string                `json:"group"`
	Stats     *pythonBenchmarkStats  `json:"stats"`
	Options   map[string]interface{} `json:"options"`
	ExtraInfo string                 `json:"extra_info"`
}

// pythonBenchmarkStats represents the stats for a pytest-benchmark benchmark
type pythonBenchmarkStats struct {
	Min         float64 `json:"min"`
	Max         float64 `json:"max"`
	Mean        float64 `json:"mean"`
	StdDev      float64 `json:"stddev"`
	Median      float64 `json:"median"`
	Rounds      int64   `json:"rounds"`
	IQR         float64 `json:"iqr"`
	Q1          float64 `json:"q1"`
	Q3          float64 `json:"q3"`
	IQROutliers int64   `json:"iqr_outliers"`
	Stddevs     int64   `json:"stddevs"`
	Outliers    string  `json:"outliers"`
	Ops         float64 `json:"ops"`
	Total       float64 `json:"total"`
}

// Parse parses pytest-benchmark JSON output
// Expected format: JSON with "benchmarks" array containing benchmark results
func (p *PythonParser) Parse(output []byte) (*BenchmarkSuite, error) {
	var data pythonBenchmarkJSON
	if err := json.Unmarshal(output, &data); err != nil {
		return nil, &ParseError{
			Message: fmt.Sprintf("failed to parse JSON: %v", err),
			Input:   string(output),
		}
	}

	suite := &BenchmarkSuite{
		Language:  "python",
		Timestamp: time.Now(),
		Results:   make([]*BenchmarkResult, 0),
		Metadata:  make(map[string]string),
	}

	// Parse datetime if available
	if data.Datetime != "" {
		suite.Metadata["datetime"] = data.Datetime
	}
	if data.Version != "" {
		suite.Metadata["version"] = data.Version
	}

	// Process each benchmark
	for i, bench := range data.Benchmarks {
		// Skip benchmarks without stats
		if bench.Stats == nil {
			continue
		}

		// Validate that we have at least a rounds field to confirm the stats are valid
		if bench.Stats.Rounds == 0 {
			// Skip if no rounds recorded
			continue
		}

		// Convert time from seconds to nanoseconds
		// pytest-benchmark reports times in seconds
		timeSec := bench.Stats.Mean
		timeNs := int64(timeSec * 1e9)
		if timeNs < 0 {
			return nil, &ParseError{
				Line:    i + 1,
				Message: fmt.Sprintf("invalid mean time: %f", timeSec),
				Input:   bench.FullName,
			}
		}

		// Convert stddev from seconds to nanoseconds
		stdDevNs := int64(bench.Stats.StdDev * 1e9)
		if stdDevNs < 0 {
			return nil, &ParseError{
				Line:    i + 1,
				Message: fmt.Sprintf("invalid stddev: %f", bench.Stats.StdDev),
				Input:   bench.FullName,
			}
		}

		// Use benchmark name (without full path)
		name := bench.Name
		if name == "" {
			name = bench.FullName
		}

		// Create benchmark result
		result := &BenchmarkResult{
			Name:       name,
			Language:   "python",
			Time:       time.Duration(timeNs) * time.Nanosecond,
			Iterations: bench.Stats.Rounds,
			StdDev:     time.Duration(stdDevNs) * time.Nanosecond,
			Metadata:   make(map[string]string),
		}

		// Add throughput if available
		if bench.Stats.Ops > 0 {
			result.Throughput = &Throughput{
				Value: bench.Stats.Ops,
				Unit:  "ops/s",
			}
		}

		// Add additional stats to metadata
		result.Metadata["min"] = fmt.Sprintf("%f", bench.Stats.Min)
		result.Metadata["max"] = fmt.Sprintf("%f", bench.Stats.Max)
		result.Metadata["median"] = fmt.Sprintf("%f", bench.Stats.Median)
		result.Metadata["iqr"] = fmt.Sprintf("%f", bench.Stats.IQR)
		if bench.Stats.Q1 > 0 {
			result.Metadata["q1"] = fmt.Sprintf("%f", bench.Stats.Q1)
		}
		if bench.Stats.Q3 > 0 {
			result.Metadata["q3"] = fmt.Sprintf("%f", bench.Stats.Q3)
		}

		suite.Results = append(suite.Results, result)
	}

	if len(suite.Results) == 0 {
		return nil, &ParseError{
			Message: "no valid benchmark results found in JSON",
		}
	}

	return suite, nil
}
