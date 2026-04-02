//go:build integration
// +build integration

package e2e_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	binName := "kpv"
	if runtime.GOOS == "windows" {
		binName = "kpv.exe"
	}
	binPath := filepath.Join(tmpDir, binName)

	cmd := exec.Command("go", "build", "-o", binPath, "../../main.go")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	require.NoError(t, err, "Failed to build kpv binary")
	return binPath
}

func runCmd(t *testing.T, binPath string, args ...string) (string, string, error) {
	return runCmdEnv(t, binPath, nil, args...)
}

func runCmdEnv(t *testing.T, binPath string, env []string, args ...string) (string, string, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(binPath, args...)
	if len(env) > 0 {
		cmd.Env = append(os.Environ(), env...)
	}
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		t.Logf("Cmd failed: %v\nStdout: %s\nStderr: %s", err, stdout.String(), stderr.String())
	}
	return stdout.String(), stderr.String(), err
}

func TestE2E_VaultLifecycle(t *testing.T) {
	binPath := buildBinary(t)
	vaultDir := t.TempDir()
	vaultPath := filepath.Join(vaultDir, "test.kdbx")
	passPath := filepath.Join(vaultDir, "test.key")

	// Create a password file to use throughout the test to avoid OS keyring prompts
	err := os.WriteFile(passPath, []byte("supersecret"), 0600)
	require.NoError(t, err)

	// 1. Init vault
	stdout, _, err := runCmd(t, binPath, "init", "--vault", vaultPath, "--vault-password-file", passPath)
	require.NoError(t, err)
	assert.Contains(t, stdout, "created KeePass vault")

	// 2. Set secret
	_, stderr, err := runCmd(t, binPath, "secrets", "set", "--key", "api-key", "--value", "my-secret-token", "--vault", vaultPath, "--vault-password-file", passPath)
	require.NoError(t, err)
	assert.Contains(t, stderr, "api-key set")

	// 3. Get secret
	stdout, _, err = runCmd(t, binPath, "secrets", "get", "--key", "api-key", "--vault", vaultPath, "--vault-password-file", passPath)
	require.NoError(t, err)
	assert.Equal(t, "my-secret-token\n", stdout)

	// 4. Rm secret
	_, stderr, err = runCmd(t, binPath, "secrets", "rm", "--key", "api-key", "--yes", "--vault", vaultPath, "--vault-password-file", passPath)
	require.NoError(t, err)
	assert.Contains(t, stderr, "deleted secret: api-key")

	// 5. Get secret (should fail)
	_, _, err = runCmd(t, binPath, "secrets", "get", "--key", "api-key", "--vault", vaultPath, "--vault-password-file", passPath)
	require.Error(t, err)
}

