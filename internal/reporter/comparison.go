package reporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/jpequegn/benchflow/internal/comparator"
)

// ComparisonReporter generates comparison reports in various formats
type ComparisonReporter interface {
	GenerateMarkdown(result *comparator.ComparisonResult) (string, error)
	GenerateHTML(result *comparator.ComparisonResult) (string, error)
	GenerateJSON(result *comparator.ComparisonResult) (string, error)
}

// BasicComparisonReporter implements ComparisonReporter
type BasicComparisonReporter struct{}

// NewBasicComparisonReporter creates a new BasicComparisonReporter
func NewBasicComparisonReporter() *BasicComparisonReporter {
	return &BasicComparisonReporter{}
}

// GenerateMarkdown generates a Markdown comparison report
func (bcr *BasicComparisonReporter) GenerateMarkdown(result *comparator.ComparisonResult) (string, error) {
	if result == nil || len(result.Benchmarks) == 0 {
		return "# Comparison Report\n\nNo benchmarks to compare.\n", nil
	}

	var buf bytes.Buffer

	// Header
	buf.WriteString("# Performance Comparison Report\n\n")

	// Summary section
	buf.WriteString("## Summary\n\n")
	buf.WriteString(fmt.Sprintf("- **Total Comparisons**: %d\n", result.Summary.TotalComparisons))
	buf.WriteString(fmt.Sprintf("- **Regressions**: %d\n", result.Summary.Regressions))
	buf.WriteString(fmt.Sprintf("- **Improvements**: %d\n", result.Summary.Improvements))
	buf.WriteString(fmt.Sprintf("- **Average Delta**: %.2f%%\n", result.Summary.AverageDelta))
	buf.WriteString(fmt.Sprintf("- **Max Delta**: %.2f%%\n", result.Summary.MaxDelta))
	buf.WriteString(fmt.Sprintf("- **Min Delta**: %.2f%%\n", result.Summary.MinDelta))
	buf.WriteString(fmt.Sprintf("- **Significant Changes**: %d\n\n", result.Summary.SignificantChanges))

	// Regressions section
	if len(result.Regressions) > 0 {
		buf.WriteString("## ⚠️ Regressions\n\n")
		for _, name := range result.Regressions {
			buf.WriteString(fmt.Sprintf("- `%s`\n", name))
		}
		buf.WriteString("\n")
	}

	// Improvements section
	if len(result.Improvements) > 0 {
		buf.WriteString("## ✅ Improvements\n\n")
		for _, name := range result.Improvements {
			buf.WriteString(fmt.Sprintf("- `%s`\n", name))
		}
		buf.WriteString("\n")
	}

	// Detailed results table
	buf.WriteString("## Detailed Results\n\n")
	buf.WriteString(bcr.generateMarkdownTable(result.Benchmarks))

	return buf.String(), nil
}

// generateMarkdownTable creates a Markdown table for benchmark comparisons
func (bcr *BasicComparisonReporter) generateMarkdownTable(comparisons []*comparator.BenchmarkComparison) string {
	if len(comparisons) == 0 {
		return ""
	}

	var buf bytes.Buffer

	// Table header
	buf.WriteString("| Benchmark | Language | Baseline | Current | Delta | Status | P-Value | Effect Size |\n")
	buf.WriteString("|-----------|----------|----------|---------|-------|--------|---------|-------------|\n")

	// Sort comparisons by name
	sorted := make([]*comparator.BenchmarkComparison, len(comparisons))
	copy(sorted, comparisons)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})

	for _, comp := range sorted {
		status := "→"
		if comp.IsRegression {
			status = "🔴"
		} else if comp.TimeDelta < 0 {
			status = "🟢"
		}

		baselineNs := comp.Baseline.Time.Nanoseconds()
		currentNs := comp.Current.Time.Nanoseconds()

		buf.WriteString(fmt.Sprintf("| %s | %s | %d ns | %d ns | %.2f%% | %s | %.4f | %.2f |\n",
			comp.Name,
			comp.Language,
			baselineNs,
			currentNs,
			comp.TimeDelta,
			status,
			comp.TTestPValue,
			comp.EffectSize,
		))
	}

	return buf.String()
}

