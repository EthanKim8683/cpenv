package workspace_test

import (
	"os"
	"testing"

	problemv1 "github.com/EthanKim8683/cpenv/gen/problem/v1"
	"github.com/EthanKim8683/cpenv/internal/workspace"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspace(t *testing.T) {
	t.Parallel()

	t.Run("round trip", func(t *testing.T) {
		t.Parallel()

		artifact := "artifact"

		fs := afero.NewMemMapFs()

		created, err := workspace.Create(fs, &problemv1.Problem{Id: "id"})
		require.NoError(t, err)

		require.NoError(t, afero.WriteFile(fs, artifact, []byte("artifact"), 0644))

		require.NoError(t, created.Clear())

		_, err = fs.Stat(artifact)
		assert.ErrorIs(t, err, os.ErrNotExist)

		opened, err := workspace.Open(fs)
		require.NoError(t, err)
		assert.Equal(t, created.Problem(), opened.Problem())
	})
}
