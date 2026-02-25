package utils_test

import (
	"path/filepath"
	"testing"

	"github.com/frostyeti/kpv/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestResolveVaultPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "Absolute Path",
			input:    "/tmp/test.kdbx",
			expected: "/tmp/test.kdbx",
			wantErr:  false,
		},
		{
			name:     "File URI",
			input:    "file:///tmp/test.kdbx",
			expected: "/tmp/test.kdbx",
			wantErr:  false,
		},
		{
			name:     "Kpv URI",
			input:    "kpv:///tmp/test.kdbx",
			expected: "/tmp/test.kdbx",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := utils.ResolveVaultPath(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Use exact match for absolute paths
				if filepath.IsAbs(tt.expected) {
					assert.Equal(t, tt.expected, res.Path)
				} else {
					assert.Contains(t, filepath.ToSlash(res.Path), tt.expected)
				}
			}
		})
	}
}

func TestGenerateSecretWithOptions(t *testing.T) {
	secret, err := utils.GenerateSecretWithOptions(32, false, false, false, false, "", "")
	assert.NoError(t, err)
	assert.Len(t, secret, 32)

	// Test chars override
	secret, err = utils.GenerateSecretWithOptions(10, false, false, false, false, "", "a")
	assert.NoError(t, err)
	assert.Len(t, secret, 10)
	assert.Equal(t, "aaaaaaaaaa", secret)
}
