package cmd

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadBenchmarkSuite_JSON(t *testing.T) {
	// Create temporary JSON file
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "benchmarks.json")

	jsonContent := `{
  "summary": {
    "total_comparisons": 2,
    "regressions": 0,
    "improvements": 1
  },
  "benchmarks": [
    {
      "name": "sort",
      "language": "go",
      "baseline_time_ns": 1000,
      "std_dev_ns": 50,
      "iterations": 100
    },
    {
      "name": "search",
      "language": "go",
      "baseline_time_ns": 500,
      "std_dev_ns": 25,
      "iterations": 100
    }
  ]
}`

	if err := os.WriteFile(jsonFile, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	suite, err := LoadBenchmarkSuite(jsonFile)
	if err != nil {
		t.Fatalf("LoadBenchmarkSuite failed: %v", err)
	}

	if suite == nil {
		t.Fatal("LoadBenchmarkSuite returned nil")
	}

	if len(suite.Results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(suite.Results))
	}

	if suite.Results[0].Name != "sort" {
		t.Errorf("Expected first benchmark name 'sort', got %q", suite.Results[0].Name)
	}

	if suite.Results[0].Time != 1000*time.Nanosecond {
		t.Errorf("Expected time 1000ns, got %v", suite.Results[0].Time)
	}

	if suite.Results[0].Language != "go" {
		t.Errorf("Expected language 'go', got %q", suite.Results[0].Language)
	}

	if suite.Results[0].StdDev != 50*time.Nanosecond {
		t.Errorf("Expected stddev 50ns, got %v", suite.Results[0].StdDev)
	}

	if suite.Results[0].Iterations != 100 {
		t.Errorf("Expected iterations 100, got %d", suite.Results[0].Iterations)
	}
}

func TestLoadBenchmarkSuite_CSV(t *testing.T) {
	// Create temporary CSV file
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "benchmarks.csv")

	csvContent := `name,language,time_ns,std_dev_ns,iterations
sort,go,1000,50,100
search,go,500,25,100`

	if err := os.WriteFile(csvFile, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	suite, err := LoadBenchmarkSuite(csvFile)
	if err != nil {
		t.Fatalf("LoadBenchmarkSuite failed: %v", err)
	}

	if suite == nil {
		t.Fatal("LoadBenchmarkSuite returned nil")
	}

	if len(suite.Results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(suite.Results))
	}

	if suite.Results[0].Name != "sort" {
		t.Errorf("Expected first benchmark name 'sort', got %q", suite.Results[0].Name)
	}

	if suite.Results[0].Time != 1000*time.Nanosecond {
		t.Errorf("Expected time 1000ns, got %v", suite.Results[0].Time)
	}
}

func TestLoadBenchmarkSuite_UnsupportedFormat(t *testing.T) {
	tmpDir := t.TempDir()
	txtFile := filepath.Join(tmpDir, "benchmarks.txt")

	if err := os.WriteFile(txtFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := LoadBenchmarkSuite(txtFile)
	if err == nil {
		t.Fatal("Expected error for unsupported format")
	}
}

func TestLoadBenchmarkSuite_FileNotFound(t *testing.T) {
	_, err := LoadBenchmarkSuite("/nonexistent/path/benchmarks.json")
	if err == nil {
		t.Fatal("Expected error for missing file")
	}
}

func TestLoadBenchmarkSuite_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "invalid.json")

	if err := os.WriteFile(jsonFile, []byte("{invalid json}"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := LoadBenchmarkSuite(jsonFile)
	if err == nil {
		t.Fatal("Expected error for invalid JSON")
	}
}

func TestLoadBenchmarkSuite_JSONMissingBenchmarks(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "empty.json")

	if err := os.WriteFile(jsonFile, []byte("{}"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := LoadBenchmarkSuite(jsonFile)
	if err == nil {
		t.Fatal("Expected error for missing benchmarks field")
	}
}

func TestLoadBenchmarkSuite_JSONNoBenchmarks(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "empty.json")

	if err := os.WriteFile(jsonFile, []byte("{\"benchmarks\": []}"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := LoadBenchmarkSuite(jsonFile)
	if err == nil {
		t.Fatal("Expected error for empty benchmarks")
	}
}

func TestLoadBenchmarkSuite_CSVMissingColumns(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "incomplete.csv")

	// Missing required 'time_ns' column
	csvContent := `name,language
sort,go
search,go`

	if err := os.WriteFile(csvFile, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := LoadBenchmarkSuite(csvFile)
	if err == nil {
		t.Fatal("Expected error for missing required column")
	}
}

func TestLoadBenchmarkSuite_CSVInvalidNumber(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "invalid.csv")

	csvContent := `name,language,time_ns,std_dev_ns,iterations
sort,go,invalid,50,100`

	if err := os.WriteFile(csvFile, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := LoadBenchmarkSuite(csvFile)
	if err == nil {
		t.Fatal("Expected error for invalid number in CSV")
	}
}

func TestLoadBenchmarkSuite_JSONMissingTime(t *testing.T) {
	tmpDir := t.TempDir()
	jsonFile := filepath.Join(tmpDir, "invalid.json")

	jsonContent := `{
  "benchmarks": [
    {
      "name": "sort",
      "language": "go"
    }
  ]
}`

	if err := os.WriteFile(jsonFile, []byte(jsonContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := LoadBenchmarkSuite(jsonFile)
	if err == nil {
		t.Fatal("Expected error for missing baseline_time_ns")
	}
}

func TestLoadBenchmarkSuite_CSVOptionalFields(t *testing.T) {
	tmpDir := t.TempDir()
	csvFile := filepath.Join(tmpDir, "minimal.csv")

	// Only required columns
	csvContent := `name,language,time_ns
sort,go,1000
search,rust,500`

	if err := os.WriteFile(csvFile, []byte(csvContent), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	suite, err := LoadBenchmarkSuite(csvFile)
	if err != nil {
		t.Fatalf("LoadBenchmarkSuite failed: %v", err)
	}

	if len(suite.Results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(suite.Results))
	}

	// Check optional fields are zero/empty
	if suite.Results[0].StdDev != 0 {
		t.Errorf("Expected zero stddev, got %v", suite.Results[0].StdDev)
	}

	if suite.Results[0].Iterations != 0 {
		t.Errorf("Expected zero iterations, got %d", suite.Results[0].Iterations)
	}
}
