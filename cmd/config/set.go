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
var setCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set one config value in the config file",
	Long: `Set one config value in the config file.

Examples:
  # Gets the default service name
  osv config set <key> <value>
  
  osv config set service myservice`,

	Run: func(cmd *cobra.Command, args []string) {

		if len(args) < 2 {
			utils.Fail("key and value must be provided")
			return
		}

		key := args[0]
		value := args[1]

		if key == "" {
			utils.Fail("key must be not be empty")
			return
		}

		conf, err := utils.GetConfig()
		if err != nil {
			utils.Failf("failed to load config: %v", err)
			return
		}

		if key == "alias" {
			utils.Fail("Use the aliases subcommand to edit an alias")
			return
		}

		conf.Set(key, value)
		if err := conf.Save(); err != nil {
			utils.Failf("failed to save config: %v", err)
			return
		}

		utils.Okf("%s updated", key)
		os.Exit(0)
	},
}

func init() {
	ConfigCmd.AddCommand(setCmd)
}
