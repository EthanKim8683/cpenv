package daemon

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	focusv1 "github.com/EthanKim8683/cpenv/internal/gen/focus/v1"
	focusv1connect "github.com/EthanKim8683/cpenv/internal/gen/focus/v1/focusv1connect"
)

type FocusService struct {
	hub *hub[*focusv1.FocusRequest, *focusv1.FocusResponse]
}

func (s *FocusService) Focus(ctx context.Context, req *focusv1.FocusRequest) (*focusv1.FocusResponse, error) {
	res, err := s.hub.tryRequest(ctx, req.GetEnvId(), req)
	if err != nil {
		if errors.Is(err, ErrNoReceiver) {
			return nil, connect.NewError(connect.CodeFailedPrecondition, err)
		}
		return nil, err
	}
	return res, nil
}

func (s *FocusService) Claim(ctx context.Context, req *focusv1.ClaimRequest) (*focusv1.ClaimResponse, error) {
	msg, err := s.hub.claim(ctx, req.GetEnvId())
	if err != nil {
		return nil, err
	}
	return &focusv1.ClaimResponse{
		ReplyId: msg.replyID,
		Path:    msg.req.GetPath(),
	}, nil
}

func (s *FocusService) Reply(ctx context.Context, req *focusv1.ReplyRequest) (*focusv1.ReplyResponse, error) {
	reply := &focusv1.FocusResponse{Error: req.Error}
	if err := s.hub.reply(req.GetReplyId(), reply); err != nil {
		if errors.Is(err, ErrReplyNotFound) {
			return nil, connect.NewError(connect.CodeNotFound, err)
		}
		return nil, err
	}
	return &focusv1.ReplyResponse{}, nil
}

var _ focusv1connect.FocusServiceHandler = (*FocusService)(nil)

func NewFocusService() *FocusService {
	return &FocusService{hub: newHub[*focusv1.FocusRequest, *focusv1.FocusResponse]()}
}
