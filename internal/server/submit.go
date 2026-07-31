package server

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"

	"connectrpc.com/connect"
	submitv1 "github.com/EthanKim8683/cpenv/gen/submit/v1"
	submitv1connect "github.com/EthanKim8683/cpenv/gen/submit/v1/submitv1connect"
)

type subscriber struct {
	mu     sync.Mutex
	stream *connect.ServerStream[submitv1.SubscribeResponse]
	cancel context.CancelCauseFunc
}

func (s *subscriber) send(msg *submitv1.SubscribeResponse) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.stream.Send(msg); err != nil {
		s.cancel(fmt.Errorf("send: %w", err))
		return err
	}
	return nil
}

type SubmitService struct {
	mu        sync.Mutex
	subs      map[string]map[*subscriber]struct{}
	callbacks map[string]chan string
}

func (s *SubmitService) addSubscriber(problemID string, sub *subscriber) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.subs[problemID]; !ok {
		s.subs[problemID] = make(map[*subscriber]struct{})
	}
	s.subs[problemID][sub] = struct{}{}
}

func (s *SubmitService) removeSubscriber(problemID string, sub *subscriber) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.subs[problemID], sub)
	if len(s.subs[problemID]) == 0 {
		delete(s.subs, problemID)
	}
}

func (s *SubmitService) anySubscriber(problemID string) *subscriber {
	s.mu.Lock()
	defer s.mu.Unlock()
	for sub := range s.subs[problemID] {
		return sub
	}
	return nil
}

func (s *SubmitService) makeCallback() (string, chan string) {
	cbID := uuid.New().String()
	cb := make(chan string, 1)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.callbacks[cbID] = cb
	return cbID, cb
}

func (s *SubmitService) takeCallback(cbID string) chan string {
	s.mu.Lock()
	defer s.mu.Unlock()
	cb, ok := s.callbacks[cbID]
	if !ok {
		return nil
	}
	delete(s.callbacks, cbID)
	return cb
}

func (s *SubmitService) Submit(ctx context.Context, req *submitv1.SubmitRequest) (*submitv1.SubmitResponse, error) {
	if !filepath.IsAbs(req.Path) {
		return nil, fmt.Errorf("submit %q: path %q is not absolute", req.ProblemId, req.Path)
	}

	data, err := os.ReadFile(req.Path)
	if err != nil {
		return nil, fmt.Errorf("submit %q: %w", req.ProblemId, err)
	}

	cbID, cb := s.makeCallback()
	defer s.takeCallback(cbID)

	sub := s.anySubscriber(req.ProblemId)
	if sub == nil {
		return nil, fmt.Errorf("submit %q: no subscribers", req.ProblemId)
	}

	if err := sub.send(&submitv1.SubscribeResponse{
		CallbackId: cbID,
		Path:       req.Path,
		Data:       data,
	}); err != nil {
		return nil, fmt.Errorf("submit %q: subscriber: %w", req.ProblemId, err)
	}

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("submit %q: %w", req.ProblemId, ctx.Err())
	case errMsg := <-cb:
		if errMsg != "" {
			return nil, fmt.Errorf("submit %q: extension: %s", req.ProblemId, errMsg)
		}
	}

	return &submitv1.SubmitResponse{}, nil
}

func (s *SubmitService) Subscribe(ctx context.Context, req *submitv1.SubscribeRequest, stream *connect.ServerStream[submitv1.SubscribeResponse]) error {
	errCtx, cancel := context.WithCancelCause(context.Background())
	defer cancel(nil)

	sub := &subscriber{
		stream: stream,
		cancel: cancel,
	}
	s.addSubscriber(req.ProblemId, sub)
	defer s.removeSubscriber(req.ProblemId, sub)

	select {
	case <-ctx.Done():
		return nil
	case <-errCtx.Done():
		return fmt.Errorf("subscribe %q: %w", req.ProblemId, context.Cause(errCtx))
	}
}

func (s *SubmitService) Callback(_ context.Context, req *submitv1.CallbackRequest) (*submitv1.CallbackResponse, error) {
	cb := s.takeCallback(req.CallbackId)
	if cb == nil {
		return nil, fmt.Errorf("callback %q: not found", req.CallbackId)
	}

	cb <- req.Error

	return &submitv1.CallbackResponse{}, nil
}

var _ submitv1connect.SubmitServiceHandler = (*SubmitService)(nil)

func NewSubmitService() *SubmitService {
	return &SubmitService{
		subs:      make(map[string]map[*subscriber]struct{}),
		callbacks: make(map[string]chan string),
	}
}
