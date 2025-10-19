package comparator

import (
	"math"
	"testing"
	"time"

	"github.com/jpequegn/benchflow/internal/parser"
)

func TestNewBasicComparator(t *testing.T) {
	comp := NewBasicComparator()
	if comp == nil {
		t.Error("NewBasicComparator() returned nil")
	}
	if comp.ConfidenceLevel != 0.95 {
		t.Errorf("ConfidenceLevel = %v, want 0.95", comp.ConfidenceLevel)
	}
	if comp.RegressionThreshold != 1.05 {
		t.Errorf("RegressionThreshold = %v, want 1.05", comp.RegressionThreshold)
	}
}

func TestCompare_BasicComparison(t *testing.T) {
	comp := NewBasicComparator()

	baseline := &parser.BenchmarkSuite{
		Language: "go",
		Results: []*parser.BenchmarkResult{
			{
				Name:       "sort",
				Language:   "go",
				Time:       1000 * time.Nanosecond,
				Iterations: 100,
				StdDev:     50 * time.Nanosecond,
			},
		},
	}

	current := &parser.BenchmarkSuite{
		Language: "go",
		Results: []*parser.BenchmarkResult{
			{
				Name:       "sort",
				Language:   "go",
				Time:       950 * time.Nanosecond,
				Iterations: 100,
				StdDev:     45 * time.Nanosecond,
			},
		},
	}

	result := comp.Compare(baseline, current)

	if result == nil {
		t.Fatal("Compare() returned nil")
	}

	if len(result.Benchmarks) != 1 {
		t.Errorf("len(Benchmarks) = %d, want 1", len(result.Benchmarks))
	}

	comparison := result.Benchmarks[0]
	if comparison.Name != "sort" {
		t.Errorf("Name = %q, want %q", comparison.Name, "sort")
	}

	// 950 vs 1000: -5% improvement
	expectedDelta := -5.0
	if math.Abs(comparison.TimeDelta-expectedDelta) > 0.1 {
		t.Errorf("TimeDelta = %v, want %v", comparison.TimeDelta, expectedDelta)
	}

	if comparison.IsRegression {
		t.Error("IsRegression = true, want false (this is an improvement)")
	}
}

func TestCompare_Regression(t *testing.T) {
	comp := NewBasicComparator()
	comp.RegressionThreshold = 1.05 // 5% regression threshold

	baseline := &parser.BenchmarkSuite{
		Language: "rust",
		Results: []*parser.BenchmarkResult{
			{
				Name:       "search",
				Language:   "rust",
				Time:       1000 * time.Nanosecond,
				Iterations: 100,
				StdDev:     50 * time.Nanosecond,
			},
		},
	}

	current := &parser.BenchmarkSuite{
		Language: "rust",
		Results: []*parser.BenchmarkResult{
			{
				Name:       "search",
				Language:   "rust",
				Time:       1100 * time.Nanosecond, // 10% slower
				Iterations: 100,
				StdDev:     60 * time.Nanosecond,
			},
		},
	}

	result := comp.Compare(baseline, current)

	if len(result.Benchmarks) != 1 {
		t.Fatalf("len(Benchmarks) = %d, want 1", len(result.Benchmarks))
	}

	comparison := result.Benchmarks[0]
	if !comparison.IsRegression {
		t.Error("IsRegression = false, want true (10% slower exceeds 5% threshold)")
	}

	if len(result.Regressions) != 1 {
		t.Errorf("len(Regressions) = %d, want 1", len(result.Regressions))
	}
}

func TestCompare_NoRegression_WithinThreshold(t *testing.T) {
	comp := NewBasicComparator()
	comp.RegressionThreshold = 1.05 // 5% regression threshold

	baseline := &parser.BenchmarkSuite{
		Language: "python",
		Results: []*parser.BenchmarkResult{
			{
				Name:       "filter",
				Language:   "python",
				Time:       10000 * time.Nanosecond,
				Iterations: 50,
				StdDev:     100 * time.Nanosecond,
			},
		},
	}

	current := &parser.BenchmarkSuite{
		Language: "python",
		Results: []*parser.BenchmarkResult{
			{
				Name:       "filter",
				Language:   "python",
				Time:       10200 * time.Nanosecond, // 2% slower
				Iterations: 50,
				StdDev:     110 * time.Nanosecond,
			},
		},
	}

	result := comp.Compare(baseline, current)

	if len(result.Benchmarks) != 1 {
		t.Fatalf("len(Benchmarks) = %d, want 1", len(result.Benchmarks))
	}

	comparison := result.Benchmarks[0]
	if comparison.IsRegression {
		t.Error("IsRegression = true, want false (2% slower is within 5% threshold)")
	}

	if len(result.Regressions) != 0 {
		t.Errorf("len(Regressions) = %d, want 0", len(result.Regressions))
	}
}

