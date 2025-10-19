package reporter

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/jpequegn/benchflow/internal/comparator"
	"github.com/jpequegn/benchflow/internal/parser"
)

func createTestComparisonResult() *comparator.ComparisonResult {
	result := &comparator.ComparisonResult{
		Benchmarks: []*comparator.BenchmarkComparison{
			{
				Name:                "sort",
				Language:            "go",
				Baseline:            &parser.BenchmarkResult{Time: 1000 * time.Nanosecond},
				Current:             &parser.BenchmarkResult{Time: 950 * time.Nanosecond},
				TimeDelta:           -5.0,
				IsRegression:        false,
				IsSignificant:       true,
				ConfidenceLevel:     0.95,
				TTestPValue:         0.02,
				EffectSize:          0.8,
				RegressionThreshold: 1.05,
			},
			{
				Name:                "search",
				Language:            "go",
				Baseline:            &parser.BenchmarkResult{Time: 500 * time.Nanosecond},
				Current:             &parser.BenchmarkResult{Time: 600 * time.Nanosecond},
				TimeDelta:           20.0,
				IsRegression:        true,
				IsSignificant:       true,
				ConfidenceLevel:     0.95,
				TTestPValue:         0.01,
				EffectSize:          1.2,
				RegressionThreshold: 1.05,
			},
		},
		Summary: comparator.ComparisonSummary{
			TotalComparisons:   2,
			Regressions:       1,
			Improvements:      1,
			AverageDelta:      7.5,
			MaxDelta:          20.0,
			MinDelta:          -5.0,
			SignificantChanges: 2,
		},
		Regressions:  []string{"search"},
		Improvements: []string{"sort"},
		Statistics: comparator.ComparisonStats{
			ConfidenceLevel:     0.95,
			SignificanceLevel:   0.05,
			RegressionThreshold: 1.05,
		},
	}
	return result
}

func TestNewBasicComparisonReporter(t *testing.T) {
	reporter := NewBasicComparisonReporter()
	if reporter == nil {
		t.Error("NewBasicComparisonReporter() returned nil")
	}
}

func TestGenerateMarkdown(t *testing.T) {
	reporter := NewBasicComparisonReporter()
	result := createTestComparisonResult()

	markdown, err := reporter.GenerateMarkdown(result)
	if err != nil {
		t.Fatalf("GenerateMarkdown() returned error: %v", err)
	}

	if markdown == "" {
		t.Error("GenerateMarkdown() returned empty string")
	}

	// Check for key sections
	if !strings.Contains(markdown, "# Performance Comparison Report") {
		t.Error("Markdown missing header")
	}

	if !strings.Contains(markdown, "## Summary") {
		t.Error("Markdown missing Summary section")
	}

	if !strings.Contains(markdown, "Total Comparisons") {
		t.Error("Markdown missing Total Comparisons")
	}

	// Check for Regressions section (should be present since result.Regressions is not empty)
	hasRegressions := strings.Contains(markdown, "Regressions")
	hasImprovements := strings.Contains(markdown, "Improvements")

	if !hasRegressions {
		t.Error("Markdown should contain information about regressions")
	}

	if !hasImprovements {
		t.Error("Markdown should contain information about improvements")
	}

	if !strings.Contains(markdown, "## Detailed Results") {
		t.Error("Markdown missing Detailed Results section")
	}

	// Check for benchmark names
	if !strings.Contains(markdown, "sort") {
		t.Error("Markdown missing 'sort' benchmark")
	}

	if !strings.Contains(markdown, "search") {
		t.Error("Markdown missing 'search' benchmark")
	}
}

func TestGenerateMarkdown_EmptyResult(t *testing.T) {
	reporter := NewBasicComparisonReporter()
	result := &comparator.ComparisonResult{
		Benchmarks: make([]*comparator.BenchmarkComparison, 0),
	}

	markdown, err := reporter.GenerateMarkdown(result)
	if err != nil {
		t.Fatalf("GenerateMarkdown(empty) returned error: %v", err)
	}

	if !strings.Contains(markdown, "No benchmarks") {
		t.Error("Markdown should mention no benchmarks")
	}
}

func TestGenerateMarkdown_NilResult(t *testing.T) {
	reporter := NewBasicComparisonReporter()

	markdown, err := reporter.GenerateMarkdown(nil)
	if err != nil {
		t.Fatalf("GenerateMarkdown(nil) returned error: %v", err)
	}

	if !strings.Contains(markdown, "No benchmarks") {
		t.Error("Markdown should mention no benchmarks for nil result")
	}
}

func TestGenerateHTML(t *testing.T) {
	reporter := NewBasicComparisonReporter()
	result := createTestComparisonResult()

	html, err := reporter.GenerateHTML(result)
	if err != nil {
		t.Fatalf("GenerateHTML() returned error: %v", err)
	}

	if html == "" {
		t.Error("GenerateHTML() returned empty string")
	}

	// Check for HTML structure
	if !strings.Contains(html, "<!DOCTYPE html>") {
		t.Error("HTML missing DOCTYPE")
	}

	if !strings.Contains(html, "<title>") {
		t.Error("HTML missing title tag")
	}

	if !strings.Contains(html, "<table>") {
		t.Error("HTML missing table")
	}

	if !strings.Contains(html, "<thead>") {
		t.Error("HTML missing table header")
	}

	if !strings.Contains(html, "Benchmark") {
		t.Error("HTML missing Benchmark column")
	}

	// Check for benchmark names
	if !strings.Contains(html, "sort") {
		t.Error("HTML missing 'sort' benchmark")
	}

	if !strings.Contains(html, "search") {
		t.Error("HTML missing 'search' benchmark")
	}

	// Check for CSS styling
	if !strings.Contains(html, "background-color") {
		t.Error("HTML missing CSS styling")
	}
}

