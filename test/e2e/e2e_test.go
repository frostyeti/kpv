//go:build integration
// +build integration

package e2e_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildBinary(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "kpv")

	cmd := exec.Command("go", "build", "-o", binPath, "../../main.go")
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	require.NoError(t, err, "Failed to build kpv binary")
	return binPath
}

func runCmd(t *testing.T, binPath string, args ...string) (string, string, error) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	cmd := exec.Command(binPath, args...)
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
