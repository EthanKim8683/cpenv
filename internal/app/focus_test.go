package app_test

import (
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

	templatesDir := "./testdata"
	envsDir := filepath.Join(t.TempDir(), "envs")
	statePath := filepath.Join(t.TempDir(), "state.json")

	stateStore := state.NewFileStore(statePath)

	require.NoError(t, stateStore.Save(&state.State{
		Focus: &focusv1.Focus{
			Problem: &problemv1.Problem{
				Id: "id",
			},
		},
	}))

	app := &app.App{
		Cfg: &config.Config{
			TemplatesDir: templatesDir,
			EnvsDir:      envsDir,
		},
		StateStore: stateStore,
	}

	dir, err := app.Focus("")
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(envsDir, "id"), dir)
}
