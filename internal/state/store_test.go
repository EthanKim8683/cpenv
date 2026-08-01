package state_test

import (
	"path/filepath"
	"testing"

	focusv1 "github.com/EthanKim8683/cpenv/gen/focus/v1"
	problemv1 "github.com/EthanKim8683/cpenv/gen/problem/v1"
	"github.com/EthanKim8683/cpenv/internal/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFileStore(t *testing.T) {
	path := filepath.Join(t.TempDir(), "state.json")

	store := state.NewFileStore(path)

	load, err := store.Load()
	require.NoError(t, err)
	assert.Empty(t, load)

	save := &state.State{
		Focus: &focusv1.Focus{
			Problem: &problemv1.Problem{
				Id: t.Name(),
			},
		},
	}
	require.NoError(t, store.Save(save))

	load, err = store.Load()
	require.NoError(t, err)
	assert.Equal(t, save, load)
}
