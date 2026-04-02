package cmd

import (
	"github.com/spf13/cobra"
)

var secretsCmd = &cobra.Command{
	Use:   "secrets",
	Short: "Manage secrets in KeePass vault",
	Long:  `Manage secrets stored in a KeePass vault database.`,
}

func cloneCommand(base *cobra.Command) *cobra.Command {
	clone := *base
	clone.Aliases = nil
	clone.Hidden = false
	clone.Use = base.Use
	clone.Short = base.Short
	clone.Long = base.Long
	clone.Example = base.Example
	clone.Run = base.Run
	clone.RunE = base.RunE
	clone.PreRun = base.PreRun
	clone.PreRunE = base.PreRunE
	clone.PersistentPreRun = base.PersistentPreRun
	clone.PersistentPreRunE = base.PersistentPreRunE
	clone.PostRun = base.PostRun
	clone.PostRunE = base.PostRunE
	clone.Args = base.Args
	clone.Flags().AddFlagSet(base.Flags())
	clone.PersistentFlags().AddFlagSet(base.PersistentFlags())
	return &clone
}

func init() {
	RootCmd.AddCommand(secretsCmd)
}
