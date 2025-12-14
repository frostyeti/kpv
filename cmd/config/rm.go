/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package config

import (
	"os"

	"github.com/frostyeti/kpv/internal/utils"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var rmCmd = &cobra.Command{
	Use:   "rm <key>",
	Short: "Remove one config value from the config file",
	Long: `Remove one config value from the config file.

Examples:
  # Gets the default service name
  osv config set <key> <value>
  
  osv config set service myservice`,

	Run: func(cmd *cobra.Command, args []string) {

		if len(args) != 1 {
			utils.Fail("key must be provided")
			return
		}

		key := args[0]

		if key == "" {
			utils.Fail("key must be not be empty")
			return
		}

		if key == "alias" {
			utils.Fail("Use the aliases subcommand to edit an alias")
			return
		}

		conf, err := utils.GetConfig()
		if err != nil {
			utils.Failf("failed to load config: %v", err)
			return
		}

		conf.Remove(key)
		if err := conf.Save(); err != nil {
			utils.Failf("failed to save config: %v", err)
			return
		}

		utils.Okf("%s removed", key)
		os.Exit(0)
	},
}

func init() {
	ConfigCmd.AddCommand(rmCmd)
}
