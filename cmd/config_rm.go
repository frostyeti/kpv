package cmd

import (
	"github.com/frostyeti/kpv/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configRmCmd = &cobra.Command{
	Use:   "rm [key]",
	Short: "Remove a configuration setting",
	Long:  `Remove a configuration setting from the kpv config file.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		err := utils.LoadConfig()
		if err != nil {
			cmd.PrintErrln("Error loading config:", err)
			return
		}

		viper.Set(key, nil)

		err = viper.WriteConfig()
		if err != nil {
			cmd.PrintErrln("Error writing config:", err)
			return
		}

		cmd.Printf("Configuration key '%s' removed successfully.\n", key)
	},
}

func init() {
	configCmd.AddCommand(configRmCmd)
}
