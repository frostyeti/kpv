package cmd

import (
	"encoding/json"
	"os"
	"strings"

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
			utils.Failf("Error loading config: %v\n", err)
			return
		}

		settings := viper.AllSettings()

		// Navigate and delete the key
		parts := strings.Split(key, ".")
		deleteKey(settings, parts)

		// Viper doesn't have an Unset method, and WriteConfig merges,
		// so we write the JSON ourselves to completely remove the key.
		file := viper.ConfigFileUsed()
		if file != "" {
			b, err := json.MarshalIndent(settings, "", "  ")
			if err != nil {
				utils.Failf("Error marshalling config: %v\n", err)
			}
			err = os.WriteFile(file, b, 0600)
			if err != nil {
				utils.Failf("Error writing config: %v\n", err)
			}
		}

		utils.Okf("Configuration key '%s' removed successfully.", key)
	},
}

func deleteKey(m map[string]interface{}, parts []string) {
	if len(parts) == 0 {
		return
	}

	key := parts[0]
	if len(parts) == 1 {
		delete(m, key)
		return
	}

	if nested, ok := m[key].(map[string]interface{}); ok {
		deleteKey(nested, parts[1:])
		if len(nested) == 0 {
			delete(m, key)
		}
	}
}

func init() {
	configCmd.AddCommand(configRmCmd)
}
