package reporter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/jpequegn/benchflow/internal/analyzer"
)

// TrendReporter generates trend analysis reports
type TrendReporter interface {
	GenerateTrendMarkdown(trends []*analyzer.TrendResult, anomalies []*analyzer.Anomaly) (string, error)
	GenerateTrendHTML(trends []*analyzer.TrendResult, anomalies []*analyzer.Anomaly) (string, error)
	GenerateTrendJSON(trends []*analyzer.TrendResult, anomalies []*analyzer.Anomaly) (string, error)
}

// BasicTrendReporter implements TrendReporter
type BasicTrendReporter struct{}

// NewBasicTrendReporter creates a new trend reporter
func NewBasicTrendReporter() *BasicTrendReporter {
	return &BasicTrendReporter{}
}

// GenerateTrendMarkdown generates a Markdown trend report
func (btr *BasicTrendReporter) GenerateTrendMarkdown(trends []*analyzer.TrendResult, anomalies []*analyzer.Anomaly) (string, error) {
	var buf bytes.Buffer

	buf.WriteString("# Performance Trend Analysis Report\n\n")

	// Summary section
	buf.WriteString("## Summary\n\n")
	improving := 0
	degrading := 0
	stable := 0

	for _, t := range trends {
		switch t.Direction {
		case "improving":
			improving++
		case "degrading":
			degrading++
		case "stable":
			stable++
		}
	}

	buf.WriteString(fmt.Sprintf("- **Total Benchmarks**: %d\n", len(trends)))
	buf.WriteString(fmt.Sprintf("- **Improving**: %d (🟢)\n", improving))
	buf.WriteString(fmt.Sprintf("- **Degrading**: %d (🔴)\n", degrading))
	buf.WriteString(fmt.Sprintf("- **Stable**: %d (→)\n", stable))
	if len(anomalies) > 0 {
		buf.WriteString(fmt.Sprintf("- **Anomalies Detected**: %d ⚠️\n", len(anomalies)))
	}
	buf.WriteString("\n")

	// Trends section
	if len(trends) > 0 {
		buf.WriteString("## Trend Analysis\n\n")
		buf.WriteString("| Benchmark | Language | Direction | Change | Slope | R² | Data Points |\n")
		buf.WriteString("|-----------|----------|-----------|--------|-------|-----|-------------|\n")

		// Sort by benchmark name
		sorted := make([]*analyzer.TrendResult, len(trends))
		copy(sorted, trends)
		sort.Slice(sorted, func(i, j int) bool {
			if sorted[i].BenchmarkName != sorted[j].BenchmarkName {
				return sorted[i].BenchmarkName < sorted[j].BenchmarkName
			}
			return sorted[i].Language < sorted[j].Language
		})

		for _, t := range sorted {
			directionEmoji := "→"
			if t.Direction == "improving" {
				directionEmoji = "🟢"
			} else if t.Direction == "degrading" {
				directionEmoji = "🔴"
			}

			buf.WriteString(fmt.Sprintf("| %s | %s | %s %s | %.2f%% | %.2f | %.3f | %d |\n",
				t.BenchmarkName,
				t.Language,
				directionEmoji,
				t.Direction,
				t.ChangePercent,
				t.Slope,
				t.RSquared,
				t.DataPoints,
			))
		}
		buf.WriteString("\n")
	}

	// Anomalies section
	if len(anomalies) > 0 {
		buf.WriteString("## Detected Anomalies\n\n")

		// Group by benchmark
		byBenchmark := make(map[string][]*analyzer.Anomaly)
		for _, a := range anomalies {
			key := a.BenchmarkName + ":" + a.Language
			byBenchmark[key] = append(byBenchmark[key], a)
		}

		for key, anoms := range byBenchmark {
			parts := strings.Split(key, ":")
			buf.WriteString(fmt.Sprintf("### %s (%s)\n\n", parts[0], parts[1]))

			for _, a := range anoms {
				severityEmoji := "⚠️"
				if a.Severity == "critical" {
					severityEmoji = "🚨"
				} else if a.Severity == "high" {
					severityEmoji = "⛔"
				}

				buf.WriteString(fmt.Sprintf("- **%s** %s: %.2f%% deviation (Z-score: %.2f)\n",
					a.Timestamp.Format("2006-01-02 15:04"),
					severityEmoji,
					a.ZScore*100/3, // Approximate percentage
					a.ZScore,
				))
				if a.IsRegression {
					buf.WriteString("  ⚠️ Regression detected\n")
				}
			}
			buf.WriteString("\n")
		}
	}

	// Legend
	buf.WriteString("## Legend\n\n")
	buf.WriteString("- **Direction**: 🟢 improving, 🔴 degrading, → stable\n")
	buf.WriteString("- **Change**: Percentage change from first to last measurement\n")
	buf.WriteString("- **Slope**: Change per day (ns/day)\n")
	buf.WriteString("- **R²**: Trend confidence (0-1, higher = more reliable)\n")
	buf.WriteString("- **Data Points**: Number of measurements in trend\n")

	return buf.String(), nil
}

