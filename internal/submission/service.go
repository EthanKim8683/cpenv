package submission

import (
	"context"

	submissionv1 "github.com/EthanKim8683/cpenv/internal/gen/submission/v1"
	"github.com/EthanKim8683/cpenv/internal/gen/submission/v1/submissionv1connect"
)

type Service struct {
	Store  *Store
	SaveCh chan<- *submissionv1.Submission
}

func (s *Service) Save(ctx context.Context, req *submissionv1.SaveRequest) (*submissionv1.SaveResponse, error) {
	if err := s.Store.save(req.GetSubmissions()); err != nil {
		return nil, err
	}

	if ch := s.SaveCh; ch != nil {
		for _, sub := range req.GetSubmissions() {
			select {
			case ch <- sub:
			default:
			}
		}
	}

	return &submissionv1.SaveResponse{}, nil
}

var _ submissionv1connect.SubmissionServiceHandler = (*Service)(nil)
