package extension

import (
	"context"

	"github.com/EthanKim8683/cpenv/internal/cli"
	submitv1 "github.com/EthanKim8683/cpenv/internal/gen/submit/v1"
	"github.com/EthanKim8683/cpenv/internal/gen/submit/v1/submitv1connect"
)

type SubmitService struct {
	hub *hub[*submitv1.SubmitRequest, *submitv1.SubmitResponse]
}

func (s *SubmitService) Submit(ctx context.Context, req *submitv1.SubmitRequest) (*submitv1.SubmitResponse, error) {
	res, err := s.hub.request(ctx, req.GetProblemId(), req)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *SubmitService) Claim(ctx context.Context, req *submitv1.ClaimRequest) (*submitv1.ClaimResponse, error) {
	msg, err := s.hub.claim(ctx, req.GetProblemId())
	if err != nil {
		return nil, err
	}
	return &submitv1.ClaimResponse{
		ReplyId:  msg.replyID,
		FileName: msg.req.GetFileName(),
		Content:  msg.req.GetContent(),
	}, nil
}

func (s *SubmitService) Reply(ctx context.Context, req *submitv1.ReplyRequest) (*submitv1.ReplyResponse, error) {
	reply := &submitv1.SubmitResponse{Error: req.GetError()}
	if err := s.hub.reply(req.GetReplyId(), reply); err != nil {
		return nil, err
	}
	return &submitv1.ReplyResponse{}, nil
}

var _ submitv1connect.SubmitServiceHandler = (*SubmitService)(nil)

func NewSubmitService() *SubmitService {
	return &SubmitService{hub: newHub[*submitv1.SubmitRequest, *submitv1.SubmitResponse]()}
}

type Submitter struct {
	client submitv1connect.SubmitServiceClient
}

func (r *Submitter) Submit(ctx context.Context, req *submitv1.SubmitRequest) error {
	res, err := r.client.Submit(ctx, req)
	if err != nil {
		return err
	}

	if errMsg := res.GetError(); errMsg != "" {
		return &ExtensionError{Msg: errMsg}
	}
	return nil
}

var _ cli.Submitter = (*Submitter)(nil)
