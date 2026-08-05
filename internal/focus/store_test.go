package focus

import (
	"path/filepath"
	"testing"

	focusv1 "github.com/EthanKim8683/cpenv/internal/gen/focus/v1"
	problemv1 "github.com/EthanKim8683/cpenv/internal/gen/problem/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	t.Parallel()

	t.Run("round trip", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "focus.json")
		focus := &focusv1.Focus{
			Problem: &problemv1.Problem{Id: "id"},
			Error:   "error",
		}

		store := &Store{path: path}
		require.NoError(t, store.save(focus))
		gotFocus, err := store.Load()
		assert.NoError(t, err)
		assert.Equal(t, focus, gotFocus)
	})

	t.Run("get problem", func(t *testing.T) {
		t.Parallel()

		t.Run("got problem", func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "focus.json")
			problem := &problemv1.Problem{Id: "id"}

			store := &Store{path: path}
			require.NoError(t, store.save(&focusv1.Focus{Problem: problem}))
			gotProblem, err := store.Problem()
			assert.NoError(t, err)
			assert.Equal(t, problem, gotProblem)
		})

		t.Run("focus error", func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "focus.json")
			errMsg := "error"

			store := &Store{path: path}
			require.NoError(t, store.save(&focusv1.Focus{Error: errMsg}))
			_, err := store.Problem()
			focusErr := &FocusError{}
			assert.ErrorAs(t, err, &focusErr)
			assert.Equal(t, errMsg, focusErr.Message)
		})
	})
}
