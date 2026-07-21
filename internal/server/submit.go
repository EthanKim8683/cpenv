package server

import (
	"context"
	"errors"
	"sync"

	"connectrpc.com/connect"
	submitv1 "github.com/EthanKim8683/cpenv/gen/submit/v1"
	"github.com/EthanKim8683/cpenv/gen/submit/v1/submitv1connect"
)

type sub struct {
	stream *connect.ServerStream[submitv1.SubscribeResponse]
	cancel context.CancelCauseFunc
}

type SubmitService struct {
	mu   sync.Mutex
	subs map[string]*sub
}

func (s *SubmitService) Submit(ctx context.Context, req *submitv1.SubmitRequest) (*submitv1.SubmitResponse, error) {
	s.mu.Lock()
	sub, ok := s.subs[req.TabId]
	s.mu.Unlock()
	if !ok {
		return nil, connect.NewError(connect.CodeNotFound, errors.New("subscriber not found"))
	}
	if err := sub.stream.Send(&submitv1.SubscribeResponse{}); err != nil {
		return nil, err
	}
	return &submitv1.SubmitResponse{}, nil
}

func (s *SubmitService) Subscribe(ctx context.Context, req *submitv1.SubscribeRequest, stream *connect.ServerStream[submitv1.SubscribeResponse]) error {
	ctx, cancel := context.WithCancelCause(ctx)
	sub := &sub{
		stream: stream,
		cancel: cancel,
	}

	s.mu.Lock()
	if sub, ok := s.subs[req.TabId]; ok {
		sub.cancel(connect.NewError(connect.CodeCanceled, errors.New("subscriber replaced")))
	}
	s.subs[req.TabId] = sub
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if s.subs[req.TabId] == sub {
			delete(s.subs, req.TabId)
		}
	}()

	if err := stream.Send(&submitv1.SubscribeResponse{}); err != nil {
		return err
	}

	<-ctx.Done()
	return context.Cause(ctx)
}

var _ submitv1connect.SubmitServiceHandler = (*SubmitService)(nil)

func NewSubmitService() *SubmitService {
	return &SubmitService{
		subs: make(map[string]*sub),
	}
}
