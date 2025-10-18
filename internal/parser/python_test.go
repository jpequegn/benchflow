package parser

import (
	"os"
	"testing"
	"time"
)

func TestPythonParser_Language(t *testing.T) {
	parser := NewPythonParser()
	if got := parser.Language(); got != "python" {
		t.Errorf("Language() = %v, want %v", got, "python")
	}
}

func TestPythonParser_Parse_BasicJSON(t *testing.T) {
	input := []byte(`{
  "machine_info": {
    "host": "MacBook-Pro.local",
    "python": "3.11.6",
    "implementation": "CPython"
  },
  "benchmarks": [
    {
      "name": "test_sort",
      "fullname": "tests/test_perf.py::test_sort",
      "params": null,
      "group": null,
      "extra_info": "",
      "options": {
        "disable_gc": false,
        "warmup": false
      },
      "stats": {
        "min": 0.0001234,
        "max": 0.0005678,
        "mean": 0.0002456,
        "stddev": 0.0000123,
        "rounds": 100,
        "median": 0.0002400,
        "iqr": 0.0000050,
        "q1": 0.0002200,
        "q3": 0.0002250,
        "iqr_outliers": 5,
        "stddevs": 2,
        "outliers": "0 0 0 0",
        "ops": 4071.66,
        "total": 0.024560
      }
    },
    {
      "name": "test_search",
      "fullname": "tests/test_perf.py::test_search",
      "params": null,
      "group": null,
      "extra_info": "",
      "options": {
        "disable_gc": false,
        "warmup": false
      },
      "stats": {
        "min": 0.0000567,
        "max": 0.0001234,
        "mean": 0.0000890,
        "stddev": 0.0000089,
        "rounds": 50,
        "median": 0.0000850,
        "iqr": 0.0000030,
        "q1": 0.0000800,
        "q3": 0.0000830,
        "iqr_outliers": 2,
        "stddevs": 1,
        "outliers": "0 0 0 0",
        "ops": 11235.96,
        "total": 0.004450
      }
    }
  ],
  "datetime": "2025-10-18T14:30:00",
  "version": "4.0.1"
}`)

	parser := NewPythonParser()
	suite, err := parser.Parse(input)

	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	if suite == nil {
		t.Fatal("Parse() returned nil suite")
	}

	// Check suite metadata
	if suite.Language != "python" {
		t.Errorf("Suite.Language = %v, want %v", suite.Language, "python")
	}

	// Check number of results
	if len(suite.Results) != 2 {
		t.Fatalf("len(Results) = %d, want %d", len(suite.Results), 2)
	}

	// Verify first benchmark
	first := suite.Results[0]
	if first.Name != "test_sort" {
		t.Errorf("Results[0].Name = %v, want %v", first.Name, "test_sort")
	}
	if first.Language != "python" {
		t.Errorf("Results[0].Language = %v, want %v", first.Language, "python")
	}
	// Mean is 0.0002456 seconds = 245600 nanoseconds
	expectedTime := time.Duration(245600) * time.Nanosecond
	if first.Time != expectedTime {
		t.Errorf("Results[0].Time = %v, want %v", first.Time, expectedTime)
	}
	// Iterations should be rounds
	if first.Iterations != 100 {
		t.Errorf("Results[0].Iterations = %d, want %d", first.Iterations, 100)
	}
	// StdDev is 0.0000123 seconds = 12300 nanoseconds
	expectedStdDev := time.Duration(12300) * time.Nanosecond
	if first.StdDev != expectedStdDev {
		t.Errorf("Results[0].StdDev = %v, want %v", first.StdDev, expectedStdDev)
	}

	// Verify throughput
	if first.Throughput == nil {
		t.Fatal("Results[0].Throughput is nil")
	}
	if first.Throughput.Value != 4071.66 {
		t.Errorf("Results[0].Throughput.Value = %v, want %v", first.Throughput.Value, 4071.66)
	}
	if first.Throughput.Unit != "ops/s" {
		t.Errorf("Results[0].Throughput.Unit = %v, want %v", first.Throughput.Unit, "ops/s")
	}

	// Verify metadata
	if _, ok := first.Metadata["min"]; !ok {
		t.Error("Results[0].Metadata missing 'min'")
	}
	if _, ok := first.Metadata["max"]; !ok {
		t.Error("Results[0].Metadata missing 'max'")
	}

	// Verify second benchmark
	second := suite.Results[1]
	if second.Name != "test_search" {
		t.Errorf("Results[1].Name = %v, want %v", second.Name, "test_search")
	}
	if second.Iterations != 50 {
		t.Errorf("Results[1].Iterations = %d, want %d", second.Iterations, 50)
	}
}

