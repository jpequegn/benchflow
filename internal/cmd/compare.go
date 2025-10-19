package cmd

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/jpequegn/benchflow/internal/comparator"
	"github.com/jpequegn/benchflow/internal/reporter"
	"github.com/spf13/cobra"
)

// compareCmd represents the compare command
var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare benchmark results",
	Long: `Compare benchmark results between baseline and current runs using statistical analysis.

Detects performance regressions, improvements, and statistically significant changes.
Supports JSON and CSV input formats.

Example:
  benchflow compare --baseline baseline.json --current current.json
  benchflow compare --baseline baseline.json --current current.json --format html --output report.html
  benchflow compare -b main.json -c feature.json -f markdown`,
	RunE: compareBenchmarks,
}

func init() {
	rootCmd.AddCommand(compareCmd)

	// Compare-specific flags
	compareCmd.Flags().StringP("baseline", "b", "", "path to baseline benchmark results (JSON or CSV) (required)")
	compareCmd.Flags().StringP("current", "c", "", "path to current benchmark results (JSON or CSV) (required)")
	compareCmd.Flags().Float64P("threshold", "t", 1.05, "regression threshold multiplier (default: 1.05 = 5% slower)")
	compareCmd.Flags().Float64P("confidence", "C", 0.95, "statistical confidence level (default: 0.95 = 95%)")
	compareCmd.Flags().StringP("format", "f", "markdown", "output format: markdown, html, or json (default: markdown)")
	compareCmd.Flags().StringP("output", "o", "", "output file path (default: stdout)")

	_ = compareCmd.MarkFlagRequired("baseline")
	_ = compareCmd.MarkFlagRequired("current")
}

func compareBenchmarks(cmd *cobra.Command, args []string) error {
	// Get flags
	baselinePath, _ := cmd.Flags().GetString("baseline")
	currentPath, _ := cmd.Flags().GetString("current")
	threshold, _ := cmd.Flags().GetFloat64("threshold")
	confidence, _ := cmd.Flags().GetFloat64("confidence")
	format, _ := cmd.Flags().GetString("format")
	outputPath, _ := cmd.Flags().GetString("output")

	// Validate format
	if format != "markdown" && format != "html" && format != "json" {
		return fmt.Errorf("invalid format: %s (must be markdown, html, or json)", format)
	}

	// Validate confidence level
	if confidence <= 0 || confidence >= 1 {
		return fmt.Errorf("confidence level must be between 0 and 1 (e.g., 0.95 for 95%%)")
	}

	// Validate threshold
	if threshold <= 1.0 {
		return fmt.Errorf("threshold must be greater than 1.0 (e.g., 1.05 for 5%% regression)")
	}

	slog.Info("Loading benchmark suites",
		"baseline", baselinePath,
		"current", currentPath)

	// Load baseline suite
	baselineSuite, err := LoadBenchmarkSuite(baselinePath)
	if err != nil {
		return fmt.Errorf("failed to load baseline: %w", err)
	}

	slog.Info("Loaded baseline suite", "benchmarks", len(baselineSuite.Results))

	// Load current suite
	currentSuite, err := LoadBenchmarkSuite(currentPath)
	if err != nil {
		return fmt.Errorf("failed to load current suite: %w", err)
	}

	slog.Info("Loaded current suite", "benchmarks", len(currentSuite.Results))

	// Create comparator
	comp := comparator.NewBasicComparator()
	comp.RegressionThreshold = threshold
	comp.ConfidenceLevel = confidence

	slog.Info("Performing comparison",
		"threshold", threshold,
		"confidence", confidence)

	// Compare suites
	result := comp.Compare(baselineSuite, currentSuite)

	slog.Info("Comparison complete",
		"total", result.Summary.TotalComparisons,
		"regressions", result.Summary.Regressions,
		"improvements", result.Summary.Improvements,
		"significant", result.Summary.SignificantChanges)

	// Generate report
	var report string
	var err2 error

	compReporter := reporter.NewBasicComparisonReporter()

	switch format {
	case "markdown":
		report, err2 = compReporter.GenerateMarkdown(result)
	case "html":
		report, err2 = compReporter.GenerateHTML(result)
	case "json":
		report, err2 = compReporter.GenerateJSON(result)
	}

	if err2 != nil {
		return fmt.Errorf("failed to generate %s report: %w", format, err2)
	}

	// Output report
	if outputPath != "" {
		if err := os.WriteFile(outputPath, []byte(report), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		slog.Info("Report written", "path", outputPath)
		fmt.Fprintf(os.Stderr, "Report saved to: %s\n", outputPath)
	} else {
		fmt.Println(report)
	}

	// Print summary to stderr
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "═══════════════════════════════════════════\n")
	fmt.Fprintf(os.Stderr, "  Comparison Summary\n")
	fmt.Fprintf(os.Stderr, "═══════════════════════════════════════════\n")
	fmt.Fprintf(os.Stderr, "Total Comparisons: %d\n", result.Summary.TotalComparisons)
	fmt.Fprintf(os.Stderr, "Regressions:      %d\n", result.Summary.Regressions)
	fmt.Fprintf(os.Stderr, "Improvements:     %d\n", result.Summary.Improvements)
	fmt.Fprintf(os.Stderr, "Significant:      %d\n", result.Summary.SignificantChanges)
	fmt.Fprintf(os.Stderr, "Average Delta:    %.2f%%\n", result.Summary.AverageDelta)
	fmt.Fprintf(os.Stderr, "Max Delta:        %.2f%%\n", result.Summary.MaxDelta)
	fmt.Fprintf(os.Stderr, "Min Delta:        %.2f%%\n", result.Summary.MinDelta)
	fmt.Fprintf(os.Stderr, "═══════════════════════════════════════════\n")

	// Exit with error if regressions detected
	if result.Summary.Regressions > 0 {
		fmt.Fprintf(os.Stderr, "\n⚠️  Performance regressions detected!\n")
		for _, name := range result.Regressions {
			fmt.Fprintf(os.Stderr, "  • %s\n", name)
		}
		return fmt.Errorf("performance regressions detected (%d)", result.Summary.Regressions)
	}

	return nil
}
