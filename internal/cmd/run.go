package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run benchmarks from configuration",
	Long: `Run all benchmarks defined in the configuration file.

Example:
  benchflow run --config benchflow.yaml`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Running benchmarks...")
		// TODO: Implement benchmark execution (Phase 3)
		return fmt.Errorf("not yet implemented - coming in Phase 3")
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Run-specific flags
	runCmd.Flags().StringP("name", "n", "", "run specific benchmark by name")
	runCmd.Flags().IntP("parallel", "p", 4, "number of parallel benchmark executions")
	runCmd.Flags().DurationP("timeout", "t", 0, "timeout for each benchmark (0 = no timeout)")
}
