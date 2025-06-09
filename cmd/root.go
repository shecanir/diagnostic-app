package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

var planFlag string

var rootCmd = &cobra.Command{
	Use:   "diagnostic",
	Short: "Run DNS diagnostic tool",
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&planFlag, "plan", "p", "", "Select plan (Free or Pro)")
}

func printSelectedPlan() {
	fmt.Println("Selected plan:", planFlag)
}
