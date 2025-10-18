# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Benchflow is a cross-language benchmark aggregator with parallel execution and visualization. It's built in Go to leverage goroutines for concurrent benchmark runs across multiple languages (Rust, Python, Go, Node.js).

**Current Status**: All 7 phases complete! Foundation (Phase 1), Rust parser (Phase 2), parallel execution engine (Phase 3), aggregation & storage (Phase 4), HTML report generation (Phase 5), multi-language support (Phase 6), and Node.js parser (Phase 7) are fully implemented and tested. Supports Rust, Python, Go, and Node.js benchmarks with interactive HTML reports, Chart.js visualizations, Nebula UI dark theme, and responsive design. Production ready! 🚀

## Development Commands

### Build & Run
```bash
# Build the CLI
go build -o benchflow ./cmd/benchflow

# Run with go run (during development)
go run ./cmd/benchflow [args]

# Install locally
go install ./cmd/benchflow
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/parser
go test ./internal/executor

# Run single test
go test ./internal/parser -run TestParseBencher

# Verbose output
go test -v ./...
```

### Code Quality
```bash
# Format code
go fmt ./...

# Lint (requires golangci-lint)
golangci-lint run

# Vet
go vet ./...

# Run all quality checks
go fmt ./... && go vet ./... && go test ./...
```

## Architecture

Benchflow follows Go's standard project layout with clear separation of concerns:

### Core Pipeline Flow
```
User Config (YAML)
    ↓
CLI (cmd/benchflow) - Entry point, command routing
    ↓
Executor (internal/executor) - Concurrent benchmark execution via goroutines
    ↓
Parser (internal/parser) - Language-specific output parsing
    ↓
Aggregator (internal/aggregator) - Normalize results, calculate statistics
    ↓
Storage (internal/storage) - Historical tracking (SQLite)
    ↓
Reporter (internal/reporter) - Generate HTML/JSON/CSV output
```

### Package Responsibilities

**`cmd/benchflow/`** - CLI entry point
- Framework: cobra or urfave-cli (TBD in Phase 1)
- Commands: `run`, `compare`, `report`
- Configuration loading via viper (YAML/TOML)

**`internal/parser/`** - Multi-language benchmark parsers
- Rust: cargo bench bencher format + criterion
- Python: pytest-benchmark JSON
- Go: testing.B output
- Interface-based design for extensibility
- See Phase 2 issue for Rust parser details

**`internal/executor/`** - Concurrent execution engine
- Worker pool pattern with goroutines
- Context-based timeout/cancellation
- Process spawning via os/exec
- Real-time output streaming to parser
- Error handling and retry logic
- Reference: https://github.com/jpequegn/parakeet-podcast-processor parallel transcription

**`internal/aggregator/`** - Result normalization
- Unified result format across languages
- Statistical calculations (mean, median, std dev)
- Comparison logic for baseline vs current
- Regression detection with configurable thresholds

**`internal/storage/`** - Historical tracking
- SQLite for persistence
- Time-series benchmark data
- Query interface for trend analysis

**`internal/reporter/`** - Output generation
- HTML reports with html/template + Chart.js
- Dark mode support (Nebula UI style)
- Self-contained output (embedded CSS/JS via embed package)
- JSON/CSV export for CI/CD integration

**`pkg/benchflow/`** - Public API
- Stable interfaces for external consumption
- Versioned according to semantic versioning

**`testdata/`** - Test fixtures
- Sample benchmark outputs for each language
- Edge case data for parser validation

**`examples/`** - Configuration examples
- Sample benchflow.yaml files
- Multi-language benchmark setups

## Implementation Phases

All 7 phases complete (tracked in GitHub Issues):

