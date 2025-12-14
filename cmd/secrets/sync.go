/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/frostyeti/kpv/internal/keepass"
	"github.com/frostyeti/kpv/internal/utils"
	"github.com/spf13/cobra"
)

// SecretSync represents the sync format for a secret
type SecretSync struct {
	Value    *string                  `json:"value,omitempty"`
	Username *string                  `json:"username,omitempty"`
	URL      *string                  `json:"url,omitempty"`
	Notes    *string                  `json:"notes,omitempty"`
	Strings  map[string]*CustomString `json:"strings,omitempty"`
	Tags     map[string]string        `json:"tags,omitempty"` // For KeePass tags - values are ignored, only keys are used
	Delete   *bool                    `json:"delete,omitempty"`

	// Secret generation options
	Ensure    bool   `json:"ensure,omitempty"`    // If true and value is empty and doesn't exist, generate secret
	Size      int    `json:"size,omitempty"`      // Password size (default: 32)
	NoUpper   bool   `json:"noUpper,omitempty"`   // Exclude uppercase letters
	NoLower   bool   `json:"noLower,omitempty"`   // Exclude lowercase letters
	NoDigits  bool   `json:"noDigits,omitempty"`  // Exclude digits
	NoSpecial bool   `json:"noSpecial,omitempty"` // Exclude special characters
	Special   string `json:"special,omitempty"`   // Custom special characters
	Chars     string `json:"chars,omitempty"`     // Custom character set (overrides all other options)
}

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync secrets to KeePass vault, updating only when values differ",
	Long: `Sync secrets to a KeePass vault from a JSON file.

Note: Currently only supports JSON sync format. Support for syncing with
other KeePass files (.kdbx) will be added in a future release.

This command intelligently updates secrets only when their values differ from
what's currently stored in the vault. Property values can be either:
  1. A string (simple secret value - updates only if different)
  2. An object with the following properties:
     - value: The secret password/value (updates if different)
     - username: The username (updates if different)
     - url: The URL (updates if different)
     - notes: Notes/comments (updates if different)
     - strings: Custom non-standard fields with encryption metadata (updates if different)
       Each field is an object with 'value' and 'encrypted' properties
     - tags: Map of tag names (values ignored, only keys used for KeePass tags)
     - delete: If true, delete the secret (ignores other properties)
     - ensure: If true and value is empty and doesn't exist, generate secret (optional)
     - size, noUpper, noLower, noDigits, noSpecial, special, chars: Generation options

Omitted/null properties are not updated in the vault.

Examples:
  # Sync secrets from a file
  kpv sync --json --file secrets.json

  # Sync from stdin
  cat secrets.json | kpv sync --json --stdin

  # Dry run to see what would change
  kpv sync --json --file secrets.json --dry-run

Example JSON format:
  {
    "simple-secret": "new-value",
    "complex-secret": {
      "value": "secret-password",
      "username": "admin",
      "url": "https://example.com",
      "notes": "Updated notes",
      "strings": {
        "api-key": {
          "value": "sk-abc123",
          "encrypted": true
        }
      },
      "tags": {
        "production": "",
        "important": ""
      }
    },
    "secret-to-delete": {
      "delete": true
    },
    "ensure-generated": {
      "ensure": true,
      "size": 32
    }
  }`,

	Run: func(cmd *cobra.Command, args []string) {
		useJSON, _ := cmd.Flags().GetBool("json")
		file, _ := cmd.Flags().GetString("file")
		stdin, _ := cmd.Flags().GetBool("stdin")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		// Validate that json flag is set
		if !useJSON {
			utils.Errf("--json flag is required. Other sync formats will be supported in the future.\n")
			return
		}

		// Validate that exactly one input method is specified
		if file == "" && !stdin {
			utils.Errf("must specify either --file or --stdin\n")
			return
		}

		if file != "" && stdin {
			utils.Errf("--file and --stdin are mutually exclusive\n")
			return
		}

		// Read JSON data
		var jsonData []byte
		var err error
		if file != "" {
			jsonData, err = os.ReadFile(file)
			if err != nil {
				utils.Errf("reading file %s failed: %v\n", file, err)
				return
			}
		} else {
			jsonData, err = os.ReadFile(os.Stdin.Name())
			if err != nil {
				utils.Errf("reading from stdin failed: %v\n", err)
				return
			}
		}

		// Parse JSON - first try as raw map to handle both string and object values
		var rawData map[string]json.RawMessage
		err = json.Unmarshal(jsonData, &rawData)
		if err != nil {
			utils.Errf("parsing JSON failed: %v\n", err)
			return
		}

		// Convert to SecretSync objects
		secretsToSync := make(map[string]SecretSync)
		for key, rawValue := range rawData {
			var simpleValue string
			// Try to unmarshal as string first
			if err := json.Unmarshal(rawValue, &simpleValue); err == nil {
				secretsToSync[key] = SecretSync{
					Value: &simpleValue,
				}
			} else {
				// Try to unmarshal as SecretSync object
				var secretObj SecretSync
				if err := json.Unmarshal(rawValue, &secretObj); err != nil {
					utils.Errf("parsing secret failed %s: %v\n", key, err)
					return
				}
				secretsToSync[key] = secretObj
			}
		}

		// Open the KeePass vault
		kdbx, _, err := utils.OpenKeePass(cmd)
		if err != nil {
			utils.Failf("opening vault failed: %v\n", err)
			return
		}

		// Sync each secret
		createdCount := 0
		updatedCount := 0
		deletedCount := 0
		skippedCount := 0
		failCount := 0

		for name, secretSync := range secretsToSync {
			// Handle deletion
			if secretSync.Delete != nil && *secretSync.Delete {
				if dryRun {
					fmt.Printf("[DRY RUN] Would delete secret: %s\n", name)
					deletedCount++
					continue
				}

				entry := kdbx.FindEntry(name)
				if entry == nil {
					fmt.Fprintf(os.Stderr, "Secret %s not found, skipping deletion\n", name)
					continue
				}

				root := kdbx.Root()
				root.RmEntry(entry)

				fmt.Printf("Deleted secret: %s\n", name)
				deletedCount++
				continue
			}

			// Get existing secret
			existing := kdbx.FindEntry(name)
			secretExists := existing != nil
			isNew := !secretExists

			// Handle ensure logic for new secrets
			var valueToUse *string
			if secretSync.Ensure && (secretSync.Value == nil || *secretSync.Value == "") && !secretExists {
				// Generate secret
				size := secretSync.Size
				if size == 0 {
					size = 32
				}

				generated, genErr := utils.GenerateSecretWithOptions(
					size,
					secretSync.NoUpper,
					secretSync.NoLower,
					secretSync.NoDigits,
					secretSync.NoSpecial,
					secretSync.Special,
					secretSync.Chars,
				)

				if genErr != nil {
					utils.Errf("generating secret for %s: %v\n", name, genErr)
					failCount++
					continue
				}

				valueToUse = &generated

				fmt.Fprintf(os.Stderr, "Generated secret for %s\n", name)
			} else {
				valueToUse = secretSync.Value
			}

			// Determine if we need to update
			needsUpdate := isNew

			if !isNew {
				// Check if value differs
				if valueToUse != nil && *valueToUse != existing.GetPassword() {
					needsUpdate = true
				}

				// Check if username differs
				if secretSync.Username != nil && *secretSync.Username != existing.GetUsername() {
					needsUpdate = true
				}

				// Check if URL differs
				if secretSync.URL != nil && *secretSync.URL != existing.GetUrl() {
					needsUpdate = true
				}

				// Check if notes differ
				if secretSync.Notes != nil {
					existingNotes := existing.GetContent("Notes")
					if *secretSync.Notes != existingNotes {
						needsUpdate = true
					}
				}

				// Check if custom strings differ
				if secretSync.Strings != nil {
					for k, v := range secretSync.Strings {
						if v != nil {
							existingValue := existing.GetContent(k)
							if existingValue != v.Value {
								needsUpdate = true
								break
							}
						}
					}
				}

				// Check if tags differ (for KeePass tags array)
				if secretSync.Tags != nil {
					existingTags := existing.Tags()
					if len(secretSync.Tags) != len(existingTags) {
						needsUpdate = true
					} else {
						// Check if all tag keys exist
						for k := range secretSync.Tags {
							if _, found := existingTags[k]; !found {
								needsUpdate = true
								break
							}
						}
					}
				}
			}

			if !needsUpdate {
				skippedCount++
				continue
			}

			if dryRun {
				if isNew {
					fmt.Printf("[DRY RUN] Would create secret: %s\n", name)
					createdCount++
				} else {
					fmt.Printf("[DRY RUN] Would update secret: %s\n", name)
					updatedCount++
				}
				continue
			}

			// Create or update entry
			kdbx.UpsertEntry(name, func(entry *keepass.Entry) {
				// Set value if provided
				if valueToUse != nil {
					entry.SetPassword(*valueToUse)
				}

				// Set username if provided
				if secretSync.Username != nil {
					entry.SetUsername(*secretSync.Username)
				}

				// Set URL if provided
				if secretSync.URL != nil {
					entry.SetUrl(*secretSync.URL)
				}

				// Set notes if provided
				if secretSync.Notes != nil {
					entry.SetNotes(*secretSync.Notes)
				}

				// Set custom strings with encryption
				if secretSync.Strings != nil {
					for k, v := range secretSync.Strings {
						if v != nil {
							if v.Encrypted {
								entry.SetProtectedValue(k, v.Value)
							} else {
								entry.SetValue(k, v.Value)
							}
						}
					}
				}

				// Set KeePass tags (only keys are used, values are ignored)
				if secretSync.Tags != nil {
					for k := range secretSync.Tags {
						entry.AddTag(k)
					}
				}
			})

			if isNew {
				fmt.Printf("Created secret: %s\n", name)
				createdCount++
			} else {
				fmt.Printf("Updated secret: %s\n", name)
				updatedCount++
			}
		}

		// Save the vault if not a dry run and there were changes
		if !dryRun && (createdCount > 0 || updatedCount > 0 || deletedCount > 0) {
			err = kdbx.Save()
			if err != nil {
				utils.Failf("Error saving vault: %v\n", err)
				return
			}
		}

		// Print summary
		fmt.Println()
		if dryRun {
			fmt.Println("[DRY RUN] Summary:")
		}
		fmt.Printf("Created: %d\n", createdCount)
		fmt.Printf("Updated: %d\n", updatedCount)
		fmt.Printf("Deleted: %d\n", deletedCount)
		fmt.Printf("Skipped (no changes): %d\n", skippedCount)
		if failCount > 0 {
			fmt.Printf("Failed: %d\n", failCount)
		}
	},
}

func init() {
	secretsCmd.AddCommand(syncCmd)

	syncCmd.Flags().Bool("json", false, "Sync using JSON format (required for now)")
	syncCmd.Flags().StringP("file", "f", "", "Input JSON file path")
	syncCmd.Flags().Bool("stdin", false, "Read JSON from stdin")
	syncCmd.Flags().Bool("dry-run", false, "Show what would be changed without making changes")
}
