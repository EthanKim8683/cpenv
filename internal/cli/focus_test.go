package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/EthanKim8683/cpenv/internal/config"
	activeproblemv1 "github.com/EthanKim8683/cpenv/internal/gen/active_problem/v1"
	"github.com/EthanKim8683/cpenv/internal/gen/active_problem/v1/active_problemv1connect"
	problemv1 "github.com/EthanKim8683/cpenv/internal/gen/problem/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubActiveProblemClient struct {
	activeProblem *activeproblemv1.ActiveProblem
}

func (c *stubActiveProblemClient) Save(context.Context, *activeproblemv1.SaveRequest) (*activeproblemv1.SaveResponse, error) {
	return nil, nil
}

func (c *stubActiveProblemClient) Load(context.Context, *activeproblemv1.LoadRequest) (*activeproblemv1.LoadResponse, error) {
	return &activeproblemv1.LoadResponse{ActiveProblem: c.activeProblem}, nil
}

var _ active_problemv1connect.ActiveProblemServiceClient = (*stubActiveProblemClient)(nil)

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
			ActiveProblemClient: &stubActiveProblemClient{
				activeProblem: &activeproblemv1.ActiveProblem{Problem: problem},
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
			ActiveProblemClient: &stubActiveProblemClient{
				activeProblem: &activeproblemv1.ActiveProblem{Problem: problem},
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
			ActiveProblemClient: &stubActiveProblemClient{
				activeProblem: &activeproblemv1.ActiveProblem{Problem: problem},
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
