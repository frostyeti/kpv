/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/frostyeti/kpv/internal/utils"
	"github.com/spf13/cobra"
)

// SecretExport represents the export format for a secret
type SecretExport struct {
	Value    string                  `json:"value"`
	Username string                  `json:"username,omitempty"`
	URL      string                  `json:"url,omitempty"`
	Notes    string                  `json:"notes,omitempty"`
	Strings  map[string]CustomString `json:"strings,omitempty"`
	Tags     map[string]string       `json:"tags,omitempty"`
}

// exportCmd represents the export command
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export secrets from KeePass vault to JSON",
	Long: `Export all entries from a KeePass vault to a JSON file.

Note: Currently only supports JSON export format. Support for exporting to
other KeePass files (.kdbx) will be added in a future release.

The exported JSON file contains an object where each property represents a secret.
Each secret value is an object containing:
  - value: The secret password/value
  - username: The username (optional)
  - url: The URL (optional)
  - notes: Notes/comments (optional)
  - strings: Custom non-standard fields (optional)
    Each custom field is an object with:
    - value: The field value
    - encrypted: Whether the field is encrypted/protected
  - tags: KeePass tags (optional, values are empty strings)

Examples:
  # Export all secrets to a file
  kpv export --json --file secrets.json

  # Export to stdout
  kpv export --json

  # Export with pretty formatting
  kpv export --json --file secrets.json --pretty`,

	Run: func(cmd *cobra.Command, args []string) {
		useJSON, _ := cmd.Flags().GetBool("json")
		file, _ := cmd.Flags().GetString("file")
		pretty, _ := cmd.Flags().GetBool("pretty")

		// Validate that json flag is set
		if !useJSON {
			cmd.PrintErrf("Error: --json flag is required. Other export formats will be supported in the future.\n")
			return
		}

		// Open the KeePass vault
		kdbx, _, err := utils.OpenKeePass(cmd)
		if err != nil {
			cmd.PrintErrf("Error opening vault: %v\n", err)
			return
		}

		// Export all entries
		exported := make(map[string]SecretExport)

		root := kdbx.Root()
		if root == nil {
			cmd.PrintErrf("Error: vault has no content\n")
			return
		}

		if len(root.Entries) == 0 {
			fmt.Fprintf(os.Stderr, "Warning: no entries found in vault\n")
			// Still output empty JSON
		}

		// Iterate through all entries
		for _, entry := range root.Entries {
			title := entry.GetTitle()
			if title == "" {
				continue
			}

			export := SecretExport{
				Value: entry.GetPassword(),
			}

			if username := entry.GetContent("Username"); username != "" {
				export.Username = username
			}

			if url := entry.GetContent("URL"); url != "" {
				export.URL = url
			}

			if notes := entry.GetContent("Notes"); notes != "" {
				export.Notes = notes
			}

			// Export custom attributes as strings with encryption metadata
			if len(entry.Values) > 0 {
				export.Strings = make(map[string]CustomString)
				for _, value := range entry.Values {
					key := value.Key
					// Skip standard fields
					if key == "Title" || key == "Username" || key == "Password" || key == "URL" || key == "Notes" {
						continue
					}
					if value.Value.Content != "" {
						export.Strings[key] = CustomString{
							Value:     value.Value.Content,
							Encrypted: value.Value.Protected.Bool,
						}
					}
				}
				if len(export.Strings) == 0 {
					export.Strings = nil
				}
			}

			// Export KeePass tags
			tags := entry.Tags
			if tags != "" {
				export.Tags = make(map[string]string)
				tagNames := strings.Split(tags, ",")
				for _, tagName := range tagNames {
					tagName = strings.TrimSpace(tagName)
					if tagName != "" {
						export.Tags[tagName] = ""
					}
				}
				if len(export.Tags) == 0 {
					export.Tags = nil
				}
			}

			exported[title] = export
		}

		// Marshal to JSON
		var jsonData []byte
		if pretty {
			jsonData, err = json.MarshalIndent(exported, "", "  ")
		} else {
			jsonData, err = json.Marshal(exported)
		}

		if err != nil {
			cmd.PrintErrf("Error marshaling secrets to JSON: %v\n", err)
			return
		}

		// Write to file or stdout
		if file != "" {
			err = os.WriteFile(file, jsonData, 0600)
			if err != nil {
				cmd.PrintErrf("Error writing to file %s: %v\n", file, err)
				return
			}
			fmt.Fprintf(os.Stderr, "Exported %d secret(s) to %s\n", len(exported), file)
		} else {
			fmt.Println(string(jsonData))
			fmt.Fprintf(os.Stderr, "Exported %d secret(s)\n", len(exported))
		}
	},
}

func init() {
	secretsCmd.AddCommand(exportCmd)

	exportCmd.Flags().Bool("json", false, "Export in JSON format (required for now)")
	exportCmd.Flags().StringP("file", "f", "", "Output JSON file path (default: stdout)")
	exportCmd.Flags().Bool("pretty", false, "Pretty-print JSON output")
}
