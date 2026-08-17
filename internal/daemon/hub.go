package daemon

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
)

var (
	ErrNoReceiver    = errors.New("no receiver")
	ErrReplyNotFound = errors.New("not found")
)

type message[Req any] struct {
	req     Req
	replyID uint32
}

type claimChRef[Req any] struct {
	ch    chan *message[Req]
	count int
}

type hub[Req, Reply any] struct {
	mu         sync.Mutex
	claimChs   map[string]*claimChRef[Req]
	replyChs   sync.Map
	replyIDSeq atomic.Uint32
}

func (h *hub[Req, Reply]) acquireClaimCh(subject string) chan *message[Req] {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.claimChs[subject]; !ok {
		h.claimChs[subject] = &claimChRef[Req]{ch: make(chan *message[Req])}
	}
	ref := h.claimChs[subject]
	ref.count++
	return ref.ch
}

func (h *hub[Req, Reply]) releaseClaimCh(subject string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	ref := h.claimChs[subject]
	ref.count--
	if ref.count == 0 {
		delete(h.claimChs, subject)
	}
}

func (h *hub[Req, Reply]) makeReplyCh() (uint32, chan Reply) {
	id := h.replyIDSeq.Add(1)
	ch := make(chan Reply, 1)
	h.replyChs.Store(id, ch)
	return id, ch
}

func (h *hub[Req, Reply]) takeReplyCh(id uint32) (chan Reply, bool) {
	ch, ok := h.replyChs.LoadAndDelete(id)
	if !ok {
		return nil, false
	}
	return ch.(chan Reply), true
}

func (h *hub[Req, Reply]) claim(ctx context.Context, subject string) (*message[Req], error) {
	ch := h.acquireClaimCh(subject)
	defer h.releaseClaimCh(subject)

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("claim %q: receive request: %w", subject, ctx.Err())
	case msg := <-ch:
		return msg, nil
	}
}

func (h *hub[Req, Reply]) doRequest(ctx context.Context, subject string, req Req, wait bool) (Reply, error) {
	var zero Reply

	replyID, replyCh := h.makeReplyCh()
	defer h.takeReplyCh(replyID)

	claimCh := h.acquireClaimCh(subject)
	defer h.releaseClaimCh(subject)

	msg := &message[Req]{req: req, replyID: replyID}
	if wait {
		select {
		case <-ctx.Done():
			return zero, fmt.Errorf("send: %w", ctx.Err())
		case claimCh <- msg:
		}
	} else {
		select {
		case claimCh <- msg:
		default:
			return zero, ErrNoReceiver
		}
	}

	select {
	case <-ctx.Done():
		return zero, fmt.Errorf("receive reply %d: %w", replyID, ctx.Err())
	case reply := <-replyCh:
		return reply, nil
	}
}

func (h *hub[Req, Reply]) request(ctx context.Context, subject string, req Req) (Reply, error) {
	reply, err := h.doRequest(ctx, subject, req, true)
	if err != nil {
		var zero Reply
		return zero, fmt.Errorf("request %q: %w", subject, err)
	}
	return reply, nil
}

func (h *hub[Req, Reply]) tryRequest(ctx context.Context, subject string, req Req) (Reply, error) {
	reply, err := h.doRequest(ctx, subject, req, false)
	if err != nil {
		var zero Reply
		return zero, fmt.Errorf("try request %q: %w", subject, err)
	}
	return reply, nil
}

func (h *hub[Req, Reply]) reply(id uint32, reply Reply) error {
	ch, ok := h.takeReplyCh(id)
	if !ok {
		return fmt.Errorf("reply %d: %w", id, ErrReplyNotFound)
	}
	ch <- reply
	return nil
}

func newHub[Req, Reply any]() *hub[Req, Reply] {
	return &hub[Req, Reply]{claimChs: make(map[string]*claimChRef[Req])}
}
