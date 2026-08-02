package state_test

import (
	"testing"

	focusv1 "github.com/EthanKim8683/cpenv/gen/focus/v1"
	problemv1 "github.com/EthanKim8683/cpenv/gen/problem/v1"
	"github.com/EthanKim8683/cpenv/internal/state"
	"github.com/stretchr/testify/assert"
)

func TestState(t *testing.T) {
	t.Parallel()

	t.Run("focused problem", func(t *testing.T) {
		t.Parallel()

		problem := &problemv1.Problem{
			Id: t.Name(),
		}

		state := &state.State{
			Focus: &focusv1.Focus{
				Problem: problem,
			},
		}

		problem, err := state.FocusedProblem()
		assert.NoError(t, err)
		assert.Equal(t, problem, problem)
	})

	t.Run("extension error", func(t *testing.T) {
		t.Parallel()

		errMsg := t.Name()

		state := &state.State{
			Focus: &focusv1.Focus{
				Error: errMsg,
			},
		}

		problem, err := state.FocusedProblem()
		assert.Empty(t, problem)
		assert.Error(t, err)
		assert.ErrorContains(t, err, errMsg)
	})
}
