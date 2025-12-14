package aliases

import (
	"os"
	"strings"

	"github.com/frostyeti/kpv/internal/utils"
	"github.com/spf13/cobra"
)

var rmCmd = &cobra.Command{
	Use:   "rm [alias_name]",
	Short: "remove a keepass path alias",
	Long:  `Remove a keepass path alias.`,
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
		var newAliases []string
		found := false
		for _, line := range aliases {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			name := strings.TrimSpace(parts[0])
			if name == aliasName {
				found = true
				continue
			}
			newAliases = append(newAliases, line)
		}

		if !found {
			os.Stderr.WriteString("alias not found\n")
			os.Exit(1)
		}

		conf.Set("aliases", strings.Join(newAliases, "\n"))
		if err := conf.Save(); err != nil {
			utils.Failf("saving config: %v", err)
			return
		}

		utils.Okf("alias %s removed", aliasName)
		os.Exit(0)
	},
}
