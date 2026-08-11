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

		w1, err := createWorkspace(dir, problem)
		require.NoError(t, err)
		require.NoError(t, w1.close())

		w2, err := openWorkspace(dir)
		require.NoError(t, err)
		assert.Equal(t, problem, w2.problem)
		require.NoError(t, w1.close())
	})

	t.Run("idempotent close", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		problem1 := &problemv1.Problem{Id: "id1"}
		problem2 := &problemv1.Problem{Id: "id2"}

		w1, err := createWorkspace(dir, problem1)
		require.NoError(t, err)
		require.NoError(t, w1.close())

		w2, err := createWorkspace(dir, problem2)
		require.NoError(t, err)
		require.NoError(t, w2.close())

		require.NoError(t, w1.close())

		w3, err := openWorkspace(dir)
		require.NoError(t, err)
		assert.Equal(t, problem2, w3.problem)
	})
}
