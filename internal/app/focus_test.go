package app_test

import (
	"os"
	"path/filepath"
	"testing"

	focusv1 "github.com/EthanKim8683/cpenv/gen/focus/v1"
	problemv1 "github.com/EthanKim8683/cpenv/gen/problem/v1"
	"github.com/EthanKim8683/cpenv/internal/app"
	"github.com/EthanKim8683/cpenv/internal/config"
	"github.com/EthanKim8683/cpenv/internal/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFocus(t *testing.T) {
	t.Parallel()

	wd, err := os.Getwd()
	require.NoError(t, err)

	tmplDir := filepath.Join(wd, "testdata", "known-templates")
	wsDir := filepath.Join(t.TempDir(), "workspaces")

	t.Run("uses focused problem", func(t *testing.T) {
		t.Parallel()

		problemID := t.Name()
		statePath := filepath.Join(t.TempDir(), "state.json")

		stateStore := state.NewFileStore(statePath)
		require.NoError(t, stateStore.Save(&state.State{
			Focus: &focusv1.Focus{
				Problem: &problemv1.Problem{
					Id: problemID,
				},
			},
		}))

		app := app.NewApp(&config.Config{
			StatePath:     statePath,
			TemplatesDir:  tmplDir,
			WorkspacesDir: wsDir,
		}, nil)

		dir, err := app.Focus("1.star")
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(wsDir, problemID), dir)
	})

	t.Run("uses default template", func(t *testing.T) {
		t.Parallel()

		problemID := t.Name()
		statePath := filepath.Join(t.TempDir(), "state.json")

		stateStore := state.NewFileStore(statePath)
		require.NoError(t, stateStore.Save(&state.State{
			Focus: &focusv1.Focus{
				Problem: &problemv1.Problem{
					Id: problemID,
				},
			},
			Template: filepath.Join(tmplDir, "1.star"),
		}))

		app := app.NewApp(&config.Config{
			StatePath:     statePath,
			TemplatesDir:  tmplDir,
			WorkspacesDir: wsDir,
		}, nil)

		dir, err := app.Focus("")
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(wsDir, problemID), dir)

		tmpl, err := os.ReadFile(filepath.Join(dir, "template"))
		assert.NoError(t, err)
		assert.Equal(t, "1", string(tmpl))
	})

	t.Run("uses any template if no default", func(t *testing.T) {
		t.Parallel()

		problemID := t.Name()
		statePath := filepath.Join(t.TempDir(), "state.json")

		wd, err := os.Getwd()
		require.NoError(t, err)

		stateStore := state.NewFileStore(statePath)
		require.NoError(t, stateStore.Save(&state.State{
			Focus: &focusv1.Focus{
				Problem: &problemv1.Problem{
					Id: problemID,
				},
			},
			Template: filepath.Join(wd, "testdata", "unknown-templates", "3.star"),
		}))

		app := app.NewApp(&config.Config{
			StatePath:     statePath,
			TemplatesDir:  tmplDir,
			WorkspacesDir: wsDir,
		}, nil)

		dir, err := app.Focus("")
		assert.NoError(t, err)

		_, err = os.Stat(filepath.Join(dir, "template"))
		assert.NoError(t, err)
	})

	t.Run("updates state", func(t *testing.T) {
		t.Parallel()

		problemID := t.Name()
		statePath := filepath.Join(t.TempDir(), "state.json")

		stateStore := state.NewFileStore(statePath)
		require.NoError(t, stateStore.Save(&state.State{
			Focus: &focusv1.Focus{
				Problem: &problemv1.Problem{
					Id: problemID,
				},
			},
		}))

		app := app.NewApp(&config.Config{
			StatePath:     statePath,
			TemplatesDir:  tmplDir,
			WorkspacesDir: wsDir,
		}, nil)

		_, err := app.Focus("1.star")
		assert.NoError(t, err)

		state, err := stateStore.Load()
		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(tmplDir, "1.star"), state.Template)
	})
}
