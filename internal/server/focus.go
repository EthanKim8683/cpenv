package server

import (
	"context"
	"fmt"
	"sync"

	focusv1 "github.com/EthanKim8683/cpenv/gen/focus/v1"
	"github.com/EthanKim8683/cpenv/gen/focus/v1/focusv1connect"
	"github.com/EthanKim8683/cpenv/internal/state"
)

type FocusService struct {
	mu         sync.Mutex
	StateStore state.Store
}

func (s *FocusService) SetFocus(_ context.Context, req *focusv1.SetFocusRequest) (*focusv1.SetFocusResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, err := s.StateStore.Load()
	if err != nil {
		return nil, fmt.Errorf("set focus: %w", err)
	}

	state.Focus = req.Focus

	if err := s.StateStore.Save(state); err != nil {
		return nil, fmt.Errorf("set focus: %w", err)
	}

	return &focusv1.SetFocusResponse{}, nil
}

var _ focusv1connect.FocusServiceHandler = (*FocusService)(nil)
