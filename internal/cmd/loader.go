package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jpequegn/benchflow/internal/parser"
)

// LoadBenchmarkSuite loads a benchmark suite from a file (JSON or CSV)
func LoadBenchmarkSuite(filePath string) (*parser.BenchmarkSuite, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	// Determine file format by extension
	if strings.HasSuffix(filePath, ".json") {
		return loadBenchmarkFromJSON(file)
	} else if strings.HasSuffix(filePath, ".csv") {
		return loadBenchmarkFromCSV(file)
	}

	return nil, fmt.Errorf("unsupported file format: %s (must be .json or .csv)", filePath)
}

// loadBenchmarkFromJSON loads benchmark suite from JSON format
// Expected format matches the reporter JSON output structure
func loadBenchmarkFromJSON(r io.Reader) (*parser.BenchmarkSuite, error) {
	var data map[string]interface{}
	if err := json.NewDecoder(r).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Extract benchmarks array
	benchmarksData, ok := data["benchmarks"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid JSON format: missing or invalid 'benchmarks' field")
	}

	suite := &parser.BenchmarkSuite{
		Results: make([]*parser.BenchmarkResult, 0, len(benchmarksData)),
	}

	for _, bData := range benchmarksData {
		bMap, ok := bData.(map[string]interface{})
		if !ok {
			continue
		}

		result, err := parseBenchmarkFromJSON(bMap)
		if err != nil {
			return nil, fmt.Errorf("failed to parse benchmark: %w", err)
		}

		if result != nil {
			suite.Results = append(suite.Results, result)
			if suite.Language == "" {
				suite.Language = result.Language
			}
		}
	}

	if len(suite.Results) == 0 {
		return nil, fmt.Errorf("no valid benchmarks found in JSON")
	}

	return suite, nil
}

// parseBenchmarkFromJSON parses a single benchmark from JSON map
func parseBenchmarkFromJSON(data map[string]interface{}) (*parser.BenchmarkResult, error) {
	result := &parser.BenchmarkResult{}

	// Parse name
	if name, ok := data["name"].(string); ok {
		result.Name = name
	}

	// Parse language
	if lang, ok := data["language"].(string); ok {
		result.Language = lang
	}

	// Parse baseline_time_ns (required)
	baselineTimeNs, ok := data["baseline_time_ns"].(float64)
	if !ok {
		return nil, fmt.Errorf("missing or invalid baseline_time_ns")
	}
	result.Time = time.Duration(int64(baselineTimeNs))

	// Parse iterations if present
	if iter, ok := data["iterations"].(float64); ok {
		result.Iterations = int64(iter)
	}

	// Parse standard deviation if present
	if stdDev, ok := data["std_dev_ns"].(float64); ok {
		result.StdDev = time.Duration(int64(stdDev))
	}

	return result, nil
}

// loadBenchmarkFromCSV loads benchmark suite from CSV format
// Expected columns: name, language, time_ns, std_dev_ns, iterations
func loadBenchmarkFromCSV(r io.Reader) (*parser.BenchmarkSuite, error) {
	reader := csv.NewReader(r)

	// Read header
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Map column indices
	columnIndex := make(map[string]int)
	for i, col := range header {
		columnIndex[strings.TrimSpace(col)] = i
	}

	// Verify required columns
	requiredCols := []string{"name", "language", "time_ns"}
	for _, col := range requiredCols {
		if _, ok := columnIndex[col]; !ok {
			return nil, fmt.Errorf("missing required column: %s", col)
		}
	}

	suite := &parser.BenchmarkSuite{
		Results: make([]*parser.BenchmarkResult, 0),
	}

	// Read rows
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read CSV row: %w", err)
		}

		result := &parser.BenchmarkResult{}

		// Parse name
		if idx, ok := columnIndex["name"]; ok && idx < len(record) {
			result.Name = strings.TrimSpace(record[idx])
		}

		// Parse language
		if idx, ok := columnIndex["language"]; ok && idx < len(record) {
			result.Language = strings.TrimSpace(record[idx])
		}

		// Parse time_ns
		if idx, ok := columnIndex["time_ns"]; ok && idx < len(record) {
			timeNs, err := strconv.ParseInt(strings.TrimSpace(record[idx]), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid time_ns value: %w", err)
			}
			result.Time = time.Duration(timeNs)
		}

		// Parse std_dev_ns if present
		if idx, ok := columnIndex["std_dev_ns"]; ok && idx < len(record) {
			if val := strings.TrimSpace(record[idx]); val != "" {
				stdDevNs, err := strconv.ParseInt(val, 10, 64)
				if err == nil {
					result.StdDev = time.Duration(stdDevNs)
				}
			}
		}

		// Parse iterations if present
		if idx, ok := columnIndex["iterations"]; ok && idx < len(record) {
			if val := strings.TrimSpace(record[idx]); val != "" {
				iter, err := strconv.ParseInt(val, 10, 64)
				if err == nil {
					result.Iterations = iter
				}
			}
		}

		suite.Results = append(suite.Results, result)
		if suite.Language == "" {
			suite.Language = result.Language
		}
	}

	if len(suite.Results) == 0 {
		return nil, fmt.Errorf("no valid benchmarks found in CSV")
	}

	return suite, nil
}
