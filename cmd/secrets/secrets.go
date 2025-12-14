package cmd

import (
	"github.com/spf13/cobra"

	"github.com/frostyeti/kpv/cmd"
)

var secretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Manage secrets in KeePass vault",
	Long:  `Manage secrets stored in a KeePass vault database.`,
}

func init() {
	cmd.RootCmd.AddCommand(secretsCmd)
}
