# Statistical Concepts Guide

This guide explains the statistical methods used in benchflow's comparative analysis.

## Table of Contents

- [Why Statistics Matter](#why-statistics-matter)
- [T-Tests and P-Values](#t-tests-and-p-values)
- [Confidence Levels](#confidence-levels)
- [Effect Size (Cohen's d)](#effect-size-cohens-d)
- [Confidence Intervals](#confidence-intervals)
- [Interpreting Results](#interpreting-results)
- [Common Mistakes](#common-mistakes)

## Why Statistics Matter

Benchmark measurements are inherently noisy:

```
Baseline measurements: [1000, 1005, 998, 1002, 1001] ns
Average: 1001 ns, but measurements vary by ±5 ns

Current measurements: [1010, 1008, 1012, 1005, 1015] ns
Average: 1010 ns, but measurements vary by ±5 ns
```

Is this a real 0.9% regression, or just measurement noise?

**Statistical testing answers this question.**

## T-Tests and P-Values

### The T-Test

The t-test determines if two groups of measurements are significantly different.

**Formula (simplified):**
```
t-statistic = (mean₁ - mean₂) / standard_error
```

**Interpretation:**
- Larger |t| = bigger difference relative to noise
- t-statistic gets converted to a p-value

### P-Values

The p-value is the probability of observing this difference **if there's actually no difference**:

**P-value Interpretation:**
```
p = 0.01  → 1% chance this difference is just noise → SIGNIFICANT
p = 0.05  → 5% chance this difference is just noise → SIGNIFICANT
p = 0.10  → 10% chance this difference is just noise → NOT SIGNIFICANT
p = 0.50  → 50% chance this difference is just noise → NOT SIGNIFICANT
```

### Significance Thresholds

Standard thresholds in research:

| Confidence | Threshold | Interpretation |
|-----------|-----------|----------------|
| 95% (default) | p < 0.05 | 95% confident change is real |
| 99% | p < 0.01 | 99% confident change is real |
| 90% | p < 0.10 | 90% confident change is real |

### Practical Example

```
Baseline: 1000 ns ± 50 ns (10 runs)
Current:  1100 ns ± 55 ns (10 runs)
Difference: 100 ns (10% slower)

T-test result:
  t-statistic = 4.2
  p-value = 0.0008

Conclusion:
  P < 0.05 → SIGNIFICANT at 95% confidence
  "We're 99.92% confident this is a real regression"
```

## Confidence Levels

Confidence level represents how stringent your statistical test is:

### 95% Confidence (Default)

```
Confidence Level: 0.95
Significance Level (α): 1 - 0.95 = 0.05
P-value Threshold: 0.05
```

**Meaning:** Accept 5% chance of false positive

**Usage:**
- Balanced approach for most projects
- Good default choice
- Catches real changes while tolerating some noise

**Example:**
```bash
benchflow compare -b main.json -c current.json --confidence 0.95
```

### 99% Confidence

```
Confidence Level: 0.99
Significance Level (α): 1 - 0.99 = 0.01
P-value Threshold: 0.01
```

**Meaning:** Accept 1% chance of false positive

**Usage:**
- High-stakes performance decisions
- Low tolerance for false alarms
- Requires larger differences to be significant

**Example:**
```bash
benchflow compare -b main.json -c current.json --confidence 0.99
```

### 90% Confidence

```
Confidence Level: 0.90
Significance Level (α): 1 - 0.90 = 0.10
P-value Threshold: 0.10
```

**Meaning:** Accept 10% chance of false positive

**Usage:**
- Quick feedback during development
- Higher false positive rate
- Smaller changes detected

**Example:**
```bash
benchflow compare -b main.json -c current.json --confidence 0.90
```

## Effect Size (Cohen's d)

Effect size measures the **magnitude** of performance change, independent of sample size.

### Interpretation

| Cohen's d | Magnitude | Practical Meaning |
|-----------|-----------|------------------|
| < 0.2 | Negligible | Imperceptible to users |
| 0.2 - 0.5 | Small | Slight but real change |
| 0.5 - 0.8 | Medium | Noticeable, meaningful |
| > 0.8 | Large | Substantial change |

### Example

```
Change: 1000 ns → 1050 ns (+5%)
Cohen's d: 0.3 (small effect)
→ Statistically significant but small practical impact

Change: 1000 ns → 2000 ns (+100%)
Cohen's d: 2.1 (large effect)
→ Statistically significant AND major practical impact
```

### Formula (Simplified)

```
Cohen's d = (mean₂ - mean₁) / pooled_std_dev
```

**Why It Matters:**
- P-value can be small with large sample size even for tiny differences
- Effect size shows if the difference matters practically
- Both are important for interpretation

### Example Report

```
| Benchmark | Delta | P-Value | Effect Size | Assessment |
|-----------|-------|---------|-------------|------------|
| sorting   | +1.5% | 0.04    | 0.15        | Significant but negligible effect |
| hashing   | +15%  | 0.01    | 1.2         | Significant and large effect |
| parsing   | -2%   | 0.08    | 0.18        | Not significant, negligible effect |
```

## Confidence Intervals

A confidence interval provides a range where the true value likely falls.

### Example

```
Baseline measurements: [1000, 1005, 998, 1002, 1001] ns
Calculated 95% CI: [995 ns, 1005 ns]

Meaning:
  "We're 95% confident the true mean is between 995 and 1005 ns"
```

### Interpretation

**Non-Overlapping Intervals:**
```
Baseline: [995 ns, 1005 ns]
Current:  [1045 ns, 1055 ns]
→ Changes don't overlap → Likely significant change
```

**Overlapping Intervals:**
```
Baseline: [990 ns, 1010 ns]
Current:  [1000 ns, 1020 ns]
→ Intervals overlap → Change may not be significant
```

### CLI Usage

```bash
benchflow compare -b main.json -c current.json --confidence 0.95
```

## Interpreting Results

### Decision Matrix

| P-Value | Effect Size | Regression? | Action |
|---------|-------------|------------|--------|
| < 0.05 | > 0.8 | Yes | 🛑 Fix before merging |
| < 0.05 | 0.2 - 0.8 | Yes | ⚠️  Review and consider |
| < 0.05 | < 0.2 | Yes | ✅ Accept, statistically insignificant |
| ≥ 0.05 | Any | No | ✅ Not statistically significant |

### Real-World Example

```
Benchmark: "json_parsing"
Baseline: 5000 ns ± 200 ns (100 runs)
Current:  5100 ns ± 180 ns (100 runs)
Delta: +2%

Statistical Results:
  P-Value: 0.03 (< 0.05, significant!)
  Cohen's d: 0.5 (medium effect)
  95% CI: [5020 ns, 5180 ns]

Decision:
  ✓ Statistically significant (p = 0.03)
  ✓ Medium effect size (d = 0.5)
  ✓ Confidence interval doesn't include baseline
  → This is a REAL regression worth investigating
```

## Common Mistakes

### Mistake #1: Treating P-Value as Probability of Hypothesis

❌ **Wrong Interpretation:**
```
P-value = 0.05
"There's a 5% chance my change is real"
```

✅ **Correct Interpretation:**
```
P-value = 0.05
"If there's no real change, there's a 5% chance
 I'd see this data by random chance"
```

### Mistake #2: Ignoring Effect Size

❌ **Wrong:**
```
P-value = 0.01 (highly significant!)
→ "This is a big problem"

(But Cohen's d = 0.1, effect is negligible)
```

✅ **Correct:**
```
P-value = 0.01 (statistically significant)
Cohen's d = 0.1 (negligible practical effect)
→ "Statistically significant but practically irrelevant"
```

### Mistake #3: Choosing Arbitrary Thresholds

❌ **Wrong:**
```
benchflow compare ... --threshold 1.001  # 0.1%
→ False positives galore!
```

✅ **Correct:**
```
# Choose based on your project needs:
# - 1.02 (2%): High performance focus
# - 1.05 (5%): Standard/default
# - 1.10 (10%): Lenient/noisy benchmarks
```

### Mistake #4: Low Sample Size

❌ **Wrong:**
```
Single run: Baseline=1000 ns, Current=1050 ns
→ Can't distinguish from noise!
```

✅ **Correct:**
```
Multiple runs (10+):
  Baseline: [998, 1002, 1001, 1003, 999, 1000, 1004, 1002, 1001, 1003]
  Current: [1045, 1052, 1048, 1051, 1049, 1050, 1048, 1051, 1049, 1052]
  → Clear, statistically significant difference
```

### Mistake #5: Comparing Unstable Systems

❌ **Wrong:**
```
Run benchmarks with many programs open
 → High variance, missed real regressions
```

✅ **Correct:**
```
Close other applications
 → Lower variance, more reliable statistics
```

## Guidelines for Different Scenarios

### Conservative (High Confidence Required)

Use when:
- Performance-critical systems
- Mission-critical applications
- Production releases

Settings:
```bash
benchflow compare -b main.json -c current.json \
  --confidence 0.99 \      # 99% confidence
  --threshold 1.05         # 5% regression threshold
```

### Balanced (Default)

Use for:
- Most projects
- Normal development workflow
- Regular reviews

Settings:
```bash
benchflow compare -b main.json -c current.json \
  --confidence 0.95 \      # 95% confidence (default)
  --threshold 1.05         # 5% regression threshold (default)
```

### Permissive (Quick Feedback)

Use when:
- Early development
- High-variance benchmarks
- Quick prototyping

Settings:
```bash
benchflow compare -b main.json -c current.json \
  --confidence 0.90 \      # 90% confidence
  --threshold 1.10         # 10% regression threshold
```

## References

### T-Distribution Calculations

Benchflow uses rational approximation for normal distribution CDF:

```go
// For small sample sizes, t-distribution approximation
// Uses coefficients from Abramowitz and Stegun
b1 := 0.319381530
b2 := -0.356563782
b3 := 1.781477937
b4 := -1.821255978
b5 := 1.330274429
```

### Cohen's d Calculation

```go
// Pooled standard deviation
pooledStdDev := sqrt((n1-1)*var1 + (n2-1)*var2) / (n1 + n2 - 2)

// Effect size
d := (mean2 - mean1) / pooledStdDev
```

### Further Reading

- [Statistical Rethinking](https://xcelab.net/rm/statistical-rethinking/) by Richard McElreath
- [How to Lie with Statistics](https://en.wikipedia.org/wiki/How_to_Lie_with_Statistics) by Darrell Huff
- [t-test Explanation](https://en.wikipedia.org/wiki/Student%27s_t-test)
- [P-values Explained](https://www.nature.com/articles/d41586-019-00857-9)

## See Also

- [Comparative Analysis Guide](COMPARISON.md)
- [API Reference](API_COMPARATOR.md)
- [GitHub Actions Integration](CI_CD_INTEGRATION.md)
