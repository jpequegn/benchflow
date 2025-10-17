package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// compareCmd represents the compare command
var compareCmd = &cobra.Command{
	Use:   "compare",
	Short: "Compare benchmark results",
	Long: `Compare benchmark results between baseline and current runs.

Example:
  benchflow compare --baseline v1.0.0 --current HEAD`,
	RunE: func(cmd *cobra.Command, args []string) error {
		baseline, _ := cmd.Flags().GetString("baseline")
		current, _ := cmd.Flags().GetString("current")

		fmt.Printf("Comparing benchmarks: %s vs %s\n", baseline, current)
		// TODO: Implement comparison logic (Phase 4)
		return fmt.Errorf("not yet implemented - coming in Phase 4")
	},
}

func init() {
	rootCmd.AddCommand(compareCmd)

	// Compare-specific flags
	compareCmd.Flags().StringP("baseline", "b", "", "baseline benchmark results (required)")
	compareCmd.Flags().StringP("current", "c", "", "current benchmark results (required)")
	compareCmd.Flags().Float64P("threshold", "t", 1.05, "regression threshold (1.05 = 5% slower)")

	compareCmd.MarkFlagRequired("baseline")
	compareCmd.MarkFlagRequired("current")
}
