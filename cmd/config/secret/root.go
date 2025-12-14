package secret

import (
	"github.com/frostyeti/kpv/cmd/config"
	"github.com/spf13/cobra"
)

var secretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Manage secrets in the OS secret store for KeePass databases",
	Long:  `Manage secrets stored in the OS secret store for KeePass databases.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func init() {
	config.ConfigCmd.AddCommand(secretCmd)
}
