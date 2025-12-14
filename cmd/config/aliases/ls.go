package aliases

import (
	"fmt"
	"os"
	"strings"

	"github.com/frostyeti/kpv/internal/utils"
	"github.com/gobwas/glob"
	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls [filter]",
	Short: "list keepass path aliases",
	Long: `List keepass path aliases.

Examples:
  # lists all aliases
  kpv config aliases ls
  
  # lists aliases matching a pattern
  kpv config aliases ls "db-*-prod"`,

	Run: func(cmd *cobra.Command, args []string) {
		filter := ""
		if len(args) > 0 {
			filter = args[0]
		}

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

		var matcher glob.Glob
		aliases := strings.Split(aliasesRaw, "\n")
		if filter != "" {
			matcher, err = glob.Compile(filter)
			if err != nil {
				utils.Failf("compiling glob pattern failed: %v\n", err)
				return
			}

			for _, line := range aliases {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) != 2 {
					continue
				}
				name := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				if matcher.Match(name) {
					os.Stdout.WriteString(fmt.Sprintf("%s=%s", name, value))
					os.Exit(0)
				}
			}
			os.Exit(0)
		}

		for _, line := range aliases {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			name := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			os.Stdout.WriteString(fmt.Sprintf("%s=%s", name, value))

		}

		os.Exit(0)
	},
}

func init() {
	aliasesCmd.AddCommand(getCmd)
}
