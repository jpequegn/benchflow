package comparator

import (
	"math"
	"sort"

	"github.com/jpequegn/benchflow/internal/parser"
)

// Comparator defines the interface for comparing benchmark results
type Comparator interface {
	// Compare compares a baseline suite against a current suite
	Compare(baseline, current *parser.BenchmarkSuite) *ComparisonResult

	// GetSignificance determines if the difference between two results is statistically significant
	GetSignificance(baseline, current *parser.BenchmarkResult, confidenceLevel float64) (significant bool, pValue float64)

	// CalculateConfidenceInterval calculates the confidence interval for benchmark results
	CalculateConfidenceInterval(results []*parser.BenchmarkResult, confidenceLevel float64) (lower, upper float64)
}

// ComparisonResult represents the result of comparing two benchmark suites
type ComparisonResult struct {
	// Benchmarks contains comparison data for each benchmark
	Benchmarks []*BenchmarkComparison

	// Summary contains aggregate statistics about the comparison
	Summary ComparisonSummary

	// Regressions lists names of benchmarks that regressed
	Regressions []string

	// Improvements lists names of benchmarks that improved
	Improvements []string

	// Statistics contains detailed statistics about the comparison
	Statistics ComparisonStats
}

// BenchmarkComparison represents a single benchmark comparison
type BenchmarkComparison struct {
	// Name is the benchmark name
	Name string

	// Language is the programming language
	Language string

	// Baseline is the baseline benchmark result
	Baseline *parser.BenchmarkResult

	// Current is the current benchmark result
	Current *parser.BenchmarkResult

	// TimeDelta is the time change in percentage (negative = faster, positive = slower)
	TimeDelta float64

	// IsRegression indicates if this is a performance regression
	IsRegression bool

	// IsSignificant indicates if the difference is statistically significant
	IsSignificant bool

	// ConfidenceLevel is the confidence level used (e.g., 0.95 for 95%)
	ConfidenceLevel float64

	// TTestPValue is the p-value from the t-test
	TTestPValue float64

	// EffectSize is Cohen's d effect size
	EffectSize float64

	// RegressionThreshold is the threshold for regression detection
	RegressionThreshold float64
}

// ComparisonSummary contains aggregate summary statistics
type ComparisonSummary struct {
	// TotalComparisons is the total number of comparisons
	TotalComparisons int

	// Regressions is the count of regressions
	Regressions int

	// Improvements is the count of improvements
	Improvements int

	// AverageDelta is the average time delta percentage
	AverageDelta float64

	// MaxDelta is the maximum time delta percentage
	MaxDelta float64

	// MinDelta is the minimum time delta percentage
	MinDelta float64

	// SignificantChanges is the count of statistically significant changes
	SignificantChanges int
}

// ComparisonStats contains detailed statistical information
type ComparisonStats struct {
	// ConfidenceLevel is the confidence level used
	ConfidenceLevel float64

	// SignificanceLevel is 1 - ConfidenceLevel
	SignificanceLevel float64

	// RegressionThreshold is the threshold for regression detection
	RegressionThreshold float64
}

// BasicComparator implements the Comparator interface
type BasicComparator struct {
	// ConfidenceLevel is the desired confidence level (default: 0.95)
	ConfidenceLevel float64

	// RegressionThreshold is the multiplier for regression detection (default: 1.05 = 5%)
	RegressionThreshold float64
}

// NewBasicComparator creates a new BasicComparator with default settings
func NewBasicComparator() *BasicComparator {
	return &BasicComparator{
		ConfidenceLevel:     0.95,
		RegressionThreshold: 1.05,
	}
}

