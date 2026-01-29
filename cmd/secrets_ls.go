/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/frostyeti/kpv/internal/utils"
	"github.com/gobwas/glob"
	"github.com/spf13/cobra"
)

// secretsLsCmd represents the ls command
var secretsLsCmd = &cobra.Command{
	Use:     "ls [filter]",
	Aliases: []string{"list"},
	Short:   "List secrets in KeePass vault",
	Long: `List all secrets in a KeePass vault.

Optionally provide a glob pattern to filter the results. The pattern supports
standard glob syntax including wildcards (* and ?).

Examples:
  # List all secrets
  kpv ls

  # List secrets matching a pattern
  kpv ls "app-*"

  # List secrets with wildcards
  kpv ls "*-prod"
  kpv ls "db-*-password"

  # List secrets using the alias
  kpv list "api-key-*"

  # Use a specific vault
  kpv ls --vault myvault`,

	Run: func(cmd *cobra.Command, args []string) {
		var filterPattern string
		if len(args) > 0 {
			filterPattern = args[0]
		}

		kdbx, _, err := utils.OpenKeePass(cmd)
		if err != nil {
			utils.Failf("opening KeePass vault failed: %v\n", err)
			return
		}

		// Compile glob pattern if provided
		var matcher glob.Glob
		if filterPattern != "" {
			matcher, err = glob.Compile(filterPattern)
			if err != nil {
				utils.Failf("compiling glob pattern failed: %v\n", err)
				return
			}
		}

		// Get all entries from the root group
		root := kdbx.Root()
		count := 0
		matchCount := 0

		// Collect all entry titles
		var allEntries []string
		if root != nil && len(root.Entries) > 0 {
			for _, entry := range root.Entries {
				title := entry.GetTitle()
				if title != "" {
					allEntries = append(allEntries, title)
				}
			}
		}

		count = len(allEntries)

		for _, name := range allEntries {
			// Apply filter if specified
			if matcher != nil {
				if !matcher.Match(name) {
					continue
				}
			}

			matchCount++
			fmt.Println(name)
		}

		// Print summary to stderr so it doesn't interfere with piping
		if filterPattern != "" {
			fmt.Fprintf(os.Stderr, "\nShowing %d of %d secrets (filtered by: %s)\n", matchCount, count, filterPattern)
		} else {
			fmt.Fprintf(os.Stderr, "\nTotal secrets: %d\n", count)
		}
	},
}

func init() {
	secretsCmd.AddCommand(secretsLsCmd)
}
