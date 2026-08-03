package state_test

import (
	"path/filepath"
	"strconv"
	"sync"
	"testing"

	focusv1 "github.com/EthanKim8683/cpenv/gen/focus/v1"
	problemv1 "github.com/EthanKim8683/cpenv/gen/problem/v1"
	"github.com/EthanKim8683/cpenv/internal/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStore(t *testing.T) {
	t.Parallel()

	t.Run("round trip", func(t *testing.T) {
		t.Parallel()

		focus := &focusv1.Focus{Problem: &problemv1.Problem{Id: "id"}}
		template := "template"

		store := state.NewFileStore(filepath.Join(t.TempDir(), "state.json"))

		loaded, err := store.Load()
		require.NoError(t, err)
		assert.Empty(t, loaded)

		require.NoError(t, store.Update(func(st *state.State) error {
			st.Focus = focus
			st.Template = template
			return nil
		}))

		loaded, err = store.Load()
		require.NoError(t, err)
		assert.Equal(t, focus, loaded.Focus)
		assert.Equal(t, template, loaded.Template)
	})

	t.Run("concurrent updates", func(t *testing.T) {
		t.Parallel()

		path := filepath.Join(t.TempDir(), "state.json")

		var mu sync.Mutex
		loaded := make(map[string]struct{})
		saved := make(map[string]struct{})

		wg := sync.WaitGroup{}
		for i := range 20 {
			wg.Go(func() {
				store := state.NewFileStore(path)

				require.NoError(t, store.Update(func(s *state.State) error {
					mu.Lock()
					defer mu.Unlock()

					if tmpl := s.Template; tmpl != "" {
						loaded[tmpl] = struct{}{}
					}

					tmpl := strconv.Itoa(i)
					s.Template = tmpl
					saved[tmpl] = struct{}{}

					return nil
				}))
			})
		}
		wg.Wait()

		store := state.NewFileStore(path)

		st, err := store.Load()
		require.NoError(t, err)

		if tmpl := st.Template; tmpl != "" {
			loaded[tmpl] = struct{}{}
		}

		assert.Equal(t, loaded, saved)
	})
}