func TestPythonParser_Parse_FromFile(t *testing.T) {
	data, err := os.ReadFile("../../testdata/python/pytest_benchmark_basic.json")
	if err != nil {
		t.Skipf("Skipping test - testdata file not found: %v", err)
		return
	}

	parser := NewPythonParser()
	suite, err := parser.Parse(data)

	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	if len(suite.Results) != 2 {
		t.Errorf("len(Results) = %d, want %d", len(suite.Results), 2)
	}
}

func TestPythonParser_Parse_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantResults int
		wantErr     bool
		wantName    string
		wantOps     float64
	}{
		{
			name: "zero time",
			input: `{
  "benchmarks": [
    {
      "name": "test_zero",
      "fullname": "test.py::test_zero",
      "stats": {
        "min": 0.0,
        "max": 0.0,
        "mean": 0.0,
        "stddev": 0.0,
        "rounds": 1,
        "median": 0.0,
        "iqr": 0.0,
        "ops": 0.0,
        "total": 0.0
      }
    }
  ],
  "datetime": "2025-10-18T00:00:00",
  "version": "4.0.1"
}`,
			wantResults: 1,
			wantErr:     false,
			wantName:    "test_zero",
			wantOps:     0.0,
		},
		{
			name: "large numbers",
			input: `{
  "benchmarks": [
    {
      "name": "test_slow",
      "fullname": "test.py::test_slow",
      "stats": {
        "min": 1.234567,
        "max": 5.678901,
        "mean": 3.456789,
        "stddev": 1.234567,
        "rounds": 10,
        "median": 3.4,
        "iqr": 0.5,
        "ops": 0.289,
        "total": 34.56789
      }
    }
  ],
  "datetime": "2025-10-18T00:00:00",
  "version": "4.0.1"
}`,
			wantResults: 1,
			wantErr:     false,
			wantName:    "test_slow",
			wantOps:     0.289,
		},
		{
			name: "high ops value",
			input: `{
  "benchmarks": [
    {
      "name": "test_fast",
      "fullname": "test.py::test_fast",
      "stats": {
        "min": 0.000001,
        "max": 0.000005,
        "mean": 0.000002,
        "stddev": 0.0000001,
        "rounds": 1000,
        "median": 0.000002,
        "iqr": 0.0000005,
        "ops": 500000.0,
        "total": 0.002
      }
    }
  ],
  "datetime": "2025-10-18T00:00:00",
  "version": "4.0.1"
}`,
			wantResults: 1,
			wantErr:     false,
			wantName:    "test_fast",
			wantOps:     500000.0,
		},
		{
			name:        "invalid JSON",
			input:       `{invalid json`,
			wantResults: 0,
			wantErr:     true,
		},
		{
			name: "empty benchmarks array",
			input: `{
  "benchmarks": [],
  "datetime": "2025-10-18T00:00:00",
  "version": "4.0.1"
}`,
			wantResults: 0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewPythonParser()
			suite, err := parser.Parse([]byte(tt.input))

			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			if len(suite.Results) != tt.wantResults {
				t.Errorf("len(Results) = %d, want %d", len(suite.Results), tt.wantResults)
			}

			if tt.wantResults > 0 {
				if suite.Results[0].Name != tt.wantName {
					t.Errorf("Results[0].Name = %v, want %v", suite.Results[0].Name, tt.wantName)
				}
				if suite.Results[0].Throughput != nil {
					if suite.Results[0].Throughput.Value != tt.wantOps {
						t.Errorf("Results[0].Throughput.Value = %v, want %v", suite.Results[0].Throughput.Value, tt.wantOps)
					}
				}
			}
		})
	}
}

func TestPythonParser_Parse_SkipsMissingStats(t *testing.T) {
	input := []byte(`{
  "benchmarks": [
    {
      "name": "test_with_stats",
      "fullname": "test.py::test_with_stats",
      "stats": {
        "min": 0.001,
        "max": 0.005,
        "mean": 0.003,
        "stddev": 0.0001,
        "rounds": 100,
        "median": 0.003,
        "iqr": 0.0001,
        "ops": 333.33,
        "total": 0.3
      }
    },
    {
      "name": "test_missing_stats",
      "fullname": "test.py::test_missing_stats"
    },
    {
      "name": "test_partial_stats",
      "fullname": "test.py::test_partial_stats",
      "stats": {
        "min": 0.001,
        "max": 0.005,
        "mean": 0.002,
        "stddev": 0.0001,
        "rounds": 0
      }
    }
  ],
  "datetime": "2025-10-18T00:00:00",
  "version": "4.0.1"
}`)

	parser := NewPythonParser()
	suite, err := parser.Parse(input)

	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	// Should only parse the benchmark with complete stats (rounds > 0)
	if len(suite.Results) != 1 {
		t.Errorf("len(Results) = %d, want %d (should skip incomplete stats)", len(suite.Results), 1)
	}

	if suite.Results[0].Name != "test_with_stats" {
		t.Errorf("Results[0].Name = %v, want %v", suite.Results[0].Name, "test_with_stats")
	}
}

