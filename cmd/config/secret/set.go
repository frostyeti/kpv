package secret

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/99designs/keyring"
	"github.com/frostyeti/kpv/internal/utils"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var setCmd = &cobra.Command{
	Use:   "set <path> [secret]",
	Short: "set the secret for a keepass db in the os secret store",
	Long:  `Set the secret for a keepass db in the os secret store.`,
	Run: func(cmd *cobra.Command, args []string) {

		value := ""
		if len(args) < 1 {
			cmd.Help()
			return
		} else if len(args) >= 2 {
			value = args[1]
		}

		path := args[0]

		if strings.HasPrefix(path, "file://") {
			path = strings.TrimPrefix(path, "file:///")
			path = "kpv:///" + path
		} else if !strings.HasPrefix(path, "kpv:///") {
			if !filepath.IsAbs(path) {
				absPath, err := filepath.Abs(path)
				if err != nil {
					utils.Failf("resolving absolute path: %v", err)
					return
				}
				path = absPath
			}

			path = "kpv:///" + path
		}

		file, _ := cmd.Flags().GetString("file")
		stdin, _ := cmd.Flags().GetBool("stdin")
		envVar, _ := cmd.Flags().GetString("env")
		prompt, _ := cmd.Flags().GetBool("prompt")

		if value == "" {
			if file != "" {
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

		ring, err := utils.OpenKeyring()
		if err != nil {
			utils.Failf("getting secret ring: %v", err)
			return
		}

		err = ring.Set(keyring.Item{
			Key:  path,
			Data: []byte(value),
		})

		if err != nil {
			utils.Failf("setting secret in keyring: %v", err)
			return
		}

		utils.Okf("%s updated", path)
		os.Exit(0)
	},
}

func init() {
	setCmd.Flags().StringP("file", "f", "", "Path to a file containing the secret to set")
	setCmd.Flags().BoolP("stdin", "s", false, "Read the secret from stdin")
	setCmd.Flags().StringP("env", "e", "", "Read the secret from the specified environment variable")
	setCmd.Flags().BoolP("prompt", "p", false, "Prompt for the secret interactively")

	secretCmd.AddCommand(setCmd)
}
