/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package config

import (
	"os"
	"strings"

	"github.com/frostyeti/kpv/internal/utils"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// getCmd represents the get command
var setCmd = &cobra.Command{
	Use:   "set <key> [value]",
	Short: "Set one config value in the config file",
	Long: `Set one config value in the config file.

Examples:
  # Gets the default service name
  kpv config set <key> <value>
  kpv config set service myservice
  
  # Set using inline value
  kpv config set keychain.name --value login
  
  # Set from a file
  kpv config set keychain.name --file ./keychain_name.txt

  # Set from an environment variable
  kpv config set keychain.name --env KEYCHAIN_NAME_ENV

  # Set from stdin
  echo "login" | kpv config set keychain.name --stdin`,

	Run: func(cmd *cobra.Command, args []string) {

		if len(args) < 2 {
			utils.Fail("key and value must be provided")
			return
		}

		key := args[0]
		value := args[1]

		inlineValue, _ := cmd.Flags().GetString("value")
		file, _ := cmd.Flags().GetString("file")
		stdin, _ := cmd.Flags().GetBool("stdin")
		envVar, _ := cmd.Flags().GetString("env")
		prompt, _ := cmd.Flags().GetBool("prompt")

		if value == "" {
			if inlineValue != "" {
				value = inlineValue
			} else if file != "" {
				data, err := os.ReadFile(file)
				if err != nil {
					utils.Failf("reading secret from file: %v", err)
					return
				}
				value = string(data)
			} else if stdin {
				data, err := os.ReadFile("/dev/stdin")
				if err != nil {
					utils.Failf("reading secret from stdin: %v", err)
					return
				}
				value = string(data)
			} else if envVar != "" {
				value = os.Getenv(envVar)
				if value == "" {
					utils.Failf("environment variable %s is not set or empty", envVar)
					return
				}
			} else if prompt {
				// Prompt for secret using os.Stdin
				password := ""

				for password == "" {
					os.Stdout.WriteString("Enter secret: ")
					bytes, _ := term.ReadPassword(int(os.Stdin.Fd()))
					password = strings.TrimSpace(string(bytes))

					if password == "" {
						utils.Warnf("secret cannot be empty, please try again\n")
					}
				}

				value = password
			} else {
				utils.Failf("no secret provided")
				return
			}
		}

		if key == "" {
			utils.Fail("key must be not be empty")
			return
		}

		conf, err := utils.GetConfig()
		if err != nil {
			utils.Failf("failed to load config: %v", err)
			return
		}

		if key == "alias" {
			utils.Fail("Use the aliases subcommand to edit an alias")
			return
		}

		conf.Set(key, value)
		if err := conf.Save(); err != nil {
			utils.Failf("failed to save config: %v", err)
			return
		}

		utils.Okf("%s updated", key)
		os.Exit(0)
	},
}

func init() {
	ConfigCmd.AddCommand(setCmd)
	setCmd.Flags().StringP("value", "v", "", "The secret value (exclusive with --file, --env, --stdin, --generate)")
	setCmd.Flags().StringP("file", "f", "", "Path to file containing the secret value (exclusive with --value, --env, --stdin, --generate)")
	setCmd.Flags().StringP("env", "e", "", "Environment variable name containing the secret value (exclusive with --value, --file, --stdin, --generate)")
	setCmd.Flags().BoolP("stdin", "s", false, "Read the secret value from stdin (exclusive with --value, --file, --env, --generate)")
}
