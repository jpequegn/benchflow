# Benchflow

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

## Core Features (Planned)

### Phase 1: Foundation
- [ ] Go project structure with proper modules
- [ ] CLI framework (cobra/urfave-cli)
- [ ] Configuration file support (YAML/TOML)
- [ ] Basic Rust cargo bench parser

### Phase 2: Execution Engine
- [ ] Parallel benchmark execution with goroutines
- [ ] Process management and timeout handling
- [ ] Output streaming and logging
- [ ] Error handling and retry logic

### Phase 3: Multi-language Support
- [ ] Python pytest-benchmark integration
- [ ] Go testing.B benchmark support
- [ ] Custom benchmark format support

### Phase 4: Reporting & Visualization
- [ ] JSON/CSV export
- [ ] HTML report generation
- [ ] Historical trend tracking
- [ ] Comparison views

### Phase 5: Advanced Features
- [ ] Web dashboard with real-time updates
- [ ] CI/CD integration helpers
- [ ] Statistical analysis (mean, median, std dev, outliers)
- [ ] Regression detection with configurable thresholds

## Technology Stack

- **Language**: Go 1.21+
- **CLI Framework**: TBD (cobra vs urfave-cli)
- **Testing**: Go standard testing package + testify
- **Web**: net/http + html/template (or htmx for interactivity)
- **Parsing**: Custom parsers for each benchmark format
- **Concurrency**: Goroutines + channels + context

## Project Structure

```
benchflow/
├── cmd/
│   └── benchflow/          # CLI entry point
├── internal/
│   ├── executor/           # Benchmark execution engine
│   ├── parser/             # Format parsers (Rust, Python, Go)
│   ├── aggregator/         # Result aggregation
│   ├── reporter/           # Report generation
│   └── storage/            # Historical data storage
├── pkg/
│   └── benchflow/          # Public API
├── web/                    # Web dashboard (optional)
├── examples/               # Example configurations
└── testdata/               # Test fixtures
```

## Quick Start (Future)

```bash
# Install
go install github.com/jpequegn/benchflow/cmd/benchflow@latest

# Run benchmarks
benchflow run --config benchflow.yaml

# Compare implementations
benchflow compare --baseline v1.0.0 --current HEAD

# Generate report
benchflow report --format html --output report.html
```

## Development Status

🚧 **Planning Phase** - Repository initialized, implementation roadmap in GitHub Issues

## License

MIT
