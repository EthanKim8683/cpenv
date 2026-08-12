package extension

import (
	"context"

	"github.com/EthanKim8683/cpenv/internal/cli"
	submitv1 "github.com/EthanKim8683/cpenv/internal/gen/submit/v1"
	"github.com/EthanKim8683/cpenv/internal/gen/submit/v1/submitv1connect"
)

type SubmitService struct {
	hub *hub[*submitv1.RequestRequest, *submitv1.RequestResponse]
}

func (s *SubmitService) Request(ctx context.Context, req *submitv1.RequestRequest) (*submitv1.RequestResponse, error) {
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
	reply := &submitv1.RequestResponse{Error: req.GetError()}
	if err := s.hub.reply(req.GetReplyId(), reply); err != nil {
		return nil, err
	}
	return &submitv1.ReplyResponse{}, nil
}

var _ submitv1connect.SubmitServiceHandler = (*SubmitService)(nil)

func NewSubmitService() *SubmitService {
	return &SubmitService{hub: newHub[*submitv1.RequestRequest, *submitv1.RequestResponse]()}
}

type SubmitRequester struct {
	client submitv1connect.SubmitServiceClient
}

func (r *SubmitRequester) Request(ctx context.Context, problemID string, fileName string, content []byte) error {
	res, err := r.client.Request(ctx, &submitv1.RequestRequest{
		ProblemId: problemID,
		FileName:  fileName,
		Content:   content,
	})
	if err != nil {
		return err
	}

	if errMsg := res.GetError(); errMsg != "" {
		return &ExtensionError{Msg: errMsg}
	}
	return nil
}

var _ cli.SubmitRequester = (*SubmitRequester)(nil)
