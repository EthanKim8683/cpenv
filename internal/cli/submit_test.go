package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveSolution(t *testing.T) {
	t.Parallel()

	cwd := filepath.Join(t.TempDir(), "cwd")
	relPath := "sol.cpp"
	cwdRelPath := filepath.Join(cwd, "sol.cpp")
	absPath := filepath.Join(t.TempDir(), "sol.cpp")

	require.NoError(t, os.MkdirAll(cwd, 0755))
	require.NoError(t, os.WriteFile(cwdRelPath, nil, 0644))
	require.NoError(t, os.WriteFile(absPath, nil, 0644))

	t.Run("no solutions", func(t *testing.T) {
		t.Parallel()

		s, err := resolveSolution(t.TempDir(), "")
		require.Error(t, err)
		assert.ErrorContains(t, err, "no sol.* files")
		assert.Nil(t, s)
	})

	t.Run("relative to cwd", func(t *testing.T) {
		t.Parallel()

		s, err := resolveSolution(cwd, relPath)
		require.NoError(t, err)
		assert.Equal(t, cwdRelPath, s.path)
		assert.Equal(t, []byte(""), s.content)
	})

	t.Run("absolute path", func(t *testing.T) {
		t.Parallel()

		s, err := resolveSolution(t.TempDir(), absPath)
		require.NoError(t, err)
		assert.Equal(t, absPath, s.path)
		assert.Equal(t, []byte(""), s.content)
	})
}
