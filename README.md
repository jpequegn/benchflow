# Benchflow

[![CI](https://github.com/jpequegn/benchflow/actions/workflows/ci.yml/badge.svg)](https://github.com/jpequegn/benchflow/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/jpequegn/benchflow)](https://goreportcard.com/report/github.com/jpequegn/benchflow)
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

# Run benchmarks (coming in Phase 3)
benchflow run --config benchflow.yaml

# Compare results (coming in Phase 4)
benchflow compare --baseline v1.0.0 --current HEAD

# Generate report (coming in Phase 5)
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

### 🚧 Phase 3: Parallel Benchmark Execution Engine (Planned)
- [ ] Concurrent execution with goroutines
- [ ] Process management and timeout handling
- [ ] Output streaming to parser
- [ ] Error handling and retry logic

### 🚧 Phase 4: Result Aggregation & Storage (Planned)
- [ ] Unified result format
- [ ] Statistical calculations
- [ ] JSON/CSV export
- [ ] SQLite historical tracking
- [ ] Comparison logic
- [ ] Regression detection

### 🚧 Phase 5: HTML Report Generation (Planned)
- [ ] HTML template structure
- [ ] Chart.js integration
- [ ] Trend visualization
- [ ] Responsive design
- [ ] Self-contained reports

### 🚧 Phase 6: Multi-language Support (Planned)
- [ ] Python pytest-benchmark parser
- [ ] Go testing.B benchmark parser
- [ ] Auto-detection of benchmark type

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
│   ├── parser/             # Benchmark parsers (Rust complete)
│   ├── executor/           # Execution engine (Phase 3)
│   ├── aggregator/         # Result aggregation (Phase 4)
│   ├── reporter/           # Report generation (Phase 5)
│   └── storage/            # Historical storage (Phase 4)
├── pkg/
│   └── benchflow/          # Public API (future)
├── examples/               # Example configurations
├── testdata/               # Test fixtures
│   └── rust/              # Rust benchmark samples
├── .github/
│   └── workflows/          # CI/CD workflows
└── CLAUDE.md              # Development documentation
```

## Current Features

### Parser Interface

```go
type Parser interface {
    Parse(output []byte) (*BenchmarkSuite, error)
    Language() string
}
```

### Rust Parser (Bencher Format)

Parses cargo bench output:

```
test bench_sort ... bench:   1,234 ns/iter (+/- 56)
```

Features:
- Extracts benchmark name, time (ns), and standard deviation
- Handles comma-separated numbers
- Skips failed and ignored tests
- Tolerates compiler warnings

Example usage:

```go
parser := parser.NewRustParser()
suite, err := parser.Parse(cargoBenchOutput)
if err != nil {
    log.Fatal(err)
}

for _, result := range suite.Results {
    fmt.Printf("%s: %v ± %v\n", result.Name, result.Time, result.StdDev)
}
```

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

**Current Phase**: Phase 2 Complete (Rust Parser) ✅
**Next Phase**: Phase 3 (Parallel Execution Engine)

See [GitHub Issues](https://github.com/jpequegn/benchflow/issues) for detailed roadmap.

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Acknowledgments

Built with:
- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Viper](https://github.com/spf13/viper) - Configuration management

---

**Status**: 🚧 Active Development | ✅ Phases 1-2 Complete | 🎯 Phase 3 In Progress
