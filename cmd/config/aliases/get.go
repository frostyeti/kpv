package aliases

import (
	"os"
	"strings"

	"github.com/frostyeti/kpv/internal/utils"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get [alias_name]",
	Short: "Gets the path for the alias",
	Long: `Get the value of a configuration alias.

Examples:
  # Gets the value of the alias 'myalias'
  kpv config aliases get myalias`,

	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			utils.Failf("alias name is required")
			return
		}

		aliasName := args[0]

		conf, err := utils.GetConfig()
		if err != nil {
			utils.Failf("getting config: %v", err)
			return
		}

		aliasesRaw, ok := conf.Get("aliases")
		if !ok {
			utils.Failf("no aliases configured")
			return
		}

		aliases := strings.Split(aliasesRaw, "\n")
		for _, line := range aliases {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			name := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if name == aliasName {
				os.Stdout.WriteString(value + "\n")
				os.Exit(0)
			}
		}

		utils.Failf("alias %s not found", aliasName)
	},
}

func init() {
	aliasesCmd.AddCommand(getCmd)
}