func TestE2E_SecretsExtended(t *testing.T) {
	binPath := buildBinary(t)
	vaultDir := t.TempDir()
	vaultPath := filepath.Join(vaultDir, "test.kdbx")
	passPath := filepath.Join(vaultDir, "test.key")

	err := os.WriteFile(passPath, []byte("supersecret"), 0600)
	require.NoError(t, err)

	_, _, err = runCmd(t, binPath, "init", "--vault", vaultPath, "--vault-password-file", passPath)
	require.NoError(t, err)

	// Ensure
	stdout, _, err := runCmd(t, binPath, "secrets", "ensure", "--key", "generated-key", "--vault", vaultPath, "--vault-password-file", passPath)
	require.NoError(t, err)
	assert.NotEmpty(t, stdout)
	assert.Len(t, stdout, 17) // 16 chars + newline

	// Set secret
	_, _, err = runCmd(t, binPath, "secrets", "set", "--key", "my-entry", "--value", "main-secret", "--vault", vaultPath, "--vault-password-file", passPath)
	require.NoError(t, err)

	// Set string
	_, _, err = runCmd(t, binPath, "secrets", "set-string", "--key", "my-entry", "--field", "custom-field", "--value", "custom-value", "--vault", vaultPath, "--vault-password-file", passPath)
	require.NoError(t, err)

	// Get string
	stdout, _, err = runCmd(t, binPath, "secrets", "get-string", "--key", "my-entry", "--field", "custom-field", "--vault", vaultPath, "--vault-password-file", passPath)
	require.NoError(t, err)
	assert.Equal(t, "custom-value\n", stdout)

	// Ls
	stdout, _, err = runCmd(t, binPath, "secrets", "ls", "--vault", vaultPath, "--vault-password-file", passPath)
	require.NoError(t, err)
	assert.Contains(t, stdout, "generated-key")
	assert.Contains(t, stdout, "my-entry")

	// Export
	exportPath := filepath.Join(vaultDir, "export.json")
	_, _, err = runCmd(t, binPath, "secrets", "export", "--json", "--file", exportPath, "--vault", vaultPath, "--vault-password-file", passPath)
	require.NoError(t, err)

	exportContent, err := os.ReadFile(exportPath)
	require.NoError(t, err)
	assert.Contains(t, string(exportContent), "main-secret")

	// Rm my-entry
	_, _, err = runCmd(t, binPath, "secrets", "rm", "--key", "my-entry", "--yes", "--vault", vaultPath, "--vault-password-file", passPath)
	require.NoError(t, err)

	// Sync
	_, _, err = runCmd(t, binPath, "secrets", "sync", "--json", "--file", exportPath, "--vault", vaultPath, "--vault-password-file", passPath)
	require.NoError(t, err)

	// Verify my-entry is back
	stdout, _, err = runCmd(t, binPath, "secrets", "get", "--key", "my-entry", "--vault", vaultPath, "--vault-password-file", passPath)
	require.NoError(t, err)
	assert.Equal(t, "main-secret\n", stdout)

	// Import (clean start)
	vaultPath2 := filepath.Join(vaultDir, "test2.kdbx")
	_, _, err = runCmd(t, binPath, "init", "--vault", vaultPath2, "--vault-password-file", passPath)
	require.NoError(t, err)

	_, _, err = runCmd(t, binPath, "secrets", "import", "--json", "--file", exportPath, "--vault", vaultPath2, "--vault-password-file", passPath)
	require.NoError(t, err)

	stdout, _, err = runCmd(t, binPath, "secrets", "get", "--key", "my-entry", "--vault", vaultPath2, "--vault-password-file", passPath)
	require.NoError(t, err)
	assert.Equal(t, "main-secret\n", stdout)
}