// Compare compares a baseline suite against a current suite
func (bc *BasicComparator) Compare(baseline, current *parser.BenchmarkSuite) *ComparisonResult {
	result := &ComparisonResult{
		Benchmarks:   make([]*BenchmarkComparison, 0),
		Regressions: make([]string, 0),
		Improvements: make([]string, 0),
		Statistics: ComparisonStats{
			ConfidenceLevel:     bc.ConfidenceLevel,
			SignificanceLevel:   1 - bc.ConfidenceLevel,
			RegressionThreshold: bc.RegressionThreshold,
		},
	}

	if baseline == nil || current == nil || len(baseline.Results) == 0 || len(current.Results) == 0 {
		return result
	}

	// Create a map of baseline results by name for quick lookup
	baselineMap := make(map[string]*parser.BenchmarkResult)
	for _, br := range baseline.Results {
		baselineMap[br.Name] = br
	}

	// Compare each current result with its baseline
	for _, currentResult := range current.Results {
		baselineResult, found := baselineMap[currentResult.Name]
		if !found {
			// No baseline for this benchmark, skip it
			continue
		}

		if baselineResult.Language != currentResult.Language {
			// Different languages, skip
			continue
		}

		// Calculate comparison
		comparison := bc.compareResults(baselineResult, currentResult)
		result.Benchmarks = append(result.Benchmarks, comparison)

		// Track regressions and improvements
		if comparison.IsRegression {
			result.Regressions = append(result.Regressions, comparison.Name)
		} else if comparison.TimeDelta < 0 {
			result.Improvements = append(result.Improvements, comparison.Name)
		}
	}

	// Calculate summary statistics
	result.Summary = bc.calculateSummary(result)

	return result
}

// compareResults compares two individual benchmark results
func (bc *BasicComparator) compareResults(baseline, current *parser.BenchmarkResult) *BenchmarkComparison {
	comparison := &BenchmarkComparison{
		Name:                 current.Name,
		Language:             current.Language,
		Baseline:             baseline,
		Current:              current,
		ConfidenceLevel:      bc.ConfidenceLevel,
		RegressionThreshold:  bc.RegressionThreshold,
	}

	// Calculate time delta percentage (negative = faster, positive = slower)
	if baseline.Time == 0 {
		comparison.TimeDelta = 0
	} else {
		comparison.TimeDelta = ((float64(current.Time) - float64(baseline.Time)) / float64(baseline.Time)) * 100
	}

	// Determine if this is a regression based on threshold
	timeRatio := float64(current.Time) / float64(baseline.Time)
	comparison.IsRegression = timeRatio > bc.RegressionThreshold

	// Calculate statistical significance
	comparison.IsSignificant, comparison.TTestPValue = bc.GetSignificance(baseline, current, bc.ConfidenceLevel)

	// Calculate effect size
	comparison.EffectSize = CohensDEffect(
		[]float64{float64(baseline.Time)},
		[]float64{float64(current.Time)},
	)

	return comparison
}

// calculateSummary calculates summary statistics from comparisons
func (bc *BasicComparator) calculateSummary(result *ComparisonResult) ComparisonSummary {
	summary := ComparisonSummary{
		TotalComparisons: len(result.Benchmarks),
		Regressions:     len(result.Regressions),
		Improvements:    len(result.Improvements),
	}

	if len(result.Benchmarks) == 0 {
		return summary
	}

	// Calculate average, max, and min deltas
	deltas := make([]float64, 0, len(result.Benchmarks))
	for _, comp := range result.Benchmarks {
		deltas = append(deltas, comp.TimeDelta)
		if comp.IsSignificant {
			summary.SignificantChanges++
		}
	}

	if len(deltas) > 0 {
		sort.Float64s(deltas)
		summary.MinDelta = deltas[0]
		summary.MaxDelta = deltas[len(deltas)-1]

		// Calculate average
		sum := 0.0
		for _, d := range deltas {
			sum += d
		}
		summary.AverageDelta = sum / float64(len(deltas))
	}

	return summary
}

// GetSignificance determines if the difference between two results is statistically significant
// Uses a simple t-test with the assumption that we have minimal data
func (bc *BasicComparator) GetSignificance(baseline, current *parser.BenchmarkResult, confidenceLevel float64) (bool, float64) {
	if baseline == nil || current == nil || baseline.Time == 0 || current.Time == 0 {
		return false, 1.0
	}

	// For simplicity, we'll use a very basic approach:
	// Calculate the relative difference and use standard deviation
	baselineTime := float64(baseline.Time)
	currentTime := float64(current.Time)
	baselineStdDev := float64(baseline.StdDev)
	currentStdDev := float64(current.StdDev)

	// If we don't have standard deviations, estimate them
	if baselineStdDev == 0 {
		baselineStdDev = baselineTime * 0.05 // Assume 5% variance
	}
	if currentStdDev == 0 {
		currentStdDev = currentTime * 0.05
	}

	// Calculate pooled standard deviation
	pooledStdDev := math.Sqrt((baselineStdDev*baselineStdDev + currentStdDev*currentStdDev) / 2)

	// Calculate t-statistic (simplified with n=1)
	if pooledStdDev == 0 {
		pooledStdDev = baselineTime * 0.01 // Avoid division by zero
	}

	tStat := (currentTime - baselineTime) / pooledStdDev

	// Approximate p-value from t-statistic
	// For practical purposes with small sample sizes, we use a simple threshold
	pValue := 2 * (1 - normalCDF(math.Abs(tStat)))

	// Determine significance at the given confidence level
	alpha := 1 - confidenceLevel
	isSignificant := pValue < alpha

	return isSignificant, pValue
}

