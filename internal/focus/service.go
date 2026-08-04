package focus

import (
	"context"
	"fmt"

	focusv1 "github.com/EthanKim8683/cpenv/gen/focus/v1"
	"github.com/EthanKim8683/cpenv/gen/focus/v1/focusv1connect"
)

type Service struct {
	Store *Store
}

func (s *Service) SetFocus(ctx context.Context, req *focusv1.SetFocusRequest) (*focusv1.SetFocusResponse, error) {
	if err := s.Store.save(req.GetFocus()); err != nil {
		return nil, fmt.Errorf("set focus: %w", err)
	}
	return &focusv1.SetFocusResponse{}, nil
}

var _ focusv1connect.FocusServiceHandler = (*Service)(nil)
