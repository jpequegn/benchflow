# Benchflow Quick Start Guide

Get up and running with benchflow in 5 minutes!

## Installation (1 minute)

```bash
# Clone repository
git clone https://github.com/jpequegn/benchflow.git
cd benchflow

# Build binary
go build -o benchflow ./cmd/benchflow

# Verify
./benchflow --version
```

## Setup (2 minutes)

Create a file called `benchflow.yaml`:

```yaml
benchmarks:
  - name: "example-rust"
    language: rust
    command: "cargo bench --bench my_bench"
    timeout: 5m

  - name: "example-python"
    language: python
    command: "pytest --benchmark-only tests/"
    timeout: 3m

execution:
  parallel: 2
  retry: 1

output:
  formats: [html, json]
  directory: ./reports

storage:
  enabled: true
  path: ./benchflow.db
```

## Run Benchmarks (1 minute)

```bash
# Run all benchmarks
./benchflow run --config benchflow.yaml

# Or specific benchmark
./benchflow run --name example-rust

# Verbose output
./benchflow --verbose run --config benchflow.yaml
```

## View Results (1 minute)

```bash
# Results are in ./reports/
# - report.html - Interactive dashboard (open in browser)
# - results.json - Raw data
# - results.csv - Spreadsheet format

# Open report (macOS)
open reports/report.html

# Open report (Linux)
xdg-open reports/report.html

# Open report (Windows)
start reports/report.html
```

---

## Supported Benchmark Formats

### Rust
```bash
# Command
cargo bench

# Example config
language: rust
command: "cargo bench --bench my_bench"
```

### Python
```bash
# Command
pytest --benchmark-only

# Example config
language: python
command: "pytest --benchmark-only tests/"
```

### Go
```bash
# Command
go test -bench=.

# Example config
language: go
command: "go test -bench=. ./..."
```

### Node.js
```bash
# Command (Benchmark.js)
npm run benchmark

# Example config
language: nodejs
command: "npm run benchmark"
```

---

## Common Commands

```bash
# Show help
./benchflow --help

# Show version
./benchflow --version

# Run with custom parallelism
./benchflow run --config benchflow.yaml --parallel 4

# Run with timeout
./benchflow run --config benchflow.yaml --timeout 10m

# Compare results
./benchflow compare --baseline HEAD~1 --current HEAD

# Generate report
./benchflow report --format html --output report.html
```

---

## Troubleshooting

**Problem**: `go: command not found`
- **Solution**: Install Go 1.24+ from https://golang.org/dl

**Problem**: Binary not found after build
- **Solution**: Use full path: `./benchflow` (current directory) or `go run ./cmd/benchflow`

**Problem**: Configuration not found
- **Solution**: Use correct path: `--config ./benchflow.yaml` or check current directory

**Problem**: No results generated
- **Solution**: Check that benchmarks actually exist and are passing. Run with `--verbose` flag

---

## Next Steps

- Read [INSTALL.md](INSTALL.md) for detailed setup guide
- Check [README.md](README.md) for full feature list
- See [examples/](examples/) for example configurations
- Review [CLAUDE.md](CLAUDE.md) for development guide

---

**For detailed documentation**, see:
- **Installation**: [INSTALL.md](INSTALL.md)
- **Features**: [README.md](README.md)
- **Development**: [CLAUDE.md](CLAUDE.md)
