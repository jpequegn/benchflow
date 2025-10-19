package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// reportCmd represents the report command
var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate benchmark reports",
	Long: `Generate benchmark reports in various formats (HTML, JSON, CSV).

Example:
  benchflow report --format html --output report.html`,
	RunE: func(cmd *cobra.Command, args []string) error {
		format, _ := cmd.Flags().GetString("format")
		output, _ := cmd.Flags().GetString("output")
		input, _ := cmd.Flags().GetString("input")

		if input == "" {
			return fmt.Errorf("input flag is required (use --input to specify benchmark results)")
		}

		fmt.Printf("Generating %s report from %s: %s\n", format, input, output)
		fmt.Println("Use 'benchflow run' to generate benchmark results")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(reportCmd)

	// Report-specific flags
	reportCmd.Flags().StringP("format", "f", "html", "report format (html, json, csv)")
	reportCmd.Flags().StringP("output", "o", "", "output file path (required)")
	reportCmd.Flags().StringP("input", "i", "", "input benchmark results file")

	_ = reportCmd.MarkFlagRequired("output")
	_ = reportCmd.MarkFlagRequired("input")
}
