package server

import (
	"context"

	observationv1 "github.com/EthanKim8683/cpenv/gen/observation/v1"
	observationv1connect "github.com/EthanKim8683/cpenv/gen/observation/v1/observationv1connect"
)

type ObservationService struct {
	OnReportContest func(ctx context.Context, req *observationv1.ReportContestRequest) error
	OnReportProblem func(ctx context.Context, req *observationv1.ReportProblemRequest) error
	OnFocusTab      func(ctx context.Context, req *observationv1.FocusTabRequest) error
}

func (s *ObservationService) ReportContest(ctx context.Context, req *observationv1.ReportContestRequest) (*observationv1.ReportContestResponse, error) {
	if err := s.OnReportContest(ctx, req); err != nil {
		return nil, err
	}
	return &observationv1.ReportContestResponse{}, nil
}

func (s *ObservationService) ReportProblem(ctx context.Context, req *observationv1.ReportProblemRequest) (*observationv1.ReportProblemResponse, error) {
	if err := s.OnReportProblem(ctx, req); err != nil {
		return nil, err
	}
	return &observationv1.ReportProblemResponse{}, nil
}

func (s *ObservationService) FocusTab(ctx context.Context, req *observationv1.FocusTabRequest) (*observationv1.FocusTabResponse, error) {
	if err := s.OnFocusTab(ctx, req); err != nil {
		return nil, err
	}
	return &observationv1.FocusTabResponse{}, nil
}

var _ observationv1connect.ObservationServiceHandler = (*ObservationService)(nil)
