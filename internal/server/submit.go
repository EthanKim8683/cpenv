package server

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"sync"

	"github.com/google/uuid"

	"connectrpc.com/connect"
	submitv1 "github.com/EthanKim8683/cpenv/gen/submit/v1"
	submitv1connect "github.com/EthanKim8683/cpenv/gen/submit/v1/submitv1connect"
)

type hub struct {
	mu        sync.Mutex
	subs      map[string]map[chan *submitv1.SubscribeResponse]struct{}
	callbacks map[string]chan error
}

func (h *hub) addSubscriber(problemID string) chan *submitv1.SubscribeResponse {
	ch := make(chan *submitv1.SubscribeResponse)
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.subs[problemID]; !ok {
		h.subs[problemID] = make(map[chan *submitv1.SubscribeResponse]struct{})
	}
	h.subs[problemID][ch] = struct{}{}
	return ch
}

func (h *hub) removeSubscriber(problemID string, ch chan *submitv1.SubscribeResponse) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.subs[problemID], ch)
	if len(h.subs[problemID]) == 0 {
		delete(h.subs, problemID)
	}
}

func (h *hub) send(ctx context.Context, problemID string, msg *submitv1.SubscribeResponse) error {
	h.mu.Lock()
	chs := slices.Collect(maps.Keys(h.subs[problemID]))
	h.mu.Unlock()

	if len(chs) == 0 {
		return errors.New("no subscribers")
	}

	cases := make([]reflect.SelectCase, 0, len(chs)+1)
	cases = append(cases, reflect.SelectCase{
		Dir:  reflect.SelectRecv,
		Chan: reflect.ValueOf(ctx.Done()),
	})
	for _, ch := range chs {
		cases = append(cases, reflect.SelectCase{
			Dir:  reflect.SelectSend,
			Chan: reflect.ValueOf(ch),
			Send: reflect.ValueOf(msg),
		})
	}
	chosen, _, _ := reflect.Select(cases)
	if chosen == 0 {
		return ctx.Err()
	}

	return nil
}

func (h *hub) addCallback() (string, chan error) {
	cbID := uuid.New().String()
	ch := make(chan error, 1)
	h.mu.Lock()
	defer h.mu.Unlock()
	h.callbacks[cbID] = ch
	return cbID, ch
}

func (h *hub) takeCallback(cbID string) chan error {
	h.mu.Lock()
	defer h.mu.Unlock()
	ch, ok := h.callbacks[cbID]
	if !ok {
		return nil
	}
	delete(h.callbacks, cbID)
	return ch
}

func (h *hub) submit(ctx context.Context, problemID, fileName string, content []byte) (<-chan error, error) {
	cbID, cbCh := h.addCallback()
	context.AfterFunc(ctx, func() {
		h.takeCallback(cbID)
	})

	if err := h.send(ctx, problemID, &submitv1.SubscribeResponse{
		CallbackId: cbID,
		FileName:   fileName,
		Content:    content,
	}); err != nil {
		h.takeCallback(cbID)
		return nil, err
	}

	return cbCh, nil
}

func (h *hub) subscribe(ctx context.Context, problemID string) (<-chan *submitv1.SubscribeResponse, error) {
	subCh := h.addSubscriber(problemID)
	context.AfterFunc(ctx, func() {
		h.removeSubscriber(problemID, subCh)
	})
	return subCh, nil
}

func (h *hub) callback(cbID string, cbErr error) error {
	cbCh := h.takeCallback(cbID)
	if cbCh == nil {
		return errors.New("not found")
	}
	cbCh <- cbErr
	return nil
}

func newHub() *hub {
	return &hub{
		subs:      make(map[string]map[chan *submitv1.SubscribeResponse]struct{}),
		callbacks: make(map[string]chan error),
	}
}

type SubmitService struct {
	hub *hub
}

func (s *SubmitService) Submit(ctx context.Context, req *submitv1.SubmitRequest) (*submitv1.SubmitResponse, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ch, err := s.hub.submit(ctx, req.ProblemId, req.FileName, req.Content)
	if err != nil {
		return nil, fmt.Errorf("submit: %w", err)
	}

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("submit: %w", ctx.Err())
	case err := <-ch:
		if err != nil {
			return nil, fmt.Errorf("submit: %w", err)
		}
	}
	return &submitv1.SubmitResponse{}, nil
}

func (s *SubmitService) Subscribe(ctx context.Context, req *submitv1.SubscribeRequest, stream *connect.ServerStream[submitv1.SubscribeResponse]) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	ch, err := s.hub.subscribe(ctx, req.ProblemId)
	if err != nil {
		return fmt.Errorf("subscribe: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case msg := <-ch:
			if err := stream.Send(msg); err != nil {
				return fmt.Errorf("subscribe: %w", err)
			}
		}
	}
}

func (s *SubmitService) Callback(_ context.Context, req *submitv1.CallbackRequest) (*submitv1.CallbackResponse, error) {
	var cbErr error
	if req.Error != "" {
		cbErr = fmt.Errorf("extension: %s", req.Error)
	}

	if err := s.hub.callback(req.CallbackId, cbErr); err != nil {
		return nil, err
	}
	return &submitv1.CallbackResponse{}, nil
}

var _ submitv1connect.SubmitServiceHandler = (*SubmitService)(nil)

func NewSubmitService() *SubmitService {
	return &SubmitService{
		hub: newHub(),
	}
}
