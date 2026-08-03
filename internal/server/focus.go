package server

import (
	"context"
	"fmt"

	focusv1 "github.com/EthanKim8683/cpenv/gen/focus/v1"
	"github.com/EthanKim8683/cpenv/gen/focus/v1/focusv1connect"
	"github.com/EthanKim8683/cpenv/internal/state"
)

type FocusService struct {
	StateStore state.Store
}

func (s *FocusService) SetFocus(_ context.Context, req *focusv1.SetFocusRequest) (*focusv1.SetFocusResponse, error) {
	if err := s.StateStore.Update(func(st *state.State) error {
		st.Focus = req.Focus
		return nil
	}); err != nil {
		return nil, fmt.Errorf("set focus: %w", err)
	}

	return &focusv1.SetFocusResponse{}, nil
}

var _ focusv1connect.FocusServiceHandler = (*FocusService)(nil)
