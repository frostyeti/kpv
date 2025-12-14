/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/frostyeti/kpv/internal/keepass"
	"github.com/frostyeti/kpv/internal/utils"
	"github.com/spf13/cobra"
)

// SecretImport represents the import format for a secret
type SecretImport struct {
	Value    string                  `json:"value"`
	Username string                  `json:"username,omitempty"`
	URL      string                  `json:"url,omitempty"`
	Notes    string                  `json:"notes,omitempty"`
	Strings  map[string]CustomString `json:"strings,omitempty"`
	Binaries map[string]CustomString `json:"binaries,omitempty"`

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

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import secrets from JSON to KeePass vault",
	Long: `Import secrets from a JSON file to a KeePass vault.

Note: Currently only supports JSON import format. Support for importing from
other KeePass files (.kdbx) will be added in a future release.

The JSON file should contain an object where each property represents a secret.
Property values can be either:
  1. A string (simple secret value)
  2. An object with the following properties:
     - value: The secret password/value (optional if ensure=true)
     - username: The username (optional)
     - url: The URL (optional)
     - notes: Notes/comments (optional)
     - strings: Custom non-standard fields (optional)
       Each custom field is an object with:
       - value: The field value
       - encrypted: Whether the field should be encrypted/protected
     - ensure: If true and value is empty and doesn't exist, generate secret (optional)
     - size: Generated password size, default 32 (optional)
     - noUpper: Exclude uppercase letters from generated password (optional)
     - noLower: Exclude lowercase letters from generated password (optional)
     - noDigits: Exclude digits from generated password (optional)
     - noSpecial: Exclude special characters from generated password (optional)
     - special: Custom special characters for generated password (optional)
     - chars: Custom character set, overrides all other options (optional)

Examples:
  # Import secrets from a file
  kpv import --json --file secrets.json

  # Import from stdin
  cat secrets.json | kpv import --json --stdin

Example JSON format:
  {
    "simple-secret": "simple-value",
    "complex-secret": {
      "value": "secret-password",
      "username": "admin",
      "url": "https://example.com",
      "notes": "Production database",
      "strings": {
        "environment": {
          "value": "production",
          "encrypted": false
        },
        "api-key": {
          "value": "sk-1234567890",
          "encrypted": true
        }
      },
	  "binaries": {
	     "file.pfx": {
			"value": "BASE64_ENCODED_DATA_HERE",
			"encrypted": true 			
		 }
    },
    "generated-secret": {
      "ensure": true,
      "size": 32,
      "noSpecial": true,
      "username": "api",
      "strings": {
        "generated": {
          "value": "true",
          "encrypted": false
        }
      }
    }
  }`,

	Run: func(cmd *cobra.Command, args []string) {
		useJSON, _ := cmd.Flags().GetBool("json")
		file, _ := cmd.Flags().GetString("file")
		stdin, _ := cmd.Flags().GetBool("stdin")

		// Validate that json flag is set
		if !useJSON {
			utils.Failf("--json flag is required. Other import formats will be supported in the future.\n")
			return
		}

		// Validate that exactly one input method is specified
		if file == "" && !stdin {
			utils.Failf("must specify either --file or --stdin\n")
			return
		}

		if file != "" && stdin {
			utils.Failf("--file and --stdin are mutually exclusive\n")
			return
		}

		// Read JSON data
		var jsonData []byte
		var err error
		if file != "" {
			jsonData, err = os.ReadFile(file)
			if err != nil {
				utils.Failf("reading file %s failed: %v\n", file, err)
				return
			}
		} else {
			jsonData, err = os.ReadFile(os.Stdin.Name())
			if err != nil {
				utils.Failf("reading from stdin failed: %v\n", err)
				return
			}
		}

		// Parse JSON - first try as raw map to handle both string and object values
		var rawData map[string]json.RawMessage
		err = json.Unmarshal(jsonData, &rawData)
		if err != nil {
			utils.Failf("parsing JSON failed: %v\n", err)
			return
		}

		// Convert to SecretImport objects
		secretsToImport := make(map[string]SecretImport)
		for key, rawValue := range rawData {
			var simpleValue string
			// Try to unmarshal as string first
			if err := json.Unmarshal(rawValue, &simpleValue); err == nil {
				secretsToImport[key] = SecretImport{
					Value: simpleValue,
				}
			} else {
				// Try to unmarshal as SecretImport object
				var secretObj SecretImport
				if err := json.Unmarshal(rawValue, &secretObj); err != nil {
					utils.Failf("parsing secret %s failed: %v\n", key, err)
					return
				}
				secretsToImport[key] = secretObj
			}
		}

		// Open the KeePass vault
		kdbx, _, err := utils.OpenKeePass(cmd)
		if err != nil {
			utils.Failf("opening KeePass vault failed: %v\n", err)
			return
		}

		// Import each secret
		successCount := 0
		failCount := 0

		for name, secretImport := range secretsToImport {
			// Handle ensure logic: if ensure=true, value is empty, and secret doesn't exist, generate it
			secretValue := secretImport.Value

			if secretImport.Ensure && secretValue == "" {
				// Check if secret already exists
				existing := kdbx.FindEntry(name)
				if existing == nil {
					// Secret doesn't exist, generate it
					size := secretImport.Size
					if size == 0 {
						size = 32 // Default size
					}

					generated, genErr := utils.GenerateSecretWithOptions(
						size,
						secretImport.NoUpper,
						secretImport.NoLower,
						secretImport.NoDigits,
						secretImport.NoSpecial,
						secretImport.Special,
						secretImport.Chars,
					)

					if genErr != nil {
						utils.Failf("generating secret for %s failed: %v\n", name, genErr)
						failCount++
						continue
					}

					secretValue = generated
					fmt.Fprintf(os.Stderr, "Generated secret for %s\n", name)
				} else {
					// Secret exists, skip
					fmt.Fprintf(os.Stderr, "Secret %s already exists, skipping generation\n", name)
					continue
				}
			}

			// Validate that we have a value to set
			if secretValue == "" {
				utils.Errf("secret %s has no value and ensure is not enabled or generation failed\n", name)
				failCount++
				continue
			}

			// Create or update entry
			kdbx.UpsertEntry(name, func(entry *keepass.Entry) {
				entry.SetPassword(secretValue)
				if secretImport.Username != "" {
					entry.SetUsername(secretImport.Username)
				}
				if secretImport.URL != "" {
					entry.SetUrl(secretImport.URL)
				}
				if secretImport.Notes != "" {
					entry.SetNotes(secretImport.Notes)
				}
				// Set custom strings with encryption
				for k, v := range secretImport.Strings {
					if v.Encrypted {
						entry.SetProtectedValue(k, v.Value)
					} else {
						entry.SetValue(k, v.Value)
					}
				}

				for k, v := range secretImport.Binaries {
					encoded := v.Value
					data, err := base64.StdEncoding.DecodeString(encoded)
					if err != nil {
						utils.Errf("decoding binary %s for secret %s failed: %v\n", k, name, err)
						continue
					}

					found := false

					for _, be := range entry.Binaries {
						if be.Name == k {
							// Update existing binary
							binary := kdbx.FindBinary(be.Value.ID)
							if binary != nil {
								binary.SetContent(data)
								if v.Encrypted {
									binary.MemoryProtection = 1
								} else {
									binary.MemoryProtection = 0
								}
							}
							found = true
						}
					}

					if !found {
						// Add new binary
						binary := kdbx.AddBinary(data)
						if v.Encrypted {
							binary.MemoryProtection = 1
						} else {
							binary.MemoryProtection = 0
						}
						entry.Binaries = append(entry.Binaries, binary.CreateReference(name))
					}
				}
			})

			successCount++
			fmt.Printf("Imported secret: %s\n", name)
		} // Save the vault
		err = kdbx.Save()
		if err != nil {
			utils.Failf("saving KeePass vault failed: %v\n", err)
			return
		}

		// Print summary
		if failCount > 0 {
			fmt.Printf("\nImported %d of %d secret(s) (%d failed)\n", successCount, len(secretsToImport), failCount)
		} else {
			fmt.Printf("\nSuccessfully imported %d secret(s)\n", successCount)
		}
	},
}

func init() {
	secretsCmd.AddCommand(importCmd)

	importCmd.Flags().Bool("json", false, "Import from JSON format (required for now)")
	importCmd.Flags().StringP("file", "f", "", "Input JSON file path")
	importCmd.Flags().Bool("stdin", false, "Read JSON from stdin")
}
