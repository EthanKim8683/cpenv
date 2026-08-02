package app

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultSolFile(t *testing.T) {
	t.Parallel()

	workingDir := filepath.Join("testdata", "workspace")

	a := &App{
		WorkingDir: workingDir,
	}

	solFile, err := a.defaultSolFile()
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(workingDir, "sol.cpp"), solFile)
}

func TestResolveSolFile(t *testing.T) {
	t.Parallel()

	t.Run("relative path", func(t *testing.T) {
		t.Parallel()

		workingDir := t.TempDir()
		relSolFile := "relative"

		a := &App{
			WorkingDir: workingDir,
		}

		solFile, err := a.resolveSolFile(relSolFile)
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(workingDir, relSolFile), solFile)
	})

	t.Run("absolute path", func(t *testing.T) {
		t.Parallel()

		absSolFile := filepath.Join(t.TempDir(), "absolute")

		a := &App{}

		solFile, err := a.resolveSolFile(absSolFile)
		assert.NoError(t, err)
		assert.Equal(t, absSolFile, solFile)
	})

	t.Run("empty path", func(t *testing.T) {
		t.Parallel()

		workingDir := filepath.Join("testdata", "workspace")

		a := &App{
			WorkingDir: workingDir,
		}

		solFile, err := a.resolveSolFile("")
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(workingDir, "sol.cpp"), solFile)
	})
}
