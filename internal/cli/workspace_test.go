package cli

import (
	"testing"

	problemv1 "github.com/EthanKim8683/cpenv/internal/gen/problem/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWorkspace(t *testing.T) {
	t.Parallel()

	t.Run("round trip", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		problem := &problemv1.Problem{Id: "id"}

		_, err := initWorkspace(dir, problem)
		require.NoError(t, err)

		w, err := openWorkspace(dir)
		require.NoError(t, err)
		assert.Equal(t, problem, w.problem)
	})

	t.Run("clear", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		problem := &problemv1.Problem{Id: "id"}

		w1, err := initWorkspace(dir, problem)
		require.NoError(t, err)

		require.NoError(t, w1.clear())

		w2, err := openWorkspace(dir)
		require.NoError(t, err)
		assert.Equal(t, problem, w2.problem)
	})
}