func TestE2E_Exec(t *testing.T) {
	binPath := buildBinary(t)
	vaultDir := t.TempDir()
	vaultPath := filepath.Join(vaultDir, "test.kdbx")
	passPath := filepath.Join(vaultDir, "test.key")
	keyFilePath := filepath.Join(vaultDir, "keys.txt")

	err := os.WriteFile(passPath, []byte("supersecret"), 0600)
	require.NoError(t, err)

	_, _, err = runCmd(t, binPath, "init", "--vault", vaultPath, "--vault-password-file", passPath)
	require.NoError(t, err)

	// Set some secrets
	_, _, err = runCmd(t, binPath, "secrets", "set", "--key", "api-token", "--value", "secret-token", "--vault", vaultPath, "--vault-password-file", passPath)
	require.NoError(t, err)
	_, _, err = runCmd(t, binPath, "secrets", "set", "--key", "db-pass", "--value", "secret-db", "--vault", vaultPath, "--vault-password-file", passPath)
	require.NoError(t, err)
	_, _, err = runCmd(t, binPath, "secrets", "set", "--key", "ignore-me", "--value", "not-loaded", "--vault", vaultPath, "--vault-password-file", passPath)
	require.NoError(t, err)

	// 1. Exec with --key loading specific secrets
	stdout, _, err := runCmd(t, binPath, "exec", "--vault", vaultPath, "--vault-password-file", passPath, "--key", "api-token", "--", "env")
	require.NoError(t, err)
	assert.Contains(t, stdout, "API_TOKEN=secret-token")
	assert.NotContains(t, stdout, "DB_PASS=")

	// 2. Exec with --key multiple times
	stdout, _, err = runCmd(t, binPath, "exec", "--vault", vaultPath, "--vault-password-file", passPath, "-k", "api-token", "-k", "db-pass", "env")
	require.NoError(t, err)
	assert.Contains(t, stdout, "API_TOKEN=secret-token")
	assert.Contains(t, stdout, "DB_PASS=secret-db")
	assert.NotContains(t, stdout, "IGNORE_ME=")

	// 3. Exec with --key-file
	err = os.WriteFile(keyFilePath, []byte("api-token\n# comment\ndb-pass\n"), 0600)
	require.NoError(t, err)
	stdout, _, err = runCmd(t, binPath, "exec", "--vault", vaultPath, "--vault-password-file", passPath, "--key-file", keyFilePath, "--", "env")
	require.NoError(t, err)
	assert.Contains(t, stdout, "API_TOKEN=secret-token")
	assert.Contains(t, stdout, "DB_PASS=secret-db")
	assert.NotContains(t, stdout, "IGNORE_ME=")

	// 4. Exec loading ALL secrets (no keys provided)
	stdout, _, err = runCmd(t, binPath, "exec", "--vault", vaultPath, "--vault-password-file", passPath, "env")
	require.NoError(t, err)
	assert.Contains(t, stdout, "API_TOKEN=secret-token")
	assert.Contains(t, stdout, "DB_PASS=secret-db")
	assert.Contains(t, stdout, "IGNORE_ME=not-loaded")
}

func TestE2E_Config(t *testing.T) {
	binPath := buildBinary(t)
	configDir := t.TempDir()

	// Environment configured to isolate config changes
	env := []string{
		"XDG_CONFIG_HOME=" + configDir, // Linux
		"LOCALAPPDATA=" + configDir,    // Windows Local
		"APPDATA=" + configDir,         // Windows Roaming
		"HOME=" + configDir,            // macOS fallback
		"USERPROFILE=" + configDir,
	}

	// Set config
	_, _, err := runCmdEnv(t, binPath, env, "config", "set", "test.key", "test-value")
	require.NoError(t, err)

	// Get config
	stdout, stderr, err := runCmdEnv(t, binPath, env, "config", "get", "test.key")
	t.Logf("get test.key stdout: %q, stderr: %q", stdout, stderr)
	require.NoError(t, err)
	assert.Equal(t, "test-value\n", stdout)

	// Ls config
	stdout, _, err = runCmdEnv(t, binPath, env, "config", "ls")
	require.NoError(t, err)
	assert.Contains(t, stdout, "test.key=test-value")

	// Rm config
	_, _, err = runCmdEnv(t, binPath, env, "config", "rm", "test.key")
	require.NoError(t, err)

	// Verify rm
	stdout, _, err = runCmdEnv(t, binPath, env, "config", "get", "test.key")
	require.NoError(t, err)
	assert.Contains(t, stdout, "not found")

	// Aliases
	_, _, err = runCmdEnv(t, binPath, env, "config", "aliases", "set", "myalias", "/tmp/vault.kdbx")
	require.NoError(t, err)

	stdout, _, err = runCmdEnv(t, binPath, env, "config", "aliases", "get", "myalias")
	require.NoError(t, err)
	assert.Equal(t, "/tmp/vault.kdbx\n", stdout)

	stdout, _, err = runCmdEnv(t, binPath, env, "config", "aliases", "ls")
	require.NoError(t, err)
	assert.Contains(t, stdout, "myalias=/tmp/vault.kdbx")

	_, _, err = runCmdEnv(t, binPath, env, "config", "aliases", "rm", "myalias")
	require.NoError(t, err)
}
