# Comparator API Reference

Complete API documentation for benchflow's comparative analysis engine.

## Table of Contents

- [Comparator Interface](#comparator-interface)
- [BasicComparator](#basiccomparator)
- [Data Structures](#data-structures)
- [Error Handling](#error-handling)
- [Usage Examples](#usage-examples)

## Comparator Interface

```go
type Comparator interface {
    // Compare compares a baseline suite against a current suite
    Compare(baseline, current *BenchmarkSuite) *ComparisonResult

    // GetSignificance determines if the difference between two results is statistically significant
    GetSignificance(baseline, current *BenchmarkResult, confidenceLevel float64) (significant bool, pValue float64)

    // CalculateConfidenceInterval calculates the confidence interval for benchmark results
    CalculateConfidenceInterval(results []*BenchmarkResult, confidenceLevel float64) (lower, upper float64)
}
```

### Compare Method

Compares a baseline BenchmarkSuite against a current BenchmarkSuite.

**Signature:**
```go
func (bc *BasicComparator) Compare(baseline, current *BenchmarkSuite) *ComparisonResult
```

**Parameters:**
- `baseline`: Reference benchmark results (usually from stable branch/release)
- `current`: New benchmark results to compare

**Returns:**
- `*ComparisonResult`: Complete comparison analysis

**Behavior:**
- Matches benchmarks by name and language
- Skips benchmarks without baseline match
- Skips benchmarks from different languages
- Calculates regressions, improvements, and statistical significance
- Returns empty result if inputs are nil

**Example:**
```go
comp := comparator.NewBasicComparator()
result := comp.Compare(baselineSuite, currentSuite)
if result.Summary.Regressions > 0 {
    // Handle regressions
}
```

### GetSignificance Method

Determines if the difference between two benchmark results is statistically significant.

**Signature:**
```go
func (bc *BasicComparator) GetSignificance(baseline, current *BenchmarkResult, confidenceLevel float64) (bool, float64)
```

**Parameters:**
- `baseline`: Reference benchmark result
- `current`: Current benchmark result
- `confidenceLevel`: Statistical confidence level (e.g., 0.95 for 95%)

**Returns:**
- `bool`: Whether the change is statistically significant
- `float64`: P-value from statistical test

**Confidence Levels:**
- 0.95 (95%): α = 0.05, p-value threshold
- 0.99 (99%): α = 0.01, more stringent

**Example:**
```go
significant, pValue := comp.GetSignificance(baseline, current, 0.95)
if significant {
    fmt.Printf("Significant change detected (p=%.3f)\n", pValue)
}
```

### CalculateConfidenceInterval Method

Calculates confidence interval for a set of benchmark results.

**Signature:**
```go
func (bc *BasicComparator) CalculateConfidenceInterval(results []*BenchmarkResult, confidenceLevel float64) (lower, upper float64)
```

**Parameters:**
- `results`: Array of benchmark results
- `confidenceLevel`: Confidence level (0.95 = 95%)

**Returns:**
- `lower`: Lower bound of confidence interval (nanoseconds)
- `upper`: Upper bound of confidence interval (nanoseconds)

**Example:**
```go
lower, upper := comp.CalculateConfidenceInterval(results, 0.95)
fmt.Printf("95%% confidence interval: [%d ns, %d ns]\n", lower, upper)
```

## BasicComparator

Default implementation of the Comparator interface.

### Constructor

```go
func NewBasicComparator() *BasicComparator
```

Creates a new BasicComparator with default settings:
- `ConfidenceLevel`: 0.95 (95%)
- `RegressionThreshold`: 1.05 (5% slower)

**Example:**
```go
comp := comparator.NewBasicComparator()
comp.RegressionThreshold = 1.02  // 2% sensitivity
comp.ConfidenceLevel = 0.99      // 99% confidence
result := comp.Compare(baseline, current)
```

### Configuration Fields

```go
type BasicComparator struct {
    ConfidenceLevel     float64  // Confidence level for statistics (default: 0.95)
    RegressionThreshold float64  // Multiplier for regression detection (default: 1.05)
}
```

**ConfidenceLevel:**
- Range: 0 < level < 1
- Default: 0.95 (95%)
- Higher values = more stringent significance requirement
- Typical values: 0.90, 0.95, 0.99

**RegressionThreshold:**
- Range: > 1.0
- Default: 1.05 (5%)
- 1.01 = 1% slower is regression
- 1.10 = 10% slower is regression
- Lower values = more sensitive to regressions

## Data Structures

### ComparisonResult

Complete result of a benchmark comparison.

```go
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
```

**Usage:**
```go
result := comp.Compare(baseline, current)
fmt.Printf("Total: %d, Regressions: %d, Improvements: %d\n",
    result.Summary.TotalComparisons,
    result.Summary.Regressions,
    result.Summary.Improvements)

for _, comp := range result.Benchmarks {
    fmt.Printf("%s: %.2f%% %s\n", comp.Name, comp.TimeDelta, comp.Status())
}
```

### BenchmarkComparison

Comparison data for a single benchmark.

```go
type BenchmarkComparison struct {
    // Name is the benchmark name
    Name string

    // Language is the programming language
    Language string

    // Baseline is the baseline benchmark result
    Baseline *BenchmarkResult

    // Current is the current benchmark result
    Current *BenchmarkResult

    // TimeDelta is the time change in percentage
    // Negative = faster, positive = slower
    TimeDelta float64

    // IsRegression indicates if this is a performance regression
    IsRegression bool

    // IsSignificant indicates if the difference is statistically significant
    IsSignificant bool

    // ConfidenceLevel is the confidence level used
    ConfidenceLevel float64

    // TTestPValue is the p-value from the t-test
    TTestPValue float64

    // EffectSize is Cohen's d effect size
    EffectSize float64

    // RegressionThreshold is the threshold for regression detection
    RegressionThreshold float64
}
```

**Fields:**
- `TimeDelta`: Percentage change (-5.0 = 5% faster, +10.0 = 10% slower)
- `IsRegression`: True if slower than threshold
- `IsSignificant`: True if p-value < α (where α = 1 - confidence level)
- `TTestPValue`: Statistical test p-value (0.05 typical threshold)
- `EffectSize`: Cohen's d (0.8 = large effect)

### ComparisonSummary

Aggregate statistics from all comparisons.

```go
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
```

**Interpretation:**
```
TotalComparisons:  Number of benchmarks compared
Regressions:       Count of regressions (performance got worse)
Improvements:      Count of improvements (performance got better)
AverageDelta:      Average % change across all benchmarks
MaxDelta:          Largest % change (could be improvement or regression)
MinDelta:          Smallest % change (could be improvement or regression)
SignificantChanges: Count where p-value < α (statistically reliable)
```

### ComparisonStats

Statistical information about the comparison.

```go
type ComparisonStats struct {
    // ConfidenceLevel is the confidence level used
    ConfidenceLevel float64

    // SignificanceLevel is 1 - ConfidenceLevel
    SignificanceLevel float64

    // RegressionThreshold is the threshold for regression detection
    RegressionThreshold float64
}
```

**Relationship:**
```
ConfidenceLevel = 0.95
SignificanceLevel = 1 - 0.95 = 0.05 (5%)
P-value threshold = 0.05 for significance
```

## Error Handling

The comparator handles various error conditions gracefully:

### Nil Inputs

```go
// Returns empty result, no error
result := comp.Compare(nil, current)  // Benchmarks will be empty
result := comp.Compare(baseline, nil) // Benchmarks will be empty
```

### Empty Results

```go
// Returns empty result, no error
baseline := &BenchmarkSuite{Results: []*BenchmarkResult{}}
result := comp.Compare(baseline, current) // No comparisons
```

### Invalid Confidence Levels

The comparator assumes valid confidence levels (0 < level < 1). Validation should occur at CLI/API boundary:

```go
if confidence <= 0 || confidence >= 1 {
    return fmt.Errorf("confidence must be between 0 and 1")
}
```

### Zero/Infinite Values

Statistical calculations handle edge cases:

```go
// Time = 0: Returns TimeDelta = 0
// StdDev = 0: Assumes 5% variance for estimate
// Identical values: GetSignificance returns (false, 1.0)
```

## Usage Examples

### Basic Comparison

```go
package main

import (
    "fmt"
    "github.com/jpequegn/benchflow/internal/comparator"
    "github.com/jpequegn/benchflow/internal/parser"
    "time"
)

func main() {
    // Create baseline suite
    baseline := &parser.BenchmarkSuite{
        Language: "go",
        Results: []*parser.BenchmarkResult{
            {
                Name:       "sort",
                Language:   "go",
                Time:       1000 * time.Nanosecond,
                StdDev:     50 * time.Nanosecond,
                Iterations: 100,
            },
        },
    }

    // Create current suite
    current := &parser.BenchmarkSuite{
        Language: "go",
        Results: []*parser.BenchmarkResult{
            {
                Name:       "sort",
                Language:   "go",
                Time:       950 * time.Nanosecond,
                StdDev:     45 * time.Nanosecond,
                Iterations: 100,
            },
        },
    }

    // Compare
    comp := comparator.NewBasicComparator()
    result := comp.Compare(baseline, current)

    // Print summary
    fmt.Printf("Comparisons: %d\n", result.Summary.TotalComparisons)
    fmt.Printf("Regressions: %d\n", result.Summary.Regressions)
    fmt.Printf("Improvements: %d\n", result.Summary.Improvements)
    fmt.Printf("Average Delta: %.2f%%\n", result.Summary.AverageDelta)
}
```

### Custom Configuration

```go
comp := comparator.NewBasicComparator()
comp.RegressionThreshold = 1.02  // 2% threshold (sensitive)
comp.ConfidenceLevel = 0.99      // 99% confidence (stringent)

result := comp.Compare(baseline, current)
if result.Summary.Regressions > 0 {
    fmt.Println("Performance regressions detected!")
}
```

### Statistical Analysis

```go
comp := comparator.NewBasicComparator()

// Check significance of a single change
baseline := baselineResults[0]
current := currentResults[0]

significant, pValue := comp.GetSignificance(baseline, current, 0.95)
fmt.Printf("Change is significant: %v (p=%.4f)\n", significant, pValue)

// Calculate confidence interval
interval := comp.CalculateConfidenceInterval(allResults, 0.95)
fmt.Printf("95%% CI: [%.0f ns, %.0f ns]\n", interval[0], interval[1])
```

### Batch Processing

```go
comp := comparator.NewBasicComparator()
comp.RegressionThreshold = 1.05

for _, baseline := range baselines {
    current := currentMap[baseline.Name]
    if current == nil {
        continue
    }

    suite := &parser.BenchmarkSuite{Results: []*parser.BenchmarkResult{baseline}}
    currentSuite := &parser.BenchmarkSuite{Results: []*parser.BenchmarkResult{current}}

    result := comp.Compare(suite, currentSuite)
    if result.Summary.Regressions > 0 {
        fmt.Printf("Regression in %s\n", baseline.Name)
    }
}
```

## Performance Characteristics

- **Time Complexity**: O(n log n) for sorting in summary calculation
- **Space Complexity**: O(n) for storing all comparisons
- **Typical Latency**: <100ms for 1000 benchmarks
- **Memory Usage**: ~100 bytes per comparison

## Thread Safety

The BasicComparator is thread-safe for read operations. For concurrent comparisons, create separate instances:

```go
// ✅ Safe: Each goroutine has its own comparator
for _, item := range items {
    go func(it Item) {
        comp := comparator.NewBasicComparator()
        result := comp.Compare(it.baseline, it.current)
        // process result
    }(item)
}

// ❌ Not safe: Shared comparator state
var comp = comparator.NewBasicComparator()
for _, item := range items {
    go func(it Item) {
        result := comp.Compare(it.baseline, it.current)
        // May have race conditions
    }(item)
}
```

## See Also

- [Comparative Analysis Guide](COMPARISON.md)
- [Statistical Concepts](STATISTICS.md)
- [GitHub Actions Integration](CI_CD_INTEGRATION.md)
