package analyzer

import "time"

// HistoricalComparison represents a stored comparison from history
type HistoricalComparison struct {
	ID               int64
	BenchmarkName    string
	Language         string
	BaselineTimeNs   int64
	CurrentTimeNs    int64
	TimeDeltaPercent float64
	IsRegression     bool
	CommitHash       string
	BranchName       string
	Author           string
	CreatedAt        time.Time
}

// TrendResult represents the result of trend analysis
type TrendResult struct {
	BenchmarkName string
	Language      string
	Direction     string    // "improving", "degrading", "stable"
	Slope         float64   // ns/commit (change per commit)
	RSquared      float64   // Trend confidence (0-1)
	ChangePercent float64   // % change over period
	PeriodDays    int       // Days covered
	DataPoints    int       // Number of measurements
	StartTime     time.Time // First measurement
	EndTime       time.Time // Last measurement
	StartValue    float64   // First measurement value
	EndValue      float64   // Last measurement value
}

// Anomaly represents a detected anomaly in performance data
type Anomaly struct {
	BenchmarkName string
	Language      string
	Timestamp     time.Time
	Value         float64 // Performance value (ns)
	ZScore        float64 // Standard deviation score
	Severity      string  // "critical", "high", "medium", "low"
	Message       string
	IsRegression  bool
}

// Forecast represents a performance forecast
type Forecast struct {
	BenchmarkName string
	Language      string
	Period        int     // Number of periods ahead
	PredictedTime float64 // Predicted time (ns)
	LowerBound    float64 // 95% confidence lower
	UpperBound    float64 // 95% confidence upper
	Confidence    float64 // Forecast confidence (0-1)
}

// TrendAnalyzer defines the interface for trend analysis
type TrendAnalyzer interface {
	// CalculateTrend calculates trend from historical comparisons
	CalculateTrend(history []*HistoricalComparison, minDataPoints int) (*TrendResult, error)

	// DetectAnomalies detects performance anomalies
	DetectAnomalies(history []*HistoricalComparison, zScoreThreshold float64) []*Anomaly

	// ForecastPerformance forecasts future performance
	ForecastPerformance(history []*HistoricalComparison, periods int) []*Forecast
}

// BasicTrendAnalyzer implements TrendAnalyzer
type BasicTrendAnalyzer struct {
	// Configuration
	MinDataPoints   int     // Minimum data points for trend (default: 3)
	ZScoreThreshold float64 // Z-score threshold for anomalies (default: 2.0)
	ConfidenceLevel float64 // Forecast confidence (default: 0.95)
}

// NewBasicTrendAnalyzer creates a new trend analyzer
func NewBasicTrendAnalyzer() *BasicTrendAnalyzer {
	return &BasicTrendAnalyzer{
		MinDataPoints:   3,
		ZScoreThreshold: 2.0,
		ConfidenceLevel: 0.95,
	}
}
