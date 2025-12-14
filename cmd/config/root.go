/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package config

import (
	"github.com/frostyeti/kpv/cmd"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration values",
	Long:  `Manage configuration values.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	cmd.RootCmd.AddCommand(ConfigCmd)
}
