/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"cloudpost/config"
	"fmt"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all the components in the project",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Components:\n")
		for _, v := range flock.Components {
			switch r := v.(type) {
			case *config.Function:
				fmt.Printf("  \x1b[32mFunction\x1b[m %s:\n", v.GetName())
				if r.Entry == "" {
					r.Entry = "main"
				}
				fmt.Printf("    Entry: %s\n", r.Entry)
				if r.TriggerTopic != "" {
					fmt.Printf("    Trigger Topic: %s\n", r.TriggerTopic)
				}
			case *config.Container:
				fmt.Printf("  \x1b[32mContainer\x1b[m %s:\n", v.GetName())
				fmt.Printf("    Entry: %s\n", r.Entry)
				if r.TriggerTopic != "" {
					fmt.Printf("    Trigger Topic : %s\n", r.TriggerTopic)
					fmt.Printf("    Trigger Path  : %s\n", r.TriggerPath)
				}
			case *config.Bucket:
				fmt.Printf("  \x1b[32mBucket\x1b[m %s:\n", v.GetName())
			default:
				fmt.Printf("invalid flock component spec: unknown type %T\n", r)
			}
			fmt.Println()
		}
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
