package cmd

import (
	"os"
	"os/exec"

	"github.com/frostyeti/kpv/internal/utils"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec [flags] -- <command> [args...]",
	Short: "Execute a command with secrets loaded as environment variables",
	Long: `Execute a command with secrets loaded as environment variables.

To avoid leaking your entire password database into the environment, you must
explicitly specify which secrets you want to load using --key.

The secrets are formatted as SCREAMING_SNAKE_CASE environment variables.
For example, a secret named "my-api-token" becomes "MY_API_TOKEN".

Examples:
  # Execute 'npm start' with specific secrets
  kpv exec --key db-password --key api-token -- npm start

  # Combine with vault path aliases
  kpv exec -V prod-vault -k prod-db-password -- ./deploy.sh
`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		keys, _ := cmd.Flags().GetStringSlice("key")

		if len(keys) == 0 {
			utils.Failf("at least one --key must be provided to avoid loading the entire vault")
			return
		}

		kdbx, _, err := utils.OpenKeePass(cmd)
		if err != nil {
			utils.Failf("opening KeePass vault failed: %v", err)
			return
		}

		// Retrieve all requested values
		values := map[string]string{}
		for _, key := range keys {
			entry := kdbx.FindEntry(key)
			if entry == nil {
				utils.Failf("secret %s not found", key)
				return
			}
			values[key] = entry.GetPassword()
		}

		// Prepare the environment for the child process
		env := os.Environ()
		for k, v := range values {
			envVar := toScreamingSnakeCase(k)
			env = append(env, envVar+"="+v)
		}

		// Prepare the command
		execArgs := args[1:]
		childCmd := exec.Command(args[0], execArgs...)
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

	execCmd.Flags().StringSliceP("key", "k", []string{}, "Name of secret(s) to load (can be specified multiple times)")
}
