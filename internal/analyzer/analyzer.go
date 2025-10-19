package analyzer

import (
	"fmt"
	"math"
	"sort"
)

// CalculateTrend calculates linear regression trend from historical data
func (bta *BasicTrendAnalyzer) CalculateTrend(history []*HistoricalComparison, minDataPoints int) (*TrendResult, error) {
	if len(history) < minDataPoints {
		return nil, fmt.Errorf("insufficient data points: %d < %d", len(history), minDataPoints)
	}

	if len(history) == 0 {
		return nil, fmt.Errorf("no historical data")
	}

	// Sort by timestamp
	sorted := make([]*HistoricalComparison, len(history))
	copy(sorted, history)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].CreatedAt.Before(sorted[j].CreatedAt)
	})

	// Extract values (using CurrentTimeNs for trend)
	n := float64(len(sorted))
	var sumX, sumY, sumXY, sumX2, sumY2 float64
	times := make([]float64, len(sorted))

	startTime := sorted[0].CreatedAt
	for i, comp := range sorted {
		// X = days since start
		x := float64(comp.CreatedAt.Sub(startTime).Hours() / 24)
		y := float64(comp.CurrentTimeNs)

		times[i] = x
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
		sumY2 += y * y
	}

	// Calculate linear regression
	denominator := n*sumX2 - sumX*sumX
	if math.Abs(denominator) < 1e-10 {
		return nil, fmt.Errorf("cannot calculate trend: no variance in x")
	}

	slope := (n*sumXY - sumX*sumY) / denominator
	intercept := (sumY - slope*sumX) / n

	// Calculate R-squared
	ssRes := 0.0
	ssTot := 0.0
	meanY := sumY / n

	for _, comp := range sorted {
		predicted := intercept + slope*float64(comp.CreatedAt.Sub(startTime).Hours()/24)
		actual := float64(comp.CurrentTimeNs)
		ssRes += math.Pow(actual-predicted, 2)
		ssTot += math.Pow(actual-meanY, 2)
	}

	rSquared := 1.0
	if ssTot > 0 {
		rSquared = 1.0 - (ssRes / ssTot)
	}

	// Clamp to [0, 1]
	if rSquared < 0 {
		rSquared = 0
	}
	if rSquared > 1 {
		rSquared = 1
	}

	// Determine direction
	direction := "stable"
	absSlope := math.Abs(slope)
	if absSlope > 1.0 { // > 1 ns/day change
		if slope > 0 {
			direction = "degrading"
		} else {
			direction = "improving"
		}
	}

	// Calculate period
	endTime := sorted[len(sorted)-1].CreatedAt
	periodDays := int(endTime.Sub(startTime).Hours() / 24)
	if periodDays == 0 {
		periodDays = 1
	}

	// Calculate overall change
	startValue := float64(sorted[0].CurrentTimeNs)
	endValue := float64(sorted[len(sorted)-1].CurrentTimeNs)
	changePercent := 0.0
	if startValue > 0 {
		changePercent = ((endValue - startValue) / startValue) * 100
	}

	return &TrendResult{
		BenchmarkName: sorted[0].BenchmarkName,
		Language:      sorted[0].Language,
		Direction:     direction,
		Slope:         slope,
		RSquared:      rSquared,
		ChangePercent: changePercent,
		PeriodDays:    periodDays,
		DataPoints:    len(sorted),
		StartTime:     startTime,
		EndTime:       endTime,
		StartValue:    startValue,
		EndValue:      endValue,
	}, nil
}

