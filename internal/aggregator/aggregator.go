package aggregator

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/jpequegn/benchflow/internal/parser"
)

// DefaultAggregator implements the Aggregator interface
type DefaultAggregator struct{}

// NewAggregator creates a new aggregator instance
func NewAggregator() *DefaultAggregator {
	return &DefaultAggregator{}
}

// Aggregate aggregates a benchmark suite into statistics
func (a *DefaultAggregator) Aggregate(suite *parser.BenchmarkSuite) (*AggregatedSuite, error) {
	if suite == nil {
		return nil, fmt.Errorf("suite cannot be nil")
	}

	if len(suite.Results) == 0 {
		return nil, fmt.Errorf("suite has no results")
	}

	aggregated := &AggregatedSuite{
		Results:   make([]*AggregatedResult, 0, len(suite.Results)),
		Metadata:  suite.Metadata,
		Timestamp: suite.Timestamp,
		Stats:     &SuiteStats{},
	}

	// Aggregate each benchmark result
	for _, result := range suite.Results {
		aggResult := &AggregatedResult{
			Name:       result.Name,
			Language:   result.Language,
			Mean:       result.Time,
			Median:     result.Time, // For single iteration, mean = median
			Min:        result.Time,
			Max:        result.Time,
			StdDev:     result.StdDev,
			Iterations: result.Iterations,
			Timestamp:  suite.Timestamp,
		}

		aggregated.Results = append(aggregated.Results, aggResult)
	}

	// Calculate suite statistics
	aggregated.Stats = a.calculateSuiteStats(aggregated.Results)

	return aggregated, nil
}

// calculateSuiteStats calculates overall statistics for the suite
func (a *DefaultAggregator) calculateSuiteStats(results []*AggregatedResult) *SuiteStats {
	if len(results) == 0 {
		return &SuiteStats{}
	}

	stats := &SuiteStats{
		TotalBenchmarks: len(results),
	}

	// Find fastest and slowest benchmarks
	fastest := results[0]
	slowest := results[0]

	for _, r := range results {
		stats.TotalDuration += r.Mean

		if r.Mean < fastest.Mean {
			fastest = r
		}
		if r.Mean > slowest.Mean {
			slowest = r
		}
	}

	stats.FastestBench = fastest.Name
	stats.FastestTime = fastest.Mean
	stats.SlowestBench = slowest.Name
	stats.SlowestTime = slowest.Mean

	return stats
}

// Compare compares two aggregated suites
func (a *DefaultAggregator) Compare(baseline, current *AggregatedSuite, threshold float64) (*ComparisonSuite, error) {
	if baseline == nil || current == nil {
		return nil, fmt.Errorf("baseline and current suites cannot be nil")
	}

	// Create a map of baseline results for quick lookup
	baselineMap := make(map[string]*AggregatedResult)
	for _, r := range baseline.Results {
		baselineMap[r.Name] = r
	}

	comparison := &ComparisonSuite{
		Comparisons: make([]*Comparison, 0),
		Threshold:   threshold,
		Timestamp:   time.Now(),
		Metadata:    make(map[string]string),
	}

	// Compare each current result with baseline
	for _, currentResult := range current.Results {
		baselineResult, exists := baselineMap[currentResult.Name]
		if !exists {
			// Benchmark doesn't exist in baseline, skip
			continue
		}

		comp := a.compareResults(baselineResult, currentResult, threshold)
		comparison.Comparisons = append(comparison.Comparisons, comp)

		// Update counts
		if comp.Regression {
			comparison.RegressionCount++
		} else if comp.Improvement {
			comparison.ImprovementCount++
		} else {
			comparison.UnchangedCount++
		}
	}

	return comparison, nil
}

// compareResults compares two aggregated results
func (a *DefaultAggregator) compareResults(baseline, current *AggregatedResult, threshold float64) *Comparison {
	delta := current.Mean - baseline.Mean
	deltaPercent := 0.0

	if baseline.Mean > 0 {
		deltaPercent = (float64(delta) / float64(baseline.Mean)) * 100.0
	}

	comp := &Comparison{
		Name:         current.Name,
		Baseline:     baseline,
		Current:      current,
		Delta:        delta,
		DeltaPercent: deltaPercent,
	}

	// Determine if this is a regression or improvement
	// Positive delta means slower (regression), negative means faster (improvement)
	absPercent := math.Abs(deltaPercent)
	if absPercent > threshold {
		if delta > 0 {
			comp.Regression = true
		} else {
			comp.Improvement = true
		}
	}

	return comp
}

// Export exports aggregated results to the specified format
func (a *DefaultAggregator) Export(suite *AggregatedSuite, format ExportFormat) ([]byte, error) {
	if suite == nil {
		return nil, fmt.Errorf("suite cannot be nil")
	}

	switch format {
	case FormatJSON:
		return a.exportJSON(suite)
	case FormatCSV:
		return a.exportCSV(suite)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// exportJSON exports results as JSON
func (a *DefaultAggregator) exportJSON(suite *AggregatedSuite) ([]byte, error) {
	data, err := json.MarshalIndent(suite, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return data, nil
}

// exportCSV exports results as CSV
func (a *DefaultAggregator) exportCSV(suite *AggregatedSuite) ([]byte, error) {
	var buf strings.Builder
	writer := csv.NewWriter(&buf)

	// Write header
	header := []string{"Name", "Language", "Mean (ns)", "Median (ns)", "Min (ns)", "Max (ns)", "StdDev (ns)", "Iterations"}
	if err := writer.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write CSV header: %w", err)
	}

	// Write data rows
	for _, result := range suite.Results {
		row := []string{
			result.Name,
			result.Language,
			fmt.Sprintf("%d", result.Mean.Nanoseconds()),
			fmt.Sprintf("%d", result.Median.Nanoseconds()),
			fmt.Sprintf("%d", result.Min.Nanoseconds()),
			fmt.Sprintf("%d", result.Max.Nanoseconds()),
			fmt.Sprintf("%d", result.StdDev.Nanoseconds()),
			fmt.Sprintf("%d", result.Iterations),
		}
		if err := writer.Write(row); err != nil {
			return nil, fmt.Errorf("failed to write CSV row: %w", err)
		}
	}

	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("CSV writer error: %w", err)
	}

	return []byte(buf.String()), nil
}

// CalculateStatistics calculates statistical measures for a set of durations
func CalculateStatistics(durations []time.Duration) (mean, median, stdDev time.Duration) {
	if len(durations) == 0 {
		return 0, 0, 0
	}

	// Calculate mean
	var sum int64
	for _, d := range durations {
		sum += d.Nanoseconds()
	}
	mean = time.Duration(sum / int64(len(durations)))

	// Calculate median
	sorted := make([]time.Duration, len(durations))
	copy(sorted, durations)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i] < sorted[j]
	})

	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		median = (sorted[mid-1] + sorted[mid]) / 2
	} else {
		median = sorted[mid]
	}

	// Calculate standard deviation
	var variance float64
	for _, d := range durations {
		diff := float64(d.Nanoseconds() - mean.Nanoseconds())
		variance += diff * diff
	}
	variance /= float64(len(durations))
	stdDev = time.Duration(math.Sqrt(variance))

	return mean, median, stdDev
}
