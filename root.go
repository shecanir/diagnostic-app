package main

import (
	"github.com/spf13/cobra"
)

// PlanFlag stores the selected diagnostic plan from the CLI flag.
var PlanFlag string

var rootCmd = &cobra.Command{
	Use:   "diagnostic",
	Short: "Run DNS diagnostic tool",
	Run: func(cmd *cobra.Command, args []string) {
		runDiagnostic()
	},
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&PlanFlag, "plan", "p", "", "Select plan (Free or Pro)")
}
