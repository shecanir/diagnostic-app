package main

import (
	"github.com/spf13/cobra"
)

// PlanFlag stores the selected diagnostic plan from the CLI flag.
var PlanFlag string

var rootCmd = &cobra.Command{
	Use:   "diagnostic",
	Short: "Run DNS diagnostic tool",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&PlanFlag, "plan", "p", "", "Select plan (Free or Pro)")
}
