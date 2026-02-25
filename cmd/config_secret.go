package cmd

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/99designs/keyring"
	"github.com/frostyeti/kpv/internal/utils"
	"github.com/spf13/cobra"
)

var configSecretCmd = &cobra.Command{
	Use:   "secret",
	Short: "Manage KeePass passwords in the OS secret store",
	Long:  `Manage KeePass database passwords in the OS secret store.`,
}

var configSecretGetCmd = &cobra.Command{
	Use:   "get <path>",
	Short: "Retrieve a password from OS keyring",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]

		// Normalize path
		resolved, err := utils.ResolveVaultPath(path)
		if err != nil {
			utils.Failf("invalid path: %v", err)
		}
		key := "kpv:///" + resolved.Path

		kr, err := utils.OpenKeyring()
		if err != nil {
			utils.Failf("failed to open keyring: %v", err)
		}

		item, err := kr.Get(key)
		if err != nil {
			utils.Failf("secret for %s not found in keyring: %v", path, err)
		}

		fmt.Print(string(item.Data))
	},
}

var configSecretSetCmd = &cobra.Command{
	Use:   "set <path> [secret]",
	Short: "Store a password to OS keyring",
	Args:  cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]
		var secret string

		if len(args) == 2 {
			secret = args[1]
		}

		// Normalize path
		resolved, err := utils.ResolveVaultPath(path)
		if err != nil {
			utils.Failf("invalid path: %v", err)
		}
		key := "kpv:///" + resolved.Path

		// Check input flags
		file, _ := cmd.Flags().GetString("file")
		env, _ := cmd.Flags().GetString("env")
		stdin, _ := cmd.Flags().GetBool("stdin")

		inputMethods := 0
		if secret != "" {
			inputMethods++
		}
		if file != "" {
			inputMethods++
		}
		if env != "" {
			inputMethods++
		}
		if stdin {
			inputMethods++
		}

		if inputMethods > 1 {
			utils.Failf("options [secret], --file, --env, and --stdin are mutually exclusive")
		}

		if secret == "" && inputMethods == 0 {
			utils.Failf("a secret must be provided")
		}

		if file != "" {
			data, err := os.ReadFile(file)
			if err != nil {
				utils.Failf("failed to read file: %v", err)
			}
			secret = string(data)
		} else if env != "" {
			secret = os.Getenv(env)
			if secret == "" {
				utils.Failf("environment variable %s is empty", env)
			}
		} else if stdin {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				utils.Failf("failed to read stdin: %v", err)
			}
			secret = string(data)
		}

		secret = strings.TrimSpace(secret)

		kr, err := utils.OpenKeyring()
		if err != nil {
			utils.Failf("failed to open keyring: %v", err)
		}

		err = kr.Set(keyring.Item{
			Key:  key,
			Data: []byte(secret),
		})

		if err != nil {
			utils.Failf("failed to save secret to keyring: %v", err)
		}

		utils.Okf("secret for %s saved to keyring", path)
	},
}

var configSecretRmCmd = &cobra.Command{
	Use:   "rm <path>",
	Short: "Remove a password from OS keyring",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		path := args[0]

		// Normalize path
		resolved, err := utils.ResolveVaultPath(path)
		if err != nil {
			utils.Failf("invalid path: %v", err)
		}
		key := "kpv:///" + resolved.Path

		kr, err := utils.OpenKeyring()
		if err != nil {
			utils.Failf("failed to open keyring: %v", err)
		}

		err = kr.Remove(key)
		if err != nil {
			utils.Failf("failed to remove secret: %v", err)
		}

		utils.Okf("secret for %s removed from keyring", path)
	},
}

func init() {
	configCmd.AddCommand(configSecretCmd)
	configSecretCmd.AddCommand(configSecretGetCmd)

	configSecretSetCmd.Flags().StringP("file", "f", "", "Read secret from file")
	configSecretSetCmd.Flags().StringP("env", "e", "", "Read secret from environment variable")
	configSecretSetCmd.Flags().BoolP("stdin", "s", false, "Read secret from stdin")
	configSecretCmd.AddCommand(configSecretSetCmd)

	configSecretCmd.AddCommand(configSecretRmCmd)
}
