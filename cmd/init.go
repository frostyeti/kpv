/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/99designs/keyring"
	"github.com/frostyeti/go/secrets"
	"github.com/frostyeti/kpv/internal/keepass"
	"github.com/frostyeti/kpv/internal/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new KeePass vault",
	Long: `Initialize a new KeePass vault database file.

Creates a new .kdbx file with a generated or provided password. If no password
is specified, a secure random password is generated and stored in:
  1. OS keyring (service: kpv, key: vault path)
  2. A .key file in ~/.local/share/kpv/ (or %LOCALAPPDATA%/kpv on Windows)

Examples:
  # Initialize default vault
  kpv init

  # Initialize a named vault
  kpv init --vault myvault

  # Initialize with custom password
  kpv init --vault myvault --password "my-secure-password"

  # Initialize in global location
  kpv init --vault myvault --global

  # Initialize at specific path
  kpv init --url /path/to/vault.kdbx

  # Initialize at specific path with file URI
  kpv init --url file:///path/to/vault.kdbx`,

	Run: func(cmd *cobra.Command, args []string) {
		vault, _ := cmd.Flags().GetString("vault")
		global, _ := cmd.Flags().GetBool("global")

		// Check if password was actually provided (vs default empty string from env var)
		password := ""
		if cmd.Flags().Changed("password") {
			password, _ = cmd.Flags().GetString("password")
		}

		// Determine vault path
		var vaultPath string
		var err error

		if vault == "" {
			err := utils.LoadConfig()
			if err == nil {
				v := viper.GetString("defaults.path")
				if v != "" {
					if filepath.IsAbs(v) {
						vaultPath = v
					} else {
						v2, err := filepath.Abs(v)
						if err != nil {
							cmd.PrintErrf("Error resolving default vault path from config: %v\n", err)
							return
						}
						vaultPath = v2
					}
				}
			}

			if vaultPath == "" {
				// Default to ~/.local/share/kpv/default.kdbx
				vaultPath = utils.GetDefaultVaultPath("default.kdbx")
			}
		} else {

			// vault is specified
			if global && !filepath.IsAbs(vault) && !strings.Contains(vault, string(filepath.Separator)) && !strings.Contains(vault, "/") {
				alias := vault

				// Global flag and simple name - use default directory
				vaultName := vault
				if !strings.HasSuffix(vaultName, ".kdbx") {
					vaultName = vaultName + ".kdbx"
				}
				vaultPath = utils.GetDefaultVaultPath(vaultName)

				utils.SetAlias(alias, vaultPath)
			} else {
				// Use vault as-is (could be relative or absolute path)
				resolved, err := utils.ResolveVaultPath(vault)
				if err != nil {
					cmd.PrintErrf("Error resolving vault path: %v\n", err)
					return
				}
				vaultPath = resolved.Path
			}
		}

		// Check if file already exists
		if _, err := os.Stat(vaultPath); err == nil {
			utils.Failf("Error: vault file already exists at %s\n", vaultPath)
			return
		}

		// Generate password if not provided
		var vaultPassword string
		passwordGenerated := false
		if password != "" {
			vaultPassword = password
		} else {
			// Generate a secure password (at least 16 characters)
			builder := secrets.NewOptionsBuilder()
			builder.WithSize(32) // Use 32 for extra security
			builder.WithUpper(true)
			builder.WithLower(true)
			builder.WithDigits(true)
			builder.WithSymbols("@_-{}|#!~:^")
			builder.WithRetries(100) // Add retries

			opts := builder.Build()
			vaultPassword, err = opts.Generate()
			if err != nil {
				cmd.PrintErrf("Error generating password: %v\n", err)
				return
			}
			passwordGenerated = true
		}

		// Ensure directory exists
		dir := filepath.Dir(vaultPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			cmd.PrintErrf("Error creating directory %s: %v\n", dir, err)
			return
		}

		// Save generated password before creating vault (keepass.Create may clear it for security)
		// Make a real copy of the string, not just an alias
		passwordToSave := string([]byte(vaultPassword))

		// Create the KeePass database
		options := keepass.KdbxOptions{
			Path:      vaultPath,
			Secret:    &vaultPassword,
			Create:    true,
			CreateDir: true,
		}

		kdbx, err := keepass.Create(options)
		if err != nil {
			cmd.PrintErrf("Error creating KeePass vault: %v\n", err)
			return
		}

		if kdbx == nil || !kdbx.IsOpen() {
			cmd.PrintErrf("Error: vault was not created successfully\n")
			return
		}

		fmt.Printf("Successfully created KeePass vault at: %s\n", vaultPath)

		// Save password to OS keyring
		if passwordGenerated {
			kr, err := utils.OpenKeyring()
			if err == nil {
				err = kr.Set(keyring.Item{
					Key:  "kpv://" + vaultPath,
					Data: []byte(passwordToSave),
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not save password to OS keyring: %v\n", err)
				} else {
					fmt.Printf("Password saved to OS keyring (service: kpv, key: %s)\n", vaultPath)
				}
			} else {
				fmt.Fprintf(os.Stderr, "Warning: could not open OS keyring: %v\n", err)
			}

			// Save password to .key file in kpv config directory
			configDir := utils.GetDefaultVaultPath("")
			if err := os.MkdirAll(configDir, 0755); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: could not create config directory: %v\n", err)
			} else {
				// Use vault base name for key file
				baseName := filepath.Base(vaultPath)
				keyFileName := baseName[:len(baseName)-len(filepath.Ext(baseName))] + ".key"
				keyFilePath := filepath.Join(configDir, keyFileName)

				err = os.WriteFile(keyFilePath, []byte(passwordToSave), 0600)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not save password to key file: %v\n", err)
				} else {
					fmt.Printf("\nIMPORTANT: Password has been saved to: %s\n", keyFilePath)
					fmt.Println("Please back up this password and consider removing the .key file after saving it securely.")
					fmt.Printf("\nGenerated password: %s\n", passwordToSave)
				}
			}
		} else {
			kr, err := utils.OpenKeyring()
			if err == nil {
				err = kr.Set(keyring.Item{
					Key:  "kpv://" + vaultPath,
					Data: []byte(passwordToSave),
				})
				if err != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not save password to OS keyring: %v\n", err)
				} else {
					fmt.Printf("Password saved to OS keyring (service: kpv, key: %s)\n", "kpv://"+vaultPath)
				}
			} else {
				fmt.Fprintf(os.Stderr, "Warning: could not store password to OS keyring: %v\n", err)
			}
		}
	},
}

func init() {
	RootCmd.AddCommand(initCmd)

	initCmd.Flags().Bool("global", false, "Create vault in global location (~/.local/share/kpv or %LOCALAPPDATA%/kpv)")
}
