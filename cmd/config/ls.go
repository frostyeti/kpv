/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package config

import (
	"os"

	"github.com/frostyeti/kpv/internal/utils"
	"github.com/gobwas/glob"
	"github.com/spf13/cobra"
)

// getCmd represents the get command
var lsCmd = &cobra.Command{
	Use:   "ls [filter]",
	Short: "List all config values from the config file",
	Long: `List all config values from the config file.

Examples:
  # List all config names
  kpv config ls
  
  # List config names matching a glob pattern
  kpv config ls "keychain.*"`,

	Run: func(cmd *cobra.Command, args []string) {

		filter := ""
		if len(args) > 0 {
			filter = args[0]
		}

		conf, err := utils.GetConfig()
		if err != nil {
			utils.Failf("getting config path: %v", err)
			return
		}

		keys := conf.Keys()

		var matcher glob.Glob
		if filter != "" {
			matcher, err = glob.Compile(filter)
			if err != nil {
				utils.Failf("compiling glob pattern failed: %v\n", err)
				return
			}

			for _, key := range keys {
				if matcher.Match(key) {
					os.Stdout.WriteString(key + "\n")
				}
			}
			os.Exit(0)
		}

		for _, key := range keys {
			os.Stdout.WriteString(key + "\n")
		}

		os.Exit(0)
	},
}

func init() {
	ConfigCmd.AddCommand(lsCmd)
}
