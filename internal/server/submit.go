package server

import (
	"context"
	"sync"

	"connectrpc.com/connect"
	submitv1 "github.com/EthanKim8683/cpenv/gen/submit/v1"
	"github.com/EthanKim8683/cpenv/gen/submit/v1/submitv1connect"
)

type SubmitService struct {
	mu      sync.Mutex
	streams map[*connect.ServerStream[submitv1.SubscribeResponse]]struct{}
}

func (s *SubmitService) Submit(ctx context.Context, req *submitv1.SubmitRequest) (*submitv1.SubmitResponse, error) {
	return &submitv1.SubmitResponse{}, nil
}

func (s *SubmitService) Subscribe(ctx context.Context, req *submitv1.SubscribeRequest, stream *connect.ServerStream[submitv1.SubscribeResponse]) error {
	s.mu.Lock()
	s.streams[stream] = struct{}{}
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		delete(s.streams, stream)
	}()

	if err := stream.Send(&submitv1.SubscribeResponse{}); err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}

var _ submitv1connect.SubmitServiceHandler = (*SubmitService)(nil)

func NewSubmitService() *SubmitService {
	return &SubmitService{
		streams: make(map[*connect.ServerStream[submitv1.SubscribeResponse]]struct{}),
	}
}
