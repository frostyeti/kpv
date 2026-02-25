package cmd

import (
	"fmt"
	"sort"

	"github.com/frostyeti/kpv/internal/utils"
	"github.com/gobwas/glob"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var configAliasesCmd = &cobra.Command{
	Use:   "aliases",
	Short: "Manage KeePass path aliases",
	Long:  `Manage KeePass path aliases stored in config.`,
}

var configAliasesSetCmd = &cobra.Command{
	Use:   "set <name> <path>",
	Short: "Set a KeePass path alias",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		path := args[1]

		err := utils.LoadConfig()
		if err != nil {
			utils.Failf("failed to load config: %v", err)
		}

		err = utils.SetAlias(name, path)
		if err != nil {
			utils.Failf("failed to set alias: %v", err)
		}

		utils.Okf("alias %s set to %s", name, path)
	},
}

var configAliasesGetCmd = &cobra.Command{
	Use:   "get <name>",
	Short: "Get a KeePass path alias",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		err := utils.LoadConfig()
		if err != nil {
			utils.Failf("failed to load config: %v", err)
		}

		aliases := viper.GetStringMapString("aliases")
		if path, ok := aliases[name]; ok {
			fmt.Println(path)
		} else {
			utils.Failf("alias %s not found", name)
		}
	},
}

var configAliasesLsCmd = &cobra.Command{
	Use:   "ls [filter]",
	Short: "List KeePass path aliases",
	Run: func(cmd *cobra.Command, args []string) {
		err := utils.LoadConfig()
		if err != nil {
			utils.Failf("failed to load config: %v", err)
		}

		aliases := viper.GetStringMapString("aliases")
		if len(aliases) == 0 {
			return
		}

		var filter glob.Glob
		if len(args) > 0 {
			filter, err = glob.Compile(args[0])
			if err != nil {
				utils.Failf("invalid filter: %v", err)
			}
		}

		var keys []string
		for k := range aliases {
			if filter == nil || filter.Match(k) {
				keys = append(keys, k)
			}
		}

		sort.Strings(keys)
		for _, k := range keys {
			fmt.Printf("%s=%s\n", k, aliases[k])
		}
	},
}

var configAliasesRmCmd = &cobra.Command{
	Use:   "rm <name>",
	Short: "Remove a KeePass path alias",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		err := utils.LoadConfig()
		if err != nil {
			utils.Failf("failed to load config: %v", err)
		}

		aliases := viper.GetStringMapString("aliases")
		if _, ok := aliases[name]; !ok {
			utils.Failf("alias %s not found", name)
		}

		// Delete the alias from the map
		delete(aliases, name)

		// Set it back to viper (which merges, so it doesn't strictly delete)
		// To properly delete, we need to rewrite the map
		viper.Set("aliases", aliases)
		err = viper.WriteConfig()
		if err != nil {
			utils.Failf("failed to remove alias: %v", err)
		}

		utils.Okf("alias %s removed", name)
	},
}

func init() {
	configCmd.AddCommand(configAliasesCmd)
	configAliasesCmd.AddCommand(configAliasesSetCmd)
	configAliasesCmd.AddCommand(configAliasesGetCmd)
	configAliasesCmd.AddCommand(configAliasesLsCmd)
	configAliasesCmd.AddCommand(configAliasesRmCmd)
}
