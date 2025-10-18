package reporter

import (
	"io"

	"github.com/jpequegn/benchflow/internal/aggregator"
)

// ReportFormat represents the output format for reports
type ReportFormat string

const (
	FormatHTML ReportFormat = "html"
	FormatJSON ReportFormat = "json"
	FormatCSV  ReportFormat = "csv"
)

// ReportType represents the type of report to generate
type ReportType string

const (
	TypeSummary    ReportType = "summary"    // Single suite summary
	TypeComparison ReportType = "comparison" // Baseline vs current
	TypeTrend      ReportType = "trend"      // Historical trends
)

// ReportOptions configures report generation
type ReportOptions struct {
	Title       string       // Report title
	Format      ReportFormat // Output format
	Type        ReportType   // Report type
	DarkMode    bool         // Enable dark mode theme
	ShowCharts  bool         // Include charts (HTML only)
	ShowDetails bool         // Include detailed results
}

// Reporter defines the interface for report generation
type Reporter interface {
	// GenerateSummary generates a report for a single benchmark suite
	GenerateSummary(suite *aggregator.AggregatedSuite, opts *ReportOptions, writer io.Writer) error

	// GenerateComparison generates a comparison report
	GenerateComparison(comparison *aggregator.ComparisonSuite, opts *ReportOptions, writer io.Writer) error

	// GenerateTrend generates a trend report from historical data
	GenerateTrend(history []*aggregator.AggregatedResult, opts *ReportOptions, writer io.Writer) error
}

// TemplateData represents data passed to HTML templates
type TemplateData struct {
	Title       string
	Suite       *aggregator.AggregatedSuite
	Comparison  *aggregator.ComparisonSuite
	History     []*aggregator.AggregatedResult
	DarkMode    bool
	ShowCharts  bool
	ShowDetails bool
	ChartData   *ChartData
}

// ChartData represents data for Chart.js visualizations
type ChartData struct {
	Labels     []string       // X-axis labels
	Datasets   []ChartDataset // Chart datasets
	ChartType  string         // Chart type (bar, line, etc.)
	ChartTitle string         // Chart title
	YAxisLabel string         // Y-axis label
	XAxisLabel string         // X-axis label
}

// ChartDataset represents a single dataset for charts
type ChartDataset struct {
	Label           string    // Dataset label
	Data            []float64 // Data values
	BackgroundColor string    // Bar/point color
	BorderColor     string    // Line/border color
	BorderWidth     int       // Border width
}
