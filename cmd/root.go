/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"os"
	"path"

	"github.com/profMagija/cloudpost/config"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var arg_Root *string

var flock *config.Flock

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "cloudpost",
	Short: "CloudPost -- a long-awaiter, never completed, cloud all-in-one",
	Long: `CloudPost is a CLI tool and library for creating, deploying, managing, tracking,
health-checking, debugging, ruining and ultimately disposing of your cloud 
infrastructure. It will make your productivity go up, put a smile on your face,
and probably rack up your cloud bill.`,

	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		data, err := os.ReadFile(path.Join(*arg_Root, "cloudpost.yml"))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return
			} else {
				panic(err)
			}
		}

		flock = new(config.Flock)
		err = yaml.Unmarshal(data, flock)
		if err != nil {
			panic(err)
		}

		flock.Root, _ = cmd.Flags().GetString("root")
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	arg_Root = rootCmd.PersistentFlags().String("root", ".", "Root of the project")
}
