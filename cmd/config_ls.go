package cmd

import (
	"fmt"
	"sort"

	"github.com/frostyeti/kpv/internal/utils"
	"github.com/gobwas/glob"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configLsCmd = &cobra.Command{
	Use:   "ls [filter]",
	Short: "List configuration keys",
	Run: func(cmd *cobra.Command, args []string) {
		err := utils.LoadConfig()
		if err != nil {
			utils.Failf("failed to load config: %v", err)
		}

		keys := viper.AllKeys()
		if len(keys) == 0 {
			return
		}

		var filter glob.Glob
		if len(args) > 0 {
			filter, err = glob.Compile(args[0])
			if err != nil {
				utils.Failf("invalid filter: %v", err)
			}
		}

		var matchedKeys []string
		for _, k := range keys {
			if filter == nil || filter.Match(k) {
				matchedKeys = append(matchedKeys, k)
			}
		}

		sort.Strings(matchedKeys)
		for _, k := range matchedKeys {
			val := viper.Get(k)
			// Avoid printing large objects or complex types, just convert to string
			fmt.Printf("%s=%v\n", k, val)
		}
	},
}

func init() {
	configCmd.AddCommand(configLsCmd)
}
