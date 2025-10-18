# Phase 7: Node.js Benchmark Parser Implementation Plan

## Overview
Implement Node.js benchmark parser for Benchmark.js text format output. This completes the multi-language support by adding JavaScript/Node.js benchmarks to benchflow.

**GitHub Issue**: #15
**Branch**: `feat/issue-15-nodejs-parser`
**Target Coverage**: 80%+

## 1. Parser Architecture

### File Structure
```
internal/parser/
├── nodejs.go          # Node.js parser implementation
├── nodejs_test.go     # Comprehensive test suite
└── [existing files]

testdata/nodejs/
├── benchmark_js_basic.txt        # Basic benchmarks
├── benchmark_js_edge_cases.txt   # Edge cases
└── benchmark_js_with_errors.txt  # With "Fastest" line
```

### Implementation Pattern
Follow the established pattern from Python and Go parsers:

```go
package parser

import (
    "bufio"
    "bytes"
    "fmt"
    "regexp"
    "strconv"
    "strings"
    "time"
)

// NodeJSParser implements Parser for Benchmark.js output
type NodeJSParser struct{}

// NewNodeJSParser creates a new Node.js benchmark parser
func NewNodeJSParser() *NodeJSParser {
    return &NodeJSParser{}
}

// Language returns the language this parser supports
func (p *NodeJSParser) Language() string {
    return "nodejs"
}

// Parse parses Benchmark.js text output
func (p *NodeJSParser) Parse(output []byte) (*BenchmarkSuite, error) {
    // Implementation here
}
```

## 2. Parsing Logic

### Regex Pattern
```go
// Pattern: name x ops/sec ±percentage% (runs sampled)
benchRegex := regexp.MustCompile(
    `^(.+?)\s+x\s+([\d,]+)\s+ops/sec\s+±([\d.]+)%\s+\((\d+)\s+runs?\s+sampled\)`,
)
```

### Parse Steps

1. **Initialize Suite**
   - Create BenchmarkSuite with Language="nodejs"
   - Initialize Results slice

2. **Line-by-line Processing**
   - Use bufio.Scanner for line-by-line parsing
   - Trim whitespace from each line
   - Skip empty lines and non-benchmark lines

3. **Extract Metrics**
   - **Capture Group 1**: Benchmark name (e.g., "Array#forEach")
   - **Capture Group 2**: ops/sec with commas (e.g., "1,234,567")
   - **Capture Group 3**: Margin of error % (e.g., "1.23")
   - **Capture Group 4**: Runs sampled (e.g., "90")

4. **Convert Metrics**
   - Remove commas from ops/sec: `"1,234,567"` → `1234567`
   - Parse as float64
   - Calculate time per operation: `timeSec = 1.0 / opsPerSec`
   - Convert to nanoseconds: `timeNs = timeSec * 1e9`
   - Approximate stddev from margin of error

5. **Create Result**
   ```go
   result := &BenchmarkResult{
       Name:       name,
       Language:   "nodejs",
       Time:       time.Duration(timeNs),
       Iterations: samplesInt,
       StdDev:     approximateStdDev(marginOfError, samplesInt),
       Throughput: &Throughput{
           Value: opsPerSec,
           Unit:  "ops/s",
       },
       Metadata: map[string]string{
           "margin_of_error": fmt.Sprintf("%.2f%%", marginOfError),
       },
   }
   ```

6. **Validate and Return**
   - Return error if no results found
   - Return ParseError if parsing fails

### Edge Cases to Handle

1. **Comma-separated numbers**: `1,234,567` → `1234567`
   - Use `strings.ReplaceAll(str, ",", "")`

2. **"Fastest is..." lines**: Skip gracefully
   - Lines won't match regex pattern, automatically skipped

3. **Special characters in names**
   - Support: `Array#forEach`, `String#indexOf`, `Object.prototype.keys`
   - Regex captures everything before ` x`

4. **Large numbers**: `123,456,789 ops/sec`
   - Standard float64 parsing handles this

5. **Very small margins of error**: `±0.01%`
   - Handle with float64 precision

6. **Edge margin values**: `±100.00%`
   - Edge case in test data, should parse correctly

### Standard Deviation Calculation

```go
func approximateStdDev(marginOfError, samples float64) time.Duration {
    // Relative Margin of Error (RME) ≈ 1.96 * StdErr / mean
    // StdErr = mean / sqrt(samples) approximately
    // For approximation: StdDev ≈ RME * mean / sqrt(samples) * sqrt(samples)
    // Simplified: StdDev ≈ RME * mean / 1.96

    // Since we're working with time, use proportional estimate
    // This is approximate; actual calculation may vary
    proportion := marginOfError / 100.0
    return time.Duration(float64(timeNanoseconds) * proportion / 1.96)
}
```

## 3. Test Strategy

### Test File: `nodejs_test.go`

#### TestNewNodeJSParser
```go
func TestNewNodeJSParser(t *testing.T) {
    parser := NewNodeJSParser()
    assert.NotNil(t, parser)
}
```

#### TestLanguage
```go
func TestLanguage(t *testing.T) {
    parser := NewNodeJSParser()
    assert.Equal(t, "nodejs", parser.Language())
}
```

#### TestParseBasic (from benchmark_js_basic.txt)
- 3 benchmarks
- Verify names, ops/sec values, iterations
- Check "Fastest is..." line is skipped

