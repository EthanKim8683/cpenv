package server

import (
	"context"
	"fmt"
	"os"
	"path"
	"sync"

	"connectrpc.com/connect"
	submitv1 "github.com/EthanKim8683/cpenv/gen/submit/v1"
	submitv1connect "github.com/EthanKim8683/cpenv/gen/submit/v1/submitv1connect"
)

type SubmitService struct {
	mu    sync.Mutex
	pools map[string]map[chan string]struct{}
}

func (s *SubmitService) Submit(ctx context.Context, req *submitv1.SubmitRequest) (*submitv1.SubmitResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !path.IsAbs(req.Path) {
		return nil, fmt.Errorf("submit: path is not absolute: %s", req.Path)
	}

	pool := s.pools[req.ProblemId]
	if len(pool) == 0 {
		return nil, fmt.Errorf("submit: no subscribers: %s", req.ProblemId)
	}

	for ch := range pool {
		ch <- req.Path
		break
	}

	return &submitv1.SubmitResponse{}, nil
}

func (s *SubmitService) Subscribe(ctx context.Context, req *submitv1.SubscribeRequest, stream *connect.ServerStream[submitv1.SubscribeResponse]) error {
	ch := make(chan string)

	s.mu.Lock()
	if _, ok := s.pools[req.ProblemId]; !ok {
		s.pools[req.ProblemId] = make(map[chan string]struct{})
	}
	pool := s.pools[req.ProblemId]
	pool[ch] = struct{}{}
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(pool, ch)
		if len(pool) == 0 {
			delete(s.pools, req.ProblemId)
		}
		s.mu.Unlock()
	}()

	for {
		select {
		case path := <-ch:
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			if err := stream.Send(&submitv1.SubscribeResponse{
				Content:  content,
				FileName: path,
			}); err != nil {
				return err
			}
		case <-ctx.Done():
			return nil
		}
	}
}

var _ submitv1connect.SubmitServiceHandler = (*SubmitService)(nil)

func NewSubmitService() *SubmitService {
	return &SubmitService{
		pools: make(map[string]map[chan string]struct{}),
	}
}
