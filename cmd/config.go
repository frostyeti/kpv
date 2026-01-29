package cmd

import "github.com/spf13/cobra"

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage kpv configuration",
	Long:  `Manage kpv configuration settings stored in the config file.`,
}

func init() {
	RootCmd.AddCommand(configCmd)
}
