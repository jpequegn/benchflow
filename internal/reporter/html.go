package reporter

import (
	"embed"
	"fmt"
	"html/template"
	"io"
	"strings"
	"time"

	"github.com/jpequegn/benchflow/internal/aggregator"
)

//go:embed templates/*
var templateFS embed.FS

// HTMLReporter generates HTML reports with embedded CSS and JavaScript
type HTMLReporter struct {
	templates *template.Template
}

// NewHTMLReporter creates a new HTML reporter
func NewHTMLReporter() (*HTMLReporter, error) {
	// Parse all templates
	tmpl, err := template.New("").Funcs(templateFuncs()).ParseFS(templateFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("failed to parse templates: %w", err)
	}

	return &HTMLReporter{
		templates: tmpl,
	}, nil
}

// GenerateSummary generates an HTML summary report
func (r *HTMLReporter) GenerateSummary(suite *aggregator.AggregatedSuite, opts *ReportOptions, writer io.Writer) error {
	if suite == nil {
		return fmt.Errorf("suite cannot be nil")
	}

	if opts == nil {
		opts = &ReportOptions{
			Title:       "Benchmark Report",
			DarkMode:    true,
			ShowCharts:  true,
			ShowDetails: true,
		}
	}

	// Prepare chart data
	chartData := r.prepareSummaryChartData(suite)

	// Prepare template data
	data := &TemplateData{
		Title:       opts.Title,
		Suite:       suite,
		DarkMode:    opts.DarkMode,
		ShowCharts:  opts.ShowCharts,
		ShowDetails: opts.ShowDetails,
		ChartData:   chartData,
	}

	// Execute template
	if err := r.templates.ExecuteTemplate(writer, "summary.html", data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// GenerateComparison generates an HTML comparison report
func (r *HTMLReporter) GenerateComparison(comparison *aggregator.ComparisonSuite, opts *ReportOptions, writer io.Writer) error {
	if comparison == nil {
		return fmt.Errorf("comparison cannot be nil")
	}

	if opts == nil {
		opts = &ReportOptions{
			Title:       "Benchmark Comparison",
			DarkMode:    true,
			ShowCharts:  true,
			ShowDetails: true,
		}
	}

	// Prepare chart data
	chartData := r.prepareComparisonChartData(comparison)

	// Prepare template data
	data := &TemplateData{
		Title:       opts.Title,
		Comparison:  comparison,
		DarkMode:    opts.DarkMode,
		ShowCharts:  opts.ShowCharts,
		ShowDetails: opts.ShowDetails,
		ChartData:   chartData,
	}

	// Execute template
	if err := r.templates.ExecuteTemplate(writer, "comparison.html", data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// GenerateTrend generates an HTML trend report
func (r *HTMLReporter) GenerateTrend(history []*aggregator.AggregatedResult, opts *ReportOptions, writer io.Writer) error {
	if len(history) == 0 {
		return fmt.Errorf("history cannot be empty")
	}

	if opts == nil {
		opts = &ReportOptions{
			Title:       "Benchmark Trends",
			DarkMode:    true,
			ShowCharts:  true,
			ShowDetails: true,
		}
	}

	// Prepare chart data
	chartData := r.prepareTrendChartData(history)

	// Prepare template data
	data := &TemplateData{
		Title:       opts.Title,
		History:     history,
		DarkMode:    opts.DarkMode,
		ShowCharts:  opts.ShowCharts,
		ShowDetails: opts.ShowDetails,
		ChartData:   chartData,
	}

	// Execute template
	if err := r.templates.ExecuteTemplate(writer, "trend.html", data); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// prepareSummaryChartData prepares chart data for summary reports
func (r *HTMLReporter) prepareSummaryChartData(suite *aggregator.AggregatedSuite) *ChartData {
	labels := make([]string, 0, len(suite.Results))
	data := make([]float64, 0, len(suite.Results))

	for _, result := range suite.Results {
		labels = append(labels, result.Name)
		// Convert to milliseconds for better readability
		data = append(data, float64(result.Mean.Nanoseconds())/1_000_000.0)
	}

	return &ChartData{
		Labels:     labels,
		ChartType:  "bar",
		ChartTitle: "Benchmark Results",
		YAxisLabel: "Time (ms)",
		XAxisLabel: "Benchmark",
		Datasets: []ChartDataset{
			{
				Label:           "Mean Time",
				Data:            data,
				BackgroundColor: "#1F4E8C",
				BorderColor:     "#2762B3",
				BorderWidth:     1,
			},
		},
	}
}

// prepareComparisonChartData prepares chart data for comparison reports
func (r *HTMLReporter) prepareComparisonChartData(comparison *aggregator.ComparisonSuite) *ChartData {
	labels := make([]string, 0, len(comparison.Comparisons))
	baselineData := make([]float64, 0, len(comparison.Comparisons))
	currentData := make([]float64, 0, len(comparison.Comparisons))

	for _, comp := range comparison.Comparisons {
		labels = append(labels, comp.Name)
		baselineData = append(baselineData, float64(comp.Baseline.Mean.Nanoseconds())/1_000_000.0)
		currentData = append(currentData, float64(comp.Current.Mean.Nanoseconds())/1_000_000.0)
	}

	return &ChartData{
		Labels:     labels,
		ChartType:  "bar",
		ChartTitle: "Baseline vs Current",
		YAxisLabel: "Time (ms)",
		XAxisLabel: "Benchmark",
		Datasets: []ChartDataset{
			{
				Label:           "Baseline",
				Data:            baselineData,
				BackgroundColor: "#28A745",
				BorderColor:     "#1F8435",
				BorderWidth:     1,
			},
			{
				Label:           "Current",
				Data:            currentData,
				BackgroundColor: "#1F4E8C",
				BorderColor:     "#2762B3",
				BorderWidth:     1,
			},
		},
	}
}

// prepareTrendChartData prepares chart data for trend reports
func (r *HTMLReporter) prepareTrendChartData(history []*aggregator.AggregatedResult) *ChartData {
	labels := make([]string, 0, len(history))
	data := make([]float64, 0, len(history))

	// Reverse history so oldest is first (left to right on chart)
	for i := len(history) - 1; i >= 0; i-- {
		result := history[i]
		labels = append(labels, result.Timestamp.Format("Jan 2 15:04"))
		data = append(data, float64(result.Mean.Nanoseconds())/1_000_000.0)
	}

	return &ChartData{
		Labels:     labels,
		ChartType:  "line",
		ChartTitle: "Performance Trend",
		YAxisLabel: "Time (ms)",
		XAxisLabel: "Date",
		Datasets: []ChartDataset{
			{
				Label:           "Mean Time",
				Data:            data,
				BackgroundColor: "rgba(31, 78, 140, 0.2)",
				BorderColor:     "#1F4E8C",
				BorderWidth:     2,
			},
		},
	}
}

// templateFuncs returns custom template functions
func templateFuncs() template.FuncMap {
	return template.FuncMap{
		"sub": func(a, b int) int {
			return a - b
		},
		"formatDuration": func(d time.Duration) string {
			if d < time.Microsecond {
				return fmt.Sprintf("%d ns", d.Nanoseconds())
			} else if d < time.Millisecond {
				return fmt.Sprintf("%.2f μs", float64(d.Nanoseconds())/1000.0)
			} else if d < time.Second {
				return fmt.Sprintf("%.2f ms", float64(d.Nanoseconds())/1_000_000.0)
			}
			return fmt.Sprintf("%.2f s", d.Seconds())
		},
		"formatPercent": func(f float64) string {
			return fmt.Sprintf("%.2f%%", f)
		},
		"formatTimestamp": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
		"plusSign": func(d time.Duration) string {
			if d > 0 {
				return "+"
			}
			return ""
		},
		"regressionClass": func(comp *aggregator.Comparison) string {
			if comp.Regression {
				return "regression"
			} else if comp.Improvement {
				return "improvement"
			}
			return "unchanged"
		},
		"statusIcon": func(comp *aggregator.Comparison) string {
			if comp.Regression {
				return "⚠️"
			} else if comp.Improvement {
				return "✅"
			}
			return "➖"
		},
		"toJSON": func(v interface{}) string {
			// Simple JSON serialization for chart data
			switch val := v.(type) {
			case []string:
				quoted := make([]string, len(val))
				for i, s := range val {
					quoted[i] = fmt.Sprintf(`"%s"`, s)
				}
				return "[" + strings.Join(quoted, ",") + "]"
			case []float64:
				strs := make([]string, len(val))
				for i, f := range val {
					strs[i] = fmt.Sprintf("%.6f", f)
				}
				return "[" + strings.Join(strs, ",") + "]"
			default:
				return "[]"
			}
		},
	}
}
