/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	localrunner "cloudpost/local_runner"

	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the project using LocalRunner",
	Long:  `Run the project using LocalRunner`,
	Run: func(cmd *cobra.Command, args []string) {
		localrunner.Init(flock)
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
