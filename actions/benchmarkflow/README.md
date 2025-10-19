# Benchflow Performance Testing Action

A GitHub Action for running benchmarks across multiple languages and detecting performance regressions in CI/CD workflows.

## Features

- **Multi-language support**: Rust, Python, Go, Node.js
- **Automatic PR comments**: Performance delta on each PR
- **Regression detection**: Alert on performance regressions
- **Multiple output formats**: HTML, JSON, CSV
- **Artifact management**: Automatic report upload and retention
- **Baseline comparison**: Compare against main branch or custom baseline
- **Configurable thresholds**: Set custom regression detection thresholds

## Quick Start

Add to your workflow:

```yaml
- name: Run Performance Benchmarks
  uses: jpequegn/benchmarkflow-action@v1
  with:
    config: ./benchflow.yaml
    comment-pr: true
    detect-regression: true
```

## Inputs

| Input | Description | Default | Required |
|-------|-------------|---------|----------|
| `config` | Path to benchflow.yaml | `./benchflow.yaml` | No |
| `parallel` | Parallel executions | `4` | No |
| `timeout` | Timeout per benchmark | `10m` | No |
| `format` | Output format (html, json, csv) | `html` | No |
| `comment-pr` | Comment on PR with results | `true` | No |
| `detect-regression` | Detect performance regressions | `true` | No |
| `regression-threshold` | Regression threshold (1.05 = 5% slower) | `1.05` | No |
| `baseline-branch` | Branch for baseline comparison | `main` | No |
| `upload-artifact` | Upload reports as artifact | `true` | No |

## Outputs

| Output | Description |
|--------|-------------|
| `report-path` | Path to generated reports |
| `regression-detected` | Whether regression was detected |
| `performance-delta` | Performance change percentage |
| `report-url` | URL to artifact download |

## Usage Examples

### Basic Usage (PR Benchmarking)

```yaml
name: Performance Tests

on:
  pull_request:
    paths: ['src/**', 'benches/**', 'benchflow.yaml']

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Run Benchmarks
        uses: jpequegn/benchmarkflow-action@v1
        with:
          config: ./benchflow.yaml
          parallel: 4
          comment-pr: true
```

### Advanced Usage (Main Branch Tracking)

```yaml
name: Daily Benchmarks

on:
  push:
    branches: [main]
  schedule:
    - cron: '0 2 * * *'  # Daily at 2 AM UTC

jobs:
  benchmark:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.24'

      - name: Run Benchmarks
        uses: jpequegn/benchmarkflow-action@v1
        with:
          config: ./benchflow.yaml
          parallel: 8
          format: html
          detect-regression: true
          regression-threshold: '1.05'
          upload-artifact: true

      - name: Notify on Regression
        if: failure()
        run: |
          echo "Performance regression detected!"
```

### Multiple Languages

```yaml
steps:
  - uses: actions/setup-go@v4
    with:
      go-version: '1.24'

  - uses: actions-rs/toolchain@v1
    with:
      toolchain: stable

  - uses: actions/setup-python@v4
    with:
      python-version: '3.11'

  - uses: actions/setup-node@v4
    with:
      node-version: '20'

  - name: Run Benchmarks
    uses: jpequegn/benchmarkflow-action@v1
    with:
      config: ./benchflow.yaml
      parallel: 4
```

## Configuration File (benchflow.yaml)

See [benchflow documentation](../../README.md) for full configuration details.

Example with multiple languages:

```yaml
benchmarks:
  - name: rust-algo
    language: rust
    command: cargo bench
    timeout: 5m

  - name: python-data
    language: python
    command: pytest --benchmark-only
    timeout: 3m

  - name: go-service
    language: go
    command: go test -bench=.
    timeout: 2m

  - name: nodejs-web
    language: nodejs
    command: npm run benchmark
    timeout: 2m

execution:
  parallel: 4
  retry: 1

output:
  formats: [html, json, csv]
  directory: ./reports

storage:
  enabled: true
  path: ./benchflow.db
```

## Workflow Patterns

### Pattern 1: PR Performance Review

Every PR automatically gets performance review with:
- Performance delta compared to main
- HTML report as artifact
- Automatic comment with results
- Regression alert if threshold exceeded

### Pattern 2: Daily Trend Tracking

Scheduled workflow runs benchmarks daily:
- Stores historical data in SQLite
- Tracks performance trends over time
- Alerts on significant regressions
- Archives reports for 90 days

### Pattern 3: Commit-based Tracking

On every commit to main:
- Run full benchmark suite
- Store baseline for future comparisons
- Alert team on regressions
- Archive for trend analysis

## PR Comment Format

When `comment-pr: true`, the action posts:

```
## 📊 Performance Benchmark Results

✅ **No regressions detected**

- 📈 [View full report](#)
- 🔗 [Artifacts](#)
- ⏱️ Completed successfully
```

Or if regression detected:

```
## 📊 Performance Benchmark Results

⚠️ **Performance regression detected!**

Performance change: +5.2%

- 📈 [View full report](#)
- 🔗 [Artifacts](#)
- ⚠️ Please address performance regression
```

## Requirements

The runner must have:
- Go 1.24+ (for benchflow binary)
- Language toolchains for benchmarks you run:
  - Rust: `rustc` + `cargo`
  - Python: `python` + `pip`
  - Go: `go`
  - Node.js: `node` + `npm`

Or use language setup actions (see examples above).

## Tips

1. **Use fetch-depth: 0** for proper baseline comparison:
   ```yaml
   - uses: actions/checkout@v4
     with:
       fetch-depth: 0
   ```

2. **Run on specific paths** to avoid unnecessary executions:
   ```yaml
   paths: ['src/**', 'benches/**', 'benchflow.yaml']
   ```

3. **Use matrix for multiple OS/configs**:
   ```yaml
   strategy:
     matrix:
       os: [ubuntu-latest, macos-latest]
       parallel: [4, 8]
   ```

4. **Archive for trend analysis**:
   ```yaml
   retention-days: 90  # Keep for 90 days
   ```

5. **Set appropriate threshold**:
   - 1.05 = 5% slower (strict)
   - 1.10 = 10% slower (moderate)
   - 1.20 = 20% slower (lenient)

## Troubleshooting

**Issue**: Action fails with "benchflow command not found"
- **Solution**: Ensure Go 1.24+ is installed before running action

**Issue**: PR comment not appearing
- **Solution**: Check GITHUB_TOKEN has `pull-requests: write` permission

**Issue**: Regression not detected
- **Solution**: Ensure baseline branch exists and `detect-regression: true`

**Issue**: Reports not uploaded as artifact
- **Solution**: Check `upload-artifact: true` and permissions

## Further Documentation

- [Benchflow README](../../README.md)
- [Installation Guide](../../INSTALL.md)
- [Quick Start](../../QUICK_START.md)

## Support

- Issues: [GitHub Issues](https://github.com/jpequegn/benchflow/issues)
- Discussions: [GitHub Discussions](https://github.com/jpequegn/benchflow/discussions)

## License

MIT License - See LICENSE file for details
