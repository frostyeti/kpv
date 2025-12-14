/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:     "kpv",
	Version: VERSION,
	Short:   "KeePass vault cli: a tool for enabling automation secrets with KeePass",
	Long: `KeePass vault cli: a tool for enabling automation secrets with KeePass.

The "ensure" command will get or create a secret in the KeePass vault if it does not exist.

The "get" command supports getting a secret from the KeePass vault in multiple formats
including json, env, bash export, powershell and more.

kpv can leverage the os secret store to get the KeePass vault password for easier
automation access.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := RootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	flags := RootCmd.PersistentFlags()

	vault := os.Getenv("KPV_VAULT")
	password := os.Getenv("KPV_PASSWORD")
	flags.StringP("vault", "V", vault, "The KeePass vault name or path")

	flags.StringP("vault-password", "P", password, "The KeePass vault password")

	passwordFile := os.Getenv("KPV_PASSWORD_FILE")
	flags.StringP("vault-password-file", "F", passwordFile, "The to file containing the KeePass vault password. Can be set with KPV_PASSWORD_FILE env var")

	vaultOsSecret := os.Getenv("KPV_VAULT_OS_SECRET")
	flags.String("vault-os-secret", vaultOsSecret, "The OS secret store key to get the KeePass vault password from. Can be set with KPV_VAULT_OS_SECRET env var")

	RootCmd.CompletionOptions.DisableDefaultCmd = true

	RootCmd.SetHelpCommand(&cobra.Command{
		Use:    "no-help",
		Hidden: true,
	})
}
