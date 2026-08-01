package state

import (
	"fmt"

	focusv1 "github.com/EthanKim8683/cpenv/gen/focus/v1"
	problemv1 "github.com/EthanKim8683/cpenv/gen/problem/v1"
)

type State struct {
	Focus    *focusv1.Focus
	Template string
}

func (s *State) FocusedProblem() (*problemv1.Problem, error) {
	if s.Focus == nil {
		return nil, fmt.Errorf("no focus")
	}

	if errMsg := s.Focus.Error; errMsg != "" {
		return nil, fmt.Errorf("extension: %s", errMsg)
	}

	return s.Focus.Problem, nil
}
