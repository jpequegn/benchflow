package storage

import (
	"os"
	"testing"
	"time"

	"github.com/jpequegn/benchflow/internal/comparator"
	"github.com/jpequegn/benchflow/internal/parser"
)

func TestSaveComparison(t *testing.T) {
	// Create temporary database
	tmpFile, err := os.CreateTemp("", "benchflow_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	storage, err := NewSQLiteStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	if err := storage.Init(); err != nil {
		t.Fatalf("Failed to init storage: %v", err)
	}

	if err := storage.InitComparisonHistory(); err != nil {
		t.Fatalf("Failed to init history: %v", err)
	}

	// Create test comparison result
	result := &comparator.ComparisonResult{
		Benchmarks: []*comparator.BenchmarkComparison{
			{
				Name:     "sort",
				Language: "go",
				Baseline: &parser.BenchmarkResult{
					Time:   1000 * time.Nanosecond,
					StdDev: 50 * time.Nanosecond,
				},
				Current: &parser.BenchmarkResult{
					Time:   950 * time.Nanosecond,
					StdDev: 45 * time.Nanosecond,
				},
				TimeDelta:    -5.0,
				IsRegression: false,
			},
		},
	}

	metadata := map[string]string{
		"commit_hash": "abc123",
		"branch_name": "main",
		"author":      "test@example.com",
	}

	// Save comparison
	err = storage.SaveComparison(1, 2, result, metadata)
	if err != nil {
		t.Fatalf("Failed to save comparison: %v", err)
	}
}

func TestGetComparisonHistory(t *testing.T) {
	// Create temporary database
	tmpFile, err := os.CreateTemp("", "benchflow_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	storage, err := NewSQLiteStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	if err := storage.Init(); err != nil {
		t.Fatalf("Failed to init storage: %v", err)
	}

	if err := storage.InitComparisonHistory(); err != nil {
		t.Fatalf("Failed to init history: %v", err)
	}

	// Create and save multiple comparisons
	for i := 0; i < 3; i++ {
		result := &comparator.ComparisonResult{
			Benchmarks: []*comparator.BenchmarkComparison{
				{
					Name:     "sort",
					Language: "go",
					Baseline: &parser.BenchmarkResult{
						Time: 1000 * time.Nanosecond,
					},
					Current: &parser.BenchmarkResult{
						Time: time.Duration((1000+50*i)) * time.Nanosecond,
					},
					TimeDelta: float64(5 * i),
				},
			},
		}

		err := storage.SaveComparison(1, 2, result, nil)
		if err != nil {
			t.Fatalf("Failed to save comparison %d: %v", i, err)
		}
	}

	// Retrieve history
	history, err := storage.GetComparisonHistory("sort", "go", 10)
	if err != nil {
		t.Fatalf("Failed to get history: %v", err)
	}

	if len(history) != 3 {
		t.Errorf("Expected 3 comparisons, got %d", len(history))
	}

	if history[0].BenchmarkName != "sort" {
		t.Errorf("Expected benchmark name 'sort', got %q", history[0].BenchmarkName)
	}

	if history[0].Language != "go" {
		t.Errorf("Expected language 'go', got %q", history[0].Language)
	}
}

func TestGetComparisonHistoryRange(t *testing.T) {
	// Create temporary database
	tmpFile, err := os.CreateTemp("", "benchflow_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	storage, err := NewSQLiteStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	if err := storage.Init(); err != nil {
		t.Fatalf("Failed to init storage: %v", err)
	}

	if err := storage.InitComparisonHistory(); err != nil {
		t.Fatalf("Failed to init history: %v", err)
	}

	now := time.Now()

	// Create and save comparisons with different timestamps
	for i := 0; i < 3; i++ {
		result := &comparator.ComparisonResult{
			Benchmarks: []*comparator.BenchmarkComparison{
				{
					Name:     "sort",
					Language: "go",
					Baseline: &parser.BenchmarkResult{
						Time: 1000 * time.Nanosecond,
					},
					Current: &parser.BenchmarkResult{
						Time: 1000 * time.Nanosecond,
					},
					TimeDelta: 0,
				},
			},
		}

		err := storage.SaveComparison(1, 2, result, nil)
		if err != nil {
			t.Fatalf("Failed to save comparison %d: %v", i, err)
		}

		time.Sleep(100 * time.Millisecond)
	}

	// Query range - should get all 3
	start := now.Add(-1 * time.Hour)
	end := now.Add(1 * time.Hour)

	history, err := storage.GetComparisonHistoryRange("sort", "go", start, end)
	if err != nil {
		t.Fatalf("Failed to get history range: %v", err)
	}

	if len(history) != 3 {
		t.Errorf("Expected 3 comparisons in range, got %d", len(history))
	}
}

