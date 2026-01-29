package cmd

import (
	"github.com/frostyeti/kpv/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configSetCmd = &cobra.Command{
	Use:   "set [key] [value]",
	Short: "Set configuration values",
	Long:  `Set various configuration values for kpv.`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			utils.Failf("exactly two arguments required: key and value\n")
			return
		}

		key := args[0]
		value := args[1]

		err := utils.LoadConfig()
		if err != nil {
			utils.Failf("failed to load config: %v\n", err)
			return
		}

		viper.Set(key, value)
		err = viper.WriteConfig()
		if err != nil {
			utils.Failf("failed to write config: %v\n", err)
			return
		}
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
}
