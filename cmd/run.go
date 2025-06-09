package main

import (
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run diagnostic sequence",
	Run: func(cmd *cobra.Command, args []string) {
		runDiagnostic()
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
