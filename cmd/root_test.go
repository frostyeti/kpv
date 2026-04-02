package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRuntimeCommands(t *testing.T) {
	registerRuntimeCommands()

	root := RootCmd
	require.NotNil(t, root)

	get, _, err := root.Find([]string{"get"})
	require.NoError(t, err)
	require.NotNil(t, get)
	assert.Equal(t, secretsGetCmd.Short, get.Short)

	set, _, err := root.Find([]string{"set"})
	require.NoError(t, err)
	require.NotNil(t, set)
	assert.Equal(t, secretsSetCmd.Short, set.Short)

	ls, _, err := root.Find([]string{"ls"})
	require.NoError(t, err)
	require.NotNil(t, ls)
	assert.Equal(t, secretsLsCmd.Short, ls.Short)

	rm, _, err := root.Find([]string{"rm"})
	require.NoError(t, err)
	require.NotNil(t, rm)
	assert.Equal(t, secretsRmCmd.Short, rm.Short)

	version, _, err := root.Find([]string{"version"})
	require.NoError(t, err)
	require.NotNil(t, version)
	assert.True(t, version.Hidden)

	upgrade, _, err := root.Find([]string{"upgrade"})
	require.NoError(t, err)
	require.NotNil(t, upgrade)
	assert.Equal(t, "upgrade [version]", upgrade.Use)
}
