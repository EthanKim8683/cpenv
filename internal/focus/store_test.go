package focus

import (
	"path/filepath"
	"testing"

	focusv1 "github.com/EthanKim8683/cpenv/internal/gen/proto/focus/v1"
	problemv1 "github.com/EthanKim8683/cpenv/internal/gen/proto/problem/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore(t *testing.T) {
	t.Parallel()

	t.Run("round trip", func(t *testing.T) {
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
}