// GenerateTrendHTML generates an HTML trend report
func (btr *BasicTrendReporter) GenerateTrendHTML(trends []*analyzer.TrendResult, anomalies []*analyzer.Anomaly) (string, error) {
	var buf bytes.Buffer

	buf.WriteString(`<!DOCTYPE html>
<html>
<head>
	<title>Performance Trend Analysis Report</title>
	<style>
		body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif; margin: 20px; background-color: #f5f5f5; }
		.container { max-width: 1200px; margin: 0 auto; background-color: white; padding: 20px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
		h1 { color: #333; border-bottom: 2px solid #007bff; padding-bottom: 10px; }
		h2 { color: #555; margin-top: 30px; }
		.summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(150px, 1fr)); gap: 15px; margin: 20px 0; }
		.stat-box { padding: 15px; background-color: #f8f9fa; border-left: 4px solid #007bff; border-radius: 4px; }
		.stat-label { font-size: 12px; color: #666; text-transform: uppercase; }
		.stat-value { font-size: 24px; font-weight: bold; color: #333; margin-top: 5px; }
		table { width: 100%; border-collapse: collapse; margin: 20px 0; }
		th { background-color: #f8f9fa; padding: 12px; text-align: left; font-weight: 600; border-bottom: 2px solid #dee2e6; }
		td { padding: 12px; border-bottom: 1px solid #dee2e6; }
		tr:hover { background-color: #f5f5f5; }
		.improving { color: #28a745; font-weight: bold; }
		.degrading { color: #dc3545; font-weight: bold; }
		.stable { color: #666; }
		.anomaly { background-color: #fff3cd; padding: 10px; margin: 10px 0; border-left: 4px solid #ffc107; border-radius: 4px; }
		.critical { background-color: #f8d7da; border-left-color: #dc3545; }
		.high { background-color: #fff3cd; border-left-color: #fd7e14; }
		.legend { background-color: #f8f9fa; padding: 15px; border-radius: 4px; margin-top: 20px; }
	</style>
</head>
<body>
	<div class="container">
		<h1>Performance Trend Analysis Report</h1>
`)

	// Summary
	improving := 0
	degrading := 0
	stable := 0
	for _, t := range trends {
		switch t.Direction {
		case "improving":
			improving++
		case "degrading":
			degrading++
		case "stable":
			stable++
		}
	}

	buf.WriteString(`		<h2>Summary</h2>
		<div class="summary">
`)
	buf.WriteString(fmt.Sprintf(`			<div class="stat-box"><div class="stat-label">Total</div><div class="stat-value">%d</div></div>`, len(trends)))
	buf.WriteString(fmt.Sprintf(`			<div class="stat-box"><div class="stat-label">Improving 🟢</div><div class="stat-value" style="color: #28a745;">%d</div></div>`, improving))
	buf.WriteString(fmt.Sprintf(`			<div class="stat-box"><div class="stat-label">Degrading 🔴</div><div class="stat-value" style="color: #dc3545;">%d</div></div>`, degrading))
	buf.WriteString(fmt.Sprintf(`			<div class="stat-box"><div class="stat-label">Stable →</div><div class="stat-value">%d</div></div>`, stable))
	if len(anomalies) > 0 {
		buf.WriteString(fmt.Sprintf(`			<div class="stat-box"><div class="stat-label">Anomalies ⚠️</div><div class="stat-value" style="color: #ffc107;">%d</div></div>`, len(anomalies)))
	}
	buf.WriteString(`		</div>
`)

	// Trends table
	if len(trends) > 0 {
		buf.WriteString(`		<h2>Trend Analysis</h2>
		<table>
			<thead>
				<tr>
					<th>Benchmark</th>
					<th>Language</th>
					<th>Direction</th>
					<th>Change</th>
					<th>Slope</th>
					<th>R²</th>
					<th>Data Points</th>
				</tr>
			</thead>
			<tbody>
`)

		sorted := make([]*analyzer.TrendResult, len(trends))
		copy(sorted, trends)
		sort.Slice(sorted, func(i, j int) bool {
			if sorted[i].BenchmarkName != sorted[j].BenchmarkName {
				return sorted[i].BenchmarkName < sorted[j].BenchmarkName
			}
			return sorted[i].Language < sorted[j].Language
		})

		for _, t := range sorted {
			class := ""
			emoji := "→"
			if t.Direction == "improving" {
				class = "improving"
				emoji = "🟢"
			} else if t.Direction == "degrading" {
				class = "degrading"
				emoji = "🔴"
			} else {
				class = "stable"
			}

			buf.WriteString(fmt.Sprintf(`				<tr>
					<td>%s</td>
					<td>%s</td>
					<td class="%s">%s %s</td>
					<td>%.2f%%</td>
					<td>%.2f ns/day</td>
					<td>%.3f</td>
					<td>%d</td>
				</tr>
`, t.BenchmarkName, t.Language, class, emoji, t.Direction, t.ChangePercent, t.Slope, t.RSquared, t.DataPoints))
		}

		buf.WriteString(`			</tbody>
		</table>
`)
	}

	// Anomalies
	if len(anomalies) > 0 {
		buf.WriteString(`		<h2>Detected Anomalies</h2>
`)

		// Group by benchmark
		byBenchmark := make(map[string][]*analyzer.Anomaly)
		for _, a := range anomalies {
			key := a.BenchmarkName + ":" + a.Language
			byBenchmark[key] = append(byBenchmark[key], a)
		}

		for key, anoms := range byBenchmark {
			parts := strings.Split(key, ":")
			buf.WriteString(fmt.Sprintf(`		<h3>%s (%s)</h3>
`, parts[0], parts[1]))

			for _, a := range anoms {
				class := "anomaly"
				if a.Severity == "critical" {
					class = "anomaly critical"
				} else if a.Severity == "high" {
					class = "anomaly high"
				}

				buf.WriteString(fmt.Sprintf(`		<div class="%s">
			<strong>%s</strong> (Severity: %s, Z-score: %.2f)<br>
			Performance: %.0f ns (%.2f%% deviation)
`, class, a.Timestamp.Format("2006-01-02 15:04"), a.Severity, a.ZScore, a.Value, a.ZScore*100/3))

				if a.IsRegression {
					buf.WriteString(`			<br><em>⚠️ Regression detected</em>
`)
				}

				buf.WriteString(`		</div>
`)
			}
		}
	}

	// Legend
	buf.WriteString(`		<h2>Legend</h2>
		<div class="legend">
			<p><strong>Direction:</strong> 🟢 Improving (performance getting faster), 🔴 Degrading (performance getting slower), → Stable (minimal change)</p>
			<p><strong>Change:</strong> Percentage change from first to last measurement</p>
			<p><strong>Slope:</strong> Change per day in nanoseconds</p>
			<p><strong>R²:</strong> Trend confidence level (0-1, higher = more reliable trend)</p>
			<p><strong>Data Points:</strong> Number of measurements used in analysis</p>
		</div>
	</div>
</body>
</html>
`)

	return buf.String(), nil
}

