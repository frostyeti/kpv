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
  kpv config rm service
  
  kpv config rm --key service
  
  `,

	Run: func(cmd *cobra.Command, args []string) {

		key := ""
		if len(args) > 0 {
			key = args[0]
		}

		keyInline, _ := cmd.Flags().GetString("key")
		if keyInline != "" {
			key = keyInline
		}

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
	rmCmd.Flags().StringP("key", "k", "", "The config key to remove (exclusive with positional argument)")
}