func TestGenerateHTML_EmptyResult(t *testing.T) {
	reporter := NewBasicComparisonReporter()
	result := &comparator.ComparisonResult{
		Benchmarks: make([]*comparator.BenchmarkComparison, 0),
	}

	html, err := reporter.GenerateHTML(result)
	if err != nil {
		t.Fatalf("GenerateHTML(empty) returned error: %v", err)
	}

	if !strings.Contains(html, "No benchmarks") {
		t.Error("HTML should mention no benchmarks")
	}
}

func TestGenerateJSON(t *testing.T) {
	reporter := NewBasicComparisonReporter()
	result := createTestComparisonResult()

	jsonStr, err := reporter.GenerateJSON(result)
	if err != nil {
		t.Fatalf("GenerateJSON() returned error: %v", err)
	}

	if jsonStr == "" {
		t.Error("GenerateJSON() returned empty string")
	}

	// Parse JSON to verify it's valid
	var data map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		t.Fatalf("GenerateJSON() returned invalid JSON: %v", err)
	}

	// Check for key fields
	if _, ok := data["summary"]; !ok {
		t.Error("JSON missing summary field")
	}

	if _, ok := data["benchmarks"]; !ok {
		t.Error("JSON missing benchmarks field")
	}

	if _, ok := data["statistics"]; !ok {
		t.Error("JSON missing statistics field")
	}

	// Check summary structure
	summary := data["summary"].(map[string]interface{})
	if _, ok := summary["total_comparisons"]; !ok {
		t.Error("JSON summary missing total_comparisons")
	}

	if _, ok := summary["regressions"]; !ok {
		t.Error("JSON summary missing regressions")
	}

	if _, ok := summary["improvements"]; !ok {
		t.Error("JSON summary missing improvements")
	}
}

func TestGenerateJSON_EmptyResult(t *testing.T) {
	reporter := NewBasicComparisonReporter()
	result := &comparator.ComparisonResult{
		Benchmarks: make([]*comparator.BenchmarkComparison, 0),
	}

	jsonStr, err := reporter.GenerateJSON(result)
	if err != nil {
		t.Fatalf("GenerateJSON(empty) returned error: %v", err)
	}

	// Parse JSON to verify it's valid
	var data map[string]interface{}
	err = json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		t.Fatalf("GenerateJSON(empty) returned invalid JSON: %v", err)
	}
}

func TestGenerateJSON_NilResult(t *testing.T) {
	reporter := NewBasicComparisonReporter()

	jsonStr, err := reporter.GenerateJSON(nil)
	if err != nil {
		t.Fatalf("GenerateJSON(nil) returned error: %v", err)
	}

	// Should return empty object
	if jsonStr != "{}" {
		t.Errorf("GenerateJSON(nil) = %q, want {}", jsonStr)
	}
}

func TestGenerateMarkdownTable(t *testing.T) {
	reporter := NewBasicComparisonReporter()
	comparisons := []*comparator.BenchmarkComparison{
		{
			Name:     "benchmark1",
			Language: "go",
			Baseline: &parser.BenchmarkResult{Time: 1000 * time.Nanosecond},
			Current:  &parser.BenchmarkResult{Time: 950 * time.Nanosecond},
			TimeDelta: -5.0,
		},
	}

	table := reporter.generateMarkdownTable(comparisons)

	if !strings.Contains(table, "Benchmark") {
		t.Error("Table missing header")
	}

	if !strings.Contains(table, "benchmark1") {
		t.Error("Table missing benchmark name")
	}

	if !strings.Contains(table, "go") {
		t.Error("Table missing language")
	}
}

func TestMarshalBenchmarkComparisons(t *testing.T) {
	reporter := NewBasicComparisonReporter()
	comparisons := []*comparator.BenchmarkComparison{
		{
			Name:              "test",
			Language:          "rust",
			Baseline:          &parser.BenchmarkResult{Time: 1000 * time.Nanosecond},
			Current:           &parser.BenchmarkResult{Time: 1100 * time.Nanosecond},
			TimeDelta:         10.0,
			IsRegression:      true,
			IsSignificant:     true,
			TTestPValue:       0.01,
			EffectSize:        0.5,
			RegressionThreshold: 1.05,
		},
	}

	marshaled := reporter.marshalBenchmarkComparisons(comparisons)

	if len(marshaled) != 1 {
		t.Errorf("len(marshaled) = %d, want 1", len(marshaled))
	}

	comp := marshaled[0]
	if comp["name"] != "test" {
		t.Errorf("name = %v, want 'test'", comp["name"])
	}

	if comp["language"] != "rust" {
		t.Errorf("language = %v, want 'rust'", comp["language"])
	}

	if comp["is_regression"] != true {
		t.Errorf("is_regression = %v, want true", comp["is_regression"])
	}
}
