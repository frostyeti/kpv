package secret

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/frostyeti/kpv/internal/utils"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get <path>",
	Short: "Gets the secret for a keepass db from the os secret store",
	Long: `Gets the secret for a keepass db from the os secret store.
	
The secret is used to unlock the keepass db when opening it.

Examples:
  # Get secret for a keepass db
  kpv config secret get kpv:///path/to/vault.kdbx

  # Get secret for a file path (file://)
  kpv config secret get file:///path/to/vault.kdbx

  # Get secret for an absolute file path
  kpv config secret get /path/to/vault.kdbx

  # Get secret for a relative file path
  kpv config secret get ./vault.kdbx

	`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			cmd.Help()
			return
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

		// retrieve secret from keyring
		kr, err := utils.OpenKeyring()
		if err != nil {
			utils.Failf("accessing keyring: %v", err)
			return
		}

		item, err := kr.Get(path)
		if err != nil {
			utils.Failf("getting secret from keyring: %v", err)
			return
		}

		os.Stdout.WriteString(string(item.Data))
		os.Exit(0)
	},
}

func init() {
	secretCmd.AddCommand(getCmd)
}
