/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/frostyeti/go/dotenv"
	"github.com/frostyeti/kpv/internal/utils"
	"github.com/spf13/cobra"
)

// toScreamingSnakeCase converts a string to SCREAMING_SNAKE_CASE
func toScreamingSnakeCase(input string) string {
	output := ""
	for i, char := range input {
		if char >= 'A' && char <= 'Z' {
			if i > 0 {
				output += "_"
			}
			output += string(char)
		} else if char >= 'a' && char <= 'z' {
			if i > 0 && input[i-1] >= 'A' && input[i-1] <= 'Z' {
				output += "_"
			}
			output += string(char - ('a' - 'A'))
		} else if char >= '0' && char <= '9' {
			output += string(char)
		} else {
			if i > 0 && input[i-1] != '_' {
				output += "_"
			}
		}
	}
	return output
}

// secretsGetCmd represents the get command
var secretsGetCmd = &cobra.Command{
	Use:   "get <key>...",
	Short: "Get one or more secrets from KeePass vault",
	Long: `Get one or more secrets from a KeePass vault.

Examples:
  # Get a single secret
  kpv get --key my-secret

  # Get multiple secrets
  kpv get --key secret1 --key secret2

  # Get secrets with different output formats
  kpv get --key secret1 --format json
  kpv get --key secret1 --format sh
  kpv get --key secret1 --format dotenv

  # Use a specific vault
  kpv get --vault myvault --key secret1
  kpv get --vault /path/to/vault.kdbx --key secret1`,

	Run: func(cmd *cobra.Command, args []string) {
		keys, _ := cmd.Flags().GetStringSlice("key")
		format, _ := cmd.Flags().GetString("format")

		if len(args) > 0 {
			keys = append(keys, args...)
		}

		if format == "" {
			format = "text"
		}

		if len(keys) == 0 {
			utils.Failf("at least one --key must be provided")
			return
		}

		kdbx, _, err := utils.OpenKeePass(cmd)
		if err != nil {
			utils.Failf("opening KeePass vault failed: %v", err)
			return
		}

		values := map[string]string{}
		for _, key := range keys {
			entry := kdbx.FindEntry(key)
			if entry == nil {
				utils.Failf("secret %s not found", key)
				return
			}
			values[key] = entry.GetPassword()
		}

		switch format {
		case "json":
			b, err := json.MarshalIndent(values, "", "  ")
			if err != nil {
				utils.Failf("marshaling secrets to JSON failed: %v", err)
				return
			}
			fmt.Println(string(b))

		case "null-terminated", "null":
			for _, v := range values {
				fmt.Printf("%s\x00", v)
			}

		case "sh", "bash", "zsh":
			for k, v := range values {
				fmt.Printf("export %s='%s'\n", toScreamingSnakeCase(k), v)
			}

		case "powershell", "pwsh":
			for k, v := range values {
				fmt.Printf("$Env:%s='%s'\n", toScreamingSnakeCase(k), v)
			}

		case "dotenv", "env", ".env":
			doc := dotenv.NewDoc()
			for k, v := range values {
				doc.Set(toScreamingSnakeCase(k), v)
			}
			fmt.Println(doc.String())

		case "azure-pipelines", "ado", "azure-devops":
			for k, v := range values {
				fmt.Printf("##vso[task.setvariable variable=%s;]%s\n", toScreamingSnakeCase(k), v)
			}

		case "run", "runfile":
			envFile := os.Getenv("RUN_SECRETS")
			if envFile == "" {
				utils.Failf("RUN_SECRETS environment variable is not set")
				return
			}

			dir := filepath.Dir(envFile)
			if err := os.MkdirAll(dir, 0700); err != nil {
				utils.Failf("creating directory for RUN_ENV file failed: %v", err)
				return
			}

			f, err := os.OpenFile(envFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
			if err != nil {
				utils.Failf("opening RUN_SECRETS file failed: %v", err)
				return
			}
			defer f.Close()

			for k, v := range values {
				if strings.ContainsAny(v, "\r\n") {
					v2 := fmt.Sprintf("<< EOF\n%s\nEOF", v)
					_, err := f.WriteString(fmt.Sprintf("%s%s\n", toScreamingSnakeCase(k), v2))
					if err != nil {
						utils.Failf("writing to GITHUB_ENV file failed: %v", err)
						return
					}
				} else {
					_, err := f.WriteString(fmt.Sprintf("%s=%s\n", toScreamingSnakeCase(k), v))
					if err != nil {
						utils.Failf("writing to GITHUB_ENV file failed: %v", err)
						return
					}
				}
			}

		case "github", "gh-actions", "github-actions":
			envFile := os.Getenv("GITHUB_ENV")
			if envFile == "" {
				utils.Failf("GITHUB_ENV environment variable is not set")
				return
			}

			dir := filepath.Dir(envFile)
			if err := os.MkdirAll(dir, 0700); err != nil {
				utils.Failf("creating directory for GITHUB_ENV file failed: %v", err)
				return
			}

			f, err := os.OpenFile(envFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
			if err != nil {
				utils.Failf("opening GITHUB_ENV file failed: %v", err)
				return
			}
			defer f.Close()

			for k, v := range values {
				if strings.ContainsAny(v, "\r\n") {
					v2 := fmt.Sprintf("<< EOF\n%s\nEOF", v)
					_, err := f.WriteString(fmt.Sprintf("%s%s\n", toScreamingSnakeCase(k), v2))
					if err != nil {
						utils.Failf("writing to GITHUB_ENV file failed: %v", err)
						return
					}
				} else {
					_, err := f.WriteString(fmt.Sprintf("%s=%s\n", toScreamingSnakeCase(k), v))
					if err != nil {
						utils.Failf("writing to GITHUB_ENV file failed: %v", err)
						return
					}
				}

				fmt.Fprintf(os.Stdout, "::add-mask::%s", v)
			}

		case "text":
			fallthrough
		default:
			for _, v := range values {
				fmt.Println(v)
			}
		}
	},
}

func init() {

	secretsGetCmd.Flags().StringSliceP("key", "k", []string{}, "Name of secret(s) to get (can be specified multiple times)")
	secretsGetCmd.Flags().StringP("format", "f", "text", "Output format (text, json, sh, bash, zsh, powershell, pwsh, dotenv, azure-devops, github-actions, run)")
}