// DetectAnomalies detects statistical anomalies in performance data
func (bta *BasicTrendAnalyzer) DetectAnomalies(history []*HistoricalComparison, zScoreThreshold float64) []*Anomaly {
	if len(history) < 2 {
		return nil
	}

	// Sort by timestamp
	sorted := make([]*HistoricalComparison, len(history))
	copy(sorted, history)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].CreatedAt.Before(sorted[j].CreatedAt)
	})

	// Calculate statistics
	values := make([]float64, len(sorted))
	for i, comp := range sorted {
		values[i] = float64(comp.CurrentTimeNs)
	}

	mean := calculateMean(values)
	stdDev := calculateStdDev(values, mean)

	if stdDev == 0 {
		return nil // No variance, can't detect anomalies
	}

	// Detect anomalies
	var anomalies []*Anomaly
	for i, comp := range sorted {
		value := float64(comp.CurrentTimeNs)
		zScore := (value - mean) / stdDev

		if math.Abs(zScore) > zScoreThreshold {
			severity := "medium"
			if math.Abs(zScore) > 3.0 {
				severity = "critical"
			} else if math.Abs(zScore) > 2.5 {
				severity = "high"
			} else if math.Abs(zScore) > 1.5 {
				severity = "medium"
			} else {
				severity = "low"
			}

			message := fmt.Sprintf("Anomaly detected: %.2f%% deviation from mean", math.Abs(zScore)*100/3)

			anomalies = append(anomalies, &Anomaly{
				BenchmarkName: comp.BenchmarkName,
				Language:      comp.Language,
				Timestamp:     comp.CreatedAt,
				Value:         value,
				ZScore:        zScore,
				Severity:      severity,
				Message:       message,
				IsRegression:  comp.IsRegression,
			})

			// For early anomaly detection: check if this is a regression
			if i > 0 {
				prevValue := float64(sorted[i-1].CurrentTimeNs)
				if value > prevValue*1.05 {
					anomalies[len(anomalies)-1].IsRegression = true
				}
			}
		}
	}

	return anomalies
}

// ForecastPerformance forecasts future performance using linear extrapolation
func (bta *BasicTrendAnalyzer) ForecastPerformance(history []*HistoricalComparison, periods int) []*Forecast {
	if len(history) < 2 || periods <= 0 {
		return nil
	}

	// Sort by timestamp
	sorted := make([]*HistoricalComparison, len(history))
	copy(sorted, history)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].CreatedAt.Before(sorted[j].CreatedAt)
	})

	// Group by benchmark
	benchmarks := make(map[string][]*HistoricalComparison)
	for _, comp := range sorted {
		key := comp.BenchmarkName + ":" + comp.Language
		benchmarks[key] = append(benchmarks[key], comp)
	}

	var forecasts []*Forecast

	for _, comps := range benchmarks {
		if len(comps) < 2 {
			continue
		}

		// Calculate trend
		trend, err := bta.CalculateTrend(comps, 2)
		if err != nil {
			continue
		}

		// Calculate prediction standard error
		stdErr := calculateForecastStdErr(comps)

		// Generate forecasts
		for p := 1; p <= periods; p++ {
			predictedDays := float64(p)
			predictedTime := trend.EndValue + trend.Slope*predictedDays

			// Confidence interval (approximated)
			marginOfError := 1.96 * stdErr * math.Sqrt(1+1/float64(len(comps)))

			forecast := &Forecast{
				BenchmarkName: trend.BenchmarkName,
				Language:      trend.Language,
				Period:        p,
				PredictedTime: predictedTime,
				LowerBound:    predictedTime - marginOfError,
				UpperBound:    predictedTime + marginOfError,
				Confidence:    bta.ConfidenceLevel,
			}

			// Ensure bounds don't go negative
			if forecast.LowerBound < 0 {
				forecast.LowerBound = 0
			}

			forecasts = append(forecasts, forecast)
		}
	}

	return forecasts
}

// Helper functions

func calculateMean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

func calculateStdDev(values []float64, mean float64) float64 {
	if len(values) <= 1 {
		return 0
	}
	varianceSum := 0.0
	for _, v := range values {
		diff := v - mean
		varianceSum += diff * diff
	}
	variance := varianceSum / float64(len(values)-1)
	return math.Sqrt(variance)
}

func calculateForecastStdErr(history []*HistoricalComparison) float64 {
	if len(history) < 2 {
		return 0
	}

	// Calculate residual standard error from linear regression
	values := make([]float64, len(history))
	for i, comp := range history {
		values[i] = float64(comp.CurrentTimeNs)
	}

	mean := calculateMean(values)
	ssRes := 0.0

	for _, v := range values {
		diff := v - mean
		ssRes += diff * diff
	}

	mse := ssRes / float64(len(values)-1)
	return math.Sqrt(mse)
}
