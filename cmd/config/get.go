/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package config

import (
	"os"
	"strings"

	"github.com/frostyeti/kpv/internal/utils"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get one config value from the config file",
	Long: `Get one config value from the config file.

Examples:
  # Gets the default service name
  kpv config get service`,

	Run: func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			utils.Fail("key must be provided")
			return
		}

		key := args[0]

		if key == "" {
			utils.Fail("key must be not be empty")
			return
		}

		conf, err := utils.GetConfig()
		if err != nil {
			utils.Failf("getting config path: %v", err)
			return
		}

		if strings.HasPrefix(key, "kpv://") {
			ring, err := utils.OpenKeyring()
			if err != nil {
				utils.Failf("opening keyring: %v", err)
				return
			}

			item, err := ring.Get(key)
			if err != nil {
				utils.Failf("getting %s from os vault: %v", key, err)
				return
			}

			os.Stdout.WriteString(string(item.Data) + "\n")
			os.Exit(0)
		}

		value, ok := conf.Get(key)
		if !ok {
			utils.Failf("%s not set", key)
			return
		}

		os.Stdout.WriteString(value + "\n")
	},
}

func init() {
	ConfigCmd.AddCommand(getCmd)
}
