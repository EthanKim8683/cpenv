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

type offerChRef[Req any] struct {
	ch    chan *message[Req]
	count int
}

type hub[Req, Reply any] struct {
	mu         sync.Mutex
	offerChs   map[string]*offerChRef[Req]
	replyChs   sync.Map
	replyIDSeq atomic.Uint32
}

func (h *hub[Req, Reply]) acquireOfferCh(subj string) chan *message[Req] {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.offerChs[subj]; !ok {
		h.offerChs[subj] = &offerChRef[Req]{ch: make(chan *message[Req])}
	}
	ref := h.offerChs[subj]
	ref.count++
	return ref.ch
}

func (h *hub[Req, Reply]) releaseOfferCh(subj string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	ref := h.offerChs[subj]
	ref.count--
	if ref.count == 0 {
		delete(h.offerChs, subj)
	}
}

func (h *hub[Req, Reply]) offer(ctx context.Context, subj string) (*message[Req], error) {
	ch := h.acquireOfferCh(subj)
	defer h.releaseOfferCh(subj)

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("receive request: %w", ctx.Err())
	case msg := <-ch:
		return msg, nil
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

func (h *hub[Req, Reply]) doRequest(ctx context.Context, subj string, req Req, wait bool) (Reply, error) {
	var zero Reply

	rID, rCh := h.makeReplyCh()
	defer h.takeReplyCh(rID)

	oCh := h.acquireOfferCh(subj)
	defer h.releaseOfferCh(subj)

	msg := &message[Req]{req: req, replyID: rID}
	if wait {
		select {
		case <-ctx.Done():
			return zero, fmt.Errorf("send request: %w", ctx.Err())
		case oCh <- msg:
		}
	} else {
		select {
		case oCh <- msg:
		default:
			return zero, errors.New("no offers")
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
	return h.doRequest(ctx, subj, req, true)
}

func (h *hub[Req, Reply]) tryRequest(ctx context.Context, subj string, req Req) (Reply, error) {
	return h.doRequest(ctx, subj, req, false)
}

func (h *hub[Req, Reply]) reply(id uint32, reply Reply) error {
	ch, ok := h.takeReplyCh(id)
	if !ok {
		return errors.New("not found")
	}
	ch <- reply
	return nil
}

func newHub[Req, Reply any]() *hub[Req, Reply] {
	return &hub[Req, Reply]{offerChs: make(map[string]*offerChRef[Req])}
}
