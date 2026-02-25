/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/frostyeti/kpv/internal/utils"
	"github.com/spf13/cobra"
)

// secretsGetStringCmd represents the get-string command
var secretsGetStringCmd = &cobra.Command{
	Use:   "get-string",
	Short: "Get a custom string field from a KeePass entry",
	Long: `Get a custom string field from a KeePass entry.

This command retrieves non-standard string fields from entries. Standard fields
like Title, Username, Password, URL, and Notes should be retrieved using the
regular 'get' command.

Custom string fields can be either protected (encrypted) or unprotected.

Examples:
  # Get a custom string field
  kpv get-string --key my-secret --field api-key

  # Get from a specific vault
  kpv get-string --vault myvault --key my-secret --field custom-field`,

	Run: func(cmd *cobra.Command, args []string) {
		key, _ := cmd.Flags().GetString("key")
		field, _ := cmd.Flags().GetString("field")

		if key == "" {
			utils.Failf("--key must be provided")
			return
		}

		if field == "" {
			utils.Failf("--field must be provided")
			return
		}

		// Standard fields that should not be retrieved with get-string
		standardFields := map[string]bool{
			"Title":    true,
			"Username": true,
			"Password": true,
			"URL":      true,
			"Notes":    true,
		}

		if standardFields[field] {
			utils.Failf("'%s' is a standard field. Use the 'get' command instead.", field)
			return
		}

		kdbx, _, err := utils.OpenKeePass(cmd)
		if err != nil {
			utils.Failf("opening KeePass vault failed: %v", err)
			return
		}

		entry := kdbx.FindEntry(key)
		if entry == nil {
			utils.Failf("secret %s not found", key)
			return
		}

		value := entry.GetContent(field)
		if value == "" {
			utils.Failf("field '%s' not found in secret %s", field, key)
			return
		}

		fmt.Println(value)
	},
}

func init() {
	secretsCmd.AddCommand(secretsGetStringCmd)

	secretsGetStringCmd.Flags().StringP("key", "k", "", "The secret name/key (required)")
	secretsGetStringCmd.Flags().StringP("field", "f", "", "The custom field name to retrieve (required)")
}
