# Benchflow Installation & Development Guide

Complete guide for installing, building, testing, and developing benchflow.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Building](#building)
- [Testing](#testing)
- [Development Workflow](#development-workflow)
- [Troubleshooting](#troubleshooting)

## Prerequisites

### Required

- **Go 1.24 or higher**
  ```bash
  go version
  # Should output: go version go1.24.x ...
  ```

- **Git**
  ```bash
  git --version
  ```

### Optional (for development)

- **golangci-lint** (for linting)
  ```bash
  # macOS
  brew install golangci-lint

  # Linux/WSL
  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin

  # Verify
  golangci-lint --version
  ```

## Installation

### Method 1: Install from Source (Recommended)

```bash
# 1. Clone repository
git clone https://github.com/jpequegn/benchflow.git
cd benchflow

# 2. Download dependencies
go mod download
go mod verify

# 3. Install to $GOPATH/bin (usually ~/go/bin)
go install ./cmd/benchflow

# 4. Verify installation (ensure $GOPATH/bin is in your PATH)
benchflow --version
```

### Method 2: Build Locally

```bash
# Clone and navigate to directory
git clone https://github.com/jpequegn/benchflow.git
cd benchflow

# Build binary in current directory
go build -o benchflow ./cmd/benchflow

# Run it
./benchflow --version
```

### Method 3: Go Install (from GitHub)

```bash
# Install latest version directly
go install github.com/jpequegn/benchflow/cmd/benchflow@latest

# Install specific version (when releases are available)
go install github.com/jpequegn/benchflow/cmd/benchflow@v0.1.0
```

## Building

### Basic Build

```bash
# Build in current directory
go build -o benchflow ./cmd/benchflow

# Build with verbose output
go build -v -o benchflow ./cmd/benchflow

# Build for different platforms
GOOS=linux GOARCH=amd64 go build -o benchflow-linux ./cmd/benchflow
GOOS=darwin GOARCH=arm64 go build -o benchflow-macos ./cmd/benchflow
GOOS=windows GOARCH=amd64 go build -o benchflow.exe ./cmd/benchflow
```

### Build with Optimizations

```bash
# Strip debug information and reduce binary size
go build -ldflags="-s -w" -o benchflow ./cmd/benchflow

# Add version information
go build -ldflags="-X main.version=0.1.0" -o benchflow ./cmd/benchflow
```

### Verify Build

```bash
# Check binary exists
ls -lh benchflow

# Test execution
./benchflow --version
./benchflow --help

# Check dependencies
go list -m all
```

## Testing

### Run All Tests

```bash
# Basic test run
go test ./...

# Verbose output (shows all tests)
go test -v ./...

# Run tests with race detector
go test -race ./...

# Short mode (skip slow tests)
go test -short ./...
```

### Test Specific Packages

```bash
# Test parser only
go test ./internal/parser

# Test parser with verbose output
go test -v ./internal/parser

# Test CLI commands
go test ./internal/cmd
```

### Run Individual Tests

```bash
# Run specific test function
go test ./internal/parser -run TestRustParser_Parse_BasicBencher

# Run tests matching pattern
go test ./internal/parser -run TestRustParser

# Run with verbose output
go test -v ./internal/parser -run TestRustParser_Language
```

### Test Coverage

```bash
# Show coverage percentage
go test -cover ./...

# Expected output:
# ?       github.com/jpequegn/benchflow/cmd/benchflow      [no test files]
# ok      github.com/jpequegn/benchflow/internal/cmd        0.367s  coverage: 100.0% of statements
# ok      github.com/jpequegn/benchflow/internal/parser     0.285s  coverage: 82.9% of statements

# Test specific package with coverage
go test -cover ./internal/parser
```

### Generate Coverage Report

```bash
# Generate coverage profile
go test -coverprofile=coverage.out ./...

# View coverage in terminal
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out

# Generate and open HTML report (macOS)
go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html && open coverage.html
```

### Advanced Testing

```bash
# Run tests with timeout
go test -timeout 30s ./...

# Run tests in parallel
go test -parallel 4 ./...

# Run benchmarks (when available)
go test -bench=. ./...

# Generate test coverage for CI
go test -coverprofile=coverage.out -covermode=atomic ./...
```

## Development Workflow

### Initial Setup

```bash
# 1. Fork and clone
git clone https://github.com/YOUR_USERNAME/benchflow.git
cd benchflow

# 2. Add upstream remote
git remote add upstream https://github.com/jpequegn/benchflow.git

# 3. Install dependencies
go mod download
```

### Daily Development

```bash
# 1. Update from upstream
git checkout main
git pull upstream main

# 2. Create feature branch
git checkout -b feature/my-feature

# 3. Make changes and test frequently
go test ./...

# 4. Format and lint
go fmt ./...
go vet ./...
golangci-lint run

# 5. Run all quality checks
go fmt ./... && go vet ./... && go test ./... && go build ./cmd/benchflow

# 6. Commit and push
git add .
git commit -m "feat: Add my feature"
git push origin feature/my-feature
```

### Code Quality Checks

```bash
# Format all Go code
go fmt ./...

# Vet for common issues
go vet ./...

# Run linter
golangci-lint run

# Run linter with auto-fix
golangci-lint run --fix

# Check for security issues (requires gosec)
gosec ./...

# Check for updates
go list -u -m all
```

### Pre-Commit Checklist

```bash
# Run this before committing
#!/bin/bash
set -e

echo "Formatting code..."
go fmt ./...

echo "Vetting code..."
go vet ./...

echo "Running tests..."
go test ./...

echo "Running linter..."
golangci-lint run

echo "Building binary..."
go build ./cmd/benchflow

echo "✅ All checks passed!"
```

## Running Benchflow

### Quick Reference

```bash
# View help
benchflow --help

# Show version
benchflow --version

# Verbose mode
benchflow --verbose run --config benchflow.yaml
```

### Available Commands (All Implemented ✅)

#### **1. Run Benchmarks**

```bash
# Run all benchmarks from config
benchflow run --config benchflow.yaml

# Run specific benchmark by name
benchflow run --config benchflow.yaml --name rust-sort

# Run with custom parallelism
benchflow run --config benchflow.yaml --parallel 8

# Run with timeout
benchflow run --config benchflow.yaml --timeout 10m

# Verbose output
benchflow --verbose run --config benchflow.yaml
```

#### **2. Compare Results**

```bash
# Compare against git reference
benchflow compare --baseline HEAD~1 --current HEAD

# Compare version tags
benchflow compare --baseline v0.1.0 --current HEAD

# Compare with custom threshold (5% = 1.05)
benchflow compare --baseline main --current develop --threshold 1.10
```

#### **3. Generate Reports**

```bash
# Generate HTML report
benchflow report --format html --output performance.html

# Generate JSON report
benchflow report --format json --output results.json

# Generate CSV report
benchflow report --format csv --output results.csv

# Generate all formats
benchflow report --format html --format json --format csv --output results
```

### Configuration Files

#### Basic Setup

```bash
# Use custom config
benchflow run --config examples/benchflow.example.yaml

# Use config from environment
export BENCHFLOW_CONFIG=benchflow.yaml
benchflow run
```

#### Configuration Example

Create `benchflow.yaml`:

```yaml
benchmarks:
  # Rust benchmark
  - name: "rust-algorithms"
    language: rust
    command: "cargo bench"
    workdir: "./rust-project"
    timeout: 5m

  # Python benchmark
  - name: "python-data"
    language: python
    command: "pytest --benchmark-only tests/"
    workdir: "./python-project"
    timeout: 3m

  # Go benchmark
  - name: "go-services"
    language: go
    command: "go test -bench=. ./..."
    workdir: "./go-project"
    timeout: 2m

  # Node.js benchmark
  - name: "nodejs-web"
    language: nodejs
    command: "npm run benchmark"
    workdir: "./nodejs-project"
    timeout: 2m

# Execution configuration
execution:
  parallel: 4        # Run up to 4 benchmarks in parallel
  retry: 3           # Retry failed benchmarks 3 times
  failfast: false    # Continue even if some fail

# Output configuration
output:
  formats: [html, json, csv]     # Generate all formats
  directory: ./reports            # Output directory
  timestamp: true                 # Add timestamp to filenames

# Historical storage
storage:
  enabled: true
  path: ./benchflow.db
  retention_days: 90

# Performance regression detection
comparison:
  baseline: "HEAD~1"              # Baseline reference
  threshold: 1.05                 # Fail if >5% slower
  alert_on_regression: true

# Logging
verbose: false
log_file: ./benchflow.log
```

### Supported Languages

| Language | Format | Parser | Status |
|----------|--------|--------|--------|
| **Rust** | cargo bench bencher | ✅ Implemented | Complete |
| **Python** | pytest-benchmark JSON | ✅ Implemented | Complete |
| **Go** | testing.B output | ✅ Implemented | Complete |
| **Node.js** | Benchmark.js text | ✅ Implemented | Complete |

## Project Structure Reference

```
benchflow/
├── cmd/
│   └── benchflow/
│       └── main.go              # CLI entry point
├── internal/
│   ├── cmd/                     # Cobra commands
│   │   ├── root.go             # Root command
│   │   ├── run.go              # Run benchmarks
│   │   ├── compare.go          # Compare results
│   │   └── report.go           # Generate reports
│   └── parser/                  # Benchmark parsers
│       ├── types.go            # Interfaces and types
│       ├── rust.go             # Rust parser (complete)
│       └── rust_test.go        # Rust parser tests
├── testdata/
│   └── rust/                    # Test fixtures
│       ├── cargo_bench_bencher.txt
│       ├── cargo_bench_with_warnings.txt
│       └── cargo_bench_edge_cases.txt
├── examples/
│   └── benchflow.example.yaml   # Example configuration
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
├── README.md                    # Project overview
├── CLAUDE.md                    # Development docs
└── INSTALL.md                   # This file
```

## Troubleshooting

### Build Issues

**Problem**: `go build` fails with module errors

```bash
# Solution: Clean module cache and re-download
go clean -modcache
go mod download
go mod verify
```

**Problem**: Binary not found after `go install`

```bash
# Solution: Ensure $GOPATH/bin is in PATH
echo $GOPATH  # Should show path like /Users/you/go
echo $PATH    # Should include $GOPATH/bin

# Add to PATH if needed (add to ~/.bashrc or ~/.zshrc)
export PATH=$PATH:$(go env GOPATH)/bin
```

### Test Issues

**Problem**: Tests fail with import errors

```bash
# Solution: Ensure you're in the project directory
pwd  # Should be .../benchflow

# Navigate to correct directory
cd /path/to/benchflow
```

**Problem**: Coverage report not generating

```bash
# Solution: Ensure coverage.out file is created
go test -coverprofile=coverage.out ./...
ls coverage.out  # Should exist

# Then generate report
go tool cover -html=coverage.out
```

### Development Issues

**Problem**: Linter not found

```bash
# Solution: Install golangci-lint
# macOS
brew install golangci-lint

# Linux
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```

**Problem**: Tests are slow

```bash
# Solution: Run tests in parallel
go test -parallel 8 ./...

# Or use short mode
go test -short ./...
```

### Getting Help

- **Documentation**: See `CLAUDE.md` for development documentation
- **Issues**: [GitHub Issues](https://github.com/jpequegn/benchflow/issues)
- **CI Logs**: Check [GitHub Actions](https://github.com/jpequegn/benchflow/actions) for test failures

## Summary

### Quick Commands Reference

```bash
# Installation
git clone https://github.com/jpequegn/benchflow.git && cd benchflow
go install ./cmd/benchflow

# Development
go test ./...                    # Run tests
go test -cover ./...             # Run with coverage
go fmt ./... && go vet ./...     # Format and vet
golangci-lint run                # Lint
go build ./cmd/benchflow         # Build

# Usage
benchflow --version              # Show version
benchflow --help                 # Show help
```

### One-Liner Setup

```bash
git clone https://github.com/jpequegn/benchflow.git && cd benchflow && go mod download && go test ./... && go build -o benchflow ./cmd/benchflow && ./benchflow --version
```

### Implementation Status

| Phase | Description | Status | Coverage |
|-------|-------------|--------|----------|
| Phase 1 | Project Foundation & Setup | ✅ Complete | - |
| Phase 2 | Rust Benchmark Parser | ✅ Complete | 82.9% |
| Phase 3 | Parallel Benchmark Execution | ✅ Complete | 94.0% |
| Phase 4 | Result Aggregation & Storage | ✅ Complete | 94.0% |
| Phase 5 | HTML Report Generation | ✅ Complete | 75.6% |
| Phase 6 | Multi-language (Python & Go) | ✅ Complete | 88.3% |
| Phase 7 | Node.js Benchmark Parser | ✅ Complete | 81.2% |

**Overall**: 🚀 **Production Ready** - All 7 phases implemented and tested

---

**Last Updated**: 2025-10-19
**Version**: 0.2.0 (All 7 Phases Complete)
