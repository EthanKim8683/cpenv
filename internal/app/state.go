package state

import (
	"context"

	contestv1 "github.com/EthanKim8683/cpenv/gen/contest/v1"
	observationv1 "github.com/EthanKim8683/cpenv/gen/observation/v1"
	problemv1 "github.com/EthanKim8683/cpenv/gen/problem/v1"
	tabv1 "github.com/EthanKim8683/cpenv/gen/tab/v1"
)

type State struct {
	contests map[string]*contestv1.Contest
	problems map[string]*problemv1.Problem
	tab      *tabv1.Tab
}

func (s *State) ReportContestCallback(_ context.Context, req *observationv1.ReportContestRequest) error {
	s.contests[req.Contest.Id] = req.Contest
	return nil
}

func (s *State) ReportProblemCallback(_ context.Context, req *observationv1.ReportProblemRequest) error {
	s.problems[req.Problem.Id] = req.Problem
	return nil
}

func (s *State) FocusTabCallback(_ context.Context, req *observationv1.FocusTabRequest) error {
	s.tab = req.Tab
	return nil
}
