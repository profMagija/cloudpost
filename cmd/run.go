/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	localrunner "github.com/profMagija/cloudpost/local_runner"

	"github.com/spf13/cobra"
)

var run_verbose *bool

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the project using LocalRunner",
	Long:  `Run the project using LocalRunner`,
	Run: func(cmd *cobra.Command, args []string) {
		localrunner.Verbose = *run_verbose
		localrunner.Init(flock)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
	run_verbose = runCmd.Flags().BoolP("verbose", "v", false, "Use verbose logging")
}
