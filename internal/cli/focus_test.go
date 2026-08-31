package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/EthanKim8683/cpenv/internal/config"
	focusv1 "github.com/EthanKim8683/cpenv/internal/gen/focus/v1"
	"github.com/EthanKim8683/cpenv/internal/gen/focus/v1/focusv1connect"
	problemv1 "github.com/EthanKim8683/cpenv/internal/gen/problem/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubFocusClient struct {
	focus *focusv1.Focus
}

func (c *stubFocusClient) Save(context.Context, *focusv1.SaveRequest) (*focusv1.SaveResponse, error) {
	return nil, nil
}

func (c *stubFocusClient) Load(context.Context, *focusv1.LoadRequest) (*focusv1.LoadResponse, error) {
	return &focusv1.LoadResponse{Focus: c.focus}, nil
}

var _ focusv1connect.FocusServiceClient = (*stubFocusClient)(nil)

type dummyPreferences struct{}

func (p *dummyPreferences) DefaultTemplate() (string, error) {
	return "", nil
}

func (p *dummyPreferences) SetDefaultTemplate(path string) error {
	return nil
}

var _ Preferences = (*dummyPreferences)(nil)

func TestFocus(t *testing.T) {
	t.Parallel()

	homeDir := t.TempDir()
	templateName, err := filepath.Abs(filepath.Join("testdata", "focus", "flag.star"))
	require.NoError(t, err)

	cfg := &config.Config{HomeDir: homeDir}
	prefs := &dummyPreferences{}

	t.Run("existing workspace", func(t *testing.T) {
		t.Parallel()

		problem := &problemv1.Problem{Id: "existing"}

		cli := &CLI{
			Cfg: cfg,
			FocusClient: &stubFocusClient{
				focus: &focusv1.Focus{Problem: problem},
			},
		}

		dir := cli.workspaceDir(problem.GetId())
		_, err := initWorkspace(dir, problem)
		require.NoError(t, err)

		gotDir, err := cli.Focus(t.Context(), templateName)
		require.NoError(t, err)
		assert.Equal(t, dir, gotDir)

		_, err = os.Stat(filepath.Join(dir, "flag"))
		assert.ErrorIs(t, err, os.ErrNotExist)
	})

	t.Run("uninitialized workspace", func(t *testing.T) {
		t.Parallel()

		problem := &problemv1.Problem{Id: "uninitialized"}

		cli := &CLI{
			Cfg: cfg,
			FocusClient: &stubFocusClient{
				focus: &focusv1.Focus{Problem: problem},
			},
		}

		dir := cli.workspaceDir(problem.GetId())
		require.NoError(t, os.MkdirAll(dir, 0755))

		gotDir, err := cli.Focus(t.Context(), templateName)
		require.NoError(t, err)
		assert.Equal(t, dir, gotDir)

		_, err = os.Stat(filepath.Join(dir, "flag"))
		assert.ErrorIs(t, err, os.ErrNotExist)

		_, err = openWorkspace(dir)
		assert.NoError(t, err)
	})

	t.Run("nonexistent workspace", func(t *testing.T) {
		t.Parallel()

		problem := &problemv1.Problem{Id: "nonexistent"}

		cli := &CLI{
			Cfg: cfg,
			FocusClient: &stubFocusClient{
				focus: &focusv1.Focus{Problem: problem},
			},
			Preferences: prefs,
		}

		dir := cli.workspaceDir(problem.GetId())
		gotDir, err := cli.Focus(t.Context(), templateName)
		require.NoError(t, err)
		assert.Equal(t, dir, gotDir)

		_, err = os.Stat(filepath.Join(dir, "flag"))
		assert.NoError(t, err)

		_, err = openWorkspace(dir)
		assert.NoError(t, err)
	})
}
