package aliases

import (
	"github.com/frostyeti/kpv/cmd/config"
	"github.com/spf13/cobra"
)

var aliasesCmd = &cobra.Command{
	Use:   "aliases",
	Short: "Manage configuration aliases",
	Long:  `Manage configuration aliases.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	config.ConfigCmd.AddCommand(aliasesCmd)
}