// CalculateConfidenceInterval calculates the confidence interval for benchmark results
func (bc *BasicComparator) CalculateConfidenceInterval(results []*parser.BenchmarkResult, confidenceLevel float64) (lower, upper float64) {
	if len(results) == 0 {
		return 0, 0
	}

	// Calculate mean
	sum := 0.0
	for _, r := range results {
		sum += float64(r.Time)
	}
	mean := sum / float64(len(results))

	// Calculate standard deviation
	varianceSum := 0.0
	for _, r := range results {
		diff := float64(r.Time) - mean
		varianceSum += diff * diff
	}
	stdDev := math.Sqrt(varianceSum / float64(len(results)-1))

	// Calculate confidence interval using t-distribution approximation
	// For small samples, use a simple z-score approximation
	zScore := 1.96 // 95% confidence
	if confidenceLevel == 0.99 {
		zScore = 2.576 // 99% confidence
	}

	marginOfError := zScore * (stdDev / math.Sqrt(float64(len(results))))
	lower = mean - marginOfError
	upper = mean + marginOfError

	// Ensure lower bound is not negative
	if lower < 0 {
		lower = 0
	}

	return lower, upper
}

// normalCDF approximates the cumulative distribution function of the standard normal distribution
// Using a rational approximation
func normalCDF(x float64) float64 {
	// Constants for rational approximation
	b1 := 0.319381530
	b2 := -0.356563782
	b3 := 1.781477937
	b4 := -1.821255978
	b5 := 1.330274429
	p := 0.2316419
	c := 0.39894228

	if x >= 0 {
		t := 1.0 / (1.0 + p*x)
		return 1.0 - c*math.Exp(-x*x/2.0)*t*(b1+t*(b2+t*(b3+t*(b4+t*b5))))
	} else {
		t := 1.0 / (1.0 - p*x)
		return c * math.Exp(-x*x/2.0) * t * (b1+t*(b2+t*(b3+t*(b4+t*b5))))
	}
}

// CohensDEffect calculates Cohen's d effect size
func CohensDEffect(group1, group2 []float64) float64 {
	if len(group1) == 0 || len(group2) == 0 {
		return 0
	}

	// Calculate means
	mean1 := calculateMean(group1)
	mean2 := calculateMean(group2)

	// Calculate standard deviations
	std1 := calculateStdDev(group1, mean1)
	std2 := calculateStdDev(group2, mean2)

	// Calculate pooled standard deviation
	n1 := float64(len(group1))
	n2 := float64(len(group2))
	variance1 := std1 * std1
	variance2 := std2 * std2

	pooledVariance := ((n1 - 1) * variance1 + (n2 - 1) * variance2) / (n1 + n2 - 2)
	pooledStdDev := math.Sqrt(pooledVariance)

	if pooledStdDev == 0 {
		return 0
	}

	// Calculate Cohen's d
	d := (mean2 - mean1) / pooledStdDev
	return d
}

// calculateMean calculates the mean of a slice
func calculateMean(data []float64) float64 {
	if len(data) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range data {
		sum += v
	}
	return sum / float64(len(data))
}

// calculateStdDev calculates the standard deviation of a slice
func calculateStdDev(data []float64, mean float64) float64 {
	if len(data) <= 1 {
		return 0
	}
	varianceSum := 0.0
	for _, v := range data {
		diff := v - mean
		varianceSum += diff * diff
	}
	variance := varianceSum / float64(len(data)-1)
	return math.Sqrt(variance)
}