func TestPruneComparisonHistory(t *testing.T) {
	// Create temporary database
	tmpFile, err := os.CreateTemp("", "benchflow_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	storage, err := NewSQLiteStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	if err := storage.Init(); err != nil {
		t.Fatalf("Failed to init storage: %v", err)
	}

	if err := storage.InitComparisonHistory(); err != nil {
		t.Fatalf("Failed to init history: %v", err)
	}

	// Save comparison
	result := &comparator.ComparisonResult{
		Benchmarks: []*comparator.BenchmarkComparison{
			{
				Name:     "sort",
				Language: "go",
				Baseline: &parser.BenchmarkResult{
					Time: 1000 * time.Nanosecond,
				},
				Current: &parser.BenchmarkResult{
					Time: 1000 * time.Nanosecond,
				},
				TimeDelta: 0,
			},
		},
	}

	err = storage.SaveComparison(1, 2, result, nil)
	if err != nil {
		t.Fatalf("Failed to save comparison: %v", err)
	}

	// Prune with high retention should keep records
	err = storage.PruneComparisonHistory(90)
	if err != nil {
		t.Fatalf("Failed to prune: %v", err)
	}

	// Verify record still exists
	history, err := storage.GetComparisonHistory("sort", "go", 10)
	if err != nil {
		t.Fatalf("Failed to get history: %v", err)
	}

	if len(history) != 1 {
		t.Errorf("Expected 1 comparison after prune with high retention, got %d", len(history))
	}
}

func TestComparisonHistoryWithMetadata(t *testing.T) {
	// Create temporary database
	tmpFile, err := os.CreateTemp("", "benchflow_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	storage, err := NewSQLiteStorage(tmpFile.Name())
	if err != nil {
		t.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	if err := storage.Init(); err != nil {
		t.Fatalf("Failed to init storage: %v", err)
	}

	if err := storage.InitComparisonHistory(); err != nil {
		t.Fatalf("Failed to init history: %v", err)
	}

	// Save comparison with metadata
	result := &comparator.ComparisonResult{
		Benchmarks: []*comparator.BenchmarkComparison{
			{
				Name:     "sort",
				Language: "go",
				Baseline: &parser.BenchmarkResult{
					Time: 1000 * time.Nanosecond,
				},
				Current: &parser.BenchmarkResult{
					Time: 1100 * time.Nanosecond,
				},
				TimeDelta:    10.0,
				IsRegression: true,
			},
		},
	}

	metadata := map[string]string{
		"commit_hash": "abc123def456",
		"branch_name": "feature/optimizations",
		"author":      "developer@example.com",
	}

	err = storage.SaveComparison(1, 2, result, metadata)
	if err != nil {
		t.Fatalf("Failed to save comparison: %v", err)
	}

	// Retrieve and verify metadata
	history, err := storage.GetComparisonHistory("sort", "go", 10)
	if err != nil {
		t.Fatalf("Failed to get history: %v", err)
	}

	if len(history) != 1 {
		t.Fatalf("Expected 1 comparison, got %d", len(history))
	}

	comp := history[0]
	if comp.CommitHash != "abc123def456" {
		t.Errorf("Expected commit hash 'abc123def456', got %q", comp.CommitHash)
	}

	if comp.BranchName != "feature/optimizations" {
		t.Errorf("Expected branch 'feature/optimizations', got %q", comp.BranchName)
	}

	if comp.Author != "developer@example.com" {
		t.Errorf("Expected author 'developer@example.com', got %q", comp.Author)
	}

	if !comp.IsRegression {
		t.Error("Expected IsRegression to be true")
	}

	if comp.TimeDeltaPercent != 10.0 {
		t.Errorf("Expected delta 10.0, got %f", comp.TimeDeltaPercent)
	}
}
