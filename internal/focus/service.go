package focus

import (
	"context"
	"errors"
	"fmt"

	focusv1 "github.com/EthanKim8683/cpenv/internal/gen/focus/v1"
	"github.com/EthanKim8683/cpenv/internal/gen/focus/v1/focusv1connect"
)

type Service struct {
	Store *Store
}

func validateFocus(focus *focusv1.Focus) error {
	if focus.GetError() != "" {
		return nil
	}

	if focus.GetProblem() == nil {
		return errors.New("no problem")
	}
	return nil
}

func (s *Service) SetFocus(ctx context.Context, req *focusv1.SetFocusRequest) (*focusv1.SetFocusResponse, error) {
	focus := req.GetFocus()
	if err := validateFocus(focus); err != nil {
		return nil, fmt.Errorf("set focus: %w", err)
	}

	if err := s.Store.save(focus); err != nil {
		return nil, fmt.Errorf("set focus: %w", err)
	}
	return &focusv1.SetFocusResponse{}, nil
}

var _ focusv1connect.FocusServiceHandler = (*Service)(nil)