func TestCompare_MultipleResults(t *testing.T) {
	comp := NewBasicComparator()

	baseline := &parser.BenchmarkSuite{
		Language: "go",
		Results: []*parser.BenchmarkResult{
			{
				Name:       "sort",
				Language:   "go",
				Time:       1000 * time.Nanosecond,
				Iterations: 100,
				StdDev:     50 * time.Nanosecond,
			},
			{
				Name:       "search",
				Language:   "go",
				Time:       500 * time.Nanosecond,
				Iterations: 100,
				StdDev:     25 * time.Nanosecond,
			},
			{
				Name:       "insert",
				Language:   "go",
				Time:       800 * time.Nanosecond,
				Iterations: 100,
				StdDev:     40 * time.Nanosecond,
			},
		},
	}

	current := &parser.BenchmarkSuite{
		Language: "go",
		Results: []*parser.BenchmarkResult{
			{
				Name:       "sort",
				Language:   "go",
				Time:       950 * time.Nanosecond, // Improvement
				Iterations: 100,
				StdDev:     45 * time.Nanosecond,
			},
			{
				Name:       "search",
				Language:   "go",
				Time:       600 * time.Nanosecond, // Regression
				Iterations: 100,
				StdDev:     30 * time.Nanosecond,
			},
			{
				Name:       "insert",
				Language:   "go",
				Time:       800 * time.Nanosecond, // No change
				Iterations: 100,
				StdDev:     40 * time.Nanosecond,
			},
		},
	}

	result := comp.Compare(baseline, current)

	if len(result.Benchmarks) != 3 {
		t.Errorf("len(Benchmarks) = %d, want 3", len(result.Benchmarks))
	}

	if result.Summary.TotalComparisons != 3 {
		t.Errorf("TotalComparisons = %d, want 3", result.Summary.TotalComparisons)
	}

	if result.Summary.Improvements != 1 {
		t.Errorf("Improvements = %d, want 1", result.Summary.Improvements)
	}

	if result.Summary.Regressions != 1 {
		t.Errorf("Regressions = %d, want 1", result.Summary.Regressions)
	}
}

func TestCompare_NilInputs(t *testing.T) {
	comp := NewBasicComparator()

	// Test nil baseline
	result := comp.Compare(nil, &parser.BenchmarkSuite{})
	if result == nil {
		t.Error("Compare(nil, suite) returned nil")
	}
	if len(result.Benchmarks) != 0 {
		t.Errorf("Compare(nil, suite) len(Benchmarks) = %d, want 0", len(result.Benchmarks))
	}

	// Test nil current
	result = comp.Compare(&parser.BenchmarkSuite{}, nil)
	if result == nil {
		t.Error("Compare(suite, nil) returned nil")
	}
	if len(result.Benchmarks) != 0 {
		t.Errorf("Compare(suite, nil) len(Benchmarks) = %d, want 0", len(result.Benchmarks))
	}

	// Test empty results
	result = comp.Compare(
		&parser.BenchmarkSuite{Results: make([]*parser.BenchmarkResult, 0)},
		&parser.BenchmarkSuite{Results: make([]*parser.BenchmarkResult, 0)},
	)
	if len(result.Benchmarks) != 0 {
		t.Errorf("Compare(empty, empty) len(Benchmarks) = %d, want 0", len(result.Benchmarks))
	}
}

func TestCompare_LanguageMismatch(t *testing.T) {
	comp := NewBasicComparator()

	baseline := &parser.BenchmarkSuite{
		Language: "rust",
		Results: []*parser.BenchmarkResult{
			{
				Name:     "sort",
				Language: "rust",
				Time:     1000 * time.Nanosecond,
			},
		},
	}

	current := &parser.BenchmarkSuite{
		Language: "python",
		Results: []*parser.BenchmarkResult{
			{
				Name:     "sort",
				Language: "python",
				Time:     5000 * time.Nanosecond,
			},
		},
	}

	result := comp.Compare(baseline, current)

	if len(result.Benchmarks) != 0 {
		t.Errorf("len(Benchmarks) = %d, want 0 (different languages)", len(result.Benchmarks))
	}
}

func TestCompare_MissingBaseline(t *testing.T) {
	comp := NewBasicComparator()

	baseline := &parser.BenchmarkSuite{
		Language: "go",
		Results: []*parser.BenchmarkResult{
			{
				Name:     "sort",
				Language: "go",
				Time:     1000 * time.Nanosecond,
			},
		},
	}

	current := &parser.BenchmarkSuite{
		Language: "go",
		Results: []*parser.BenchmarkResult{
			{
				Name:     "sort",
				Language: "go",
				Time:     950 * time.Nanosecond,
			},
			{
				Name:     "search", // No baseline for this
				Language: "go",
				Time:     500 * time.Nanosecond,
			},
		},
	}

	result := comp.Compare(baseline, current)

	if len(result.Benchmarks) != 1 {
		t.Errorf("len(Benchmarks) = %d, want 1 (only 'sort' has baseline)", len(result.Benchmarks))
	}

	if result.Benchmarks[0].Name != "sort" {
		t.Errorf("first benchmark Name = %q, want %q", result.Benchmarks[0].Name, "sort")
	}
}

