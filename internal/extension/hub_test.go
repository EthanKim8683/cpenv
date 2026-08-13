package extension

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type collectT struct {
	t    *testing.T
	mu   sync.Mutex
	errs error
}

func (c *collectT) Errorf(format string, args ...any) {
	c.t.Helper()
	c.mu.Lock()
	defer c.mu.Unlock()
	c.errs = errors.Join(c.errs, fmt.Errorf(format, args...))
}

func (c *collectT) Err() error {
	c.t.Helper()
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.errs
}

var _ assert.TestingT = (*collectT)(nil)

func newCollectT(t *testing.T) *collectT {
	return &collectT{t: t}
}

func TestHub(t *testing.T) {
	t.Parallel()

	subject := "subject"

	t.Run("round trip", func(t *testing.T) {
		t.Parallel()

		req := "req"
		reply := "reply"

		h := newHub[string, string]()
		c := newCollectT(t)
		var wg sync.WaitGroup
		wg.Go(func() {
			gotReply, err := h.request(t.Context(), subject, req)
			assert.NoError(c, err)
			assert.Equal(c, reply, gotReply)
		})
		wg.Go(func() {
			msg, err := h.claim(t.Context(), subject)
			assert.NoError(c, err)
			assert.Equal(c, req, msg.req)
			assert.NoError(c, h.reply(msg.replyID, reply))
		})
		wg.Wait()
		require.NoError(t, c.Err())
	})

	t.Run("try request", func(t *testing.T) {
		t.Parallel()

		h := newHub[struct{}, struct{}]()
		c := newCollectT(t)
		var wg sync.WaitGroup
		wg.Go(func() {
			assert.EventuallyWithT(c, func(ec *assert.CollectT) {
				_, err := h.tryRequest(t.Context(), subject, struct{}{})
				assert.NoError(ec, err)
			}, time.Second, 100*time.Millisecond)
		})
		wg.Go(func() {
			msg, err := h.claim(t.Context(), subject)
			assert.NoError(c, err)
			assert.NoError(c, h.reply(msg.replyID, struct{}{}))
		})
		wg.Wait()
		require.NoError(t, c.Err())
	})

	t.Run("request consumes claim", func(t *testing.T) {
		t.Parallel()

		h := newHub[struct{}, struct{}]()
		c := newCollectT(t)
		var wg sync.WaitGroup
		wg.Go(func() {
			_, err := h.request(t.Context(), subject, struct{}{})
			assert.NoError(c, err)
		})
		wg.Go(func() {
			msg, err := h.claim(t.Context(), subject)
			assert.NoError(c, err)
			assert.NoError(c, h.reply(msg.replyID, struct{}{}))
		})
		wg.Wait()
		require.NoError(t, c.Err())
		_, err := h.tryRequest(t.Context(), subject, struct{}{})
		assert.ErrorContains(t, err, "no receiver")
	})

	t.Run("claim context done", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		h := newHub[struct{}, struct{}]()
		_, err := h.claim(ctx, subject)
		require.ErrorContains(t, err, "receive request")
		_, err = h.tryRequest(t.Context(), subject, struct{}{})
		assert.ErrorContains(t, err, "no receiver")
	})

	t.Run("request context done", func(t *testing.T) {
		t.Parallel()

		t.Run("send", func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(t.Context())
			cancel()
			h := newHub[struct{}, struct{}]()
			_, err := h.request(ctx, subject, struct{}{})
			require.ErrorContains(t, err, "send")
			c := newCollectT(t)
			var wg sync.WaitGroup
			wg.Go(func() {
				_, err := h.request(t.Context(), subject, struct{}{})
				assert.NoError(c, err)
			})
			wg.Go(func() {
				msg, err := h.claim(t.Context(), subject)
				assert.NoError(c, err)
				assert.NoError(c, h.reply(msg.replyID, struct{}{}))
			})
			wg.Wait()
			require.NoError(t, c.Err())
		})

		t.Run("receive", func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(t.Context())
			h := newHub[struct{}, struct{}]()
			c := newCollectT(t)
			var wg sync.WaitGroup
			wg.Go(func() {
				_, err := h.request(ctx, subject, struct{}{})
				assert.ErrorContains(c, err, "receive reply")
			})
			var msg *message[struct{}]
			wg.Go(func() {
				var err error
				msg, err = h.claim(t.Context(), subject)
				assert.NoError(c, err)
				cancel()
			})
			wg.Wait()
			require.NoError(t, c.Err())
			assert.ErrorContains(t, h.reply(msg.replyID, struct{}{}), "not found")
			_, err := h.tryRequest(t.Context(), subject, struct{}{})
			assert.ErrorContains(t, err, "no receiver")
		})
	})
}
