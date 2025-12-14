/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"io"
	"os"

	"github.com/frostyeti/kpv/internal/keepass"
	"github.com/frostyeti/kpv/internal/utils"
	"github.com/spf13/cobra"
)

// setStringCmd represents the set-string command
var setStringCmd = &cobra.Command{
	Use:   "set-string",
	Short: "Set a custom string field on a KeePass entry",
	Long: `Set a custom string field on a KeePass entry.

This command sets non-standard string fields on entries. Standard fields
like Title, Username, Password, URL, and Notes should be set using the
regular 'set' command or during entry creation with 'import'.

Custom string fields can be either protected (encrypted) or unprotected.
Use the --protected flag to encrypt the value.

You can provide the value via:
  - --value: Value from command line
  - --file: Read value from a file
  - --var: Read value from an environment variable
  - --stdin: Read value from standard input

Examples:
  # Set an unprotected custom string field
  kpv set-string --key my-secret --field api-key --value "abc123"

  # Set a protected (encrypted) custom string field
  kpv set-string --key my-secret --field secret-token --value "xyz789" --protected

  # Set from a file
  kpv set-string --key my-secret --field certificate --file cert.pem

  # Set from environment variable
  kpv set-string --key my-secret --field api-key --var API_KEY

  # Set from stdin
  echo "value" | kpv set-string --key my-secret --field custom --stdin`,

	Run: func(cmd *cobra.Command, args []string) {
		key, _ := cmd.Flags().GetString("key")
		field, _ := cmd.Flags().GetString("field")
		value, _ := cmd.Flags().GetString("value")
		file, _ := cmd.Flags().GetString("file")
		varName, _ := cmd.Flags().GetString("env")
		stdin, _ := cmd.Flags().GetBool("stdin")
		protected, _ := cmd.Flags().GetBool("protected")

		if key == "" {
			utils.Failf("--key must be provided\n")
			return
		}

		if field == "" {
			utils.Failf("--field must be provided\n")
			return
		}

		// Standard fields that should not be set with set-string
		standardFields := map[string]bool{
			"Title":    true,
			"Username": true,
			"Password": true,
			"URL":      true,
			"Notes":    true,
		}

		if standardFields[field] {
			utils.Failf("'%s' is a standard field. Use the 'set' command or import instead.\n", field)
			return
		}

		// Validate that exactly one input method is specified
		inputMethods := 0
		if value != "" {
			inputMethods++
		}
		if file != "" {
			inputMethods++
		}
		if varName != "" {
			inputMethods++
		}
		if stdin {
			inputMethods++
		}

		if inputMethods == 0 {
			utils.Failf("must specify exactly one of --value, --file, --var, or --stdin\n")
			return
		}

		if inputMethods > 1 {
			utils.Failf("options --value, --file, --var, and --stdin are mutually exclusive\n")
			return
		}

		kdbx, _, err := utils.OpenKeePass(cmd)
		if err != nil {
			utils.Failf("opening KeePass vault failed: %v\n", err)
			return
		}

		entry := kdbx.FindEntry(key)
		if entry == nil {
			utils.Failf("secret %s not found\n", key)
			return
		}

		// Get the value based on the input method
		var fieldValue string
		switch {
		case value != "":
			fieldValue = value
		case file != "":
			data, err := os.ReadFile(file)
			if err != nil {
				utils.Failf("reading file %s failed: %v\n", file, err)
				return
			}
			fieldValue = string(data)
		case varName != "":
			fieldValue = os.Getenv(varName)
			if fieldValue == "" {
				utils.Warnf("environment variable %s is empty or not set\n", varName)
			}
		case stdin:
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				utils.Failf("reading from stdin failed: %v\n", err)
				return
			}
			fieldValue = string(data)
		}

		// Update the entry with the custom field
		kdbx.UpsertEntry(key, func(e *keepass.Entry) {
			if protected {
				e.SetProtectedValue(field, fieldValue)
			} else {
				e.SetValue(field, fieldValue)
			}
		})

		err = kdbx.Save()
		if err != nil {
			utils.Failf("saving KeePass vault failed: %v\n", err)
			return
		}

		protectionStatus := "unprotected"
		if protected {
			protectionStatus = "protected"
		}
		utils.Okf("set %s field '%s' on secret: %s\n", protectionStatus, field, key)
	},
}

func init() {
	secretsCmd.AddCommand(setStringCmd)

	setStringCmd.Flags().StringP("key", "k", "", "The secret name/key (required)")
	setStringCmd.Flags().StringP("field", "z", "", "The custom field name to set (required)")
	setStringCmd.Flags().StringP("value", "v", "", "The value to set")
	setStringCmd.Flags().StringP("file", "f", "", "Read value from a file")
	setStringCmd.Flags().StringP("env", "e", "", "Read value from an environment variable")
	setStringCmd.Flags().BoolP("stdin", "s", false, "Read value from stdin")
	setStringCmd.Flags().BoolP("protected", "p", false, "Store the value as protected (encrypted)")
}