func TestPythonParser_Parse_Metadata(t *testing.T) {
	input := []byte(`{
  "benchmarks": [
    {
      "name": "test_metadata",
      "fullname": "test.py::test_metadata",
      "stats": {
        "min": 0.001,
        "max": 0.005,
        "mean": 0.003,
        "stddev": 0.0001,
        "median": 0.003,
        "q1": 0.0025,
        "q3": 0.0035,
        "rounds": 100,
        "iqr": 0.001,
        "ops": 333.33,
        "total": 0.3
      }
    }
  ],
  "datetime": "2025-10-18T14:30:00",
  "version": "4.0.1"
}`)

	parser := NewPythonParser()
	suite, err := parser.Parse(input)

	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	// Check suite metadata
	if suite.Metadata["datetime"] != "2025-10-18T14:30:00" {
		t.Errorf("Suite.Metadata['datetime'] = %v, want %v", suite.Metadata["datetime"], "2025-10-18T14:30:00")
	}
	if suite.Metadata["version"] != "4.0.1" {
		t.Errorf("Suite.Metadata['version'] = %v, want %v", suite.Metadata["version"], "4.0.1")
	}

	// Check result metadata
	result := suite.Results[0]
	if _, ok := result.Metadata["min"]; !ok {
		t.Error("Result.Metadata missing 'min'")
	}
	if _, ok := result.Metadata["max"]; !ok {
		t.Error("Result.Metadata missing 'max'")
	}
	if _, ok := result.Metadata["median"]; !ok {
		t.Error("Result.Metadata missing 'median'")
	}
	if _, ok := result.Metadata["q1"]; !ok {
		t.Error("Result.Metadata missing 'q1'")
	}
	if _, ok := result.Metadata["q3"]; !ok {
		t.Error("Result.Metadata missing 'q3'")
	}
}

func TestPythonParser_Parse_EdgeCasesFromFile(t *testing.T) {
	data, err := os.ReadFile("../../testdata/python/pytest_benchmark_edge_cases.json")
	if err != nil {
		t.Skipf("Skipping test - testdata file not found: %v", err)
		return
	}

	parser := NewPythonParser()
	suite, err := parser.Parse(data)

	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	if len(suite.Results) != 2 {
		t.Errorf("len(Results) = %d, want %d", len(suite.Results), 2)
	}

	// Check results (order: large_numbers, zero_time)
	largeResult := suite.Results[0]
	if largeResult.Name != "test_large_numbers" {
		t.Errorf("Results[0].Name = %v, want test_large_numbers", largeResult.Name)
	}
	if largeResult.Time == 0 {
		t.Error("Large numbers result.Time should not be 0")
	}

	// Check zero time case
	zeroResult := suite.Results[1]
	if zeroResult.Name != "test_zero_time" {
		t.Errorf("Results[1].Name = %v, want test_zero_time", zeroResult.Name)
	}
	if zeroResult.Time != 0 {
		t.Errorf("Zero time result.Time = %v, want 0", zeroResult.Time)
	}
}

func TestPythonParser_Parse_MalformedJSON(t *testing.T) {
	data, err := os.ReadFile("../../testdata/python/pytest_benchmark_malformed.json")
	if err != nil {
		t.Skipf("Skipping test - testdata file not found: %v", err)
		return
	}

	parser := NewPythonParser()
	suite, err := parser.Parse(data)

	if err != nil {
		t.Fatalf("Parse() error = %v, want nil", err)
	}

	// Should parse the benchmark with complete stats
	if len(suite.Results) != 1 {
		t.Errorf("len(Results) = %d, want %d (should skip incomplete)", len(suite.Results), 1)
	}

	if suite.Results[0].Name != "test_partial_stats" {
		t.Errorf("Results[0].Name = %v, want %v", suite.Results[0].Name, "test_partial_stats")
	}
}
