package server_test

import (
	"testing"

	focusv1 "github.com/EthanKim8683/cpenv/gen/focus/v1"
	problemv1 "github.com/EthanKim8683/cpenv/gen/problem/v1"
	"github.com/EthanKim8683/cpenv/internal/server"
	"github.com/EthanKim8683/cpenv/internal/state"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeStateStore struct {
	state *state.State
}

func (s *fakeStateStore) Load() (*state.State, error) {
	return s.state, nil
}

func (s *fakeStateStore) Save(state *state.State) error {
	s.state = state
	return nil
}

func TestFocusService(t *testing.T) {
	t.Parallel()

	tmpl := "template"

	store := &fakeStateStore{
		state: &state.State{
			Focus:    nil,
			Template: tmpl,
		},
	}

	svc := &server.FocusService{
		StateStore: store,
	}

	focus := &focusv1.Focus{
		Problem: &problemv1.Problem{
			Id: t.Name(),
		},
	}

	_, err := svc.SetFocus(t.Context(), &focusv1.SetFocusRequest{
		Focus: focus,
	})
	require.NoError(t, err)
	assert.Equal(t, &state.State{
		Focus:    focus,
		Template: tmpl,
	}, store.state)
}
