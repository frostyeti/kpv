package cmd

import (
	"encoding/json"
	"fmt"

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
		if value == nil {
			fmt.Printf("Configuration key '%s' not found.\n", key)
			return
		}

		if m, ok := value.(map[string]interface{}); ok {
			bytes, err := json.Marshal(m)
			if err != nil {
				utils.Failf("Error marshalling config value: %v\n", err)
				return
			}
			fmt.Printf("%s\n", string(bytes))
			return
		}

		if s, ok := value.([]interface{}); ok {
			bytes, err := json.Marshal(s)
			if err != nil {
				utils.Failf("Error marshalling config value: %v\n", err)
				return
			}
			fmt.Printf("%s\n", string(bytes))
			return
		}

		fmt.Printf("%v\n", value)
	},
}

func init() {
	configCmd.AddCommand(configGetCmd)
}
