/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/frostyeti/kpv/internal/utils"
)

// secretsRmCmd represents the rm command
var secretsRmCmd = &cobra.Command{
	Use:     "rm",
	Aliases: []string{"remove"},
	Short:   "Remove one or more secrets from KeePass vault",
	Long: `Remove (delete) one or more secrets from a KeePass vault.

You will be prompted to confirm deletion unless --yes is specified.

Examples:
  # Remove a single secret (with confirmation)
  kpv rm --key my-secret

  # Remove multiple secrets
  kpv rm --key secret1 --key secret2 --key secret3

  # Remove secrets without confirmation
  kpv rm --key my-secret --yes

  # Remove secrets with short flags
  kpv rm -k secret1 -k secret2 -y

  # Use a specific vault
  kpv rm --vault myvault --key my-secret -y`,

	Run: func(cmd *cobra.Command, args []string) {
		keys, _ := cmd.Flags().GetStringSlice("key")
		yes, _ := cmd.Flags().GetBool("yes")

		// Validate that at least one key is provided
		if len(keys) == 0 {
			utils.Failf("at least one --key must be provided\n")
			return
		}

		kdbx, _, err := utils.OpenKeePass(cmd)
		if err != nil {
			utils.Failf("opening KeePass vault failed: %v\n", err)
			return
		}

		// Prompt for confirmation unless --yes is specified
		if !yes {
			fmt.Printf("You are about to delete %d secret(s):\n", len(keys))
			for _, key := range keys {
				fmt.Printf("  - %s\n", key)
			}
			fmt.Print("\nDo you want to continue? [y/N]: ")

			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				utils.Failf("reading confirmation failed: %v\n", err)
				return
			}

			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" {
				utils.Warnf("operation cancelled\n")
				os.Exit(1)
			}
		}

		root := kdbx.Root()
		deletedCount := 0

		// Delete each secret
		for _, key := range keys {
			entry := kdbx.FindEntry(key)
			if entry == nil {
				utils.Failf("secret %s not found\n", key)
				continue
			}

			root.RmEntry(entry)
			deletedCount++
			utils.Okf("deleted secret: %s\n", key)
		}

		if deletedCount > 0 {
			err = kdbx.Save()
			if err != nil {
				utils.Failf("saving KeePass vault failed: %v\n", err)
				return
			}
		}

		if deletedCount == len(keys) {
			fmt.Printf("\nSuccessfully deleted %d secret(s)\n", deletedCount)
		} else {
			fmt.Printf("\nDeleted %d out of %d secret(s)\n", deletedCount, len(keys))
		}
	},
}

func init() {
	secretsCmd.AddCommand(secretsRmCmd)

	secretsRmCmd.Flags().StringSliceP("key", "k", []string{}, "Name of secret(s) to remove (can be specified multiple times)")
	secretsRmCmd.MarkFlagRequired("key")

	secretsRmCmd.Flags().BoolP("yes", "y", false, "Skip confirmation prompt")
}