func TestGetSignificance(t *testing.T) {
	comp := NewBasicComparator()

	baseline := &parser.BenchmarkResult{
		Time:   1000 * time.Nanosecond,
		StdDev: 10 * time.Nanosecond,
	}

	// Test case 1: Very different (should be significant)
	current := &parser.BenchmarkResult{
		Time:   2000 * time.Nanosecond,
		StdDev: 10 * time.Nanosecond,
	}

	significant, pValue := comp.GetSignificance(baseline, current, 0.95)
	if !significant {
		t.Errorf("GetSignificance(2x slower) significant = false, want true")
	}
	if pValue >= 0.05 {
		t.Errorf("GetSignificance(2x slower) pValue = %v, want < 0.05", pValue)
	}

	// Test case 2: Very similar (should not be significant)
	current2 := &parser.BenchmarkResult{
		Time:   1010 * time.Nanosecond,
		StdDev: 50 * time.Nanosecond,
	}

	significant2, pValue2 := comp.GetSignificance(baseline, current2, 0.95)
	if significant2 {
		t.Error("GetSignificance(1% slower) significant = true, want false")
	}
	if pValue2 < 0.05 {
		t.Errorf("GetSignificance(1%% slower) pValue = %v, want >= 0.05", pValue2)
	}
}

func TestCalculateConfidenceInterval(t *testing.T) {
	comp := NewBasicComparator()

	results := []*parser.BenchmarkResult{
		{Time: 1000 * time.Nanosecond},
		{Time: 1100 * time.Nanosecond},
		{Time: 900 * time.Nanosecond},
		{Time: 1050 * time.Nanosecond},
		{Time: 950 * time.Nanosecond},
	}

	lower, upper := comp.CalculateConfidenceInterval(results, 0.95)

	if lower <= 0 {
		t.Errorf("lower = %v, want > 0", lower)
	}

	if upper <= 0 {
		t.Errorf("upper = %v, want > 0", upper)
	}

	if upper <= lower {
		t.Errorf("upper (%v) should be > lower (%v)", upper, lower)
	}

	// Mean should be within the interval
	mean := 1000.0
	if mean < lower || mean > upper {
		t.Errorf("mean (%v) should be within interval [%v, %v]", mean, lower, upper)
	}
}

func TestCalculateConfidenceInterval_EmptyInput(t *testing.T) {
	comp := NewBasicComparator()

	lower, upper := comp.CalculateConfidenceInterval([]*parser.BenchmarkResult{}, 0.95)

	if lower != 0 || upper != 0 {
		t.Errorf("CalculateConfidenceInterval(empty) = (%v, %v), want (0, 0)", lower, upper)
	}
}

func TestCohensDEffect(t *testing.T) {
	// Test case 1: Large effect size
	group1 := []float64{1, 2, 3}
	group2 := []float64{10, 11, 12}

	d := CohensDEffect(group1, group2)
	if d <= 1 {
		t.Errorf("CohensDEffect (large difference) = %v, want > 1", d)
	}

	// Test case 2: Small effect size
	group3 := []float64{1, 2, 3}
	group4 := []float64{1.5, 2.5, 3.5}

	d2 := CohensDEffect(group3, group4)
	if math.Abs(d2) >= 1 {
		t.Errorf("CohensDEffect (small difference) = %v, want < 1", d2)
	}

	// Test case 3: No difference
	group5 := []float64{1, 2, 3}
	group6 := []float64{1, 2, 3}

	d3 := CohensDEffect(group5, group6)
	if d3 != 0 {
		t.Errorf("CohensDEffect (identical) = %v, want 0", d3)
	}
}

func TestCohensDEffect_EmptyInput(t *testing.T) {
	d := CohensDEffect([]float64{}, []float64{1, 2, 3})
	if d != 0 {
		t.Errorf("CohensDEffect(empty, data) = %v, want 0", d)
	}

	d2 := CohensDEffect([]float64{1, 2, 3}, []float64{})
	if d2 != 0 {
		t.Errorf("CohensDEffect(data, empty) = %v, want 0", d2)
	}
}

func TestNormalCDF(t *testing.T) {
	// Test known values
	tests := []struct {
		x        float64
		expected float64
		tolerance float64
	}{
		{0, 0.5, 0.01},     // CDF(0) = 0.5
		{1, 0.84, 0.01},    // CDF(1) ≈ 0.84
		{-1, 0.16, 0.01},   // CDF(-1) ≈ 0.16
		{2, 0.98, 0.01},    // CDF(2) ≈ 0.98
		{-2, 0.02, 0.01},   // CDF(-2) ≈ 0.02
	}

	for _, tt := range tests {
		result := normalCDF(tt.x)
		if math.Abs(result-tt.expected) > tt.tolerance {
			t.Errorf("normalCDF(%v) = %v, want ≈%v (tolerance: %v)", tt.x, result, tt.expected, tt.tolerance)
		}
	}
}
