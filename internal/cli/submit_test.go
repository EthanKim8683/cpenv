package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCLI_resolveSolution(t *testing.T) {
	t.Parallel()

	cwd := t.TempDir()

	cli := &CLI{CWD: cwd}

	t.Run("void", func(t *testing.T) {
		path, err := cli.resolveSolution("")
		assert.Error(t, err)
		assert.Empty(t, path)
	})

	t.Run("glob", func(t *testing.T) {
		relName := "sol.cpp"

		require.NoError(t, os.WriteFile(filepath.Join(cwd, relName), nil, 0644))

		path, err := cli.resolveSolution("")
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(cwd, relName), path)
	})

	t.Run("relative path", func(t *testing.T) {
		relName := "sol.cpp"

		path, err := cli.resolveSolution(relName)
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(cwd, relName), path)
	})

	t.Run("absolute path", func(t *testing.T) {
		absPath := filepath.Join(cwd, "abs.cpp")

		require.NoError(t, os.WriteFile(absPath, nil, 0644))

		path, err := cli.resolveSolution(absPath)
		assert.NoError(t, err)
		assert.Equal(t, absPath, path)
	})
}
