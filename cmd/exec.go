package cmd

import (
	"bufio"
	"os"
	"os/exec"
	"strings"

	"github.com/frostyeti/kpv/internal/utils"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec [flags] [--] <command> [args...]",
	Short: "Execute a command with secrets loaded as environment variables",
	Long: `Execute a command with secrets loaded as environment variables.

If no keys are specified, ALL secrets from the vault will be loaded.
To avoid leaking your entire password database into the environment, you can
explicitly specify which secrets you want to load using --key or --key-file.

The secrets are formatted as SCREAMING_SNAKE_CASE environment variables.
For example, a secret named "my-api-token" becomes "MY_API_TOKEN".

Examples:
  # Execute 'npm start' with specific secrets
  kpv exec --key db-password --key api-token -- npm start

  # Execute with a file containing a list of keys (one per line)
  kpv exec --key-file keys.txt -- npm start

  # Execute with ALL secrets loaded
  kpv exec npm start

  # Combine with vault path aliases
  kpv exec -V prod-vault -k prod-db-password -- ./deploy.sh
`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		keys, _ := cmd.Flags().GetStringSlice("key")
		keyFile, _ := cmd.Flags().GetString("key-file")

		if keyFile != "" {
			f, err := os.Open(keyFile)
			if err != nil {
				utils.Failf("opening key file failed: %v", err)
				return
			}
			defer f.Close()

			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" && !strings.HasPrefix(line, "#") {
					keys = append(keys, line)
				}
			}
			if err := scanner.Err(); err != nil {
				utils.Failf("reading key file failed: %v", err)
				return
			}
		}

		kdbx, _, err := utils.OpenKeePass(cmd)
		if err != nil {
			utils.Failf("opening KeePass vault failed: %v", err)
			return
		}

		// Retrieve all requested values
		values := map[string]string{}

		if len(keys) == 0 {
			// Load ALL secrets
			root := kdbx.Root()
			if root != nil && len(root.Entries) > 0 {
				for _, entry := range root.Entries {
					title := entry.GetTitle()
					if title != "" {
						values[title] = entry.GetPassword()
					}
				}
			}
		} else {
			for _, key := range keys {
				entry := kdbx.FindEntry(key)
				if entry == nil {
					utils.Failf("secret %s not found", key)
					return
				}
				values[key] = entry.GetPassword()
			}
		}

		// Prepare the environment for the child process
		env := os.Environ()
		for k, v := range values {
			envVar := toScreamingSnakeCase(k)
			env = append(env, envVar+"="+v)
		}

		// Prepare the command
		childCmd := exec.Command(args[0], args[1:]...)
		childCmd.Env = env
		childCmd.Stdin = os.Stdin
		childCmd.Stdout = os.Stdout
		childCmd.Stderr = os.Stderr

		// Run the command
		if err := childCmd.Run(); err != nil {
			// If the command exits with a non-zero status, pass that exit code through
			if exitError, ok := err.(*exec.ExitError); ok {
				os.Exit(exitError.ExitCode())
			}
			utils.Failf("command execution failed: %v", err)
			return
		}
	},
}

func init() {
	RootCmd.AddCommand(execCmd)

	// Interspersed(false) stops parsing kpv flags once the first non-flag arg (the command) is hit.
	// This makes the `--` optional when running commands that have their own flags.
	execCmd.Flags().SetInterspersed(false)

	execCmd.Flags().StringSliceP("key", "k", []string{}, "Name of secret(s) to load (can be specified multiple times)")
	execCmd.Flags().String("key-file", "", "Path to a file containing a list of keys to load (one per line)")
}
