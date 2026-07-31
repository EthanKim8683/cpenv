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

type safeStream struct {
	mu     sync.Mutex
	stream *connect.ServerStream[submitv1.SubscribeResponse]
}

func (s *safeStream) send(msg *submitv1.SubscribeResponse) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.stream.Send(msg)
}

type SubmitService struct {
	mu        sync.Mutex
	streams   map[string]map[*safeStream]struct{}
	callbacks map[string]chan string
}

func (s *SubmitService) addStream(problemID string, stream *safeStream) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.streams[problemID]; !ok {
		s.streams[problemID] = make(map[*safeStream]struct{})
	}
	s.streams[problemID][stream] = struct{}{}
}

func (s *SubmitService) removeStream(problemID string, stream *safeStream) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.streams[problemID], stream)
	if len(s.streams[problemID]) == 0 {
		delete(s.streams, problemID)
	}
}

func (s *SubmitService) anyStream(problemID string) *safeStream {
	s.mu.Lock()
	defer s.mu.Unlock()
	for stream := range s.streams[problemID] {
		return stream
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
		return nil, fmt.Errorf("submit: path is not absolute: %s", req.Path)
	}

	data, err := os.ReadFile(req.Path)
	if err != nil {
		return nil, fmt.Errorf("submit: read file: %w", err)
	}

	cbID, cb := s.makeCallback()
	defer s.takeCallback(cbID)

	stream := s.anyStream(req.ProblemId)
	if stream == nil {
		return nil, fmt.Errorf("submit: no subscribers: %s", req.ProblemId)
	}

	if err := stream.send(&submitv1.SubscribeResponse{
		CallbackId: cbID,
		Path:       req.Path,
		Data:       data,
	}); err != nil {
		s.removeStream(req.ProblemId, stream)
		return nil, fmt.Errorf("submit: forward to subscriber: %w", err)
	}

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("submit: context done: %w", ctx.Err())
	case errMsg := <-cb:
		if errMsg != "" {
			return nil, fmt.Errorf("submit: extension: %s", errMsg)
		}
	}

	return &submitv1.SubmitResponse{}, nil
}

func (s *SubmitService) Subscribe(ctx context.Context, req *submitv1.SubscribeRequest, stream *connect.ServerStream[submitv1.SubscribeResponse]) error {
	ss := &safeStream{stream: stream}
	s.addStream(req.ProblemId, ss)
	defer s.removeStream(req.ProblemId, ss)
	<-ctx.Done()
	return nil
}

func (s *SubmitService) Callback(_ context.Context, req *submitv1.CallbackRequest) (*submitv1.CallbackResponse, error) {
	cb := s.takeCallback(req.CallbackId)
	if cb == nil {
		return nil, fmt.Errorf("callback: callback not found: %s", req.CallbackId)
	}

	cb <- req.Error

	return &submitv1.CallbackResponse{}, nil
}

var _ submitv1connect.SubmitServiceHandler = (*SubmitService)(nil)

func NewSubmitService() *SubmitService {
	return &SubmitService{
		streams:   make(map[string]map[*safeStream]struct{}),
		callbacks: make(map[string]chan string),
	}
}
