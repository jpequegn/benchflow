# Comparative Analysis Guide

This guide covers benchflow's comparative analysis capabilities for detecting performance regressions, improvements, and statistically significant changes across benchmark runs.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Understanding Statistical Analysis](#understanding-statistical-analysis)
- [Report Interpretation](#report-interpretation)
- [CI/CD Integration](#cicd-integration)
- [Troubleshooting](#troubleshooting)

## Overview

Benchflow's comparative analysis engine enables teams to:
- **Detect Performance Regressions**: Automatically identify benchmarks that are slower than baseline
- **Track Improvements**: Celebrate performance improvements with statistical confidence
- **Statistical Significance**: Use t-tests and confidence intervals to distinguish real changes from noise
- **Cross-Language Comparisons**: Compare performance across multiple languages
- **Configurable Thresholds**: Set regression detection sensitivity to match your requirements

### Key Concepts

**Baseline**: The reference benchmark results (usually from main branch or previous release)
**Current**: The new benchmark results to compare against baseline
**Regression**: Performance degradation beyond the configured threshold
**Improvement**: Performance improvement detected with statistical significance
**Significant Change**: A performance change that is statistically significant at the confidence level

## Quick Start

### Basic Comparison

Compare two benchmark result files:

```bash
benchflow compare --baseline main.json --current feature-branch.json
```

### Output Formats

Generate reports in different formats:

```bash
# Markdown (default, suitable for PR comments)
benchflow compare -b main.json -c current.json

# HTML (for detailed visual inspection)
benchflow compare -b main.json -c current.json -f html -o report.html

# JSON (for programmatic analysis)
benchflow compare -b main.json -c current.json -f json
```

### Configure Regression Detection

Adjust regression detection sensitivity:

```bash
# Strict: Flag any regression >1% (sensitive to small regressions)
benchflow compare -b main.json -c current.json --threshold 1.01

# Moderate: Flag regressions >5% (default)
benchflow compare -b main.json -c current.json --threshold 1.05

# Lenient: Flag regressions >10% (only significant regressions)
benchflow compare -b main.json -c current.json --threshold 1.10
```

### Confidence Levels

Adjust statistical confidence level:

```bash
# 95% confidence (default, p-value < 0.05)
benchflow compare -b main.json -c current.json --confidence 0.95

# 99% confidence (stricter, p-value < 0.01)
benchflow compare -b main.json -c current.json --confidence 0.99
```

## Understanding Statistical Analysis

### Why Statistical Testing?

Benchmark results naturally vary due to system noise, GC pauses, and other factors. Statistical tests help distinguish real performance changes from measurement noise.

### T-Tests and P-Values

The comparator uses t-tests to determine if a performance change is statistically significant:

- **P-Value**: Probability that the observed difference occurred by chance
- **Significance Threshold**: Typically 0.05 (5%)
  - P < 0.05 → Statistically significant change (likely real)
  - P ≥ 0.05 → Not statistically significant (could be noise)

**Example:**
- Baseline: 1000 ns ± 50 ns
- Current: 1050 ns ± 40 ns
- P-Value: 0.08
- Result: 5% slower, but NOT statistically significant at 95% confidence

### Confidence Intervals

The comparator calculates confidence intervals around performance measurements:

```
Baseline: 1000 ns with 95% confidence interval [950 ns, 1050 ns]
Current:  950 ns with 95% confidence interval [920 ns, 980 ns]
```

When confidence intervals don't overlap, changes are likely significant.

### Effect Size (Cohen's d)

Measures the magnitude of performance change:

- **d < 0.2**: Negligible effect
- **0.2 ≤ d < 0.5**: Small effect
- **0.5 ≤ d < 0.8**: Medium effect
- **d ≥ 0.8**: Large effect

**Example:**
```
Baseline vs Current comparison:
- Time Delta: -5% (improvement)
- P-Value: 0.02 (significant)
- Cohen's d: 0.75 (medium-to-large effect)
→ Real, meaningful improvement
```

## Report Interpretation

### Summary Section

```
═══════════════════════════════════════════
  Comparison Summary
═══════════════════════════════════════════
Total Comparisons: 10
Regressions:      2
Improvements:     5
Significant:      7
Average Delta:    -3.50%
Max Delta:        20.00% (regression)
Min Delta:        -15.00% (improvement)
═══════════════════════════════════════════
```

**Interpretation:**
- 10 benchmarks were compared
- 2 detected regressions, 5 improvements
- 7 changes are statistically significant
- Average 3.5% performance gain

### Markdown Report Table

```markdown
| Benchmark | Language | Baseline | Current | Delta | Status | P-Value | Effect Size |
|-----------|----------|----------|---------|-------|--------|---------|-------------|
| sort      | go       | 1000 ns  | 950 ns  | -5.00% | 🟢     | 0.0200  | 0.80        |
| search    | go       | 500 ns   | 600 ns  | 20.00% | 🔴     | 0.0100  | 1.20        |
```

**Column Meanings:**
- **Baseline/Current**: Original and new timing (nanoseconds)
- **Delta**: Percentage change (negative = faster, positive = slower)
- **Status**: 🟢 improvement, 🔴 regression, → no significant change
- **P-Value**: Statistical significance (lower = more significant)
- **Effect Size**: Cohen's d magnitude of change

### HTML Report

The HTML report provides interactive features:
- **Sorting**: Click column headers to sort
- **Color Coding**: Green (improvement), Red (regression), Gray (neutral)
- **Visual Indicators**: Quickly identify problematic benchmarks
- **Statistics Display**: Full p-values, effect sizes, and confidence data

### JSON Report

Programmatic access to all comparison data:

```json
{
  "summary": {
    "total_comparisons": 2,
    "regressions": 1,
    "improvements": 1,
    "average_delta": 7.5,
    "significant_changes": 2
  },
  "benchmarks": [
    {
      "name": "sort",
      "language": "go",
      "baseline_time_ns": 1000,
      "current_time_ns": 950,
      "time_delta_percent": -5.0,
      "is_regression": false,
      "t_test_p_value": 0.02,
      "effect_size_cohens_d": 0.8
    }
  ]
}
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Performance Tests

on: [pull_request]

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Checkout main branch
        run: git fetch origin main

      - name: Run benchmarks (main)
        run: benchflow run --config benchflow.yaml --output baseline.json

      - name: Run benchmarks (current)
        run: benchflow run --config benchflow.yaml --output current.json

      - name: Compare results
        run: benchflow compare -b baseline.json -c current.json -f markdown | tee comparison.txt

      - name: Comment on PR
        uses: actions/github-script@v7
        with:
          script: |
            const fs = require('fs');
            const report = fs.readFileSync('comparison.txt', 'utf8');
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: report
            });

      - name: Fail on regression
        run: benchflow compare -b baseline.json -c current.json --threshold 1.05
        continue-on-error: true
```

### Local Development Workflow

```bash
# 1. Create baseline from main branch
git checkout main
benchflow run -o baseline.json

# 2. Switch to feature branch
git checkout feature-branch

# 3. Run current benchmarks
benchflow run -o current.json

# 4. Compare results
benchflow compare -b baseline.json -c current.json --format html --output report.html

# 5. Open report in browser
open report.html
```

### Threshold Configuration

Choose thresholds based on your project:

**High Performance Focus** (e.g., embedded systems, gaming):
- Use 1.02 (2%) to catch all regressions
- May have higher false positive rate
- Review all reported regressions carefully

**Moderate Standards** (most projects):
- Use 1.05 (5%, default) for balance
- Catches real regressions while tolerating measurement noise
- Recommended starting point

**Lenient Standards** (where performance varies):
- Use 1.10 (10%) for high-variance benchmarks
- Only flags obvious regressions
- Good for flaky benchmarks

## Troubleshooting

### No Regressions Detected, but Code is Slower

**Possible Causes:**
- Change is within measurement noise/threshold
- High variance in benchmark (standard deviation is large)
- Too lenient threshold setting

**Solutions:**
1. Lower the threshold: `--threshold 1.02` (2% instead of default 5%)
2. Increase statistical confidence: `--confidence 0.99` (99% instead of 95%)
3. Run benchmarks multiple times to reduce variance
4. Check std dev in results - high std dev = noisy benchmark

### False Positives (Flagging Improvements as Regressions)

**Possible Causes:**
- Threshold too low for noisy benchmark
- Measurement variance exceeds threshold
- Different machine/environment

**Solutions:**
1. Increase threshold: `--threshold 1.10` (10% instead of default 5%)
2. Run on same machine/environment
3. Average multiple runs before comparing
4. Review p-value and effect size - if not significant, ignore

### High P-Values Reported

**Meaning:**
- Performance change is not statistically significant
- Could be measurement noise, not real change
- Safe to ignore unless there's a clear trend

**Solutions:**
- Run more iterations to reduce variance
- Run benchmarks multiple times and average
- Check that test environment is stable (close other applications)
- Look for consistent pattern across multiple runs

### "Regression Threshold > 1.0" Error

**Cause:**
- Threshold must represent a multiplier > 1.0
- Threshold 1.05 means 5% slower is regression
- Not percentage points, but multiplication factor

**Solutions:**
```bash
# ✅ Correct
benchflow compare -b main.json -c current.json --threshold 1.05

# ❌ Incorrect
benchflow compare -b main.json -c current.json --threshold 5

# ❌ Incorrect
benchflow compare -b main.json -c current.json --threshold 0.05
```

### "Confidence Level" Error

**Cause:**
- Confidence must be between 0 and 1 (exclusive)
- 0.95 = 95%, not 95

**Solutions:**
```bash
# ✅ Correct
benchflow compare -b main.json -c current.json --confidence 0.95

# ❌ Incorrect
benchflow compare -b main.json -c current.json --confidence 95
```

## Advanced Usage

### Custom Regression Detection

```bash
# For critical performance path
# Flag anything 1% slower
benchflow compare -b main.json -c current.json --threshold 1.01

# For best-effort path
# Only flag 10%+ regressions
benchflow compare -b main.json -c current.json --threshold 1.10
```

### Cross-Language Comparison

```bash
# Compare same algorithm across languages
# Rust results in rust-results.json
# Go results in go-results.json

benchflow compare -b rust-results.json -c go-results.json
```

### Batch Processing

```bash
# Compare multiple result pairs
for baseline in v1.0/*.json; do
  current="${baseline//v1.0/main}"
  benchflow compare -b "$baseline" -c "$current" -f json >> results.jsonl
done
```

## Best Practices

1. **Run Multiple Times**: Average benchmarks over multiple runs to reduce variance
2. **Stable Environment**: Close other applications, avoid thermal throttling
3. **Same Hardware**: Compare benchmarks from the same machine when possible
4. **Review Statistical Metrics**: Always check p-value and effect size, not just threshold
5. **Document Changes**: Note why performance changed (new algorithm, optimization, etc.)
6. **Version Control**: Track benchmark results alongside code
7. **Monitor Trends**: Watch for gradual performance degradation over time
8. **Consider Workload**: Real-world workload characteristics matter more than worst-case

## See Also

- [API Reference](API_COMPARATOR.md)
- [Statistical Guide](STATISTICS.md)
- [GitHub Actions Integration](CI_CD_INTEGRATION.md)
