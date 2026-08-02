package server

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	submitv1 "github.com/EthanKim8683/cpenv/gen/submit/v1"
	submitv1connect "github.com/EthanKim8683/cpenv/gen/submit/v1/submitv1connect"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHub(t *testing.T) {
	t.Parallel()

	hub := newHub()

	t.Run("round trip", func(t *testing.T) {
		t.Parallel()

		fileName := "file name"
		content := []byte("content")
		cbErr := errors.New("callback error")

		subCh, err := hub.subscribe(t.Context(), t.Name())
		require.NoError(t, err)

		var wg sync.WaitGroup
		wg.Go(func() {
			cbCh, err := hub.submit(t.Context(), t.Name(), fileName, content)
			require.NoError(t, err)
			assert.ErrorIs(t, <-cbCh, cbErr)
		})
		wg.Go(func() {
			msg := <-subCh
			assert.Equal(t, fileName, msg.FileName)
			assert.Equal(t, content, msg.Content)

			require.NoError(t, hub.callback(msg.CallbackId, cbErr))
		})
		wg.Wait()
	})

	t.Run("no subscribers", func(t *testing.T) {
		t.Parallel()

		_, err := hub.submit(t.Context(), t.Name(), "", nil)
		assert.ErrorContains(t, err, "no subscribers")
	})

	t.Run("multiple subscribers", func(t *testing.T) {
		t.Parallel()

		subCh1, err := hub.subscribe(t.Context(), t.Name())
		require.NoError(t, err)

		subCh2, err := hub.subscribe(t.Context(), t.Name())
		require.NoError(t, err)

		var wg sync.WaitGroup
		wg.Go(func() {
			cbCh, err := hub.submit(t.Context(), t.Name(), "", nil)
			require.NoError(t, err)
			assert.ErrorContains(t, <-cbCh, "sub 1")

			cbCh, err = hub.submit(t.Context(), t.Name(), "", nil)
			require.NoError(t, err)
			assert.ErrorContains(t, <-cbCh, "sub 2")
		})
		wg.Go(func() {
			require.NoError(t, hub.callback((<-subCh1).CallbackId, fmt.Errorf("sub 1")))

			require.NoError(t, hub.callback((<-subCh2).CallbackId, fmt.Errorf("sub 2")))
		})
		wg.Wait()
	})

	t.Run("multiple submits", func(t *testing.T) {
		t.Parallel()

		subCh, err := hub.subscribe(t.Context(), t.Name())
		require.NoError(t, err)

		var wg sync.WaitGroup
		wg.Go(func() {
			cbCh, err := hub.submit(t.Context(), t.Name(), "", nil)
			require.NoError(t, err)
			assert.NoError(t, <-cbCh)
		})
		wg.Go(func() {
			cbCh, err := hub.submit(t.Context(), t.Name(), "", nil)
			require.NoError(t, err)
			assert.NoError(t, <-cbCh)
		})
		wg.Go(func() {
			require.NoError(t, hub.callback((<-subCh).CallbackId, nil))

			require.NoError(t, hub.callback((<-subCh).CallbackId, nil))
		})
		wg.Wait()
	})

	t.Run("multiple callbacks", func(t *testing.T) {
		t.Parallel()

		subCh, err := hub.subscribe(t.Context(), t.Name())
		require.NoError(t, err)

		var wg sync.WaitGroup
		wg.Go(func() {
			cbCh, err := hub.submit(t.Context(), t.Name(), "", nil)
			require.NoError(t, err)
			assert.NoError(t, <-cbCh)
		})
		wg.Go(func() {
			msg := <-subCh
			require.NoError(t, hub.callback(msg.CallbackId, nil))

			assert.ErrorContains(t, hub.callback(msg.CallbackId, nil), "not found")
		})
		wg.Wait()
	})
}

func TestSubmitService(t *testing.T) {
	t.Parallel()

	fileName := "file name"
	content := []byte("content")
	cbErr := "callback error"

	svc := NewSubmitService()
	mux := http.NewServeMux()
	mux.Handle(submitv1connect.NewSubmitServiceHandler(svc))
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	var wg sync.WaitGroup
	wg.Go(func() {
		client := submitv1connect.NewSubmitServiceClient(srv.Client(), srv.URL)

		require.EventuallyWithT(t, func(c *assert.CollectT) {
			_, err := client.Submit(t.Context(), &submitv1.SubmitRequest{
				ProblemId: t.Name(),
				FileName:  fileName,
				Content:   content,
			})
			assert.ErrorContains(c, err, cbErr)
		}, time.Second, 50*time.Millisecond)
	})
	wg.Go(func() {
		client := submitv1connect.NewSubmitServiceClient(srv.Client(), srv.URL)

		stream, err := client.Subscribe(t.Context(), &submitv1.SubscribeRequest{ProblemId: t.Name()})
		require.NoError(t, err)

		require.True(t, stream.Receive())

		msg := stream.Msg()
		assert.Equal(t, fileName, msg.FileName)
		assert.Equal(t, content, msg.Content)

		_, err = client.Callback(t.Context(), &submitv1.CallbackRequest{
			CallbackId: msg.CallbackId,
			Error:      cbErr,
		})
		require.NoError(t, err)
	})
	wg.Wait()
}