// GenerateTrendJSON generates a JSON trend report
func (btr *BasicTrendReporter) GenerateTrendJSON(trends []*analyzer.TrendResult, anomalies []*analyzer.Anomaly) (string, error) {
	// Convert trends to JSON-serializable format
	trendData := make([]map[string]interface{}, 0, len(trends))
	for _, t := range trends {
		trendData = append(trendData, map[string]interface{}{
			"benchmark_name":   t.BenchmarkName,
			"language":         t.Language,
			"direction":        t.Direction,
			"slope_ns_per_day": t.Slope,
			"r_squared":        t.RSquared,
			"change_percent":   t.ChangePercent,
			"period_days":      t.PeriodDays,
			"data_points":      t.DataPoints,
			"start_time":       t.StartTime.Format("2006-01-02T15:04:05Z"),
			"end_time":         t.EndTime.Format("2006-01-02T15:04:05Z"),
			"start_value_ns":   t.StartValue,
			"end_value_ns":     t.EndValue,
		})
	}

	// Convert anomalies to JSON-serializable format
	anomalyData := make([]map[string]interface{}, 0, len(anomalies))
	for _, a := range anomalies {
		anomalyData = append(anomalyData, map[string]interface{}{
			"benchmark_name": a.BenchmarkName,
			"language":       a.Language,
			"timestamp":      a.Timestamp.Format("2006-01-02T15:04:05Z"),
			"value_ns":       a.Value,
			"z_score":        a.ZScore,
			"severity":       a.Severity,
			"message":        a.Message,
			"is_regression":  a.IsRegression,
		})
	}

	// Summary
	improving := 0
	degrading := 0
	stable := 0
	for _, t := range trends {
		switch t.Direction {
		case "improving":
			improving++
		case "degrading":
			degrading++
		case "stable":
			stable++
		}
	}

	data := map[string]interface{}{
		"summary": map[string]interface{}{
			"total_benchmarks": len(trends),
			"improving":        improving,
			"degrading":        degrading,
			"stable":           stable,
			"anomalies_count":  len(anomalies),
		},
		"trends":    trendData,
		"anomalies": anomalyData,
	}

	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
