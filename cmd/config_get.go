package cmd

import (
	"encoding/json"

	"github.com/frostyeti/kpv/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configGetCmd = &cobra.Command{
	Use:   "get [key]",
	Short: "Get a configuration setting",
	Long:  `Get a configuration setting from the kpv config file.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]

		err := utils.LoadConfig()
		if err != nil {
			cmd.PrintErrln("Error loading config:", err)
			return
		}

		value := viper.Get(key)
		// if

		if value.(map[string]interface{}) != nil {
			map0 := value.(map[string]interface{})
			bytes, err := json.Marshal(map0)
			if err != nil {
				cmd.PrintErrln("Error marshalling config value:", err)
				return
			}
			cmd.Printf("%s\n", string(bytes))
			return
		}

		if value.([]interface{}) != nil {
			slice0 := value.([]interface{})
			bytes, err := json.Marshal(slice0)
			if err != nil {
				cmd.PrintErrln("Error marshalling config value:", err)
				return
			}
			cmd.Printf("%s\n", string(bytes))
			return
		}

		if value == nil {
			cmd.Printf("Configuration key '%s' not found.\n", key)
			return
		}

		cmd.Printf("%v\n", value)
	},
}

func init() {
	configCmd.AddCommand(configGetCmd)
}
