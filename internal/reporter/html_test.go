package reporter

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/jpequegn/benchflow/internal/aggregator"
)

func TestNewHTMLReporter(t *testing.T) {
	reporter, err := NewHTMLReporter()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if reporter == nil {
		t.Fatal("expected reporter, got nil")
	}

	if reporter.templates == nil {
		t.Fatal("expected templates to be loaded")
	}
}

func TestHTMLReporter_GenerateSummary_Success(t *testing.T) {
	reporter, err := NewHTMLReporter()
	if err != nil {
		t.Fatalf("failed to create reporter: %v", err)
	}

	suite := &aggregator.AggregatedSuite{
		Results: []*aggregator.AggregatedResult{
			{
				Name:       "bench_test",
				Language:   "rust",
				Mean:       100 * time.Millisecond,
				Median:     98 * time.Millisecond,
				Min:        90 * time.Millisecond,
				Max:        110 * time.Millisecond,
				StdDev:     10 * time.Millisecond,
				Iterations: 1000,
			},
		},
		Timestamp: time.Now(),
		Stats: &aggregator.SuiteStats{
			TotalBenchmarks: 1,
			FastestBench:    "bench_test",
			FastestTime:     100 * time.Millisecond,
			SlowestBench:    "bench_test",
			SlowestTime:     100 * time.Millisecond,
			TotalDuration:   100 * time.Millisecond,
		},
	}

	opts := &ReportOptions{
		Title:       "Test Report",
		DarkMode:    true,
		ShowCharts:  true,
		ShowDetails: true,
	}

	var buf bytes.Buffer
	err = reporter.GenerateSummary(suite, opts, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Verify HTML structure
	if !strings.Contains(output, "<!DOCTYPE html>") {
		t.Error("expected valid HTML document")
	}

	// Verify title
	if !strings.Contains(output, "Test Report") {
		t.Error("expected title in output")
	}

	// Verify benchmark data
	if !strings.Contains(output, "bench_test") {
		t.Error("expected benchmark name in output")
	}

	// Verify chart script (when ShowCharts is true)
	if !strings.Contains(output, "chart.umd.min.js") && !strings.Contains(output, "benchmarkChart") {
		t.Error("expected chart script in output")
	}
}

func TestHTMLReporter_GenerateSummary_NilSuite(t *testing.T) {
	reporter, _ := NewHTMLReporter()

	var buf bytes.Buffer
	err := reporter.GenerateSummary(nil, nil, &buf)
	if err == nil {
		t.Fatal("expected error for nil suite")
	}

	if !strings.Contains(err.Error(), "cannot be nil") {
		t.Errorf("expected 'cannot be nil' error, got: %v", err)
	}
}

func TestHTMLReporter_GenerateSummary_WithoutCharts(t *testing.T) {
	reporter, _ := NewHTMLReporter()

	suite := &aggregator.AggregatedSuite{
		Results: []*aggregator.AggregatedResult{
			{
				Name:     "bench_test",
				Language: "rust",
				Mean:     100 * time.Millisecond,
			},
		},
		Timestamp: time.Now(),
		Stats:     &aggregator.SuiteStats{TotalBenchmarks: 1},
	}

	opts := &ReportOptions{
		ShowCharts: false,
	}

	var buf bytes.Buffer
	err := reporter.GenerateSummary(suite, opts, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Should not include Chart.js when ShowCharts is false
	if strings.Contains(output, "Chart.js") {
		t.Error("expected no Chart.js when ShowCharts is false")
	}
}

func TestHTMLReporter_GenerateComparison_Success(t *testing.T) {
	reporter, _ := NewHTMLReporter()

	comparison := &aggregator.ComparisonSuite{
		Comparisons: []*aggregator.Comparison{
			{
				Name: "bench_test",
				Baseline: &aggregator.AggregatedResult{
					Mean: 100 * time.Millisecond,
				},
				Current: &aggregator.AggregatedResult{
					Mean: 120 * time.Millisecond,
				},
				Delta:        20 * time.Millisecond,
				DeltaPercent: 20.0,
				Regression:   true,
			},
		},
		Threshold:       5.0,
		RegressionCount: 1,
		Timestamp:       time.Now(),
	}

	opts := &ReportOptions{
		Title:       "Comparison Report",
		ShowCharts:  true,
		ShowDetails: true,
	}

	var buf bytes.Buffer
	err := reporter.GenerateComparison(comparison, opts, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Verify comparison data
	if !strings.Contains(output, "bench_test") {
		t.Error("expected benchmark name in output")
	}

	if !strings.Contains(output, "egression") { // Matches "Regression" or "regression"
		t.Error("expected regression flag in output")
	}

	if !strings.Contains(output, "20.00") { // Delta percent might not have % immediately after
		t.Error("expected delta percent in output")
	}
}

func TestHTMLReporter_GenerateComparison_NilComparison(t *testing.T) {
	reporter, _ := NewHTMLReporter()

	var buf bytes.Buffer
	err := reporter.GenerateComparison(nil, nil, &buf)
	if err == nil {
		t.Fatal("expected error for nil comparison")
	}
}

func TestHTMLReporter_GenerateTrend_Success(t *testing.T) {
	reporter, _ := NewHTMLReporter()

	now := time.Now()
	history := []*aggregator.AggregatedResult{
		{
			Name:      "bench_test",
			Mean:      100 * time.Millisecond,
			Timestamp: now,
		},
		{
			Name:      "bench_test",
			Mean:      110 * time.Millisecond,
			Timestamp: now.Add(-1 * time.Hour),
		},
		{
			Name:      "bench_test",
			Mean:      105 * time.Millisecond,
			Timestamp: now.Add(-2 * time.Hour),
		},
	}

	opts := &ReportOptions{
		Title:       "Trend Report",
		ShowCharts:  true,
		ShowDetails: true,
	}

	var buf bytes.Buffer
	err := reporter.GenerateTrend(history, opts, &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()

	// Verify trend data
	if !strings.Contains(output, "bench_test") {
		t.Error("expected benchmark name in output")
	}

	if !strings.Contains(output, "Performance Trend") {
		t.Error("expected trend title in output")
	}

	// Verify chart is included
	if !strings.Contains(output, "trendChart") {
		t.Error("expected trend chart in output")
	}
}

func TestHTMLReporter_GenerateTrend_EmptyHistory(t *testing.T) {
	reporter, _ := NewHTMLReporter()

	var buf bytes.Buffer
	err := reporter.GenerateTrend([]*aggregator.AggregatedResult{}, nil, &buf)
	if err == nil {
		t.Fatal("expected error for empty history")
	}

	if !strings.Contains(err.Error(), "cannot be empty") {
		t.Errorf("expected 'cannot be empty' error, got: %v", err)
	}
}

func TestHTMLReporter_PrepareSummaryChartData(t *testing.T) {
	reporter, _ := NewHTMLReporter()

	suite := &aggregator.AggregatedSuite{
		Results: []*aggregator.AggregatedResult{
			{Name: "bench_a", Mean: 100 * time.Millisecond},
			{Name: "bench_b", Mean: 200 * time.Millisecond},
		},
	}

	chartData := reporter.prepareSummaryChartData(suite)

	if len(chartData.Labels) != 2 {
		t.Errorf("expected 2 labels, got %d", len(chartData.Labels))
	}

	if len(chartData.Datasets) != 1 {
		t.Errorf("expected 1 dataset, got %d", len(chartData.Datasets))
	}

	if len(chartData.Datasets[0].Data) != 2 {
		t.Errorf("expected 2 data points, got %d", len(chartData.Datasets[0].Data))
	}

	// Check values are in milliseconds
	if chartData.Datasets[0].Data[0] != 100.0 {
		t.Errorf("expected 100.0, got %.2f", chartData.Datasets[0].Data[0])
	}
}

func TestHTMLReporter_PrepareComparisonChartData(t *testing.T) {
	reporter, _ := NewHTMLReporter()

	comparison := &aggregator.ComparisonSuite{
		Comparisons: []*aggregator.Comparison{
			{
				Name:     "bench_test",
				Baseline: &aggregator.AggregatedResult{Mean: 100 * time.Millisecond},
				Current:  &aggregator.AggregatedResult{Mean: 120 * time.Millisecond},
			},
		},
	}

	chartData := reporter.prepareComparisonChartData(comparison)

	if len(chartData.Datasets) != 2 {
		t.Errorf("expected 2 datasets (baseline + current), got %d", len(chartData.Datasets))
	}

	if chartData.Datasets[0].Label != "Baseline" {
		t.Errorf("expected Baseline label, got %s", chartData.Datasets[0].Label)
	}

	if chartData.Datasets[1].Label != "Current" {
		t.Errorf("expected Current label, got %s", chartData.Datasets[1].Label)
	}
}

func TestHTMLReporter_PrepareTrendChartData(t *testing.T) {
	reporter, _ := NewHTMLReporter()

	now := time.Now()
	history := []*aggregator.AggregatedResult{
		{Mean: 100 * time.Millisecond, Timestamp: now},
		{Mean: 110 * time.Millisecond, Timestamp: now.Add(-1 * time.Hour)},
		{Mean: 105 * time.Millisecond, Timestamp: now.Add(-2 * time.Hour)},
	}

	chartData := reporter.prepareTrendChartData(history)

	// History should be reversed (oldest first)
	if len(chartData.Labels) != 3 {
		t.Errorf("expected 3 labels, got %d", len(chartData.Labels))
	}

	// First data point should be oldest (105ms)
	if chartData.Datasets[0].Data[0] != 105.0 {
		t.Errorf("expected oldest value first (105.0), got %.2f", chartData.Datasets[0].Data[0])
	}

	// Last data point should be newest (100ms)
	lastIdx := len(chartData.Datasets[0].Data) - 1
	if chartData.Datasets[0].Data[lastIdx] != 100.0 {
		t.Errorf("expected newest value last (100.0), got %.2f", chartData.Datasets[0].Data[lastIdx])
	}
}

func TestTemplateFuncs_FormatDuration(t *testing.T) {
	funcs := templateFuncs()
	formatDuration := funcs["formatDuration"].(func(time.Duration) string)

	tests := []struct {
		input    time.Duration
		expected string
	}{
		{50 * time.Nanosecond, "50 ns"},
		{1500 * time.Nanosecond, "1.50 μs"},
		{2500 * time.Microsecond, "2.50 ms"},
		{3 * time.Second, "3.00 s"},
	}

	for _, tt := range tests {
		result := formatDuration(tt.input)
		if result != tt.expected {
			t.Errorf("formatDuration(%v) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}

func TestTemplateFuncs_FormatPercent(t *testing.T) {
	funcs := templateFuncs()
	formatPercent := funcs["formatPercent"].(func(float64) string)

	tests := []struct {
		input    float64
		expected string
	}{
		{5.5, "5.50%"},
		{-10.25, "-10.25%"},
		{0.0, "0.00%"},
	}

	for _, tt := range tests {
		result := formatPercent(tt.input)
		if result != tt.expected {
			t.Errorf("formatPercent(%.2f) = %s, expected %s", tt.input, result, tt.expected)
		}
	}
}
