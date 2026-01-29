/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/frostyeti/kpv/internal/keepass"
	"github.com/frostyeti/kpv/internal/utils"
	"github.com/spf13/cobra"
)

// secretsEnsureCmd represents the ensure command
var secretsEnsureCmd = &cobra.Command{
	Use:   "ensure [key]",
	Short: "Ensure a secret exists in KeePass vault, creating it if needed",
	Long: `Get a secret from KeePass vault, or create it if it doesn't exist.

The key can be provided as an argument or via the --key flag.

If the secret already exists, its value is returned.
If the secret does not exist, a new random secret is generated and stored.

Generation options control the created secret:
  --size       Size of the secret in characters (default: 16)
  --no-upper   Exclude uppercase letters
  --no-lower   Exclude lowercase letters
  --no-digits  Exclude digits
  --no-special Exclude special characters
  --special    Specify custom special characters (default: @_-{}|#!~:^)
  --chars      Use only these specific characters (overrides other character options)

Examples:
  # Ensure a secret exists (default 16-character random)
  kpv ensure --key my-secret

  # Ensure with custom size
  kpv ensure --key my-secret --size 32

  # Ensure with only lowercase and digits
  kpv ensure --key my-secret -U -S

  # Ensure with custom characters
  kpv ensure my-secret --chars "abc123!@#"

  # Use a specific vault
  kpv ensure --vault myvault --key my-secret`,

	Run: func(cmd *cobra.Command, args []string) {
		key, _ := cmd.Flags().GetString("key")

		if len(args) > 0 {
			if len(args[0]) > 0 {
				key = args[0]
			}
		}

		// Validate key is provided
		if key == "" {
			cmd.PrintErrf("Error: --key must be provided\n")
			os.Exit(1)
		}

		kdbx, _, err := utils.OpenKeePass(cmd)
		if err != nil {
			cmd.PrintErrf("Error opening KeePass vault: %v\n", err)
			os.Exit(1)
		}

		// Try to get the secret
		entry := kdbx.FindEntry(key)
		if entry != nil {
			// Secret exists, return its value
			fmt.Println(entry.GetPassword())
			os.Exit(0)
		}

		// Secret doesn't exist, generate a new one
		secretValue, err := generateSecret(cmd)
		if err != nil {
			cmd.PrintErrf("Error generating secret: %v\n", err)
			os.Exit(1)
		}

		if secretValue == "" {
			cmd.PrintErrf("Error: generated secret is empty\n")
			os.Exit(1)
		}

		// Set the secret
		kdbx.UpsertEntry(key, func(entry *keepass.Entry) {
			entry.SetPassword(secretValue)
		})

		err = kdbx.Save()
		if err != nil {
			cmd.PrintErrf("Error saving KeePass vault: %v\n", err)
			os.Exit(1)
		}

		// Return the generated value
		fmt.Println(secretValue)
		os.Exit(0)
	},
}

func init() {
	secretsCmd.AddCommand(secretsEnsureCmd)

	secretsEnsureCmd.Flags().StringP("key", "k", "", "The name of the secret to ensure (required)")

	// Generation options (same as set command)
	secretsEnsureCmd.Flags().Int("size", 16, "Size of the generated secret in characters")
	secretsEnsureCmd.Flags().BoolP("no-upper", "U", false, "Exclude uppercase letters from generated secret")
	secretsEnsureCmd.Flags().BoolP("no-lower", "L", false, "Exclude lowercase letters from generated secret")
	secretsEnsureCmd.Flags().BoolP("no-digits", "D", false, "Exclude digits from generated secret")
	secretsEnsureCmd.Flags().BoolP("no-special", "S", false, "Exclude special characters from generated secret")
	secretsEnsureCmd.Flags().String("special", "", "Custom special characters to use (default: @_-{}|#!~:^)")
	secretsEnsureCmd.Flags().String("chars", "", "Use only these specific characters (overrides all other character options)")
}
