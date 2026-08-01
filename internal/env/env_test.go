package env_test

import (
	"os"
	"testing"

	problemv1 "github.com/EthanKim8683/cpenv/gen/problem/v1"
	"github.com/EthanKim8683/cpenv/internal/env"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnv(t *testing.T) {
	t.Parallel()

	t.Run("create and open env", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()

		create, err := env.Create(fs, &problemv1.Problem{
			Id: t.Name(),
		})
		require.NoError(t, err)

		open, err := env.Open(fs)
		require.NoError(t, err)

		assert.Equal(t, create.Problem(), open.Problem())
	})

	t.Run("nonexistent env", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()

		_, err := env.Open(fs)
		assert.Error(t, err)
		assert.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("clear env", func(t *testing.T) {
		t.Parallel()

		fs := afero.NewMemMapFs()

		create, err := env.Create(fs, &problemv1.Problem{
			Id: "id",
		})
		require.NoError(t, err)

		require.NoError(t, afero.WriteFile(fs, "remove", []byte("remove"), 0644))

		require.NoError(t, create.Clear())

		_, err = fs.Stat("remove")
		assert.ErrorIs(t, err, os.ErrNotExist)

		open, err := env.Open(fs)
		require.NoError(t, err)

		assert.Equal(t, create.Problem(), open.Problem())
	})
}
