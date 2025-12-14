package aliases

import (
	"fmt"
	"os"
	"strings"

	"github.com/frostyeti/kpv/internal/utils"
	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set [alias_name] [path]",
	Short: "Sets a configuration alias",
	Long: `Set a configuration alias to a specified path.

Examples:
  # Sets the alias 'myalias' to the path '/path/to/entry'
  kpv config aliases set myalias /path/to/entry`,

	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			utils.Failf("alias name and path are required")
			return
		}

		aliasName := args[0]
		aliasPath := args[1]

		conf, err := utils.GetConfig()
		if err != nil {
			utils.Failf("getting config: %v", err)
			return
		}

		aliasesRaw, _ := conf.Get("aliases")
		var aliases []string
		if aliasesRaw != "" {
			aliases = strings.Split(aliasesRaw, "\n")
		}

		// Update existing alias or add new one
		found := false
		for i, line := range aliases {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			name := strings.TrimSpace(parts[0])
			if name == aliasName {
				aliases[i] = fmt.Sprintf("%s=%s", aliasName, aliasPath)
				found = true
				break
			}
		}
		if !found {
			aliases = append(aliases, fmt.Sprintf("%s=%s", aliasName, aliasPath))
		}

		conf.Set("aliases", strings.Join(aliases, "\n"))
		if err := conf.Save(); err != nil {
			utils.Failf("saving config: %v", err)
			return
		}

		os.Exit(0)
	},
}

func init() {
	aliasesCmd.AddCommand(setCmd)
}
