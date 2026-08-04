package submit

import (
	"context"
	"fmt"

	submitv1 "github.com/EthanKim8683/cpenv/gen/submit/v1"
	"github.com/EthanKim8683/cpenv/gen/submit/v1/submitv1connect"
)

type Service struct {
	hub *hub[*submitv1.SubmitRequest, *submitv1.SubmitResponse]
}

func (s *Service) Submit(ctx context.Context, req *submitv1.SubmitRequest) (*submitv1.SubmitResponse, error) {
	reply, err := s.hub.tryRequest(ctx, req.GetProblemId(), req)
	if err != nil {
		return nil, fmt.Errorf("submit: %w", err)
	}
	return &submitv1.SubmitResponse{Error: reply.Error}, nil
}

func (s *Service) RequestSubmission(ctx context.Context, req *submitv1.RequestSubmissionRequest) (*submitv1.RequestSubmissionResponse, error) {
	msg, err := s.hub.offer(ctx, req.GetProblemId())
	if err != nil {
		return nil, fmt.Errorf("request submission: %w", err)
	}

	return &submitv1.RequestSubmissionResponse{
		ReplyId:  msg.replyID,
		FileName: msg.req.GetFileName(),
		Content:  msg.req.GetContent(),
	}, nil
}

func (s *Service) CompleteSubmission(_ context.Context, req *submitv1.CompleteSubmissionRequest) (*submitv1.CompleteSubmissionResponse, error) {
	reply := &submitv1.SubmitResponse{Error: req.Error}
	if err := s.hub.reply(req.GetReplyId(), reply); err != nil {
		return nil, fmt.Errorf("complete submission: %w", err)
	}
	return &submitv1.CompleteSubmissionResponse{}, nil
}

var _ submitv1connect.SubmitServiceHandler = (*Service)(nil)

func NewService() *Service {
	return &Service{
		hub: newHub[*submitv1.SubmitRequest, *submitv1.SubmitResponse](),
	}
}
