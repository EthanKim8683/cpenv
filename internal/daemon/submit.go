package daemon

import (
	"context"
	"errors"

	"connectrpc.com/connect"
	submitv1 "github.com/EthanKim8683/cpenv/internal/gen/submit/v1"
	"github.com/EthanKim8683/cpenv/internal/gen/submit/v1/submitv1connect"
)

type SubmitService struct {
	hub *hub[*submitv1.SubmitRequest, *submitv1.SubmitResponse]
}

func (s *SubmitService) Submit(ctx context.Context, req *submitv1.SubmitRequest) (*submitv1.SubmitResponse, error) {
	res, err := s.hub.tryRequest(ctx, req.GetProblemId(), req)
	if err != nil {
		switch {
		case errors.Is(err, ErrNoReceiver):
			return nil, connect.NewError(connect.CodeFailedPrecondition, err)
		case errors.Is(err, context.Canceled):
			return nil, connect.NewError(connect.CodeCanceled, err)
		case errors.Is(err, context.DeadlineExceeded):
			return nil, connect.NewError(connect.CodeDeadlineExceeded, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return res, nil
}

func (s *SubmitService) Claim(ctx context.Context, req *submitv1.ClaimRequest) (*submitv1.ClaimResponse, error) {
	msg, err := s.hub.claim(ctx, req.GetProblemId())
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			return nil, connect.NewError(connect.CodeCanceled, err)
		case errors.Is(err, context.DeadlineExceeded):
			return nil, connect.NewError(connect.CodeDeadlineExceeded, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return &submitv1.ClaimResponse{
		ReplyId:  msg.replyID,
		FileName: msg.req.GetFileName(),
		Content:  msg.req.GetContent(),
	}, nil
}

func (s *SubmitService) Reply(ctx context.Context, req *submitv1.ReplyRequest) (*submitv1.ReplyResponse, error) {
	reply := &submitv1.SubmitResponse{Error: req.Error}
	if err := s.hub.reply(req.GetReplyId(), reply); err != nil {
		switch {
		case errors.Is(err, ErrReplyNotFound):
			return nil, connect.NewError(connect.CodeNotFound, err)
		case errors.Is(err, context.Canceled):
			return nil, connect.NewError(connect.CodeCanceled, err)
		case errors.Is(err, context.DeadlineExceeded):
			return nil, connect.NewError(connect.CodeDeadlineExceeded, err)
		}
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return &submitv1.ReplyResponse{}, nil
}

var _ submitv1connect.SubmitServiceHandler = (*SubmitService)(nil)

func NewSubmitService() *SubmitService {
	return &SubmitService{hub: newHub[*submitv1.SubmitRequest, *submitv1.SubmitResponse]()}
}
