package submission

import (
	"context"

	"connectrpc.com/connect"
	submissionv1 "github.com/EthanKim8683/cpenv/internal/gen/submission/v1"
	"github.com/EthanKim8683/cpenv/internal/gen/submission/v1/submissionv1connect"
)

type Service struct {
	defaultLimit int
	store        *store
}

func (s *Service) Save(ctx context.Context, req *submissionv1.SaveRequest) (*submissionv1.SaveResponse, error) {
	return nil, nil
}

func (s *Service) Tail(ctx context.Context, req *submissionv1.TailRequest) (*submissionv1.TailResponse, error) {
	limit := int(req.GetLimit())
	if limit == 0 {
		limit = s.defaultLimit
	}

	var subs []*submissionv1.Submission
	var err error
	if pID := req.GetProblemId(); pID != "" {
		subs, err = s.store.tailProblem(pID, limit)
	} else {
		subs, err = s.store.tail(limit)
	}
	if err != nil {
		return nil, err
	}

	return &submissionv1.TailResponse{Submissions: subs}, nil
}

func (s *Service) Subscribe(ctx context.Context, req *submissionv1.SubscribeRequest, stream *connect.ServerStream[submissionv1.SubscribeResponse]) error {
	return nil
}

var _ submissionv1connect.SubmissionServiceHandler = (*Service)(nil)
