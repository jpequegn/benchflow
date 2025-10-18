package aggregator

import (
	"encoding/csv"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/jpequegn/benchflow/internal/parser"
)

func TestAggregator_Aggregate_Success(t *testing.T) {
	agg := NewAggregator()

	suite := &parser.BenchmarkSuite{
		Language:  "rust",
		Timestamp: time.Now(),
		Results: []*parser.BenchmarkResult{
			{
				Name:       "bench_sort",
				Language:   "rust",
				Time:       100 * time.Nanosecond,
				StdDev:     10 * time.Nanosecond,
				Iterations: 1000,
			},
			{
				Name:       "bench_search",
				Language:   "rust",
				Time:       200 * time.Nanosecond,
				StdDev:     20 * time.Nanosecond,
				Iterations: 500,
			},
		},
	}

	result, err := agg.Aggregate(suite)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Results) != 2 {
		t.Errorf("expected 2 results, got %d", len(result.Results))
	}

	// Check stats
	if result.Stats.TotalBenchmarks != 2 {
		t.Errorf("expected 2 benchmarks, got %d", result.Stats.TotalBenchmarks)
	}

	if result.Stats.FastestBench != "bench_sort" {
		t.Errorf("expected fastest bench to be bench_sort, got %s", result.Stats.FastestBench)
	}

	if result.Stats.SlowestBench != "bench_search" {
		t.Errorf("expected slowest bench to be bench_search, got %s", result.Stats.SlowestBench)
	}
}

func TestAggregator_Aggregate_NilSuite(t *testing.T) {
	agg := NewAggregator()

	_, err := agg.Aggregate(nil)
	if err == nil {
		t.Fatal("expected error for nil suite")
	}

	if !strings.Contains(err.Error(), "cannot be nil") {
		t.Errorf("expected 'cannot be nil' error, got: %v", err)
	}
}

func TestAggregator_Aggregate_EmptyResults(t *testing.T) {
	agg := NewAggregator()

	suite := &parser.BenchmarkSuite{
		Language:  "rust",
		Timestamp: time.Now(),
		Results:   []*parser.BenchmarkResult{},
	}

	_, err := agg.Aggregate(suite)
	if err == nil {
		t.Fatal("expected error for empty results")
	}

	if !strings.Contains(err.Error(), "no results") {
		t.Errorf("expected 'no results' error, got: %v", err)
	}
}

func TestAggregator_Compare_Success(t *testing.T) {
	agg := NewAggregator()

	baseline := &AggregatedSuite{
		Results: []*AggregatedResult{
			{Name: "bench_sort", Mean: 100 * time.Nanosecond},
			{Name: "bench_search", Mean: 200 * time.Nanosecond},
		},
	}

	current := &AggregatedSuite{
		Results: []*AggregatedResult{
			{Name: "bench_sort", Mean: 120 * time.Nanosecond},   // 20% slower (regression)
			{Name: "bench_search", Mean: 180 * time.Nanosecond}, // 10% faster (improvement)
		},
	}

	comparison, err := agg.Compare(baseline, current, 5.0) // 5% threshold
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(comparison.Comparisons) != 2 {
		t.Errorf("expected 2 comparisons, got %d", len(comparison.Comparisons))
	}

	// Check regression detection
	if comparison.RegressionCount != 1 {
		t.Errorf("expected 1 regression, got %d", comparison.RegressionCount)
	}

	if comparison.ImprovementCount != 1 {
		t.Errorf("expected 1 improvement, got %d", comparison.ImprovementCount)
	}

	// Verify specific comparison
	sortComp := comparison.Comparisons[0]
	if !sortComp.Regression {
		t.Error("expected bench_sort to be flagged as regression")
	}

	if sortComp.DeltaPercent < 19.0 || sortComp.DeltaPercent > 21.0 {
		t.Errorf("expected delta percent ~20%%, got %.2f%%", sortComp.DeltaPercent)
	}

	searchComp := comparison.Comparisons[1]
	if !searchComp.Improvement {
		t.Error("expected bench_search to be flagged as improvement")
	}
}

func TestAggregator_Compare_WithinThreshold(t *testing.T) {
	agg := NewAggregator()

	baseline := &AggregatedSuite{
		Results: []*AggregatedResult{
			{Name: "bench_test", Mean: 100 * time.Nanosecond},
		},
	}

	current := &AggregatedSuite{
		Results: []*AggregatedResult{
			{Name: "bench_test", Mean: 102 * time.Nanosecond}, // 2% change
		},
	}

	comparison, err := agg.Compare(baseline, current, 5.0) // 5% threshold
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if comparison.UnchangedCount != 1 {
		t.Errorf("expected 1 unchanged, got %d", comparison.UnchangedCount)
	}

	comp := comparison.Comparisons[0]
	if comp.Regression || comp.Improvement {
		t.Error("expected no regression or improvement within threshold")
	}
}

