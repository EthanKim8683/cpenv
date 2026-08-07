package submit

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
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

func (h *hub[Req, Reply]) acquireClaimCh(subj string) chan *message[Req] {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.claimChs[subj]; !ok {
		h.claimChs[subj] = &claimChRef[Req]{ch: make(chan *message[Req])}
	}
	ref := h.claimChs[subj]
	ref.count++
	return ref.ch
}

func (h *hub[Req, Reply]) releaseClaimCh(subj string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	ref := h.claimChs[subj]
	ref.count--
	if ref.count == 0 {
		delete(h.claimChs, subj)
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

func (h *hub[Req, Reply]) claim(ctx context.Context, subj string) (*message[Req], error) {
	ch := h.acquireClaimCh(subj)
	defer h.releaseClaimCh(subj)

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("claim: receive request: %w", ctx.Err())
	case msg := <-ch:
		return msg, nil
	}
}

func (h *hub[Req, Reply]) doRequest(ctx context.Context, subj string, req Req, wait bool) (Reply, error) {
	var zero Reply

	rID, rCh := h.makeReplyCh()
	defer h.takeReplyCh(rID)

	cCh := h.acquireClaimCh(subj)
	defer h.releaseClaimCh(subj)

	msg := &message[Req]{req: req, replyID: rID}
	if wait {
		select {
		case <-ctx.Done():
			return zero, fmt.Errorf("send: %w", ctx.Err())
		case cCh <- msg:
		}
	} else {
		select {
		case cCh <- msg:
		default:
			return zero, errors.New("no receiver")
		}
	}

	select {
	case <-ctx.Done():
		return zero, fmt.Errorf("receive reply %d: %w", rID, ctx.Err())
	case reply := <-rCh:
		return reply, nil
	}
}

func (h *hub[Req, Reply]) request(ctx context.Context, subj string, req Req) (Reply, error) {
	reply, err := h.doRequest(ctx, subj, req, true)
	if err != nil {
		var zero Reply
		return zero, fmt.Errorf("request: %w", err)
	}
	return reply, nil
}

func (h *hub[Req, Reply]) tryRequest(ctx context.Context, subj string, req Req) (Reply, error) {
	reply, err := h.doRequest(ctx, subj, req, false)
	if err != nil {
		var zero Reply
		return zero, fmt.Errorf("try request: %w", err)
	}
	return reply, nil
}

func (h *hub[Req, Reply]) reply(id uint32, reply Reply) error {
	ch, ok := h.takeReplyCh(id)
	if !ok {
		return fmt.Errorf("reply %d: not found", id)
	}
	ch <- reply
	return nil
}

func newHub[Req, Reply any]() *hub[Req, Reply] {
	return &hub[Req, Reply]{claimChs: make(map[string]*claimChRef[Req])}
}
