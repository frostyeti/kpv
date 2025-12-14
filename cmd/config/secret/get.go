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
	Short: "get the secret for a keepass db from the os secret store",
	Long:  `Get the secret for a keepass db from the os secret store.`,
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
