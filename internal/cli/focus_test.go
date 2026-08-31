package cli

import (
	"context"
	"testing"

	"github.com/EthanKim8683/cpenv/internal/config"
	focusv1 "github.com/EthanKim8683/cpenv/internal/gen/focus/v1"
	"github.com/EthanKim8683/cpenv/internal/gen/focus/v1/focusv1connect"
	problemv1 "github.com/EthanKim8683/cpenv/internal/gen/problem/v1"
)

// TODO: write tests

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

type dummyPreferences struct {}

func (p *dummyPreferences) DefaultTemplate() (string, error) {
	return "", nil
}

func (p *dummyPreferences) SetDefaultTemplate(path string) error {
	return nil	
}

var _ Preferences = (*dummyPreferences)(nil)

func TestFocusedProblem(t *testing.T) {
	t.Parallel()

	
}

func TestFocus(t *testing.T) {
	t.Parallel()

	homeDir := t.TempDir()

	cfg := &config.Config{HomeDir: homeDir}
	preferences := &dummyPreferences{}

	t.Run("existing workspace", func(t *testing.T) {
		t.Parallel()

	focusClient := &stubFocusClient{
		focus: &focusv1.Focus{
			Problem: &problemv1.Problem{Id: "existing"},
		},
	}
	cli := &CLI{
		Cfg: cfg,
		FocusClient: focusClient,
		Preferences: preferences,
	}

		cli.Focus(t.Context(), "")
	})

	t.Run("uninitialized workspace", func(t *testing.T) {
		t.Parallel()

	focusClient := &stubFocusClient{
		focus: &focusv1.Focus{
			Problem: &problemv1.Problem{Id: "uninitialized"},
		},
	}
	cli := &CLI{
		Cfg: cfg,
		FocusClient: focusClient,
		Preferences: preferences,
	}

		cli.Focus(t.Context(), "")
	})

	t.Run("nonexistent workspace", func(t *testing.T) {
		t.Parallel()

	focusClient := &stubFocusClient{
		focus: &focusv1.Focus{
			Problem: &problemv1.Problem{Id: "nonexistent"},
		},
	}
	cli := &CLI{
		Cfg: cfg,
		FocusClient: focusClient,
		Preferences: preferences,
	}

		cli.Focus(t.Context(), "")
	})
}