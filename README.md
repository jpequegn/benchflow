# Benchflow

[![CI](https://github.com/jpequegn/benchflow/actions/workflows/ci.yml/badge.svg)](https://github.com/jpequegn/benchflow/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/jpequegn/benchflow.svg)](https://pkg.go.dev/github.com/jpequegn/benchflow)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/jpequegn/benchflow)](https://github.com/jpequegn/benchflow)

**Cross-language benchmark aggregator with parallel execution and visualization**

## Vision

A unified benchmarking tool that runs, aggregates, and visualizes performance benchmarks across multiple languages and frameworks. Built with Go for high-performance concurrent execution.

## Goals

- **Multi-language support**: Rust (cargo bench), Python (pytest-benchmark), Go (testing.B), Node.js
- **Parallel execution**: Leverage goroutines for concurrent benchmark runs
- **Unified reporting**: Aggregate results into common format (JSON, CSV, HTML)
- **Historical tracking**: Track performance trends over time
- **Comparison mode**: Compare different implementations (e.g., classical vs quantum algorithms)
- **CLI-first**: Terminal-native with rich output
- **Web dashboard**: Optional visualization interface

## Use Cases

- Compare algorithm implementations across languages
- Track performance regression over time
- Benchmark different optimization approaches
- Generate performance reports for documentation
- CI/CD integration for automated performance testing

## Installation

### Prerequisites

- Go 1.24 or higher
- Git

### Install from Source

```bash
# Clone repository
git clone https://github.com/jpequegn/benchflow.git
cd benchflow

# Build and install
go install ./cmd/benchflow

# Or build locally
go build -o benchflow ./cmd/benchflow
```

### Verify Installation

```bash
benchflow --version
# benchflow version 0.1.0

benchflow --help
```

## Quick Start

```bash
# View available commands
benchflow --help

# Run benchmarks
benchflow run --config benchflow.yaml

# Compare results
benchflow compare --baseline v1.0.0 --current HEAD

# Generate report
benchflow report --format html --output report.html
```

## Development

### Build

```bash
# Build binary
go build -o benchflow ./cmd/benchflow

# Build with verbose output
go build -v -o benchflow ./cmd/benchflow
```

### Test

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Code Quality

```bash
# Format code
go fmt ./...

# Vet for issues
go vet ./...

# Run linter (requires golangci-lint)
golangci-lint run

# Complete verification
go fmt ./... && go vet ./... && go test ./... && go build ./cmd/benchflow
```

## Implementation Status

### ✅ Phase 1: Project Foundation & Setup (Complete)
- ✅ Go project structure with proper modules
- ✅ CLI framework (cobra)
- ✅ Configuration file support (viper for YAML/TOML)
- ✅ Logging framework (slog)
- ✅ GitHub Actions CI/CD
- ✅ Initial test suite

### ✅ Phase 2: Rust Benchmark Parser (Complete)
- ✅ Parser interface for multi-language support
- ✅ Rust cargo bench bencher format parser
- ✅ Comprehensive test suite (82.9% coverage)
- ✅ Error handling and edge cases
- ✅ Full documentation

### ✅ Phase 3: Parallel Benchmark Execution Engine (Complete)
- ✅ Concurrent execution with goroutines
- ✅ Process management and timeout handling
- ✅ Output streaming to parser
- ✅ Error handling and retry logic with exponential backoff
- ✅ Comprehensive test suite (94.0% coverage)

### ✅ Phase 4: Result Aggregation & Storage (Complete)
- ✅ Unified result format across all languages
- ✅ Statistical calculations (mean, median, std dev)
- ✅ JSON/CSV export for CI/CD integration
- ✅ SQLite historical tracking and trend analysis
- ✅ Comparison logic and baseline tracking
- ✅ Regression detection with configurable thresholds
- ✅ Comprehensive test suite (94.0% coverage)

### ✅ Phase 5: HTML Report Generation (Complete)
- ✅ HTML template structure with responsive design
- ✅ Chart.js integration for interactive visualizations
- ✅ Trend visualization with historical data
- ✅ Responsive design for desktop and mobile
- ✅ Self-contained reports (embedded CSS/JS)
- ✅ Nebula UI dark theme
- ✅ Comprehensive test suite (75.6% coverage)

### ✅ Phase 6: Multi-language Support (Complete)
- ✅ Python pytest-benchmark JSON parser
- ✅ Go testing.B output parser
- ✅ Auto-detection of benchmark type
- ✅ Comprehensive test suites for both parsers

### ✅ Phase 7: Node.js Benchmark Parser (Complete)
- ✅ Benchmark.js text format parser
- ✅ Regex-based parsing (ops/sec to time conversion)
- ✅ Margin of error and sample count extraction
- ✅ Throughput metrics and standard deviation approximation
- ✅ Comprehensive test suite (81.2% coverage)
- ✅ Integration with executor and full pipeline

## Technology Stack

- **Language**: Go 1.24
- **CLI Framework**: [Cobra](https://github.com/spf13/cobra)
- **Configuration**: [Viper](https://github.com/spf13/viper) (YAML/TOML)
- **Logging**: log/slog (Go stdlib)
- **Testing**: Go standard testing + table-driven tests
- **Parsing**: Custom parsers with regexp + bufio
- **Concurrency**: Goroutines + channels + context (Phase 3)
- **Storage**: SQLite (Phase 4)
- **Reporting**: html/template + Chart.js (Phase 5)

## Project Structure

```
benchflow/
├── cmd/
│   └── benchflow/          # CLI entry point (main.go)
├── internal/
│   ├── cmd/                # CLI commands (cobra)
│   ├── parser/             # Multi-language benchmark parsers (Rust, Python, Go, Node.js)
│   ├── executor/           # Concurrent execution engine with goroutines
│   ├── aggregator/         # Result aggregation and statistics
│   ├── reporter/           # HTML/JSON/CSV report generation
│   └── storage/            # SQLite historical tracking
├── pkg/
│   └── benchflow/          # Public API (future)
├── examples/               # Example configurations
├── testdata/               # Test fixtures
│   ├── rust/              # Rust benchmark samples
│   ├── python/            # Python benchmark samples
│   ├── go/                # Go benchmark samples
│   └── nodejs/            # Node.js benchmark samples
├── .github/
│   └── workflows/          # CI/CD workflows
└── CLAUDE.md              # Development documentation
```

## Current Features

### Multi-Language Parser Support

**Rust** - Cargo bench bencher format (82.9% coverage)
- Extracts benchmark name, time (ns), and standard deviation
- Handles comma-separated numbers and large values
- Skips failed and ignored tests gracefully
- Tolerates compiler warnings

**Python** - pytest-benchmark JSON format (comprehensive coverage)
- Parses JSON output from pytest-benchmark
- Extracts mean, min, max, stddev, and iteration counts
- Handles optional fields and edge cases
- Full pytest-benchmark ecosystem support

**Go** - testing.B output format (comprehensive coverage)
- Parses Go benchmark output with ns/op metrics
- Extracts memory allocations (B/op, allocs/op)
- Supports both simple and detailed output formats
- Handles compiler optimizations gracefully

**Node.js** - Benchmark.js text format (81.2% coverage)
- Parses Benchmark.js output: `name x ops/sec ±percentage% (runs sampled)`
- Converts throughput (ops/sec) to time-based metrics
- Extracts margin of error and sample count
- Approximates standard deviation from margin of error
- Handles special characters in benchmark names

### Parser Interface

```go
type Parser interface {
    Parse(output []byte) (*BenchmarkSuite, error)
    Language() string
}
```

### Core Features

- **Parallel Execution**: Goroutine-based worker pool with configurable concurrency
- **Unified Format**: All parsers normalize to common result structure
- **Historical Tracking**: SQLite storage for trend analysis
- **Statistical Analysis**: Mean, median, stddev, min/max calculations
- **Regression Detection**: Configurable thresholds for performance regressions
- **Multiple Export Formats**: HTML (interactive), JSON, CSV
- **Interactive Reports**: Chart.js visualizations with Nebula UI dark theme

## Configuration Example

```yaml
# benchflow.yaml
benchmarks:
  - name: "rust-algorithms"
    language: rust
    command: "cargo bench --bench sort"
    timeout: 5m

  - name: "python-data"
    language: python
    command: "pytest --benchmark-only"
    timeout: 3m

  - name: "nodejs-algorithms"
    language: nodejs
    command: "npm run benchmark"
    timeout: 2m

execution:
  parallel: 4
  retry: 3

output:
  formats: [html, json, csv]
  directory: ./reports

storage:
  enabled: true
  path: ./benchflow.db
```

## Contributing

Contributions welcome! Please see:
- [GitHub Issues](https://github.com/jpequegn/benchflow/issues) for planned features
- `CLAUDE.md` for development documentation
- Run tests and linting before submitting PRs

## Development Status

**All 7 Phases Complete** ✅

- ✅ Phase 1: Project Foundation & Setup
- ✅ Phase 2: Rust Benchmark Parser
- ✅ Phase 3: Parallel Benchmark Execution Engine
- ✅ Phase 4: Result Aggregation & Storage
- ✅ Phase 5: HTML Report Generation
- ✅ Phase 6: Multi-language Support (Python & Go)
- ✅ Phase 7: Node.js Benchmark Parser (Benchmark.js)

**Next Steps**: Performance optimizations, additional language support (TypeScript, etc.), dashboard enhancements, or community features

See [GitHub Issues](https://github.com/jpequegn/benchflow/issues) for roadmap and feature requests.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

Built with:
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management

---

**Status**: ✅ All Phases Complete | 🚀 Production Ready | 📋 Future Enhancements Welcome