func TestAggregator_Compare_MissingBaseline(t *testing.T) {
	agg := NewAggregator()

	baseline := &AggregatedSuite{
		Results: []*AggregatedResult{
			{Name: "bench_old", Mean: 100 * time.Nanosecond},
		},
	}

	current := &AggregatedSuite{
		Results: []*AggregatedResult{
			{Name: "bench_new", Mean: 100 * time.Nanosecond},
		},
	}

	comparison, err := agg.Compare(baseline, current, 5.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should skip bench_new since it doesn't exist in baseline
	if len(comparison.Comparisons) != 0 {
		t.Errorf("expected 0 comparisons, got %d", len(comparison.Comparisons))
	}
}

func TestAggregator_Compare_NilSuites(t *testing.T) {
	agg := NewAggregator()

	_, err := agg.Compare(nil, &AggregatedSuite{}, 5.0)
	if err == nil {
		t.Fatal("expected error for nil baseline")
	}

	_, err = agg.Compare(&AggregatedSuite{}, nil, 5.0)
	if err == nil {
		t.Fatal("expected error for nil current")
	}
}

func TestAggregator_ExportJSON(t *testing.T) {
	agg := NewAggregator()

	suite := &AggregatedSuite{
		Results: []*AggregatedResult{
			{
				Name:       "bench_test",
				Language:   "rust",
				Mean:       100 * time.Nanosecond,
				Median:     100 * time.Nanosecond,
				Min:        90 * time.Nanosecond,
				Max:        110 * time.Nanosecond,
				StdDev:     10 * time.Nanosecond,
				Iterations: 1000,
			},
		},
		Timestamp: time.Now(),
		Stats: &SuiteStats{
			TotalBenchmarks: 1,
		},
	}

	data, err := agg.Export(suite, FormatJSON)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it's valid JSON
	var decoded AggregatedSuite
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if len(decoded.Results) != 1 {
		t.Errorf("expected 1 result, got %d", len(decoded.Results))
	}

	if decoded.Results[0].Name != "bench_test" {
		t.Errorf("expected name bench_test, got %s", decoded.Results[0].Name)
	}
}

func TestAggregator_ExportCSV(t *testing.T) {
	agg := NewAggregator()

	suite := &AggregatedSuite{
		Results: []*AggregatedResult{
			{
				Name:       "bench_test",
				Language:   "rust",
				Mean:       100 * time.Nanosecond,
				Median:     100 * time.Nanosecond,
				Min:        90 * time.Nanosecond,
				Max:        110 * time.Nanosecond,
				StdDev:     10 * time.Nanosecond,
				Iterations: 1000,
			},
		},
	}

	data, err := agg.Export(suite, FormatCSV)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify it's valid CSV
	reader := csv.NewReader(strings.NewReader(string(data)))
	records, err := reader.ReadAll()
	if err != nil {
		t.Fatalf("failed to read CSV: %v", err)
	}

	if len(records) != 2 { // header + 1 data row
		t.Errorf("expected 2 rows, got %d", len(records))
	}

	// Check header
	if records[0][0] != "Name" {
		t.Errorf("expected first column to be Name, got %s", records[0][0])
	}

	// Check data
	if records[1][0] != "bench_test" {
		t.Errorf("expected bench_test, got %s", records[1][0])
	}
}

func TestAggregator_Export_UnsupportedFormat(t *testing.T) {
	agg := NewAggregator()

	suite := &AggregatedSuite{
		Results: []*AggregatedResult{},
	}

	_, err := agg.Export(suite, ExportFormat("xml"))
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}

	if !strings.Contains(err.Error(), "unsupported format") {
		t.Errorf("expected 'unsupported format' error, got: %v", err)
	}
}

func TestAggregator_Export_NilSuite(t *testing.T) {
	agg := NewAggregator()

	_, err := agg.Export(nil, FormatJSON)
	if err == nil {
		t.Fatal("expected error for nil suite")
	}
}

func TestCalculateStatistics(t *testing.T) {
	tests := []struct {
		name           string
		durations      []time.Duration
		expectedMean   time.Duration
		expectedMedian time.Duration
	}{
		{
			name:           "empty slice",
			durations:      []time.Duration{},
			expectedMean:   0,
			expectedMedian: 0,
		},
		{
			name: "single value",
			durations: []time.Duration{
				100 * time.Nanosecond,
			},
			expectedMean:   100 * time.Nanosecond,
			expectedMedian: 100 * time.Nanosecond,
		},
		{
			name: "odd number of values",
			durations: []time.Duration{
				100 * time.Nanosecond,
				200 * time.Nanosecond,
				300 * time.Nanosecond,
			},
			expectedMean:   200 * time.Nanosecond,
			expectedMedian: 200 * time.Nanosecond,
		},
		{
			name: "even number of values",
			durations: []time.Duration{
				100 * time.Nanosecond,
				200 * time.Nanosecond,
				300 * time.Nanosecond,
				400 * time.Nanosecond,
			},
			expectedMean:   250 * time.Nanosecond,
			expectedMedian: 250 * time.Nanosecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mean, median, _ := CalculateStatistics(tt.durations)

			if mean != tt.expectedMean {
				t.Errorf("expected mean %v, got %v", tt.expectedMean, mean)
			}

			if median != tt.expectedMedian {
				t.Errorf("expected median %v, got %v", tt.expectedMedian, median)
			}
		})
	}
}

func TestCalculateStatistics_StdDev(t *testing.T) {
	durations := []time.Duration{
		100 * time.Nanosecond,
		100 * time.Nanosecond,
		100 * time.Nanosecond,
	}

	_, _, stdDev := CalculateStatistics(durations)

	// All values are the same, so stddev should be 0
	if stdDev != 0 {
		t.Errorf("expected stddev 0 for identical values, got %v", stdDev)
	}
}