// GenerateHTML generates an HTML comparison report (placeholder)
func (bcr *BasicComparisonReporter) GenerateHTML(result *comparator.ComparisonResult) (string, error) {
	if result == nil || len(result.Benchmarks) == 0 {
		return "<h1>Comparison Report</h1><p>No benchmarks to compare.</p>", nil
	}

	var buf bytes.Buffer

	// HTML header
	buf.WriteString(`<!DOCTYPE html>
<html>
<head>
	<title>Performance Comparison Report</title>
	<style>
		body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
		.container { max-width: 1200px; margin: 0 auto; background-color: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
		h1 { color: #333; border-bottom: 2px solid #007bff; padding-bottom: 10px; }
		h2 { color: #555; margin-top: 30px; }
		.summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 15px; margin: 20px 0; }
		.stat-box { padding: 15px; background-color: #f8f9fa; border-left: 4px solid #007bff; border-radius: 4px; }
		.stat-label { font-size: 12px; color: #666; text-transform: uppercase; }
		.stat-value { font-size: 24px; font-weight: bold; color: #333; margin-top: 5px; }
		table { width: 100%; border-collapse: collapse; margin: 20px 0; }
		th { background-color: #f8f9fa; padding: 12px; text-align: left; font-weight: 600; border-bottom: 2px solid #dee2e6; }
		td { padding: 12px; border-bottom: 1px solid #dee2e6; }
		tr:hover { background-color: #f5f5f5; }
		.regression { color: #dc3545; font-weight: bold; }
		.improvement { color: #28a745; font-weight: bold; }
		.icon { font-size: 20px; margin-right: 5px; }
	</style>
</head>
<body>
	<div class="container">
		<h1>Performance Comparison Report</h1>
`)

	// Summary section
	buf.WriteString(`		<h2>Summary</h2>
		<div class="summary">
`)
	buf.WriteString(fmt.Sprintf(`			<div class="stat-box"><div class="stat-label">Total Comparisons</div><div class="stat-value">%d</div></div>`, result.Summary.TotalComparisons))
	buf.WriteString(fmt.Sprintf(`			<div class="stat-box"><div class="stat-label">Regressions</div><div class="stat-value" style="color: #dc3545;">%d</div></div>`, result.Summary.Regressions))
	buf.WriteString(fmt.Sprintf(`			<div class="stat-box"><div class="stat-label">Improvements</div><div class="stat-value" style="color: #28a745;">%d</div></div>`, result.Summary.Improvements))
	buf.WriteString(fmt.Sprintf(`			<div class="stat-box"><div class="stat-label">Average Delta</div><div class="stat-value">%.2f%%</div></div>`, result.Summary.AverageDelta))
	buf.WriteString(`		</div>
`)

	// Detailed results table
	buf.WriteString(`		<h2>Detailed Results</h2>
		<table>
			<thead>
				<tr>
					<th>Benchmark</th>
					<th>Language</th>
					<th>Baseline</th>
					<th>Current</th>
					<th>Delta</th>
					<th>P-Value</th>
					<th>Effect Size</th>
				</tr>
			</thead>
			<tbody>
`)

	// Sort comparisons by name
	sorted := make([]*comparator.BenchmarkComparison, len(result.Benchmarks))
	copy(sorted, result.Benchmarks)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Name < sorted[j].Name
	})

	for _, comp := range sorted {
		statusClass := ""
		if comp.IsRegression {
			statusClass = `class="regression"`
		} else if comp.TimeDelta < 0 {
			statusClass = `class="improvement"`
		}

		baselineNs := comp.Baseline.Time.Nanoseconds()
		currentNs := comp.Current.Time.Nanoseconds()

		buf.WriteString(fmt.Sprintf(`				<tr>
					<td>%s</td>
					<td>%s</td>
					<td>%d ns</td>
					<td>%d ns</td>
					<td %s>%.2f%%</td>
					<td>%.4f</td>
					<td>%.2f</td>
				</tr>
`, comp.Name, comp.Language, baselineNs, currentNs, statusClass, comp.TimeDelta, comp.TTestPValue, comp.EffectSize))
	}

	buf.WriteString(`			</tbody>
		</table>
	</div>
</body>
</html>
`)

	return buf.String(), nil
}

// GenerateJSON generates a JSON comparison report
func (bcr *BasicComparisonReporter) GenerateJSON(result *comparator.ComparisonResult) (string, error) {
	if result == nil {
		return "{}", nil
	}

	// Create a JSON-serializable structure
	jsonData := map[string]interface{}{
		"summary": map[string]interface{}{
			"total_comparisons":   result.Summary.TotalComparisons,
			"regressions":         result.Summary.Regressions,
			"improvements":        result.Summary.Improvements,
			"average_delta":       result.Summary.AverageDelta,
			"max_delta":           result.Summary.MaxDelta,
			"min_delta":           result.Summary.MinDelta,
			"significant_changes": result.Summary.SignificantChanges,
		},
		"regressions":  result.Regressions,
		"improvements": result.Improvements,
		"benchmarks":   bcr.marshalBenchmarkComparisons(result.Benchmarks),
		"statistics": map[string]interface{}{
			"confidence_level":     result.Statistics.ConfidenceLevel,
			"significance_level":   result.Statistics.SignificanceLevel,
			"regression_threshold": result.Statistics.RegressionThreshold,
		},
	}

	// Marshal to JSON with indentation
	data, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// marshalBenchmarkComparisons converts comparisons to JSON-serializable format
func (bcr *BasicComparisonReporter) marshalBenchmarkComparisons(comparisons []*comparator.BenchmarkComparison) []map[string]interface{} {
	results := make([]map[string]interface{}, 0, len(comparisons))

	for _, comp := range comparisons {
		results = append(results, map[string]interface{}{
			"name":                 comp.Name,
			"language":             comp.Language,
			"baseline_time_ns":     comp.Baseline.Time.Nanoseconds(),
			"current_time_ns":      comp.Current.Time.Nanoseconds(),
			"time_delta_percent":   comp.TimeDelta,
			"is_regression":        comp.IsRegression,
			"is_significant":       comp.IsSignificant,
			"confidence_level":     comp.ConfidenceLevel,
			"t_test_p_value":       comp.TTestPValue,
			"effect_size_cohens_d": comp.EffectSize,
			"regression_threshold": comp.RegressionThreshold,
		})
	}

	return results
}