#### TestParseLargeNumbers (from benchmark_js_edge_cases.txt)
- Very small: 1 ops/sec
- Very large: 123,456,789 ops/sec
- Check comma handling

#### TestParseSpecialCharacters
- Names with `#` (Array#forEach)
- Names with `.` (Object.prototype.keys)
- Names with underscores (Test_name)
- Names with hyphens (test-name)

#### TestParseEdgeCases
- High margin of error: 100%
- Low margin of error: 0.01%
- Single run vs many runs

#### TestParseWithFastestLine
- Handle "Fastest is X" line
- Parse remaining benchmarks correctly

#### TestParseErrors
- Empty input → ParseError
- Malformed lines → ParseError
- No matching benchmarks → ParseError
- Partial benchmark data → ParseError

#### TestParseMetadata
- Verify metadata includes margin_of_error
- Check margin_of_error is stored correctly

#### TestConversions
- ops/sec to time: verify formula
- Margin of error to stddev: verify approximation
- Iterator count: verify preserved

### Table-Driven Tests

```go
tests := []struct {
    name          string
    input         string
    expectedCount int
    expectedFirst *BenchmarkResult
    expectedErr   bool
}{
    {
        name: "basic benchmark",
        input: "test x 1,000 ops/sec ±1.0% (10 runs sampled)",
        expectedCount: 1,
        expectedFirst: &BenchmarkResult{
            Name: "test",
            // ... other fields
        },
        expectedErr: false,
    },
    // ... more cases
}
```

## 4. Integration

### Update Internal Parser Files

1. **executor.go** - Register Node.js parser
   ```go
   // In NewExecutor or parser registry
   e.parsers["nodejs"] = parser.NewNodeJSParser()
   ```

2. **types.go** - No changes needed (already has generic Parser interface)

### Update Main CLI

1. **cmd/benchflow/** - Support nodejs language
   - Update configuration parsing to allow `language: nodejs`
   - No changes needed if already generic

## 5. Documentation Updates

### README.md

Add Node.js to Implementation Status:
```markdown
### ✅ Phase 7: Node.js Benchmark Parser (Complete)
- ✅ Benchmark.js text format parser
- ✅ Regex-based parsing (ops/sec to time conversion)
- ✅ Margin of error and sample count extraction
- ✅ Comprehensive test suite (80%+ coverage)
```

Add to Current Features:
```markdown
**Node.js** - Benchmark.js output format (XX% coverage)
- Parses Benchmark.js text format: `name x ops/sec ±percentage% (runs sampled)`
- Extracts operations per second, margin of error, and sample count
- Converts throughput metrics to unified time format
- Handles special characters in benchmark names
```

Add to Project Structure:
```markdown
├── testdata/
│   └── nodejs/              # Node.js benchmark samples
```

### CLAUDE.md

Update Implementation Phases:
```markdown
7. **Node.js Parser** - Benchmark.js format parsing (XX% coverage)
```

## 6. Testing Checklist

- [ ] Unit tests pass (all table-driven tests)
- [ ] Edge cases handled correctly
- [ ] Error handling works as expected
- [ ] Coverage ≥80%
- [ ] No golangci-lint errors
- [ ] Code formatted (go fmt)
- [ ] Integration with executor works
- [ ] Full pipeline works (exec → parse → aggregate)

## 7. Success Criteria

✅ Parses Benchmark.js text format correctly
✅ Extracts all metrics (name, ops/sec, margin, samples)
✅ Converts ops/sec to time duration
✅ Handles edge cases (large numbers, special chars, etc.)
✅ 80%+ test coverage
✅ Comprehensive documentation
✅ Works with executor and aggregator
✅ Zero golangci-lint errors

## 8. Timeline

1. **Implement parser** (30-45 min)
   - Write nodejs.go with full parsing logic
   - Handle all edge cases

2. **Comprehensive tests** (30-45 min)
   - Write nodejs_test.go
   - Table-driven tests
   - Edge case coverage

3. **Integration** (15-20 min)
   - Register in executor
   - Update documentation
   - Verify full pipeline

4. **Verification** (10-15 min)
   - Run tests and linter
   - Manual testing
   - Final documentation

**Total Estimated**: ~2 hours

## 9. References

- **Parser Interface**: `/Users/julienpequegnot/Code/benchflow/internal/parser/types.go`
- **Python Parser Example**: `/Users/julienpequegnot/Code/benchflow/internal/parser/python.go`
- **Rust Parser Example**: `/Users/julienpequegnot/Code/benchflow/internal/parser/rust.go`
- **Go Parser Example**: `/Users/julienpequegnot/Code/benchflow/internal/parser/go.go`
- **Test Data**: `/Users/julienpequegnot/Code/benchflow/testdata/nodejs/`
- **Benchmark.js Format**: Research from Issue #15

## 10. Implementation Notes

- Keep implementation consistent with existing parsers
- Maintain uniform BenchmarkResult structure
- Store ops/sec in Throughput field (unlike Rust which stores ns/iter)
- Approximate StdDev from margin of error (no exact StdDev in Benchmark.js output)
- Skip "Fastest is..." lines naturally (won't match regex)
- Use bufio.Scanner for efficient line processing
- Validate input has at least one benchmark result

---

**Ready to start implementation!** 🚀
