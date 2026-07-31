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

type job struct {
	path     string
	callback chan error
}

type worker chan *job

func (w worker) send(ctx context.Context, path string) chan error {
	cb := make(chan error, 1)
	select {
	case <-ctx.Done():
		cb <- fmt.Errorf("send: context done: %w", ctx.Err())
	case w <- &job{
		path:     path,
		callback: cb,
	}:
	}
	return cb
}

// SubmitService ...
type SubmitService struct {
	mu        sync.Mutex
	workers   map[string]map[worker]struct{}
	callbacks map[string]chan error
}

func (s *SubmitService) worker(problemID string) worker {
	s.mu.Lock()
	defer s.mu.Unlock()
	for w := range s.workers[problemID] {
		return w
	}
	return nil
}

// Submit ...
func (s *SubmitService) Submit(ctx context.Context, req *submitv1.SubmitRequest) (*submitv1.SubmitResponse, error) {
	if !filepath.IsAbs(req.Path) {
		return nil, fmt.Errorf("submit: path is not absolute: %s", req.Path)
	}

	worker := s.worker(req.ProblemId)
	if worker == nil {
		return nil, fmt.Errorf("submit: no subscribers: %s", req.ProblemId)
	}

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("submit: context done: %w", ctx.Err())
	case err := <-worker.send(ctx, req.Path):
		if err != nil {
			return nil, fmt.Errorf("submit: %w", err)
		}
	}

	return &submitv1.SubmitResponse{}, nil
}

func (s *SubmitService) addWorker(problemID string, w worker) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.workers[problemID]; !ok {
		s.workers[problemID] = make(map[worker]struct{})
	}
	s.workers[problemID][w] = struct{}{}
}

func (s *SubmitService) deleteWorker(problemID string, w worker) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.workers[problemID], w)
	if len(s.workers[problemID]) == 0 {
		delete(s.workers, problemID)
	}
}

func (s *SubmitService) setCallback(cbID string, cb chan error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.callbacks[cbID] = cb
}

func (s *SubmitService) deleteCallback(cbID string) chan error {
	s.mu.Lock()
	defer s.mu.Unlock()
	cb, ok := s.callbacks[cbID]
	if ok {
		delete(s.callbacks, cbID)
	}
	return cb
}

// Subscribe ...
func (s *SubmitService) Subscribe(ctx context.Context, req *submitv1.SubscribeRequest, stream *connect.ServerStream[submitv1.SubscribeResponse]) error {
	w := make(chan *job)
	s.addWorker(req.ProblemId, w)
	defer s.deleteWorker(req.ProblemId, w)

	handleJob := func(job *job) error {
		content, err := os.ReadFile(job.path)
		if err != nil {
			return err
		}

		cb := make(chan error, 1)
		cbID := uuid.New().String()
		s.setCallback(cbID, cb)
		defer s.deleteCallback(cbID)

		if err := stream.Send(&submitv1.SubscribeResponse{
			CallbackId: cbID,
			Content:    content,
			FileName:   job.path,
		}); err != nil {
			return err
		}

		select {
		case err := <-cb:
			return err
		case <-ctx.Done():
			return fmt.Errorf("subscribe: context done: %w", ctx.Err())
		}
	}

	for {
		select {
		case job := <-w:
			job.callback <- handleJob(job)
		case <-ctx.Done():
			return nil
		}
	}
}

// Callback ...
func (s *SubmitService) Callback(_ context.Context, req *submitv1.CallbackRequest) (*submitv1.CallbackResponse, error) {
	cb := s.deleteCallback(req.CallbackId)
	if cb == nil {
		return nil, fmt.Errorf("callback: callback not found: %s", req.CallbackId)
	}

	var err error
	if req.Error != "" {
		err = fmt.Errorf("extension: %s", req.Error)
	}

	cb <- err

	return &submitv1.CallbackResponse{}, nil
}

var _ submitv1connect.SubmitServiceHandler = (*SubmitService)(nil)

// NewSubmitService ...
func NewSubmitService() *SubmitService {
	return &SubmitService{
		workers:   make(map[string]map[worker]struct{}),
		callbacks: make(map[string]chan error),
	}
}
