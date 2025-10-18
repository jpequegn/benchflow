package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jpequegn/benchflow/internal/executor"
	"github.com/jpequegn/benchflow/internal/parser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run benchmarks from configuration",
	Long: `Run all benchmarks defined in the configuration file.

Example:
  benchflow run --config benchflow.yaml
  benchflow run --name rust-sort --parallel 2`,
	RunE: runBenchmarks,
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Run-specific flags
	runCmd.Flags().StringP("name", "n", "", "run specific benchmark by name")
	runCmd.Flags().IntP("parallel", "p", 0, "number of parallel benchmark executions (default from config)")
	runCmd.Flags().DurationP("timeout", "t", 0, "timeout for each benchmark (0 = no timeout)")
}

func runBenchmarks(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Load configuration
	configs, err := loadBenchmarkConfigs(cmd)
	if err != nil {
		return fmt.Errorf("failed to load benchmark configs: %w", err)
	}

	if len(configs) == 0 {
		return fmt.Errorf("no benchmarks configured")
	}

	slog.Info("Loaded benchmark configurations", "count", len(configs))

	// Create parser registry
	registry := executor.NewParserRegistry()
	registry.RegisterParser("rust", parser.NewRustParser())
	// TODO: Register Python and Go parsers when implemented

	// Create execution config
	execConfig := &executor.ExecutionConfig{
		Parallel: viper.GetInt("execution.parallel"),
		Retry:    viper.GetInt("execution.retry"),
		FailFast: viper.GetBool("execution.failfast"),
	}

	// Override parallel from flag if provided
	if parallel, _ := cmd.Flags().GetInt("parallel"); parallel > 0 {
		execConfig.Parallel = parallel
	}

	// Default to 4 parallel executions if not configured
	if execConfig.Parallel <= 0 {
		execConfig.Parallel = 4
	}

	slog.Info("Execution configuration",
		"parallel", execConfig.Parallel,
		"retry", execConfig.Retry,
		"failfast", execConfig.FailFast)

	// Create executor with progress handler
	progressHandler := func(event *executor.ProgressEvent) {
		switch event.Type {
		case executor.EventStarted:
			slog.Info("Started", "benchmark", event.Config.Name)
		case executor.EventRetrying:
			slog.Warn("Retrying",
				"benchmark", event.Config.Name,
				"attempt", event.Result.Attempts,
				"error", event.Error)
		case executor.EventCompleted:
			slog.Info("Completed",
				"benchmark", event.Config.Name,
				"results", len(event.Result.Suite.Results),
				"duration", event.Result.Duration.Round(time.Millisecond))
		case executor.EventFailed:
			slog.Error("Failed",
				"benchmark", event.Config.Name,
				"attempts", event.Result.Attempts,
				"error", event.Error)
		case executor.EventCancelled:
			slog.Warn("Cancelled", "benchmark", event.Config.Name)
		}
	}

	exec := executor.NewExecutor(progressHandler)

	// Execute benchmarks
	slog.Info("Starting benchmark execution...")
	startTime := time.Now()

	results, err := exec.ExecuteBatch(ctx, configs, execConfig, registry)
	duration := time.Since(startTime)

	// Print summary
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "═══════════════════════════════════════════\n")
	fmt.Fprintf(os.Stderr, "  Benchmark Execution Summary\n")
	fmt.Fprintf(os.Stderr, "═══════════════════════════════════════════\n")
	fmt.Fprintf(os.Stderr, "Total benchmarks: %d\n", len(results))
	fmt.Fprintf(os.Stderr, "Total duration: %v\n", duration.Round(time.Millisecond))

	successCount := 0
	failedCount := 0
	totalResults := 0

	for _, result := range results {
		if result.Error == nil {
			successCount++
			totalResults += len(result.Suite.Results)
		} else {
			failedCount++
		}
	}

	fmt.Fprintf(os.Stderr, "Successful: %d\n", successCount)
	fmt.Fprintf(os.Stderr, "Failed: %d\n", failedCount)
	fmt.Fprintf(os.Stderr, "Total results: %d\n", totalResults)
	fmt.Fprintf(os.Stderr, "═══════════════════════════════════════════\n\n")

	// Print detailed results
	for _, result := range results {
		if result.Error != nil {
			fmt.Fprintf(os.Stderr, "❌ %s: %v\n", result.Config.Name, result.Error)
			continue
		}

		fmt.Fprintf(os.Stderr, "✅ %s (%d results)\n", result.Config.Name, len(result.Suite.Results))
		for _, r := range result.Suite.Results {
			fmt.Fprintf(os.Stderr, "   • %s: %v (±%v)\n",
				r.Name,
				r.Time.Round(time.Nanosecond),
				r.StdDev.Round(time.Nanosecond))
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	if err != nil {
		return fmt.Errorf("batch execution failed: %w", err)
	}

	if failedCount > 0 {
		return fmt.Errorf("%d benchmark(s) failed", failedCount)
	}

	return nil
}

// loadBenchmarkConfigs loads benchmark configurations from viper
func loadBenchmarkConfigs(cmd *cobra.Command) ([]*executor.BenchmarkConfig, error) {
	// Get benchmarks from config
	var rawBenchmarks []map[string]interface{}
	if err := viper.UnmarshalKey("benchmarks", &rawBenchmarks); err != nil {
		return nil, fmt.Errorf("failed to unmarshal benchmarks: %w", err)
	}

	if len(rawBenchmarks) == 0 {
		return nil, fmt.Errorf("no benchmarks defined in configuration")
	}

	// Check if specific benchmark was requested
	nameFilter, _ := cmd.Flags().GetString("name")

	var configs []*executor.BenchmarkConfig
	for _, b := range rawBenchmarks {
		name, _ := b["name"].(string)
		language, _ := b["language"].(string)
		command, _ := b["command"].(string)
		workdir, _ := b["workdir"].(string)

		// Skip if name filter is set and doesn't match
		if nameFilter != "" && name != nameFilter {
			continue
		}

		// Parse timeout
		var timeout time.Duration
		if timeoutStr, ok := b["timeout"].(string); ok {
			timeout, _ = time.ParseDuration(timeoutStr)
		}

		// Override timeout from flag if provided
		if flagTimeout, _ := cmd.Flags().GetDuration("timeout"); flagTimeout > 0 {
			timeout = flagTimeout
		}

		config := &executor.BenchmarkConfig{
			Name:     name,
			Language: language,
			Command:  command,
			WorkDir:  workdir,
			Timeout:  timeout,
		}

		configs = append(configs, config)
	}

	if nameFilter != "" && len(configs) == 0 {
		return nil, fmt.Errorf("benchmark not found: %s", nameFilter)
	}

	return configs, nil
}