1. **Foundation** (#1) - ✅ COMPLETE - CLI framework, config, logging, tests, CI/CD
2. **Rust Parser** (#2) - ✅ COMPLETE - Bencher/criterion format parsing (82.9% coverage)
3. **Execution Engine** (#3) - ✅ COMPLETE - Goroutine-based parallel execution (94.0% coverage)
4. **Aggregation & Storage** (#4) - ✅ COMPLETE - Statistical aggregation, JSON/CSV export, SQLite storage, regression detection (94.0% aggregator, 82.2% storage coverage)
5. **HTML Reports** (#5) - ✅ COMPLETE - Interactive HTML reports with Chart.js, Nebula UI dark theme, responsive design (75.6% coverage)
6. **Multi-language** (#6) - ✅ COMPLETE - Python pytest-benchmark (#11) and Go testing.B (#12) parsers
7. **Node.js Parser** (#15) - ✅ COMPLETE - Benchmark.js text format parser (81.2% coverage)

**Current Status**: Production ready! All core functionality implemented and tested. 4 language parsers (Rust, Python, Go, Node.js) with full pipeline integration. Future enhancements: additional language support, performance optimizations, dashboard improvements.

## Key Design Patterns

### Interface-Based Parsers
Each language parser implements a common interface:
```go
type Parser interface {
    Parse(output io.Reader) (*BenchmarkResult, error)
}
```

### Worker Pool Execution
Executor uses goroutines + channels for concurrent benchmark runs:
- Worker pool with configurable concurrency
- Context for timeout/cancellation propagation
- Channel-based result collection

### Unified Result Format
All parsers normalize to a common struct for aggregation:
```go
type BenchmarkResult struct {
    Name       string
    Language   string
    Time       Duration
    Iterations int
    StdDev     Duration
    // ... additional metrics
}
```

## Configuration

Benchflow uses YAML configuration files (via viper):

```yaml
benchmarks:
  - name: "rust-sort"
    language: rust
    command: "cargo bench --bench sort"
    timeout: 5m

  - name: "python-search"
    language: python
    command: "pytest --benchmark-only"
    timeout: 3m

output:
  formats: [html, json, csv]
  directory: ./reports

storage:
  enabled: true
  path: ./benchflow.db
```

## Testing Strategy

- **Unit tests**: All parsers, aggregators, reporters (80%+ coverage goal)
- **Integration tests**: Full pipeline with testdata fixtures
- **Table-driven tests**: Go idiom for multiple test cases
- **Golden files**: Expected output comparisons for reporters
- **CI/CD**: GitHub Actions runs tests on every PR

## CLI Usage

```bash
# Run all benchmarks from config (✅ IMPLEMENTED)
benchflow run --config benchflow.yaml

# Run with custom parallelism (✅ IMPLEMENTED)
benchflow run --parallel 8

# Run specific benchmark (✅ IMPLEMENTED)
benchflow run --name rust-sort-algorithms

# Run with timeout (✅ IMPLEMENTED)
benchflow run --timeout 10m

# Compare against baseline (Phase 4)
benchflow compare --baseline v1.0.0 --current HEAD

# Generate HTML report (Phase 5)
benchflow report --format html --output report.html
```

### Implemented Features (Phase 3)

The executor supports:
- **Parallel execution**: Worker pool pattern with configurable concurrency (default: 4)
- **Context-based cancellation**: Graceful shutdown on CTRL+C or timeout
- **Automatic retry**: Configurable retry attempts with exponential backoff
- **Progress tracking**: Real-time event notifications (started, completed, failed, retrying)
- **Fail-fast mode**: Stop on first failure (optional)
- **Parser registry**: Plugin architecture for language-specific parsers
- **Comprehensive error handling**: Execution errors, parsing errors, timeouts
- **Performance**: 10 benchmarks with 5 workers complete in ~240ms vs 1s sequential

Test coverage: 94.0%

## Related Projects

- **Parakeet Podcast Processor**: Reference for parallel execution patterns
  - Path: `/Users/julienpequegnot/Code/parakeet-podcast-processor`
  - See: `p3 transcribe` for goroutine-based concurrent processing
  - Similar worker pool + error handling approach

## Dependencies (Planned)

- **CLI**: cobra or urfave-cli
- **Config**: spf13/viper
- **Logging**: log/slog (Go 1.21+ stdlib)
- **Database**: mattn/go-sqlite3
- **Testing**: stretchr/testify (assertions)
- **Web**: stdlib html/template, Chart.js (embedded)
