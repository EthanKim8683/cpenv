package server

import (
	"context"
	"fmt"

	focusv1 "github.com/EthanKim8683/cpenv/gen/focus/v1"
	"github.com/EthanKim8683/cpenv/gen/focus/v1/focusv1connect"
)

type FocusService struct {
}

func (s *FocusService) Focus(ctx context.Context, req *focusv1.FocusRequest) (*focusv1.FocusResponse, error) {
	fmt.Println("Focus", req.Problem)
	return &focusv1.FocusResponse{}, nil
}

var _ focusv1connect.FocusServiceHandler = (*FocusService)(nil)
